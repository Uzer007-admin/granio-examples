//go:build windows && !wasm && !test_web_driver
#define CINTERFACE
#define COBJMACROS
#include <windows.h>
#include <ole2.h>
#include <olectl.h>
#include <textstor.h>
#include <msctf.h>
#include <stdlib.h>
#include <string.h>
#include <stddef.h>
#include <stdio.h>
#include "ime_tsf_windows.h"
#include "_cgo_export.h"

// Each HWND owns a document/context; the native UI thread owns the TSF manager.
// ACP indices are UTF-16 code units. Neither TSF nor IMM sees password contents.
#define MAX_ACP (1<<19)
#define LOCK_MESSAGE (WM_APP+0x3C0)
typedef struct TextStore {
    ITextStoreACP acp;
    ITfContextOwnerCompositionSink composition;
    LONG refs;
    HWND hwnd;
    DWORD thread, lock, pending, mask;
    int enabled, composing, cancelling, detached, dirtySelection;
    WCHAR *text, *original;
    LONG length, originalLength, start, end, cursor, originalStart, originalEnd, compStart, compEnd;
    RECT caret;
    ITextStoreACPSink *sink;
    ITfDocumentMgr *document;
    ITfContext *context;
    TfEditCookie cookie;
    LONG lastKeyTime;
    intptr_t lastKeyScan;
    struct TextStore *next;
} TextStore;
static TextStore *stores;
static ITfThreadMgr *manager;
static ITfKeystrokeMgr *keys;
static TfClientId client;
static int coOwned;
static const IID acpIID={0x28888fe3,0xc2a0,0x483a,{0xa3,0xea,0x8c,0xb1,0xce,0x51,0xff,0x3d}};
static const IID sinkIID={0x22d44c94,0xa419,0x4542,{0xa2,0x72,0xae,0x26,0x09,0x3e,0xce,0xcf}};
static const IID compositionIID={0x5f20aa40,0xb57a,0x4f34,{0x96,0xab,0x35,0x76,0xf3,0x77,0xcc,0x79}};
static const IID unknownIID={0,0,0,{0xc0,0,0,0,0,0,0,0x46}};
#define STORE(p) ((TextStore*)(p))
#define COMPOSITION(p) ((TextStore*)((char*)(p)-offsetof(TextStore,composition)))
static int boundary(TextStore *s,LONG pos) {
    return pos>=0 && pos<=s->length && !(pos>0 && pos<s->length && s->text[pos-1]>=0xD800 && s->text[pos-1]<=0xDBFF && s->text[pos]>=0xDC00 && s->text[pos]<=0xDFFF);
}
static HRESULT readable(TextStore *s) { if(s->detached)return TS_E_NOLOCK;return (s->lock&TS_LF_READ)?S_OK:TS_E_NOLOCK; }
static HRESULT writable(TextStore *s) { if(s->detached)return TS_E_NOLOCK;return (s->lock&TS_LF_READWRITE)==TS_LF_READWRITE&&s->enabled?S_OK:TS_E_NOLOCK; }
static ULONG retain(TextStore *s){return InterlockedIncrement(&s->refs);}
static ULONG release(TextStore *s){ULONG n=InterlockedDecrement(&s->refs);if(!n){if(s->sink)ITextStoreACPSink_Release(s->sink);free(s->text);free(s->original);free(s);}return n;}
static HRESULT query(TextStore *s,REFIID iid,void **out){if(!out)return E_POINTER;*out=NULL;if(IsEqualIID(iid,&unknownIID)||IsEqualIID(iid,&acpIID))*out=&s->acp;else if(IsEqualIID(iid,&compositionIID))*out=&s->composition;if(*out){retain(s);return S_OK;}return E_NOINTERFACE;}
static HRESULT STDMETHODCALLTYPE Q(ITextStoreACP*p,REFIID i,void**o){return query(STORE(p),i,o);}
static ULONG STDMETHODCALLTYPE A(ITextStoreACP*p){return retain(STORE(p));}
static ULONG STDMETHODCALLTYPE R(ITextStoreACP*p){return release(STORE(p));}
static HRESULT STDMETHODCALLTYPE Advise(ITextStoreACP*p,REFIID iid,IUnknown *unk,DWORD mask){
    TextStore*s=STORE(p);if(!unk||!IsEqualIID(iid,&sinkIID))return E_INVALIDARG;
    ITextStoreACPSink *sink=NULL;HRESULT hr=IUnknown_QueryInterface(unk,&sinkIID,(void**)&sink);if(FAILED(hr))return hr;
    if(s->sink){IUnknown *a=NULL,*b=NULL;ITextStoreACPSink_QueryInterface(s->sink,&unknownIID,(void**)&a);ITextStoreACPSink_QueryInterface(sink,&unknownIID,(void**)&b);int same=a==b;if(a)IUnknown_Release(a);if(b)IUnknown_Release(b);if(!same){ITextStoreACPSink_Release(sink);return CONNECT_E_ADVISELIMIT;}ITextStoreACPSink_Release(s->sink);}
    s->sink=sink;s->mask=mask;return S_OK;
}
static HRESULT STDMETHODCALLTYPE Unadvise(ITextStoreACP*p,IUnknown *unk){TextStore*s=STORE(p);if(!s->sink||!unk)return CONNECT_E_NOCONNECTION;IUnknown *a=NULL,*b=NULL;ITextStoreACPSink_QueryInterface(s->sink,&unknownIID,(void**)&a);IUnknown_QueryInterface(unk,&unknownIID,(void**)&b);int same=a==b;if(a)IUnknown_Release(a);if(b)IUnknown_Release(b);if(!same)return CONNECT_E_NOCONNECTION;ITextStoreACPSink_Release(s->sink);s->sink=NULL;return S_OK;}
static HRESULT grant(TextStore*s,DWORD flags){retain(s);s->lock=flags&TS_LF_READWRITE;HRESULT hr=s->sink?ITextStoreACPSink_OnLockGranted(s->sink,s->lock):E_UNEXPECTED;s->lock=0;if(s->dirtySelection&&!s->composing&&!s->detached&&s->hwnd){goWinTSFSelection((uintptr_t)s->hwnd,(unsigned short*)s->text,s->length,s->start,s->end);s->dirtySelection=0;}release(s);return hr;}
static HRESULT STDMETHODCALLTYPE Lock(ITextStoreACP*p,DWORD flags,HRESULT *session){TextStore*s=STORE(p);if(!session)return E_POINTER;if(GetCurrentThreadId()!=s->thread)return RPC_E_WRONG_THREAD;if(!s->sink)return E_UNEXPECTED;if(!(flags&TS_LF_READ))return E_INVALIDARG;if(s->lock){if(flags&TS_LF_SYNC){*session=TS_E_SYNCHRONOUS;return S_OK;}s->pending|=flags&TS_LF_READWRITE;*session=TS_S_ASYNC;return S_OK;}*session=grant(s,flags);if(s->pending){DWORD pending=s->pending;s->pending=0;grant(s,pending);if(s->pending&&s->hwnd)PostMessageW(s->hwnd,LOCK_MESSAGE,0,0);}return S_OK;}
static HRESULT STDMETHODCALLTYPE Status(ITextStoreACP*p,TS_STATUS*out){if(!out)return E_POINTER;out->dwDynamicFlags=STORE(p)->enabled?0:TS_SD_READONLY;out->dwStaticFlags=TS_SS_NOHIDDENTEXT;return S_OK;}
static HRESULT STDMETHODCALLTYPE QueryInsert(ITextStoreACP*p,LONG a,LONG b,ULONG count,LONG*outA,LONG*outB){TextStore*s=STORE(p);if(!outA||!outB)return E_POINTER;if(!boundary(s,a)||!boundary(s,b)||b<a||count>MAX_ACP-(s->length-(b-a)))return E_INVALIDARG;*outA=a;*outB=b;return S_OK;}
static HRESULT STDMETHODCALLTYPE GetSelection(ITextStoreACP*p,ULONG index,ULONG count,TS_SELECTION_ACP*out,ULONG*fetched){TextStore*s=STORE(p);if(!fetched||(!out&&count))return E_POINTER;*fetched=0;HRESULT hr=readable(s);if(FAILED(hr))return hr;if(index!=0&&index!=TS_DEFAULT_SELECTION)return E_INVALIDARG;if(count){out->acpStart=s->start;out->acpEnd=s->end;out->style.ase=s->cursor==s->start?TS_AE_START:TS_AE_END;out->style.fInterimChar=FALSE;*fetched=1;}return S_OK;}
static HRESULT STDMETHODCALLTYPE SetSelection(ITextStoreACP*p,ULONG count,const TS_SELECTION_ACP*in){TextStore*s=STORE(p);HRESULT hr=writable(s);if(FAILED(hr))return hr;if(count!=1||!in||!boundary(s,in->acpStart)||!boundary(s,in->acpEnd)||in->acpEnd<in->acpStart)return E_INVALIDARG;s->start=in->acpStart;s->end=in->acpEnd;s->cursor=in->style.ase==TS_AE_START?s->start:s->end;s->dirtySelection=1;return S_OK;}
static HRESULT STDMETHODCALLTYPE GetText(ITextStoreACP*p,LONG a,LONG b,WCHAR*out,ULONG capacity,ULONG*copied,TS_RUNINFO*runs,ULONG runCapacity,ULONG*runCount,LONG*next){TextStore*s=STORE(p);if(!copied||!runCount||!next||(!out&&capacity)||(!runs&&runCapacity))return E_POINTER;*copied=*runCount=0;HRESULT hr=readable(s);if(FAILED(hr))return hr;if(b==-1)b=s->length;if(a<0||b<a||b>s->length)return E_INVALIDARG;ULONG n=capacity?(ULONG)(b-a)<capacity?(ULONG)(b-a):capacity:(ULONG)(b-a);if(capacity){memcpy(out,s->text+a,n*sizeof(WCHAR));*copied=n;}if(runCapacity&&n){runs->uCount=n;runs->type=TS_RT_PLAIN;*runCount=1;}*next=a+n;return S_OK;}
static void emitDiff(TextStore*s,const WCHAR*old,LONG oldLength,int kind){if(!s->hwnd||s->detached)return;LONG a=0,b=oldLength,n=s->length;while(a<b&&a<n&&old[a]==s->text[a])a++;while(b>a&&n>a&&old[b-1]==s->text[n-1]){b--;n--;}if(a>0&&a<oldLength&&old[a]>=0xDC00&&old[a]<=0xDFFF){a--;}
    if(b<oldLength&&b>0&&old[b-1]>=0xD800&&old[b-1]<=0xDBFF){b++;n++;}
    if(a==b&&a==n){if(kind==0)goWinTSFEdit((uintptr_t)s->hwnd,2,(unsigned short*)old,oldLength,NULL,0,0,0,0);return;}
    goWinTSFEdit((uintptr_t)s->hwnd,kind,(unsigned short*)old,oldLength,(unsigned short*)(s->text+a),n-a,a,b,s->cursor-a);
}
static void emitComposition(TextStore*s){if(!s->original)return;LONG a=s->compStart,b=s->compEnd;if(a<0||b<a||b>s->length){emitDiff(s,s->original,s->originalLength,1);return;}goWinTSFEdit((uintptr_t)s->hwnd,1,(unsigned short*)s->original,s->originalLength,(unsigned short*)(s->text+a),b-a,s->originalStart,s->originalEnd,s->cursor-a);}
static HRESULT STDMETHODCALLTYPE SetText(ITextStoreACP*p,DWORD flags,LONG a,LONG b,const WCHAR*text,ULONG n,TS_TEXTCHANGE*change){TextStore*s=STORE(p);HRESULT hr=writable(s);if(FAILED(hr))return hr;LONG from,to;if(!change||(!text&&n)||FAILED(QueryInsert(p,a,b,n,&from,&to)))return E_INVALIDARG;LONG length=s->length-(b-a)+n;WCHAR*next=calloc((size_t)length+1,sizeof(WCHAR));if(!next)return E_OUTOFMEMORY;memcpy(next,s->text,a*sizeof(WCHAR));if(n)memcpy(next+a,text,n*sizeof(WCHAR));memcpy(next+a+n,s->text+b,(s->length-b)*sizeof(WCHAR));WCHAR*old=s->text;LONG oldLength=s->length;s->text=next;s->length=length;s->cursor=s->start=s->end=a+n;change->acpStart=a;change->acpOldEnd=b;change->acpNewEnd=a+n;
    if(s->composing){s->compEnd+=n-(b-a);emitComposition(s);}else emitDiff(s,old,oldLength,0);free(old);return S_OK;}
static HRESULT STDMETHODCALLTYPE Formatted(ITextStoreACP*p,LONG a,LONG b,IDataObject**out){if(!out)return E_POINTER;*out=NULL;return E_NOTIMPL;}
static HRESULT STDMETHODCALLTYPE Embedded(ITextStoreACP*p,LONG a,REFGUID service,REFIID iid,IUnknown**out){if(!out)return E_POINTER;*out=NULL;return E_NOTIMPL;}
static HRESULT STDMETHODCALLTYPE QueryEmbedded(ITextStoreACP*p,const GUID*g,const FORMATETC*f,BOOL*out){if(!out)return E_POINTER;*out=FALSE;return S_OK;}
static HRESULT STDMETHODCALLTYPE InsertEmbedded(ITextStoreACP*p,DWORD f,LONG a,LONG b,IDataObject*d,TS_TEXTCHANGE*c){return E_NOTIMPL;}
static HRESULT STDMETHODCALLTYPE InsertText(ITextStoreACP*p,DWORD flags,const WCHAR*text,ULONG n,LONG*a,LONG*b,TS_TEXTCHANGE*c){TextStore*s=STORE(p);HRESULT hr=readable(s);if(FAILED(hr))return hr;if(n>MAX_ACP-(s->length-(s->end-s->start)))return E_INVALIDARG;if(flags&TS_IAS_QUERYONLY){if(!a||!b)return E_POINTER;*a=s->start;*b=s->end;return S_OK;}LONG from=s->start,to=s->end;hr=SetText(p,0,from,to,text,n,c);if(SUCCEEDED(hr)&&!(flags&TS_IAS_NOQUERY)){if(a)*a=from;if(b)*b=from+n;}return hr;}
static HRESULT STDMETHODCALLTYPE InsertEmbeddedSelection(ITextStoreACP*p,DWORD f,IDataObject*d,LONG*a,LONG*b,TS_TEXTCHANGE*c){return E_NOTIMPL;}
static HRESULT STDMETHODCALLTYPE Attrs(ITextStoreACP*p,DWORD f,ULONG n,const TS_ATTRID*a){return S_OK;}
static HRESULT STDMETHODCALLTYPE AttrsAt(ITextStoreACP*p,LONG pos,ULONG n,const TS_ATTRID*a,DWORD f){return S_OK;}
static HRESULT STDMETHODCALLTYPE AttrTransition(ITextStoreACP*p,LONG a,LONG b,ULONG n,const TS_ATTRID*ids,DWORD f,LONG*next,BOOL*found,LONG*offset){if(!next||!found||!offset)return E_POINTER;*next=b;*found=FALSE;*offset=0;return S_OK;}
static HRESULT STDMETHODCALLTYPE Retrieve(ITextStoreACP*p,ULONG count,TS_ATTRVAL*values,ULONG*fetched){if(!fetched)return E_POINTER;*fetched=0;return S_OK;}
static HRESULT STDMETHODCALLTYPE End(ITextStoreACP*p,LONG*out){if(!out)return E_POINTER;HRESULT hr=readable(STORE(p));if(SUCCEEDED(hr))*out=STORE(p)->length;return hr;}
static HRESULT STDMETHODCALLTYPE View(ITextStoreACP*p,TsViewCookie*out){if(!out)return E_POINTER;*out=1;return S_OK;}
static HRESULT STDMETHODCALLTYPE FromPoint(ITextStoreACP*p,TsViewCookie view,const POINT*point,DWORD flags,LONG*out){if(!point||!out)return E_POINTER;HRESULT hr=readable(STORE(p));if(FAILED(hr))return hr;if(view!=1)return E_INVALIDARG;POINT client=*point;ScreenToClient(STORE(p)->hwnd,&client);LONG acp=goWinTSFPoint((uintptr_t)STORE(p)->hwnd,client.x,client.y);if(acp<0)return TS_E_INVALIDPOINT;*out=acp;return S_OK;}
static HRESULT STDMETHODCALLTYPE TextExt(ITextStoreACP*p,TsViewCookie view,LONG a,LONG b,RECT*out,BOOL*clipped){if(!out||!clipped)return E_POINTER;TextStore*s=STORE(p);HRESULT hr=readable(s);if(FAILED(hr))return hr;if(view!=1||a<0||b<a||b>s->length)return E_INVALIDARG;int left,top,right,bottom;if(s->hwnd&&goWinTSFGeometry((uintptr_t)s->hwnd,a,b,&left,&top,&right,&bottom)){out->left=left;out->top=top;out->right=right;out->bottom=bottom;}else return TS_E_NOLAYOUT;POINT origin={0,0};ClientToScreen(s->hwnd,&origin);OffsetRect(out,origin.x,origin.y);*clipped=FALSE;return S_OK;}
static HRESULT STDMETHODCALLTYPE ScreenExt(ITextStoreACP*p,TsViewCookie view,RECT*out){if(!out)return E_POINTER;if(view!=1)return E_INVALIDARG;TextStore*s=STORE(p);GetClientRect(s->hwnd,out);POINT origin={0,0};ClientToScreen(s->hwnd,&origin);OffsetRect(out,origin.x,origin.y);return S_OK;}
static HRESULT STDMETHODCALLTYPE Wnd(ITextStoreACP*p,TsViewCookie view,HWND*out){if(!out)return E_POINTER;if(view!=1)return E_INVALIDARG;*out=STORE(p)->hwnd;return S_OK;}
static ITextStoreACPVtbl vtable={Q,A,R,Advise,Unadvise,Lock,Status,QueryInsert,GetSelection,SetSelection,GetText,SetText,Formatted,Embedded,QueryEmbedded,InsertEmbedded,InsertText,InsertEmbeddedSelection,Attrs,AttrsAt,AttrsAt,AttrTransition,Retrieve,End,View,FromPoint,TextExt,ScreenExt,Wnd};
static HRESULT STDMETHODCALLTYPE CQ(ITfContextOwnerCompositionSink*p,REFIID i,void**o){return query(COMPOSITION(p),i,o);}
static ULONG STDMETHODCALLTYPE CA(ITfContextOwnerCompositionSink*p){return retain(COMPOSITION(p));}
static ULONG STDMETHODCALLTYPE CR(ITfContextOwnerCompositionSink*p){return release(COMPOSITION(p));}
static int extent(ITfCompositionView*composition,LONG*start,LONG*count){ITfRange*range=NULL;ITfRangeACP*acp=NULL;int ok=SUCCEEDED(ITfCompositionView_GetRange(composition,&range))&&SUCCEEDED(ITfRange_QueryInterface(range,&IID_ITfRangeACP,(void**)&acp));if(ok)ok=SUCCEEDED(ITfRangeACP_GetExtent(acp,start,count));if(acp)ITfRangeACP_Release(acp);if(range)ITfRange_Release(range);return ok;}
static HRESULT STDMETHODCALLTYPE StartComposition(ITfContextOwnerCompositionSink*p,ITfCompositionView*v,BOOL*accept){if(!accept)return E_POINTER;TextStore*s=COMPOSITION(p);*accept=FALSE;if(!s->enabled||s->composing)return S_OK;s->original=calloc((size_t)s->length+1,sizeof(WCHAR));if(!s->original)return E_OUTOFMEMORY;memcpy(s->original,s->text,s->length*sizeof(WCHAR));s->originalLength=s->length;s->originalStart=s->start;s->originalEnd=s->end;s->compStart=s->start;s->compEnd=s->end;LONG a,n;if(v&&extent(v,&a,&n)&&a>=0&&n>=0&&a+n<=s->length){s->compStart=s->originalStart=a;s->compEnd=s->originalEnd=a+n;}s->composing=1;*accept=TRUE;return S_OK;}
static HRESULT STDMETHODCALLTYPE UpdateComposition(ITfContextOwnerCompositionSink*p,ITfCompositionView*v,ITfRange*newRange){TextStore*s=COMPOSITION(p);if(!s->composing)return S_OK;ITfRangeACP*acp=NULL;LONG a,n;if(newRange&&SUCCEEDED(ITfRange_QueryInterface(newRange,&IID_ITfRangeACP,(void**)&acp))){if(SUCCEEDED(ITfRangeACP_GetExtent(acp,&a,&n))&&a>=0&&n>=0&&a+n<=s->length){s->compStart=a;s->compEnd=a+n;}ITfRangeACP_Release(acp);}emitComposition(s);return S_OK;}
static HRESULT STDMETHODCALLTYPE EndComposition(ITfContextOwnerCompositionSink*p,ITfCompositionView*v){TextStore*s=COMPOSITION(p);if(!s->composing)return S_OK;s->composing=0;if(s->cancelling){if(s->hwnd&&!s->detached)goWinTSFEdit((uintptr_t)s->hwnd,2,NULL,0,NULL,0,0,0,0);}else emitDiff(s,s->original,s->originalLength,0);free(s->original);s->original=NULL;return S_OK;}
static ITfContextOwnerCompositionSinkVtbl compositionVtable={CQ,CA,CR,StartComposition,UpdateComposition,EndComposition};
static TextStore*find(HWND hwnd){for(TextStore*s=stores;s;s=s->next)if(s->hwnd==hwnd)return s;return NULL;}
// Ctffunc's conversion compartment is omitted by older MinGW SDK headers.
static const GUID inputModeConversion={0xccf05dd8,0x4a87,0x11d7,{0xa6,0xe2,0x00,0x06,0x5b,0x84,0x43,0x5c}};
long FyneTSFOpenForTest(uintptr_t hwnd){TextStore*s=find((HWND)hwnd);if(!s||!manager)return E_FAIL;ITfCompartmentMgr*compartments=NULL;HRESULT hr=ITfThreadMgr_QueryInterface(manager,&IID_ITfCompartmentMgr,(void**)&compartments);if(FAILED(hr))return hr;ITfCompartment*compartment=NULL;VARIANT value={0};value.vt=VT_I4;value.lVal=1;hr=ITfCompartmentMgr_GetCompartment(compartments,&GUID_COMPARTMENT_KEYBOARD_OPENCLOSE,&compartment);if(SUCCEEDED(hr)){hr=ITfCompartment_SetValue(compartment,client,&value);ITfCompartment_Release(compartment);}if(SUCCEEDED(hr)){value.lVal=9;hr=ITfCompartmentMgr_GetCompartment(compartments,&inputModeConversion,&compartment);if(SUCCEEDED(hr)){hr=ITfCompartment_SetValue(compartment,client,&value);ITfCompartment_Release(compartment);}}ITfCompartmentMgr_Release(compartments);return hr;}
static TextStore*create(HWND hwnd){TextStore*s=calloc(1,sizeof(*s));if(!s)return NULL;s->acp.lpVtbl=&vtable;s->composition.lpVtbl=&compositionVtable;s->refs=1;s->thread=GetCurrentThreadId();s->hwnd=hwnd;s->text=calloc(1,sizeof(WCHAR));if(!s->text){free(s);return NULL;}return s;}
static void releaseManager(void){if(stores)return;if(keys){ITfKeystrokeMgr_Release(keys);keys=NULL;}if(manager){ITfThreadMgr_Deactivate(manager);ITfThreadMgr_Release(manager);manager=NULL;}if(coOwned){CoUninitialize();coOwned=0;}}
int FyneTSFCreate(uintptr_t handle){HWND hwnd=(HWND)handle;if(find(hwnd))return 1;if(!manager){HRESULT hr=CoInitializeEx(NULL,COINIT_APARTMENTTHREADED);if(FAILED(hr))return 0;coOwned=1;hr=CoCreateInstance(&CLSID_TF_ThreadMgr,NULL,CLSCTX_INPROC_SERVER,&IID_ITfThreadMgr,(void**)&manager);if(FAILED(hr)||FAILED(ITfThreadMgr_Activate(manager,&client))){releaseManager();return 0;}ITfThreadMgr_QueryInterface(manager,&IID_ITfKeystrokeMgr,(void**)&keys);}
    TextStore*s=create(hwnd);if(!s){releaseManager();return 0;}HRESULT hr=ITfThreadMgr_CreateDocumentMgr(manager,&s->document);if(SUCCEEDED(hr))hr=ITfDocumentMgr_CreateContext(s->document,client,0,(IUnknown*)&s->acp,&s->context,&s->cookie);if(SUCCEEDED(hr))hr=ITfDocumentMgr_Push(s->document,s->context);if(FAILED(hr)){if(s->context)ITfContext_Release(s->context);if(s->document)ITfDocumentMgr_Release(s->document);release(s);releaseManager();return 0;}ITfDocumentMgr*previous=NULL;ITfThreadMgr_AssociateFocus(manager,hwnd,s->document,&previous);if(previous)ITfDocumentMgr_Release(previous);s->next=stores;stores=s;return 1;}
void FyneTSFUpdate(uintptr_t handle,const unsigned short*text,int length,int cursor,int start,int end,int enabled,int x,int y,int width,int height){TextStore*s=find((HWND)handle);if(!s||s->detached)return;if(length<0||length>MAX_ACP||(!text&&length)||start<0||end<start||end>length||cursor<0||cursor>length)return;int stateChanged=s->enabled!=enabled;s->enabled=enabled;
    RECT rect={x,y,x+width,y+height};int layout=memcmp(&rect,&s->caret,sizeof(rect))!=0;s->caret=rect;if(s->lock||s->composing)return;
    int changed=s->length!=length||(length&&memcmp(s->text,text,length*sizeof(WCHAR))!=0);TS_TEXTCHANGE change={0,s->length,length};if(changed){WCHAR*next=calloc((size_t)length+1,sizeof(WCHAR));if(!next)return;if(length)memcpy(next,text,length*sizeof(WCHAR));free(s->text);s->text=next;s->length=length;}
    int selected=s->start!=start||s->end!=end;s->start=start;s->end=end;s->cursor=cursor;
    if(stateChanged&&s->sink&&(s->mask&TS_AS_STATUS_CHANGE))ITextStoreACPSink_OnStatusChange(s->sink,TS_SD_READONLY);
    if(changed&&s->sink&&(s->mask&TS_AS_TEXT_CHANGE))ITextStoreACPSink_OnTextChange(s->sink,0,&change);
    if(selected&&s->sink&&(s->mask&TS_AS_SEL_CHANGE))ITextStoreACPSink_OnSelectionChange(s->sink);
    if(layout&&s->sink&&(s->mask&TS_AS_LAYOUT_CHANGE))ITextStoreACPSink_OnLayoutChange(s->sink,TS_LC_CHANGE,1);
    if(enabled&&GetFocus()==s->hwnd)ITfThreadMgr_SetFocus(manager,s->document);
}
void FyneTSFReset(uintptr_t handle){TextStore*s=find((HWND)handle);if(!s)return;s->cancelling=1;if(s->composing&&s->context){ITfContextOwnerCompositionServices*service=NULL;if(SUCCEEDED(ITfContext_QueryInterface(s->context,&IID_ITfContextOwnerCompositionServices,(void**)&service))){ITfContextOwnerCompositionServices_TerminateComposition(service,NULL);ITfContextOwnerCompositionServices_Release(service);}}if(s->composing)EndComposition(&s->composition,NULL);s->cancelling=0;s->lastKeyTime=-1;}
void FyneTSFDestroy(uintptr_t handle){TextStore**link=&stores;while(*link&&(*link)->hwnd!=(HWND)handle)link=&(*link)->next;TextStore*s=*link;if(!s)return;FyneTSFReset(handle);s->detached=1;*link=s->next;ITfDocumentMgr*previous=NULL;ITfThreadMgr_AssociateFocus(manager,s->hwnd,NULL,&previous);if(previous)ITfDocumentMgr_Release(previous);ITfDocumentMgr_Pop(s->document,TF_POPF_ALL);ITfContext_Release(s->context);ITfDocumentMgr_Release(s->document);s->context=NULL;s->document=NULL;release(s);releaseManager();}
int FyneTSFKey(uintptr_t handle,unsigned int message,uintptr_t key,intptr_t flags){TextStore*s=find((HWND)handle);if(!s)return 0;if(message==LOCK_MESSAGE){if(s->pending&&!s->lock){DWORD pending=s->pending;s->pending=0;grant(s,pending);if(s->pending)PostMessageW(s->hwnd,LOCK_MESSAGE,0,0);}return 1;}if(!s->enabled||!keys)return 0;BOOL tested=FALSE,eaten=FALSE;
    if(message==WM_KEYDOWN||message==WM_SYSKEYDOWN){s->lastKeyTime=-1;if(SUCCEEDED(ITfKeystrokeMgr_TestKeyDown(keys,key,flags,&tested))&&tested)ITfKeystrokeMgr_KeyDown(keys,key,flags,&eaten);if(eaten){s->lastKeyTime=GetMessageTime();s->lastKeyScan=flags&0x01FF0000;}}
    else if(message==WM_KEYUP||message==WM_SYSKEYUP){if(SUCCEEDED(ITfKeystrokeMgr_TestKeyUp(keys,key,flags,&tested))&&tested)ITfKeystrokeMgr_KeyUp(keys,key,flags,&eaten);}
    else if((message==WM_CHAR||message==WM_SYSCHAR)&&s->lastKeyTime==GetMessageTime()&&(flags&0x01FF0000)==s->lastKeyScan)eaten=TRUE;
    if(getenv("FYNE_LOG_IME")&&(message==WM_KEYDOWN||message==WM_KEYUP))fprintf(stderr,"FYNE-TSF: key=%u tested=%d eaten=%d enabled=%d\n",(unsigned)key,(int)tested,(int)eaten,s->enabled);
    return eaten;
}
typedef struct EditSession { ITfEditSession iface; LONG refs; ITfContext*context; } EditSession;
static ULONG STDMETHODCALLTYPE EA(ITfEditSession*p){return InterlockedIncrement(&((EditSession*)p)->refs);}
static ULONG STDMETHODCALLTYPE ER(ITfEditSession*p){EditSession*s=(EditSession*)p;ULONG n=InterlockedDecrement(&s->refs);if(!n){ITfContext_Release(s->context);free(s);}return n;}
static HRESULT STDMETHODCALLTYPE EQ(ITfEditSession*p,REFIID iid,void**out){if(!out)return E_POINTER;*out=NULL;if(IsEqualIID(iid,&unknownIID)||IsEqualIID(iid,&IID_ITfEditSession)){*out=p;EA(p);return S_OK;}return E_NOINTERFACE;}
static HRESULT STDMETHODCALLTYPE ED(ITfEditSession*p,TfEditCookie cookie){EditSession*s=(EditSession*)p;TF_SELECTION selected;ULONG count=0;HRESULT hr=ITfContext_GetSelection(s->context,cookie,TF_DEFAULT_SELECTION,1,&selected,&count);if(FAILED(hr)||count!=1)return E_FAIL;hr=ITfRange_SetText(selected.range,cookie,0,L"\x65E5\x672C",2);if(SUCCEEDED(hr)){ITfRange_Collapse(selected.range,cookie,TF_ANCHOR_END);selected.style.ase=TF_AE_END;selected.style.fInterimChar=FALSE;hr=ITfContext_SetSelection(s->context,cookie,1,&selected);}ITfRange_Release(selected.range);return hr;}
static ITfEditSessionVtbl editVtable={EQ,EA,ER,ED};
int FyneTSFEditContractTest(uintptr_t hwnd){TextStore*s=find((HWND)hwnd);if(!s||!s->enabled)return 0;EditSession*edit=calloc(1,sizeof(*edit));if(!edit)return 0;edit->iface.lpVtbl=&editVtable;edit->refs=1;edit->context=s->context;ITfContext_AddRef(s->context);HRESULT session=E_FAIL;HRESULT hr=ITfContext_RequestEditSession(s->context,client,&edit->iface,TF_ES_SYNC|TF_ES_READWRITE,&session);ER(&edit->iface);return SUCCEEDED(hr)&&SUCCEEDED(session);}

// Native ABI test: real ITextStoreACP vtable, UTF-16, access locks, bounds and
// atomic replacements. No mock Go editor can conceal a malformed COM layout.
int FyneTSFContractTest(void){TextStore*s=create(NULL);if(!s)return 0;free(s->text);s->text=calloc(6,sizeof(WCHAR));WCHAR original[]={L'a',0xD83D,0xDE00,L'z',0};memcpy(s->text,original,sizeof(original));s->length=4;s->enabled=1;LONG end;int ok=ITextStoreACP_GetEndACP(&s->acp,&end)==TS_E_NOLOCK;s->lock=TS_LF_READ;ok=ok&&SUCCEEDED(ITextStoreACP_GetEndACP(&s->acp,&end))&&end==4;WCHAR result[5]={0};ULONG copied,runs;LONG next;TS_RUNINFO run;ok=ok&&SUCCEEDED(ITextStoreACP_GetText(&s->acp,0,-1,result,5,&copied,&run,1,&runs,&next))&&copied==4&&runs==1&&next==4&&memcmp(result,original,sizeof(original))==0;TS_TEXTCHANGE change;ok=ok&&ITextStoreACP_SetText(&s->acp,0,1,3,L"x",1,&change)==TS_E_NOLOCK;s->lock=TS_LF_READWRITE;ok=ok&&ITextStoreACP_SetText(&s->acp,0,2,3,L"x",1,&change)==E_INVALIDARG;ok=ok&&SUCCEEDED(ITextStoreACP_SetText(&s->acp,0,1,3,L"\x65E5\x672C",2,&change))&&s->length==4&&s->text[1]==0x65E5&&s->text[2]==0x672C&&change.acpOldEnd==3;release(s);return ok;}
