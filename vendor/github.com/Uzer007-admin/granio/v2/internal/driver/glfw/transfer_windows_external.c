//go:build windows && !wasm && !test_web_driver
#define CINTERFACE
#define COBJMACROS
#include <windows.h>
#include <ole2.h>
#include <stdlib.h>

#include <string.h>
#include "transfer_windows.h"

// This source is used by a separate test process. The payload crosses real OLE
// IPC; it does not call Fyne's target callback or mutate the system clipboard.
typedef struct ExternalData {IDataObject data;LONG refs;CLIPFORMAT format;const void*bytes;int size;} ExternalData;
typedef struct ExternalDrag {IDropSource source;LONG refs;DWORD started;} ExternalDrag;
typedef struct ExternalFormats {IEnumFORMATETC formats;LONG refs;CLIPFORMAT format;int cursor;} ExternalFormats;
static const IID unknownIID={0,0,0,{0xc0,0,0,0,0,0,0,0x46}};
static const IID dataIID={0x0000010e,0,0,{0xc0,0,0,0,0,0,0,0x46}};
static const IID sourceIID={0x00000121,0,0,{0xc0,0,0,0,0,0,0,0x46}};
static const IID enumIID={0x00000103,0,0,{0xc0,0,0,0,0,0,0,0x46}};
static ULONG STDMETHODCALLTYPE O_A(IDataObject*p){return InterlockedIncrement(&((ExternalData*)p)->refs);}
static ULONG STDMETHODCALLTYPE O_R(IDataObject*p){ExternalData*d=(ExternalData*)p;ULONG n=InterlockedDecrement(&d->refs);if(!n){free((void*)d->bytes);free(d);}return n;}
static HRESULT STDMETHODCALLTYPE O_Q(IDataObject*p,REFIID iid,void**out){if(!out)return E_POINTER;*out=NULL;if(IsEqualIID(iid,&unknownIID)||IsEqualIID(iid,&dataIID)){*out=p;O_A(p);return S_OK;}return E_NOINTERFACE;}
static HRESULT STDMETHODCALLTYPE O_QueryData(IDataObject*p,FORMATETC*f){ExternalData*d=(ExternalData*)p;if(!f)return E_POINTER;if(f->cfFormat!=d->format)return DV_E_FORMATETC;if(!(f->tymed&TYMED_HGLOBAL))return DV_E_TYMED;if(f->dwAspect!=DVASPECT_CONTENT||f->lindex!=-1)return DV_E_DVASPECT;return S_OK;}
static HRESULT STDMETHODCALLTYPE O_Get(IDataObject*p,FORMATETC*f,STGMEDIUM*out){if(!out)return E_POINTER;memset(out,0,sizeof(*out));HRESULT hr=O_QueryData(p,f);if(FAILED(hr))return hr;ExternalData*d=(ExternalData*)p;out->hGlobal=GlobalAlloc(GMEM_MOVEABLE|GMEM_ZEROINIT,max(1,d->size));if(!out->hGlobal)return E_OUTOFMEMORY;void*bytes=GlobalLock(out->hGlobal);if(!bytes){GlobalFree(out->hGlobal);return E_OUTOFMEMORY;}if(d->size)memcpy(bytes,d->bytes,d->size);GlobalUnlock(out->hGlobal);out->tymed=TYMED_HGLOBAL;return S_OK;}
static HRESULT STDMETHODCALLTYPE O_GetHere(IDataObject*p,FORMATETC*f,STGMEDIUM*out){return E_NOTIMPL;}
static HRESULT STDMETHODCALLTYPE O_Canonical(IDataObject*p,FORMATETC*f,FORMATETC*out){if(!out)return E_POINTER;*out=*f;out->ptd=NULL;return DATA_S_SAMEFORMATETC;}
static HRESULT STDMETHODCALLTYPE O_Set(IDataObject*p,FORMATETC*f,STGMEDIUM*m,BOOL release){return E_NOTIMPL;}
static ULONG STDMETHODCALLTYPE EF_A(IEnumFORMATETC*p){return InterlockedIncrement(&((ExternalFormats*)p)->refs);}
static ULONG STDMETHODCALLTYPE EF_R(IEnumFORMATETC*p){ExternalFormats*f=(ExternalFormats*)p;ULONG n=InterlockedDecrement(&f->refs);if(!n)free(f);return n;}
static HRESULT STDMETHODCALLTYPE EF_Q(IEnumFORMATETC*p,REFIID iid,void**out){if(!out)return E_POINTER;*out=NULL;if(IsEqualIID(iid,&unknownIID)||IsEqualIID(iid,&enumIID)){*out=p;EF_A(p);return S_OK;}return E_NOINTERFACE;}
static HRESULT STDMETHODCALLTYPE EF_Next(IEnumFORMATETC*p,ULONG count,FORMATETC*out,ULONG*fetched){if(!out||(!fetched&&count!=1))return E_POINTER;if(fetched)*fetched=0;ExternalFormats*f=(ExternalFormats*)p;if(!count)return S_OK;if(f->cursor)return S_FALSE;FORMATETC value={f->format,NULL,DVASPECT_CONTENT,-1,TYMED_HGLOBAL};out[0]=value;f->cursor=1;if(fetched)*fetched=1;return count==1?S_OK:S_FALSE;}
static HRESULT STDMETHODCALLTYPE EF_Skip(IEnumFORMATETC*p,ULONG count){ExternalFormats*f=(ExternalFormats*)p;int available=!f->cursor;f->cursor=1;return count<=(ULONG)available?S_OK:S_FALSE;}
static HRESULT STDMETHODCALLTYPE EF_Reset(IEnumFORMATETC*p){((ExternalFormats*)p)->cursor=0;return S_OK;}
static IEnumFORMATETCVtbl enumVtable;
static HRESULT STDMETHODCALLTYPE EF_Clone(IEnumFORMATETC*p,IEnumFORMATETC**out){if(!out)return E_POINTER;ExternalFormats*copy=malloc(sizeof(*copy));if(!copy)return E_OUTOFMEMORY;*copy=*((ExternalFormats*)p);copy->refs=1;*out=&copy->formats;return S_OK;}
static IEnumFORMATETCVtbl enumVtable={EF_Q,EF_A,EF_R,EF_Next,EF_Skip,EF_Reset,EF_Clone};
static HRESULT STDMETHODCALLTYPE O_Enum(IDataObject*p,DWORD direction,IEnumFORMATETC**out){if(!out)return E_POINTER;*out=NULL;if(direction!=DATADIR_GET)return E_NOTIMPL;ExternalFormats*f=calloc(1,sizeof(*f));if(!f)return E_OUTOFMEMORY;f->formats.lpVtbl=&enumVtable;f->refs=1;f->format=((ExternalData*)p)->format;*out=&f->formats;return S_OK;}
static HRESULT STDMETHODCALLTYPE O_Advise(IDataObject*p,FORMATETC*f,DWORD flags,IAdviseSink*sink,DWORD*connection){if(connection)*connection=0;return OLE_E_ADVISENOTSUPPORTED;}
static HRESULT STDMETHODCALLTYPE O_Unadvise(IDataObject*p,DWORD connection){return OLE_E_ADVISENOTSUPPORTED;}
static HRESULT STDMETHODCALLTYPE O_EnumAdvise(IDataObject*p,IEnumSTATDATA**out){if(out)*out=NULL;return OLE_E_ADVISENOTSUPPORTED;}
static IDataObjectVtbl dataVtable={O_Q,O_A,O_R,O_Get,O_GetHere,O_QueryData,O_Canonical,O_Set,O_Enum,O_Advise,O_Unadvise,O_EnumAdvise};
static ULONG STDMETHODCALLTYPE S_A(IDropSource*p){return InterlockedIncrement(&((ExternalDrag*)p)->refs);}
static ULONG STDMETHODCALLTYPE S_R(IDropSource*p){ExternalDrag*s=(ExternalDrag*)p;ULONG n=InterlockedDecrement(&s->refs);if(!n)free(s);return n;}
static HRESULT STDMETHODCALLTYPE S_Q(IDropSource*p,REFIID iid,void**out){if(!out)return E_POINTER;*out=NULL;if(IsEqualIID(iid,&unknownIID)||IsEqualIID(iid,&sourceIID)){*out=p;S_A(p);return S_OK;}return E_NOINTERFACE;}
static HRESULT STDMETHODCALLTYPE S_Continue(IDropSource*p,BOOL escape,DWORD keys){ExternalDrag*s=(ExternalDrag*)p;if(escape)return DRAGDROP_S_CANCEL;return GetTickCount()-s->started>=250?DRAGDROP_S_DROP:S_OK;}
static HRESULT STDMETHODCALLTYPE S_Feedback(IDropSource*p,DWORD effect){return DRAGDROP_S_USEDEFAULTCURSORS;}
static IDropSourceVtbl sourceVtable={S_Q,S_A,S_R,S_Continue,S_Feedback};
static DWORD WINAPI driveExternalCursor(void*position){POINT point=*(POINT*)position;for(int i=0;i<12;i++){Sleep(40);SetCursorPos(point.x+(i%2),point.y);if(i==8){INPUT input={0};input.type=INPUT_MOUSE;input.mi.dwFlags=MOUSEEVENTF_LEFTUP;SendInput(1,&input,sizeof(input));}}return 0;}
int FyneExternalDropTest(uintptr_t handle,unsigned int format,const void*bytes,int size){HWND hwnd=(HWND)handle;if(size<0||size>64*1024*1024||!IsWindow(hwnd))return 0;HRESULT hr=OleInitialize(NULL);if(FAILED(hr))return 0;ExternalData*data=calloc(1,sizeof(*data));ExternalDrag*source=calloc(1,sizeof(*source));if(!data||!source){free(data);free(source);OleUninitialize();return 0;}data->data.lpVtbl=&dataVtable;data->refs=1;data->format=format;void*owned=malloc(max(1,size));if(!owned){free(data);free(source);OleUninitialize();return 0;}if(size)memcpy(owned,bytes,size);data->bytes=owned;data->size=size;source->source.lpVtbl=&sourceVtable;source->refs=1;source->started=GetTickCount();POINT old={0,0},target={0,0};RECT rect;GetCursorPos(&old);GetClientRect(hwnd,&rect);target.x=(rect.right-rect.left)/2;target.y=(rect.bottom-rect.top)/2;ClientToScreen(hwnd,&target);HWND prior=GetForegroundWindow();HWND sourceWindow=CreateWindowW(L"STATIC",L"Native external drag source",WS_OVERLAPPEDWINDOW,20,20,180,100,NULL,NULL,NULL,NULL);ShowWindow(sourceWindow,SW_SHOW);SetForegroundWindow(sourceWindow);SetCursorPos(80,70);INPUT pressed={0};pressed.type=INPUT_MOUSE;pressed.mi.dwFlags=MOUSEEVENTF_LEFTDOWN;SendInput(1,&pressed,sizeof(pressed));HANDLE movement=CreateThread(NULL,0,driveExternalCursor,&target,0,NULL);DWORD effect=0;hr=DoDragDrop(&data->data,&source->source,DROPEFFECT_COPY,&effect);if(movement){WaitForSingleObject(movement,1000);CloseHandle(movement);}INPUT released={0};released.type=INPUT_MOUSE;released.mi.dwFlags=MOUSEEVENTF_LEFTUP;SendInput(1,&released,sizeof(released));DestroyWindow(sourceWindow);if(IsWindow(prior))SetForegroundWindow(prior);SetCursorPos(old.x,old.y);O_R(&data->data);S_R(&source->source);OleUninitialize();return hr==DRAGDROP_S_DROP&&effect==DROPEFFECT_COPY;}
