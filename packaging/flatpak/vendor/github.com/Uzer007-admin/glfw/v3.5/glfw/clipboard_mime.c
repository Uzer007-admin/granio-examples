#include "glfw/src/internal.h"
#include <string.h>
#include <stdlib.h>
#if defined(_GLFW_X11) || defined(_GLFW_WAYLAND)
#include <poll.h>
#include <unistd.h>
#include <errno.h>
#include <time.h>
#endif
#if defined(_GLFW_WAYLAND)
#include "glfw/include/wayland-client-protocol.h"
#include <fcntl.h>
#include <pthread.h>
#endif
#define MIME_LIMIT (64u*1024u*1024u)
typedef struct MimePayload {char* type;unsigned char* bytes;int size;struct MimePayload* next;} MimePayload;
static MimePayload *ownedMime,*stagedMime;
static size_t stagedBytes;
static int stagedCount;
static int publishingMime;
static void freePayloads(MimePayload* p){while(p){MimePayload* next=p->next;free(p->type);free(p->bytes);free(p);p=next;}}
static MimePayload* findPayload(const char* type){if(!strcmp(type,"text/plain;charset=utf-8"))type="text/plain";for(MimePayload* p=ownedMime;p;p=p->next)if(!strcmp(type,p->type))return p;return NULL;}
void fyneClipboardClear(void){if(!publishingMime){freePayloads(ownedMime);ownedMime=NULL;}}
int fyneClipboardBegin(void){if(!_glfw.initialized)return 0;freePayloads(stagedMime);stagedMime=NULL;stagedBytes=0;stagedCount=0;return 1;}
int fyneClipboardAdd(const char* type,const void* bytes,int length){
    if(length<0||strlen(type)>255||stagedCount>=128||(size_t)length>MIME_LIMIT-stagedBytes)return 0;
    MimePayload* p=calloc(1,sizeof(*p));if(!p)return 0;p->type=malloc(strlen(type)+1);p->bytes=malloc(length+1);if(!p->type||!p->bytes){freePayloads(p);return 0;}strcpy(p->type,type);if(length)memcpy(p->bytes,bytes,length);p->bytes[length]=0;p->size=length;p->next=stagedMime;stagedMime=p;stagedBytes+=length;stagedCount++;return 1;
}
void fyneClipboardAbort(void){freePayloads(stagedMime);stagedMime=NULL;stagedBytes=0;}
#if defined(_GLFW_X11) || defined(_GLFW_WAYLAND)
static double nowSeconds(void){struct timespec now;clock_gettime(CLOCK_MONOTONIC,&now);return now.tv_sec+now.tv_nsec*1e-9;}
static int waitFD(int fd,double deadline){struct pollfd p={fd,POLLIN,0};for(;;){int timeout=(int)((deadline-nowSeconds())*1000);if(timeout<=0)return 0;int result=poll(&p,1,timeout);if(result<0&&errno==EINTR)continue;return result>0;}}
#endif

#if defined(_GLFW_X11)
typedef struct MimeSend {Window requestor;Atom property,type;unsigned char* bytes;size_t size,offset;double deadline;struct MimeSend* next;} MimeSend;
static MimeSend* sends;
static size_t sendingBytes;
static void dropSend(MimeSend** link){MimeSend* p=*link;*link=p->next;sendingBytes-=p->size;free(p->bytes);free(p);}
void fyneMimePoll(void){MimeSend** link=&sends;while(*link){if((*link)->deadline<nowSeconds())dropSend(link);else link=&(*link)->next;}}
int fyneMimeEvent(XEvent* event){
    fyneMimePoll();if(event->type!=PropertyNotify||event->xproperty.state!=PropertyDelete)return 0;
    MimeSend** link=&sends;while(*link){MimeSend* p=*link;if(p->requestor==event->xproperty.window&&p->property==event->xproperty.atom){size_t count=p->size-p->offset;if(count>65536)count=65536;XChangeProperty(_glfw.x11.display,p->requestor,p->property,p->type,8,PropModeReplace,p->bytes+p->offset,(int)count);p->offset+=count;p->deadline=nowSeconds()+3;if(!count)dropSend(link);XFlush(_glfw.x11.display);return 1;}link=&p->next;}return 0;
}
int fyneMimeSelection(const XSelectionRequestEvent* request){
    if(request->selection!=_glfw.x11.CLIPBOARD||!ownedMime||request->property==None)return 0;
    if(request->target==_glfw.x11.TARGETS){Atom targets[132];int n=0;targets[n++]=_glfw.x11.TARGETS;targets[n++]=_glfw.x11.UTF8_STRING;targets[n++]=XA_STRING;for(MimePayload* p=ownedMime;p&&n<132;p=p->next)targets[n++]=XInternAtom(_glfw.x11.display,p->type,False);XChangeProperty(_glfw.x11.display,request->requestor,request->property,XA_ATOM,32,PropModeReplace,(unsigned char*)targets,n);return 1;}
    char* name=XGetAtomName(_glfw.x11.display,request->target);if(!name)return 0;MimePayload* p=findPayload(name);if(request->target==_glfw.x11.UTF8_STRING||request->target==XA_STRING)p=findPayload("text/plain");XFree(name);if(!p)return 0;
    if(p->size<=65536){XChangeProperty(_glfw.x11.display,request->requestor,request->property,request->target,8,PropModeReplace,p->bytes,p->size);return 1;}
    fyneMimePoll();if((size_t)p->size>MIME_LIMIT-sendingBytes)return -1;
    MimeSend* send=calloc(1,sizeof(*send));if(!send)return -1;send->bytes=malloc(p->size);if(!send->bytes){free(send);return -1;}memcpy(send->bytes,p->bytes,p->size);send->size=p->size;send->requestor=request->requestor;send->property=request->property;send->type=request->target;send->deadline=nowSeconds()+3;send->next=sends;sends=send;sendingBytes+=p->size;
    XWindowAttributes attributes;XGetWindowAttributes(_glfw.x11.display,request->requestor,&attributes);XSelectInput(_glfw.x11.display,request->requestor,attributes.your_event_mask|PropertyChangeMask);unsigned long length=p->size;Atom incr=XInternAtom(_glfw.x11.display,"INCR",False);XChangeProperty(_glfw.x11.display,request->requestor,request->property,incr,32,PropModeReplace,(unsigned char*)&length,1);return 1;
}
static void* readX11(Atom target,int* length,int* status,int atoms){
    Display* display=_glfw.x11.display;Window requestor=_glfw.x11.helperWindowHandle;Atom property=XInternAtom(display,"_FYNE_MIME_CLIPBOARD",False);Atom incr=XInternAtom(display,"INCR",False);XWindowAttributes attributes;XGetWindowAttributes(display,requestor,&attributes);XSelectInput(display,requestor,attributes.your_event_mask|PropertyChangeMask);
    XDeleteProperty(display,requestor,property);XConvertSelection(display,_glfw.x11.CLIPBOARD,target,property,requestor,CurrentTime);XFlush(display);double deadline=nowSeconds()+3;XEvent event;
    while(!XCheckTypedWindowEvent(display,requestor,SelectionNotify,&event)){if(!waitFD(ConnectionNumber(display),deadline)){*status=1;return NULL;}XEventsQueued(display,QueuedAfterReading);}
    if(event.xselection.property==None){*status=1;return NULL;}
    Atom actual;int format;unsigned long count,left;unsigned char* bytes=NULL;
    XGetWindowProperty(display,requestor,property,0,(long)(MIME_LIMIT/4),False,AnyPropertyType,&actual,&format,&count,&left,&bytes);
    if(actual!=incr){size_t size=count*(format==32?sizeof(unsigned long):(size_t)format/8);if(left||size>MIME_LIMIT||format!=(atoms?32:8)){if(bytes)XFree(bytes);*status=1;return NULL;}void* out=malloc(size+1);if(!out){if(bytes)XFree(bytes);*status=1;return NULL;}memcpy(out,bytes,size);((char*)out)[size]=0;XFree(bytes);XDeleteProperty(display,requestor,property);*length=(int)size;*status=0;return out;}
    if(bytes)XFree(bytes);
    // XCheckTypedWindowEvent(SelectionNotify) leaves the INCR header's earlier
    // PropertyNewValue in the queue. Remove it before acknowledging the header;
    // otherwise it looks like a data chunk whose property has already vanished.
    while(XCheckTypedWindowEvent(display,requestor,PropertyNotify,&event)){}
    XDeleteProperty(display,requestor,property);XFlush(display);unsigned char* out=NULL;size_t size=0;
    for(;;){
        while(!XCheckTypedWindowEvent(display,requestor,PropertyNotify,&event)){if(!waitFD(ConnectionNumber(display),deadline)){free(out);*status=1;return NULL;}XEventsQueued(display,QueuedAfterReading);}
        if(event.xproperty.atom!=property||event.xproperty.state!=PropertyNewValue)continue;
        XGetWindowProperty(display,requestor,property,0,(long)(MIME_LIMIT/4),True,AnyPropertyType,&actual,&format,&count,&left,&bytes);
        if(left||format!=8||count>MIME_LIMIT-size){if(bytes)XFree(bytes);free(out);*status=1;return NULL;}
        unsigned char* resized=realloc(out,size+count+1);if(!resized){if(bytes)XFree(bytes);free(out);*status=1;return NULL;}out=resized;if(count)memcpy(out+size,bytes,count);size+=count;out[size]=0;if(bytes)XFree(bytes);XFlush(display);if(!count)break;
    }
    *length=(int)size;*status=0;return out;
}
#endif

#if defined(_GLFW_WAYLAND)
typedef struct MimeOffer {struct wl_data_offer* offer;char* types;struct MimeOffer* next;} MimeOffer;
static MimeOffer* offers;
typedef struct MimePipeSend{int fd;unsigned char* bytes;size_t size;}MimePipeSend;
static pthread_mutex_t pipeSendLock=PTHREAD_MUTEX_INITIALIZER;
static size_t pipeSendBytes;
static unsigned pipeSendCount;
static void* sendPipe(void* argument){
    MimePipeSend* send=argument;size_t offset=0;double deadline=nowSeconds()+3;
    fcntl(send->fd,F_SETFL,fcntl(send->fd,F_GETFL)|O_NONBLOCK);
    while(offset<send->size){ssize_t n=write(send->fd,send->bytes+offset,send->size-offset);if(n>0){offset+=n;continue;}if(n<0&&errno==EINTR)continue;if(n<0&&(errno==EAGAIN||errno==EWOULDBLOCK)){struct pollfd fd={send->fd,POLLOUT,0};int timeout=(int)((deadline-nowSeconds())*1000);if(timeout>0&&poll(&fd,1,timeout)>0)continue;}break;}
    close(send->fd);pthread_mutex_lock(&pipeSendLock);pipeSendBytes-=send->size;pipeSendCount--;pthread_mutex_unlock(&pipeSendLock);free(send->bytes);free(send);return NULL;
}
const char* fyneWaylandOfferTypes(struct wl_data_offer* offer){for(MimeOffer* p=offers;p;p=p->next)if(p->offer==offer)return p->types;return NULL;}
void fyneMimeOffer(struct wl_data_offer* offer,const char* type){MimeOffer* p=offers;while(p&&p->offer!=offer)p=p->next;if(!p){p=calloc(1,sizeof(*p));if(!p)return;p->offer=offer;p->next=offers;offers=p;}size_t old=p->types?strlen(p->types):0,n=strlen(type);if(old+n+1>32768)return;char* text=realloc(p->types,old+n+2);if(!text)return;p->types=text;memcpy(text+old,type,n);text[old+n]='\n';text[old+n+1]=0;}
void fyneMimeOfferDestroyMetadata(struct wl_data_offer* offer){MimeOffer** link=&offers;while(*link){MimeOffer* p=*link;if(p->offer==offer){*link=p->next;free(p->types);free(p);return;}link=&p->next;}}
int fyneMimeWaylandSend(struct wl_data_source* source,const char* type,int fd){
    if(source!=_glfw.wl.selectionSource||!ownedMime)return 0;MimePayload* p=findPayload(type);if(!p){close(fd);return 1;}
    pthread_mutex_lock(&pipeSendLock);int available=pipeSendCount<16&&(size_t)p->size<=MIME_LIMIT-pipeSendBytes;if(available){pipeSendBytes+=p->size;pipeSendCount++;}pthread_mutex_unlock(&pipeSendLock);if(!available){close(fd);return 1;}
    MimePipeSend* send=calloc(1,sizeof(*send));if(send)send->bytes=malloc(p->size+1);if(!send||!send->bytes){free(send);close(fd);pthread_mutex_lock(&pipeSendLock);pipeSendBytes-=p->size;pipeSendCount--;pthread_mutex_unlock(&pipeSendLock);return 1;}send->fd=fd;send->size=p->size;memcpy(send->bytes,p->bytes,p->size);
    pthread_t worker;if(pthread_create(&worker,NULL,sendPipe,send)==0)pthread_detach(worker);else{close(fd);pthread_mutex_lock(&pipeSendLock);pipeSendBytes-=send->size;pipeSendCount--;pthread_mutex_unlock(&pipeSendLock);free(send->bytes);free(send);}return 1;
}
void* fyneReadWaylandOffer(struct wl_data_offer* offer,const char* type,int* length,int* status){
    if(!offer){*status=1;return NULL;}
    // Request only an advertised representation. A text/plain drag source need
    // not implement the charset alias used by some clipboard producers.
    if(!strcmp(type,"text/plain")){
        const char* types=fyneWaylandOfferTypes(offer);int plain=0,utf8=0;
        if(types)for(const char* line=types;*line;){const char* end=strchr(line,'\n');size_t n=end?(size_t)(end-line):strlen(line);if(n==10&&!memcmp(line,"text/plain",10))plain=1;if(n==24&&!memcmp(line,"text/plain;charset=utf-8",24))utf8=1;line=end?end+1:line+n;}
        if(!plain&&utf8)type="text/plain;charset=utf-8";
    }
    int fds[2];if(pipe(fds)!=0){*status=1;return NULL;}
    wl_data_offer_receive(offer,type,fds[1]);wl_display_flush(_glfw.wl.display);close(fds[1]);double deadline=nowSeconds()+3;size_t size=0;unsigned char* out=NULL;
    for(;;){if(!waitFD(fds[0],deadline)){free(out);close(fds[0]);*status=1;return NULL;}unsigned char block[65536];ssize_t count=read(fds[0],block,sizeof(block));if(count<0&&errno==EINTR)continue;if(count<0||(size_t)count>MIME_LIMIT-size){free(out);close(fds[0]);*status=1;return NULL;}unsigned char* resized=realloc(out,size+count+1);if(!resized){free(out);close(fds[0]);*status=1;return NULL;}out=resized;if(count)memcpy(out+size,block,count);size+=count;out[size]=0;if(!count)break;}
    close(fds[0]);*length=(int)size;*status=0;return out;
}
#endif

int fyneClipboardPublish(void){
    freePayloads(ownedMime);ownedMime=stagedMime;stagedMime=NULL;stagedBytes=0;publishingMime=1;MimePayload* text=findPayload("text/plain");glfwSetClipboardString(NULL,text?(char*)text->bytes:"");publishingMime=0;
#if defined(_GLFW_X11)
    if(_glfw.platform.platformID==GLFW_PLATFORM_X11)return XGetSelectionOwner(_glfw.x11.display,_glfw.x11.CLIPBOARD)==_glfw.x11.helperWindowHandle;
#endif
#if defined(_GLFW_WAYLAND)
    if(_glfw.platform.platformID==GLFW_PLATFORM_WAYLAND){if(!_glfw.wl.selectionSource)return 0;for(MimePayload* p=ownedMime;p;p=p->next)wl_data_source_offer(_glfw.wl.selectionSource,p->type);wl_display_flush(_glfw.wl.display);return 1;}
#endif
    return 0;
}
void* fyneClipboardRead(const char* type,int* size,int* status){
    *size=0;*status=1;if(!_glfw.initialized)return NULL;int owner=0;
#if defined(_GLFW_X11)
    if(_glfw.platform.platformID==GLFW_PLATFORM_X11)owner=XGetSelectionOwner(_glfw.x11.display,_glfw.x11.CLIPBOARD)==_glfw.x11.helperWindowHandle;
#endif
#if defined(_GLFW_WAYLAND)
    if(_glfw.platform.platformID==GLFW_PLATFORM_WAYLAND)owner=_glfw.wl.selectionSource!=NULL;
#endif
    MimePayload* p=owner?findPayload(type):NULL;if(p){void* out=malloc(p->size+1);if(!out)return NULL;memcpy(out,p->bytes,p->size);((char*)out)[p->size]=0;*size=p->size;*status=0;return out;}
    if(owner){if(!strcmp(type,"text/plain")){const char* text="";
#if defined(_GLFW_X11)
        if(_glfw.platform.platformID==GLFW_PLATFORM_X11&&_glfw.x11.clipboardString)text=_glfw.x11.clipboardString;
#endif
#if defined(_GLFW_WAYLAND)
        if(_glfw.platform.platformID==GLFW_PLATFORM_WAYLAND&&_glfw.wl.clipboardString)text=_glfw.wl.clipboardString;
#endif
        size_t length=strlen(text);if(length>MIME_LIMIT)return NULL;char* out=malloc(length+1);if(!out)return NULL;memcpy(out,text,length+1);*size=(int)length;*status=0;return out;}return NULL;}
#if defined(_GLFW_X11)
    if(_glfw.platform.platformID==GLFW_PLATFORM_X11){Atom target=!strcmp(type,"text/plain")?_glfw.x11.UTF8_STRING:XInternAtom(_glfw.x11.display,type,False);return readX11(target,size,status,0);}
#endif
#if defined(_GLFW_WAYLAND)
    if(_glfw.platform.platformID==GLFW_PLATFORM_WAYLAND)return fyneReadWaylandOffer(_glfw.wl.selectionOffer,type,size,status);
#endif
    return NULL;
}
char* fyneClipboardFormats(void){
    char* text=calloc(1,32769);if(!text)return NULL;int owner=0;
    if(!_glfw.initialized){free(text);return NULL;}
#if defined(_GLFW_X11)
    if(_glfw.platform.platformID==GLFW_PLATFORM_X11){Window selectionOwner=XGetSelectionOwner(_glfw.x11.display,_glfw.x11.CLIPBOARD);if(selectionOwner==None)return text;owner=selectionOwner==_glfw.x11.helperWindowHandle;}
#endif
#if defined(_GLFW_WAYLAND)
    if(_glfw.platform.platformID==GLFW_PLATFORM_WAYLAND)owner=_glfw.wl.selectionSource!=NULL;
#endif
    if(owner){for(MimePayload* p=ownedMime;p;p=p->next){strcat(text,p->type);strcat(text,"\n");}if(!ownedMime)strcpy(text,"text/plain\n");return text;}
#if defined(_GLFW_X11)
    if(_glfw.platform.platformID==GLFW_PLATFORM_X11){int size,status;Atom* targets=readX11(_glfw.x11.TARGETS,&size,&status,1);if(status){free(text);return NULL;}for(int i=0;i<size/(int)sizeof(Atom);i++){char* name=XGetAtomName(_glfw.x11.display,targets[i]);if(!name)continue;const char* type=name;if(targets[i]==_glfw.x11.UTF8_STRING||targets[i]==XA_STRING)type="text/plain";if(strchr(type,'/')&&strlen(text)+strlen(type)+1<32768){strcat(text,type);strcat(text,"\n");}XFree(name);}free(targets);return text;}
#endif
#if defined(_GLFW_WAYLAND)
    if(_glfw.platform.platformID==GLFW_PLATFORM_WAYLAND){for(MimeOffer* p=offers;p;p=p->next)if(p->offer==_glfw.wl.selectionOffer&&p->types){strcpy(text,p->types);break;}return text;}
#endif
    free(text);return NULL;
}

