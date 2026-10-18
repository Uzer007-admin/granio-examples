#include "glfw/src/internal.h"
#include "_cgo_export.h"
#include <string.h>
#if defined(_GLFW_X11)
#include <time.h>
#endif
typedef struct DropTypes { _GLFWwindow* window; char* types;struct DropTypes* next;} DropTypes;
static DropTypes* dropTypes;
void fyneDropFormats(GLFWwindow* handle,const char* formats){
    DropTypes** link=&dropTypes;while(*link&&(*link)->window!=(_GLFWwindow*)handle)link=&(*link)->next;
    DropTypes* p=*link;if(!*formats){if(p){*link=p->next;free(p->types);free(p);}return;}
    if(strlen(formats)>32768)return;
    if(!p){p=calloc(1,sizeof(*p));if(!p)return;p->window=(_GLFWwindow*)handle;p->next=dropTypes;dropTypes=p;}
    char* types=malloc(strlen(formats)+1);if(!types)return;strcpy(types,formats);free(p->types);p->types=types;
}
static int accepted(_GLFWwindow* window,const char* type){
    for(DropTypes* p=dropTypes;p;p=p->next)if(p->window==window){const char* line=p->types;while(*line){const char* end=strchr(line,'\n');size_t size=end?(size_t)(end-line):strlen(line);if(strlen(type)==size&&!memcmp(line,type,size))return 1;line=end?end+1:line+size;}}
    return 0;
}
#if defined(_GLFW_X11)
typedef struct DropReceive { _GLFWwindow* window;Window source;int version;Atom type;unsigned char* bytes;size_t size;double deadline;struct DropReceive* next; } DropReceive;
static DropReceive* receiving;
static double dropTime(void){struct timespec now;clock_gettime(CLOCK_MONOTONIC,&now);return now.tv_sec+now.tv_nsec*1e-9;}
static void finished(DropReceive* drop,int accepted){if(drop->version<2)return;XEvent event={0};event.xclient.type=ClientMessage;event.xclient.window=drop->source;event.xclient.message_type=_glfw.x11.XdndFinished;event.xclient.format=32;event.xclient.data.l[0]=drop->window->x11.handle;event.xclient.data.l[1]=accepted;event.xclient.data.l[2]=accepted?_glfw.x11.XdndActionCopy:None;XSendEvent(_glfw.x11.display,drop->source,False,NoEventMask,&event);XFlush(_glfw.x11.display);}
int fyneDeliverX11Drop(_GLFWwindow*,Atom,const unsigned char*,unsigned long);
static void deliver(DropReceive* drop){
    fyneDeliverX11Drop(drop->window,drop->type,drop->bytes,drop->size);
    if(drop->type==_glfw.x11.text_uri_list){int count=0;char** paths=_glfwParseUriList((char*)drop->bytes,&count);if(paths){_glfwInputDrop(drop->window,count,(const char**)paths);for(int i=0;i<count;i++)_glfw_free(paths[i]);_glfw_free(paths);}}
    finished(drop,1);
}
static void freeReceive(DropReceive** link){DropReceive* p=*link;*link=p->next;free(p->bytes);free(p);}
void fyneDropPoll(void){DropReceive** link=&receiving;while(*link){if((*link)->deadline<dropTime()){finished(*link,0);freeReceive(link);}else link=&(*link)->next;}}
void fyneDropDispose(_GLFWwindow* window){DropReceive** link=&receiving;while(*link){if((*link)->window==window){finished(*link,0);freeReceive(link);}else link=&(*link)->next;}}
int fyneDropEvent(XEvent* event){
    fyneDropPoll();if(event->type!=PropertyNotify||event->xproperty.state!=PropertyNewValue)return 0;
    DropReceive** link=&receiving;while(*link){DropReceive* p=*link;if(p->window->x11.handle==event->xproperty.window&&event->xproperty.atom==_glfw.x11.XdndSelection){
        Atom type;int format;unsigned long count,left;unsigned char* bytes=NULL;XGetWindowProperty(_glfw.x11.display,p->window->x11.handle,_glfw.x11.XdndSelection,0,64u*1024u*1024u/4,True,AnyPropertyType,&type,&format,&count,&left,&bytes);
        if(left||format!=8||count>64u*1024u*1024u-p->size){if(bytes)XFree(bytes);finished(p,0);freeReceive(link);return 1;}
        unsigned char* resized=realloc(p->bytes,p->size+count+1);if(!resized){if(bytes)XFree(bytes);finished(p,0);freeReceive(link);return 1;}p->bytes=resized;if(count)memcpy(p->bytes+p->size,bytes,count);p->size+=count;p->bytes[p->size]=0;if(bytes)XFree(bytes);XFlush(_glfw.x11.display);
        if(!count){deliver(p);freeReceive(link);}return 1;
    }link=&p->next;}return 0;
}
int fyneHandleX11Drop(_GLFWwindow* window,const XSelectionEvent* event){
    DropReceive state={0};state.window=window;state.source=_glfw.x11.xdnd.source;state.version=_glfw.x11.xdnd.version;state.type=event->target;
    Atom actual;int format;unsigned long count,left;unsigned char* bytes=NULL;
    XGetWindowProperty(_glfw.x11.display,event->requestor,event->property,0,64u*1024u*1024u/4,False,AnyPropertyType,&actual,&format,&count,&left,&bytes);
    if(actual==XInternAtom(_glfw.x11.display,"INCR",False)){
        if(bytes)XFree(bytes);DropReceive* p=calloc(1,sizeof(*p));if(!p){finished(&state,0);return 1;}*p=state;p->deadline=dropTime()+3;p->next=receiving;receiving=p;XDeleteProperty(_glfw.x11.display,event->requestor,event->property);XFlush(_glfw.x11.display);return 1;
    }
    if(left||format!=8||count>64u*1024u*1024u){if(bytes)XFree(bytes);finished(&state,0);return 1;}
    state.bytes=bytes;state.size=count;deliver(&state);if(bytes)XFree(bytes);XDeleteProperty(_glfw.x11.display,event->requestor,event->property);return 1;
}
Atom fyneSelectDropType(_GLFWwindow* window,Atom* offered,unsigned long count){
    // Keep the caller's configured preference order, independently of source order.
    for(DropTypes* p=dropTypes;p;p=p->next)if(p->window==window){const char* line=p->types;while(*line){const char* end=strchr(line,'\n');size_t size=end?(size_t)(end-line):strlen(line);char type[256];if(size<sizeof(type)){memcpy(type,line,size);type[size]=0;Atom atom=XInternAtom(_glfw.x11.display,type,False);for(unsigned long i=0;i<count;i++)if(offered[i]==atom)return atom;}line=end?end+1:line+size;}}
    return None;
}
int fyneDeliverX11Drop(_GLFWwindow* window,Atom type,const unsigned char* bytes,unsigned long size){
    if(size>64u*1024u*1024u)return 0;char* mime=XGetAtomName(_glfw.x11.display,type);if(!mime)return 0;int deliver=accepted(window,mime);if(deliver)goFyneDataDrop((GLFWwindow*)window,mime,(void*)bytes,(int)size);XFree(mime);return deliver;
}
#endif

void fyneDropDisposeWindow(GLFWwindow* window){
#if defined(_GLFW_X11)
    if(_glfw.platform.platformID==GLFW_PLATFORM_X11)fyneDropDispose((_GLFWwindow*)window);
#endif
}
#if defined(_GLFW_WAYLAND)
const char* fyneSelectWaylandDrop(_GLFWwindow* window,const char* types){
    static char selected[256];if(!types)return NULL;
    for(DropTypes* p=dropTypes;p;p=p->next)if(p->window==window){const char* line=p->types;while(*line){const char* end=strchr(line,'\n');size_t size=end?(size_t)(end-line):strlen(line);if(size<sizeof(selected)){memcpy(selected,line,size);selected[size]=0;const char* candidate=types;while(*candidate){const char* next=strchr(candidate,'\n');size_t n=next?(size_t)(next-candidate):strlen(candidate);if(n==size&&!memcmp(line,candidate,size))return selected;candidate=next?next+1:candidate+n;}}line=end?end+1:line+size;}}
    return NULL;
}
int fyneDeliverWaylandDrop(_GLFWwindow* window,const char* type,const void* bytes,int size){if(size<0||size>64*1024*1024||!accepted(window,type))return 0;goFyneDataDrop((GLFWwindow*)window,(char*)type,(void*)bytes,size);return 1;}
#endif
