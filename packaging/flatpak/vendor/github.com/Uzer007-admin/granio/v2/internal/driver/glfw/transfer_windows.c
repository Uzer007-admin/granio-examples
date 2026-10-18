//go:build windows && !wasm && !test_web_driver
#define CINTERFACE
#define COBJMACROS
#include <windows.h>
#include <ole2.h>
#include <stdlib.h>

#include <string.h>
#include "transfer_windows.h"
#include "_cgo_export.h"
typedef struct FyneDropTarget {
    IDropTarget provider;
    LONG refs;
    HWND hwnd;
    unsigned int formats[130];
    char* mime[130];
    int count, detached, accepted;
    struct FyneDropTarget* next;
} FyneDropTarget;
static FyneDropTarget* targets;
static const IID unknownIID={0,0,0,{0xc0,0,0,0,0,0,0,0x46}};
static const IID dropIID={0x00000122,0,0,{0xc0,0,0,0,0,0,0,0x46}};
static ULONG STDMETHODCALLTYPE D_AddRef(IDropTarget* p){return InterlockedIncrement(&((FyneDropTarget*)p)->refs);}
static ULONG STDMETHODCALLTYPE D_Release(IDropTarget* p){FyneDropTarget* t=(FyneDropTarget*)p;ULONG refs=InterlockedDecrement(&t->refs);if(!refs){for(int i=0;i<t->count;i++)free(t->mime[i]);free(t);}return refs;}
static HRESULT STDMETHODCALLTYPE D_Query(IDropTarget* p,REFIID iid,void** out){if(!out)return E_POINTER;*out=NULL;if(IsEqualIID(iid,&unknownIID)||IsEqualIID(iid,&dropIID)){*out=p;D_AddRef(p);return S_OK;}return E_NOINTERFACE;}
static int choose(FyneDropTarget* target,IDataObject* data){
    if(target->detached)return -1;
    for(int i=0;i<target->count;i++){FORMATETC format={(CLIPFORMAT)target->formats[i],NULL,DVASPECT_CONTENT,-1,TYMED_HGLOBAL};if(SUCCEEDED(IDataObject_QueryGetData(data,&format)))return i;}return -1;
}
static HRESULT STDMETHODCALLTYPE D_Enter(IDropTarget* p,IDataObject* data,DWORD keys,POINTL point,DWORD* effect){if(!effect)return E_POINTER;FyneDropTarget*t=(FyneDropTarget*)p;t->accepted=choose(t,data)>=0;*effect=t->accepted?(*effect&DROPEFFECT_COPY):DROPEFFECT_NONE;return S_OK;}
static HRESULT STDMETHODCALLTYPE D_Over(IDropTarget* p,DWORD keys,POINTL point,DWORD* effect){if(!effect)return E_POINTER;FyneDropTarget*t=(FyneDropTarget*)p;if(t->detached||!t->accepted)*effect=DROPEFFECT_NONE;else *effect&=DROPEFFECT_COPY;return S_OK;}
static HRESULT STDMETHODCALLTYPE D_Leave(IDropTarget* p){((FyneDropTarget*)p)->accepted=0;return S_OK;}
static HRESULT STDMETHODCALLTYPE D_Drop(IDropTarget* p,IDataObject* data,DWORD keys,POINTL point,DWORD* effect){
    if(!effect)return E_POINTER;FyneDropTarget* t=(FyneDropTarget*)p;int selected=choose(t,data);*effect=DROPEFFECT_NONE;if(selected<0)return S_OK;
    FORMATETC format={(CLIPFORMAT)t->formats[selected],NULL,DVASPECT_CONTENT,-1,TYMED_HGLOBAL};STGMEDIUM medium;memset(&medium,0,sizeof(medium));if(FAILED(IDataObject_GetData(data,&format,&medium)))return S_OK;
    if(medium.tymed==TYMED_HGLOBAL){SIZE_T size=GlobalSize(medium.hGlobal);if(size<=64u*1024u*1024u){void* bytes=GlobalLock(medium.hGlobal);if(bytes){POINT client={point.x,point.y};ScreenToClient(t->hwnd,&client);goWinDataDrop((uintptr_t)t->hwnd,t->formats[selected],t->mime[selected],bytes,(int)size,client.x,client.y);GlobalUnlock(medium.hGlobal);*effect=DROPEFFECT_COPY;}}}
    ReleaseStgMedium(&medium);return S_OK;
}
static IDropTargetVtbl dropVtbl={D_Query,D_AddRef,D_Release,D_Enter,D_Over,D_Leave,D_Drop};
static FyneDropTarget* findTarget(HWND hwnd){for(FyneDropTarget* t=targets;t;t=t->next)if(t->hwnd==hwnd)return t;return NULL;}
int FyneDropBegin(uintptr_t handle){
    HWND hwnd=(HWND)handle;FyneDropTarget* target=findTarget(hwnd);
    if(target){for(int i=0;i<target->count;i++)free(target->mime[i]);target->count=0;return 1;}
    HRESULT initialised=OleInitialize(NULL);if(FAILED(initialised))return 0;
    target=calloc(1,sizeof(*target));if(!target){OleUninitialize();return 0;}target->provider.lpVtbl=&dropVtbl;target->refs=1;target->hwnd=hwnd;
    if(FAILED(RegisterDragDrop(hwnd,&target->provider))){D_Release(&target->provider);OleUninitialize();return 0;}
    target->next=targets;targets=target;return 1;
}
void FyneDropAdd(uintptr_t handle,unsigned int format,const char* mime){FyneDropTarget* target=findTarget((HWND)handle);if(!target||target->count>=130)return;for(int i=0;i<target->count;i++)if(target->formats[i]==format)return;char* name=malloc(strlen(mime)+1);if(!name)return;strcpy(name,mime);target->formats[target->count]=format;target->mime[target->count++]=name;}
void FyneDropRemove(uintptr_t handle){FyneDropTarget** link=&targets;while(*link&&(*link)->hwnd!=(HWND)handle)link=&(*link)->next;FyneDropTarget* target=*link;if(!target)return;*link=target->next;target->detached=1;RevokeDragDrop(target->hwnd);D_Release(&target->provider);OleUninitialize();}

