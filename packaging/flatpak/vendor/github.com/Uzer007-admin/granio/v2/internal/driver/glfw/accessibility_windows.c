//go:build accessibility && windows
#define CINTERFACE
#define COBJMACROS
#include <windows.h>
#include <commctrl.h>
#include <ole2.h>
#include <wchar.h>
#include <math.h>
#include <uiautomationcore.h>
#include <uiautomationcoreapi.h>
#include <uiautomationclient.h>
#include "accessibility_windows.h"
#include "_cgo_export.h"

// Native clients can retain a provider after its control or window disappears.
// The window owns one reference; removed providers are detached and remain valid
// COM objects returning ELEMENTNOTAVAILABLE until their final external Release.
typedef struct FyneElement FyneElement;
typedef struct FyneWindow FyneWindow;
typedef struct FyneRange FyneRange;
struct FyneElement {
    IRawElementProviderSimple simple;
    IRawElementProviderFragment fragment;
    IRawElementProviderFragmentRoot root;
    IInvokeProvider invoke;
    IValueProvider valueProvider;
    IToggleProvider toggle;
    ISelectionItemProvider selection;
    IExpandCollapseProvider expand;
    ITextProvider textProvider;
    IRangeValueProvider rangeValue;
    ISelectionProvider selectionContainer;
    IGridProvider grid;
    IGridItemProvider gridItem;
    IVirtualizedItemProvider virtualItem;
    IScrollItemProvider scrollItem;
    IItemContainerProvider itemContainer;
    LONG refs;
    HWND hwnd;
    uint64_t id, parentID;
    unsigned generation;
    int detached, isRoot, role, states, patterns;
    WCHAR *name, *value;
    double x,y,width,height;
    FyneElement *parent;
    FyneElement **children;
    int count, capacity;
    int selectionStart, selectionEnd, rows, columns, row, column, multiple;
    double number, minimum, maximum, step;
    FyneRange *ranges;
};
struct FyneWindow {
    HWND hwnd;
    FyneElement **elements;
    int count, capacity;
    unsigned generation;
    int structureChanged;
    FyneElement *root;
    FyneWindow *next;
};
static SRWLOCK providerLock = SRWLOCK_INIT;
static FyneWindow *windows;
static INIT_ONCE initOnce = INIT_ONCE_STATIC_INIT;
static HMODULE uia;
static LRESULT (WINAPI *uiaReturn)(HWND, WPARAM, LPARAM, IRawElementProviderSimple*);
static HRESULT (WINAPI *uiaHost)(HWND, IRawElementProviderSimple**);
static HRESULT (WINAPI *uiaEvent)(IRawElementProviderSimple*, EVENTID);
static HRESULT (WINAPI *uiaProperty)(IRawElementProviderSimple*, PROPERTYID, VARIANT, VARIANT);
static BOOL (WINAPI *uiaListening)(void);
static HRESULT (WINAPI *uiaStructure)(IRawElementProviderSimple*, enum StructureChangeType, int*, int);
static HRESULT (WINAPI *uiaDisconnect)(IRawElementProviderSimple*);
#define ELEMENT(p, field) ((FyneElement*)((char*)(p)-offsetof(FyneElement,field)))
#define UNAVAILABLE ((HRESULT)UIA_E_ELEMENTNOTAVAILABLE)
static const IID simpleIID = {0xd6dd68d1,0x86fd,0x4332,{0x86,0x66,0x9a,0xbe,0xde,0xa2,0xd2,0x4c}};
static const IID unknownIID = {0,0,0,{0xc0,0,0,0,0,0,0,0x46}};
static const IID fragmentIID = {0xf7063da8,0x8359,0x439c,{0x92,0x97,0xbb,0xc5,0x29,0x9a,0x7d,0x87}};
static const IID rootIID = {0x620ce2a5,0xab8f,0x40a9,{0x86,0xcb,0xde,0x3c,0x75,0x59,0x9b,0x58}};
static const IID invokeIID = {0x54fcb24b,0xe18e,0x47a2,{0xb4,0xd3,0xec,0xcb,0xe7,0x75,0x99,0xa2}};
static const IID valueIID = {0xc7935180,0x6fb3,0x4201,{0xb1,0x74,0x7d,0xf7,0x3a,0xdb,0xf6,0x4a}};
static const IID toggleIID = {0x56d00bd0,0xc4f4,0x433c,{0xa8,0x36,0x1a,0x52,0xa5,0x7e,0x08,0x92}};
static const IID selectionIID = {0x2acad808,0xb2d4,0x452d,{0xa4,0x07,0x91,0xff,0x1a,0xd1,0x67,0xb2}};
static const IID expandIID = {0xd847d3a5,0xcab0,0x4a98,{0x8c,0x32,0xec,0xb4,0x5c,0x59,0xad,0x24}};
static const IID textIID={0x3589c92c,0x63f3,0x4367,{0x99,0xbb,0xad,0xa6,0x53,0xb7,0x7c,0xf2}};
static const IID textRangeIID={0x5347ad7b,0xc355,0x46f8,{0xaf,0xf5,0x90,0x90,0x33,0x58,0x2f,0x63}};
static const IID rangeValueIID={0x36dc7aef,0x33e6,0x4691,{0xaf,0xe1,0x2b,0xe7,0x27,0x4b,0x3d,0x33}};
static const IID selectionContainerIID={0xfb8b03af,0x3bdf,0x48d4,{0xbd,0x36,0x1a,0x65,0x79,0x3b,0xe1,0x68}};
static const IID gridIID={0xb17d6187,0x0907,0x464b,{0xa1,0x68,0x0e,0xf1,0x7a,0x15,0x72,0xb1}};
static const IID gridItemIID={0xd02541f1,0xfb81,0x4d64,{0xae,0x32,0xf5,0x20,0xf8,0xa6,0xdb,0xd1}};
static const IID virtualItemIID={0xcb98b665,0x2d35,0x4fac,{0xad,0x35,0xf3,0xc6,0x0d,0x0c,0x0b,0x8b}};
static const IID scrollItemIID={0x2360c714,0x4bf1,0x4b26,{0xba,0x65,0x9b,0x21,0x31,0x61,0x27,0xeb}};
static const IID itemContainerIID={0xe747770b,0x39ce,0x4382,{0xab,0x30,0xd8,0xfb,0x3f,0x33,0x6f,0x24}};
static ULONG retain(FyneElement *e) { return InterlockedIncrement(&e->refs); }
static ULONG release(FyneElement *e) {
    ULONG count=InterlockedDecrement(&e->refs);
    if (!count) { free(e->name); free(e->value); free(e->children); free(e); }
    return count;
}
static HRESULT query(FyneElement *e, REFIID iid, void **out) {
    if (!out) return E_POINTER;
    *out=NULL;
    AcquireSRWLockShared(&providerLock);
    if (IsEqualIID(iid,&unknownIID) || IsEqualIID(iid,&simpleIID)) *out=&e->simple;
    else if (IsEqualIID(iid,&fragmentIID)) *out=&e->fragment;
    else if (IsEqualIID(iid,&rootIID) && e->isRoot) *out=&e->root;
    else if (IsEqualIID(iid,&invokeIID) && (e->patterns&1)) *out=&e->invoke;
    else if (IsEqualIID(iid,&valueIID) && (e->patterns&2)) *out=&e->valueProvider;
    else if (IsEqualIID(iid,&toggleIID) && (e->patterns&4)) *out=&e->toggle;
    else if (IsEqualIID(iid,&selectionIID) && (e->patterns&8)) *out=&e->selection;
    else if (IsEqualIID(iid,&expandIID) && (e->patterns&16)) *out=&e->expand;
    else if (IsEqualIID(iid,&textIID) && (e->patterns&64)) *out=&e->textProvider;
    else if (IsEqualIID(iid,&rangeValueIID) && (e->patterns&128)) *out=&e->rangeValue;
    else if (IsEqualIID(iid,&selectionContainerIID) && (e->patterns&256)) *out=&e->selectionContainer;
    else if (IsEqualIID(iid,&gridIID) && (e->patterns&512)) *out=&e->grid;
    else if (IsEqualIID(iid,&gridItemIID) && (e->patterns&1024)) *out=&e->gridItem;
    else if (IsEqualIID(iid,&virtualItemIID) && (e->patterns&4096)) *out=&e->virtualItem;
    else if (IsEqualIID(iid,&scrollItemIID) && (e->patterns&4096)) *out=&e->scrollItem;
    else if (IsEqualIID(iid,&itemContainerIID) && (e->patterns&2048)) *out=&e->itemContainer;
    if (*out) retain(e);
    ReleaseSRWLockShared(&providerLock);
    return *out ? S_OK:E_NOINTERFACE;
}
#define COM_BASE(prefix, type, field) \
 static HRESULT STDMETHODCALLTYPE prefix##_QI(type *p, REFIID iid, void **out) {return query(ELEMENT(p,field),iid,out);} \
 static ULONG STDMETHODCALLTYPE prefix##_AddRef(type *p) {return retain(ELEMENT(p,field));} \
 static ULONG STDMETHODCALLTYPE prefix##_Release(type *p) {return release(ELEMENT(p,field));}
COM_BASE(S,IRawElementProviderSimple,simple)
COM_BASE(F,IRawElementProviderFragment,fragment)
COM_BASE(R,IRawElementProviderFragmentRoot,root)
COM_BASE(I,IInvokeProvider,invoke)
COM_BASE(V,IValueProvider,valueProvider)
COM_BASE(T,IToggleProvider,toggle)
COM_BASE(L,ISelectionItemProvider,selection)
COM_BASE(E,IExpandCollapseProvider,expand)

static HRESULT STDMETHODCALLTYPE S_Options(IRawElementProviderSimple *p,enum ProviderOptions *out) {
    if (!out) return E_POINTER;
    *out=ProviderOptions_ServerSideProvider|ProviderOptions_UseComThreading;return S_OK;
}
static HRESULT STDMETHODCALLTYPE S_Pattern(IRawElementProviderSimple *p,PATTERNID id,IUnknown **out) {
    if (!out) return E_POINTER; *out=NULL;
    FyneElement *e=ELEMENT(p,simple);
    AcquireSRWLockShared(&providerLock);
    HRESULT result=e->detached?UNAVAILABLE:S_OK;
    if (!e->detached) {
        if (id==UIA_InvokePatternId && (e->patterns&1)) *out=(IUnknown*)&e->invoke;
        if (id==UIA_ValuePatternId && (e->patterns&2)) *out=(IUnknown*)&e->valueProvider;
        if (id==UIA_TogglePatternId && (e->patterns&4)) *out=(IUnknown*)&e->toggle;
        if (id==UIA_SelectionItemPatternId && (e->patterns&8)) *out=(IUnknown*)&e->selection;
        if (id==UIA_ExpandCollapsePatternId && (e->patterns&16)) *out=(IUnknown*)&e->expand;
        if (id==UIA_TextPatternId && (e->patterns&64)) *out=(IUnknown*)&e->textProvider;
        if (id==UIA_RangeValuePatternId && (e->patterns&128)) *out=(IUnknown*)&e->rangeValue;
        if (id==UIA_SelectionPatternId && (e->patterns&256)) *out=(IUnknown*)&e->selectionContainer;
        if (id==UIA_GridPatternId && (e->patterns&512)) *out=(IUnknown*)&e->grid;
        if (id==UIA_GridItemPatternId && (e->patterns&1024)) *out=(IUnknown*)&e->gridItem;
        if (id==UIA_VirtualizedItemPatternId && (e->patterns&4096)) *out=(IUnknown*)&e->virtualItem;
        if (id==UIA_ScrollItemPatternId && (e->patterns&4096)) *out=(IUnknown*)&e->scrollItem;
        if (id==UIA_ItemContainerPatternId && (e->patterns&2048)) *out=(IUnknown*)&e->itemContainer;
        if (*out) retain(e);
    }
    ReleaseSRWLockShared(&providerLock);return result;
}
static HRESULT STDMETHODCALLTYPE S_Property(IRawElementProviderSimple *p,PROPERTYID id,VARIANT *out) {
    if (!out) return E_POINTER;VariantInit(out);FyneElement *e=ELEMENT(p,simple);
    AcquireSRWLockShared(&providerLock);
    if (e->detached) {ReleaseSRWLockShared(&providerLock);return UNAVAILABLE;}
    switch(id) {
    case UIA_ControlTypePropertyId:out->vt=VT_I4;out->lVal=e->role;break;
    case UIA_NamePropertyId:out->vt=VT_BSTR;out->bstrVal=SysAllocString(e->name);break;
    case UIA_AutomationIdPropertyId:{WCHAR text[48];swprintf(text,48,L"fyne_%llu",(unsigned long long)e->id);out->vt=VT_BSTR;out->bstrVal=SysAllocString(text);break;}
    case UIA_HasKeyboardFocusPropertyId:out->vt=VT_BOOL;out->boolVal=(e->states&2)?VARIANT_TRUE:VARIANT_FALSE;break;
    case UIA_IsKeyboardFocusablePropertyId:out->vt=VT_BOOL;out->boolVal=(e->patterns&32)?VARIANT_TRUE:VARIANT_FALSE;break;
    case UIA_IsEnabledPropertyId:out->vt=VT_BOOL;out->boolVal=(e->states&1)?VARIANT_FALSE:VARIANT_TRUE;break;
    case UIA_IsPasswordPropertyId:out->vt=VT_BOOL;out->boolVal=(e->states&32)?VARIANT_TRUE:VARIANT_FALSE;break;
    case UIA_IsOffscreenPropertyId:out->vt=VT_BOOL;out->boolVal=(e->states&128)?VARIANT_TRUE:VARIANT_FALSE;break;
    case UIA_IsControlElementPropertyId:case UIA_IsContentElementPropertyId:out->vt=VT_BOOL;out->boolVal=VARIANT_TRUE;break;
    case UIA_FrameworkIdPropertyId:out->vt=VT_BSTR;out->bstrVal=SysAllocString(L"Fyne");break;
    }
    ReleaseSRWLockShared(&providerLock);return S_OK;
}
static HRESULT STDMETHODCALLTYPE S_Host(IRawElementProviderSimple *p,IRawElementProviderSimple **out) {
    if (!out) return E_POINTER;*out=NULL;FyneElement *e=ELEMENT(p,simple);
    AcquireSRWLockShared(&providerLock);HWND hwnd=e->hwnd;int root=e->isRoot,dead=e->detached;ReleaseSRWLockShared(&providerLock);
    if (dead) return UNAVAILABLE;
    return root&&uiaHost?uiaHost(hwnd,out):S_OK;
}
static HRESULT STDMETHODCALLTYPE F_Navigate(IRawElementProviderFragment *p,enum NavigateDirection direction,IRawElementProviderFragment **out) {
    if (!out) return E_POINTER;*out=NULL;FyneElement *e=ELEMENT(p,fragment),*target=NULL;
    AcquireSRWLockShared(&providerLock);
    if (e->detached) {ReleaseSRWLockShared(&providerLock);return UNAVAILABLE;}
    if((e->patterns&2048)&&((e->patterns&4096)||direction==NavigateDirection_FirstChild||direction==NavigateDirection_LastChild)){
        HWND hwnd=e->hwnd;uint64_t id=e->id;ReleaseSRWLockShared(&providerLock);
        uint64_t next=goWinAccessibilityVirtualNavigate((uintptr_t)hwnd,id,direction,0);
        AcquireSRWLockShared(&providerLock);
        for(FyneWindow*w=windows;w;w=w->next)if(w->hwnd==hwnd)for(int i=0;i<w->count;i++)if(w->elements[i]->id==next){*out=&w->elements[i]->fragment;retain(w->elements[i]);break;}
        ReleaseSRWLockShared(&providerLock);return S_OK;
    }
    if (direction==NavigateDirection_Parent) target=e->parent;
    else if(direction==NavigateDirection_FirstChild&&e->count)target=e->children[0];
    else if(direction==NavigateDirection_LastChild&&e->count)target=e->children[e->count-1];
    else if(e->parent) for(int i=0;i<e->parent->count;i++) if(e->parent->children[i]==e) {
        if(direction==NavigateDirection_NextSibling&&i+1<e->parent->count)target=e->parent->children[i+1];
        if(direction==NavigateDirection_PreviousSibling&&i>0)target=e->parent->children[i-1];break;
    }
    if(target){*out=&target->fragment;retain(target);}
    ReleaseSRWLockShared(&providerLock);return S_OK;
}
static HRESULT STDMETHODCALLTYPE F_Runtime(IRawElementProviderFragment *p,SAFEARRAY **out) {
    if(!out)return E_POINTER; FyneElement *e=ELEMENT(p,fragment);
    SAFEARRAYBOUND bound={3,0};*out=SafeArrayCreate(VT_I4,1,&bound);if(!*out)return E_OUTOFMEMORY;
    int values[]={UiaAppendRuntimeId,(int)(e->id>>32),(int)e->id};for(LONG i=0;i<3;i++)SafeArrayPutElement(*out,&i,&values[i]);return S_OK;
}
static HRESULT STDMETHODCALLTYPE F_Rect(IRawElementProviderFragment *p,struct UiaRect *out) {
    if(!out)return E_POINTER;FyneElement *e=ELEMENT(p,fragment);AcquireSRWLockShared(&providerLock);
    if(e->detached){ReleaseSRWLockShared(&providerLock);return UNAVAILABLE;}
    POINT origin={0,0};ClientToScreen(e->hwnd,&origin);
    out->left=origin.x+e->x;out->top=origin.y+e->y;out->width=e->width;out->height=e->height;
    ReleaseSRWLockShared(&providerLock);return S_OK;
}
static HRESULT STDMETHODCALLTYPE F_Embedded(IRawElementProviderFragment *p,SAFEARRAY **out){if(!out)return E_POINTER;*out=NULL;return S_OK;}
static HRESULT action(FyneElement *e,int operation,LPCWSTR value) {
    AcquireSRWLockShared(&providerLock);int dead=e->detached,disabled=e->states&1;HWND hwnd=e->hwnd;uint64_t id=e->id;ReleaseSRWLockShared(&providerLock);
    if(dead)return UNAVAILABLE;if(disabled)return UIA_E_ELEMENTNOTENABLED;
    goWinAccessibilityAction((uintptr_t)hwnd,id,operation,(unsigned short*)value);return S_OK;
}
static HRESULT STDMETHODCALLTYPE F_Focus(IRawElementProviderFragment *p){return action(ELEMENT(p,fragment),32,NULL);}
static HRESULT STDMETHODCALLTYPE F_Root(IRawElementProviderFragment *p,IRawElementProviderFragmentRoot **out) {
    if(!out)return E_POINTER;*out=NULL;FyneElement *e=ELEMENT(p,fragment);AcquireSRWLockShared(&providerLock);
    if(e->detached){ReleaseSRWLockShared(&providerLock);return UNAVAILABLE;}
    while(e->parent)e=e->parent;if(e->isRoot){*out=&e->root;retain(e);}ReleaseSRWLockShared(&providerLock);return S_OK;
}
static HRESULT STDMETHODCALLTYPE R_Point(IRawElementProviderFragmentRoot *p,double x,double y,IRawElementProviderFragment **out) {
    if(!out)return E_POINTER;*out=NULL;FyneElement *root=ELEMENT(p,root),*found=root;AcquireSRWLockShared(&providerLock);
    if(root->detached){ReleaseSRWLockShared(&providerLock);return UNAVAILABLE;}
    POINT pt={(LONG)x,(LONG)y};ScreenToClient(root->hwnd,&pt);
    for(;;){FyneElement *child=NULL;for(int i=0;i<found->count;i++){FyneElement *c=found->children[i];if(pt.x>=c->x&&pt.x<c->x+c->width&&pt.y>=c->y&&pt.y<c->y+c->height)child=c;}if(!child)break;found=child;}
    *out=&found->fragment;retain(found);ReleaseSRWLockShared(&providerLock);return S_OK;
}
static HRESULT STDMETHODCALLTYPE R_Focus(IRawElementProviderFragmentRoot *p,IRawElementProviderFragment **out) {
    if(!out)return E_POINTER;*out=NULL;FyneElement *root=ELEMENT(p,root);AcquireSRWLockShared(&providerLock);
    if(root->detached){ReleaseSRWLockShared(&providerLock);return UNAVAILABLE;}
    for(FyneWindow *w=windows;w;w=w->next)if(w->root==root)for(int i=0;i<w->count;i++)if(w->elements[i]->states&2){*out=&w->elements[i]->fragment;retain(w->elements[i]);break;}
    ReleaseSRWLockShared(&providerLock);return S_OK;
}
static HRESULT STDMETHODCALLTYPE I_Invoke(IInvokeProvider *p){return action(ELEMENT(p,invoke),1,NULL);}
static HRESULT STDMETHODCALLTYPE V_Set(IValueProvider *p,LPCWSTR value){return action(ELEMENT(p,valueProvider),2,value);}
static HRESULT STDMETHODCALLTYPE V_Get(IValueProvider *p,BSTR *out){if(!out)return E_POINTER;*out=NULL;FyneElement *e=ELEMENT(p,valueProvider);AcquireSRWLockShared(&providerLock);HRESULT result=e->detached?UNAVAILABLE:(e->states&32)?E_ACCESSDENIED:S_OK;if(result==S_OK)*out=SysAllocString(e->value);ReleaseSRWLockShared(&providerLock);return result;}
static HRESULT STDMETHODCALLTYPE V_ReadOnly(IValueProvider *p,BOOL *out){if(!out)return E_POINTER;FyneElement *e=ELEMENT(p,valueProvider);AcquireSRWLockShared(&providerLock);*out=(e->states&1)!=0;ReleaseSRWLockShared(&providerLock);return S_OK;}
static HRESULT STDMETHODCALLTYPE T_Toggle(IToggleProvider *p){return action(ELEMENT(p,toggle),4,NULL);}
static HRESULT STDMETHODCALLTYPE T_State(IToggleProvider *p,enum ToggleState *out){if(!out)return E_POINTER;FyneElement *e=ELEMENT(p,toggle);AcquireSRWLockShared(&providerLock);*out=(e->states&4)?ToggleState_On:ToggleState_Off;int dead=e->detached;ReleaseSRWLockShared(&providerLock);return dead?UNAVAILABLE:S_OK;}
static HRESULT STDMETHODCALLTYPE L_Select(ISelectionItemProvider *p){return action(ELEMENT(p,selection),8,NULL);}
static HRESULT STDMETHODCALLTYPE L_Add(ISelectionItemProvider *p){return action(ELEMENT(p,selection),9,NULL);}
static HRESULT STDMETHODCALLTYPE L_Remove(ISelectionItemProvider *p){return action(ELEMENT(p,selection),10,NULL);}
static HRESULT STDMETHODCALLTYPE L_Selected(ISelectionItemProvider *p,BOOL *out){if(!out)return E_POINTER;FyneElement *e=ELEMENT(p,selection);AcquireSRWLockShared(&providerLock);*out=(e->states&8)!=0;int dead=e->detached;ReleaseSRWLockShared(&providerLock);return dead?UNAVAILABLE:S_OK;}
static HRESULT STDMETHODCALLTYPE L_Container(ISelectionItemProvider *p,IRawElementProviderSimple **out){if(!out)return E_POINTER;*out=NULL;FyneElement *e=ELEMENT(p,selection);AcquireSRWLockShared(&providerLock);for(FyneElement*parent=e->parent;parent;parent=parent->parent)if(parent->patterns&256){*out=&parent->simple;retain(parent);break;}int dead=e->detached;ReleaseSRWLockShared(&providerLock);return dead?UNAVAILABLE:S_OK;}
static HRESULT expandTo(FyneElement *e,int expanded){AcquireSRWLockShared(&providerLock);int current=(e->states&16)!=0;ReleaseSRWLockShared(&providerLock);return current==expanded?S_OK:action(e,16,NULL);}
static HRESULT STDMETHODCALLTYPE E_Expand(IExpandCollapseProvider *p){return expandTo(ELEMENT(p,expand),1);}
static HRESULT STDMETHODCALLTYPE E_Collapse(IExpandCollapseProvider *p){return expandTo(ELEMENT(p,expand),0);}
static HRESULT STDMETHODCALLTYPE E_State(IExpandCollapseProvider *p,enum ExpandCollapseState *out){if(!out)return E_POINTER;FyneElement *e=ELEMENT(p,expand);AcquireSRWLockShared(&providerLock);*out=(e->states&16)?ExpandCollapseState_Expanded:ExpandCollapseState_Collapsed;int dead=e->detached;ReleaseSRWLockShared(&providerLock);return dead?UNAVAILABLE:S_OK;}

static IRawElementProviderSimpleVtbl simpleVtbl={S_QI,S_AddRef,S_Release,S_Options,S_Pattern,S_Property,S_Host};
static IRawElementProviderFragmentVtbl fragmentVtbl={F_QI,F_AddRef,F_Release,F_Navigate,F_Runtime,F_Rect,F_Embedded,F_Focus,F_Root};
static IRawElementProviderFragmentRootVtbl rootVtbl={R_QI,R_AddRef,R_Release,R_Point,R_Focus};
static IInvokeProviderVtbl invokeVtbl={I_QI,I_AddRef,I_Release,I_Invoke};
static IValueProviderVtbl valueVtbl={V_QI,V_AddRef,V_Release,V_Set,V_Get,V_ReadOnly};
static IToggleProviderVtbl toggleVtbl={T_QI,T_AddRef,T_Release,T_Toggle,T_State};
static ISelectionItemProviderVtbl selectionVtbl={L_QI,L_AddRef,L_Release,L_Select,L_Add,L_Remove,L_Selected,L_Container};
static IExpandCollapseProviderVtbl expandVtbl={E_QI,E_AddRef,E_Release,E_Expand,E_Collapse,E_State};
#include "accessibility_windows_patterns.h"
static BOOL CALLBACK initialise(PINIT_ONCE once,PVOID arg,PVOID *context) {
    uia=LoadLibraryW(L"uiautomationcore.dll");if(!uia)return TRUE;
    uiaReturn=(void*)GetProcAddress(uia,"UiaReturnRawElementProvider");uiaHost=(void*)GetProcAddress(uia,"UiaHostProviderFromHwnd");
    uiaEvent=(void*)GetProcAddress(uia,"UiaRaiseAutomationEvent");uiaStructure=(void*)GetProcAddress(uia,"UiaRaiseStructureChangedEvent");uiaDisconnect=(void*)GetProcAddress(uia,"UiaDisconnectProvider");
    uiaProperty=(void*)GetProcAddress(uia,"UiaRaiseAutomationPropertyChangedEvent");uiaListening=(void*)GetProcAddress(uia,"UiaClientsAreListening");return TRUE;
}
static LRESULT CALLBACK windowProc(HWND hwnd,UINT message,WPARAM wp,LPARAM lp,UINT_PTR id,DWORD_PTR reference) {
    FyneWindow *w=(FyneWindow*)reference;
    if(message==WM_GETOBJECT&&lp==UiaRootObjectId&&uiaReturn){AcquireSRWLockShared(&providerLock);FyneElement *root=w->root;if(root)retain(root);ReleaseSRWLockShared(&providerLock);if(root){LRESULT result=uiaReturn(hwnd,wp,lp,&root->simple);release(root);return result;}}
    return DefSubclassProc(hwnd,message,wp,lp);
}
static WCHAR* wide(const char *utf8) {int n=MultiByteToWideChar(CP_UTF8,0,utf8,-1,NULL,0);if(n<1)n=1;WCHAR *out=calloc(n,sizeof(WCHAR));if(out)MultiByteToWideChar(CP_UTF8,0,utf8,-1,out,n);return out;}
static FyneWindow* findWindow(HWND hwnd){for(FyneWindow *w=windows;w;w=w->next)if(w->hwnd==hwnd)return w;return NULL;}
void WinAccessibilityBegin(uintptr_t handle) {
    HWND hwnd=(HWND)handle;InitOnceExecuteOnce(&initOnce,initialise,NULL,NULL);
    AcquireSRWLockExclusive(&providerLock);
    FyneWindow *w=findWindow(hwnd);
    if(!w){w=calloc(1,sizeof(*w));if(!w){ReleaseSRWLockExclusive(&providerLock);return;}w->hwnd=hwnd;w->next=windows;windows=w;SetWindowSubclass(hwnd,windowProc,0xF1A11,(DWORD_PTR)w);}
    w->generation++;
    w->structureChanged=0;
    // The UI thread holds the exclusive lock until End, publishing one coherent
    // snapshot to concurrent COM readers. Add performs no COM calls.
}
void WinAccessibilityAdd(uintptr_t handle,uint64_t id,uint64_t parent,const char *name,const char *value,int role,int states,int patterns,double x,double y,double width,double height) {
    FyneWindow *w=findWindow((HWND)handle);if(!w)return;FyneElement *e=NULL;
    for(int i=0;i<w->count;i++)if(w->elements[i]->id==id){e=w->elements[i];break;}
    if(!e){
        if(w->count==w->capacity){int capacity=w->capacity?w->capacity*2:16;FyneElement **items=realloc(w->elements,capacity*sizeof(*items));if(!items)return;w->elements=items;w->capacity=capacity;}
        e=calloc(1,sizeof(*e));if(!e)return;e->refs=1;e->id=id;e->hwnd=w->hwnd;
        e->simple.lpVtbl=&simpleVtbl;e->fragment.lpVtbl=&fragmentVtbl;e->root.lpVtbl=&rootVtbl;e->invoke.lpVtbl=&invokeVtbl;e->valueProvider.lpVtbl=&valueVtbl;e->toggle.lpVtbl=&toggleVtbl;e->selection.lpVtbl=&selectionVtbl;e->expand.lpVtbl=&expandVtbl;
        e->textProvider.lpVtbl=&textVtbl;e->rangeValue.lpVtbl=&rangeValueVtbl;e->selectionContainer.lpVtbl=&selectionContainerVtbl;e->grid.lpVtbl=&gridVtbl;e->gridItem.lpVtbl=&gridItemVtbl;
        e->virtualItem.lpVtbl=&virtualItemVtbl;e->scrollItem.lpVtbl=&scrollItemVtbl;e->itemContainer.lpVtbl=&itemContainerVtbl;
        w->elements[w->count++]=e;
        w->structureChanged=1;
    }
    WCHAR *newName=wide(name),*newValue=wide(value);adjustRanges(e,e->value,newValue);free(e->name);free(e->value);e->name=newName;e->value=newValue;
    e->generation=w->generation;e->parentID=parent;e->isRoot=!parent;e->role=role;e->states=states;e->patterns=patterns;e->x=x;e->y=y;e->width=width;e->height=height;e->parent=NULL;e->count=0;
    if(!parent)w->root=e;
}
void WinAccessibilityKeep(uintptr_t handle,uint64_t id,uint64_t parent){FyneWindow*w=findWindow((HWND)handle);if(!w)return;for(int i=0;i<w->count;i++){FyneElement*e=w->elements[i];if(e->id==id){e->generation=w->generation;e->parentID=parent;e->parent=NULL;e->count=0;break;}}}
void WinAccessibilityDetails(uintptr_t handle,uint64_t id,int start,int end,int rows,int columns,int row,int column,int multiple,double value,double minimum,double maximum,double step) {
    FyneWindow*w=findWindow((HWND)handle);if(!w)return;for(int i=0;i<w->count;i++){FyneElement*e=w->elements[i];if(e->id!=id)continue;e->selectionStart=start;e->selectionEnd=end;e->rows=rows;e->columns=columns;e->row=row;e->column=column;e->multiple=multiple;e->number=value;e->minimum=minimum;e->maximum=maximum;e->step=step;break;}
}
void WinAccessibilityEnd(uintptr_t handle) {
    FyneWindow *w=findWindow((HWND)handle);if(!w)return;
    int changed=w->structureChanged,write=0;
    for(int i=0;i<w->count;i++){FyneElement *e=w->elements[i];if(e->generation!=w->generation){e->detached=1;e->parent=NULL;e->count=0;release(e);changed=1;}else w->elements[write++]=e;}
    w->count=write;
    for(int i=0;i<w->count;i++){FyneElement *e=w->elements[i];if(!e->parentID)continue;for(int n=0;n<w->count;n++){FyneElement *p=w->elements[n];if(p->id!=e->parentID)continue;e->parent=p;if(p->count==p->capacity){int capacity=p->capacity?p->capacity*2:8;FyneElement **items=realloc(p->children,capacity*sizeof(*items));if(!items)break;p->children=items;p->capacity=capacity;}p->children[p->count++]=e;break;}}
    FyneElement *root=w->root;if(root)retain(root);ReleaseSRWLockExclusive(&providerLock);
    if(root){if(changed&&uiaStructure){int id[3]={UiaAppendRuntimeId,(int)(root->id>>32),(int)root->id};uiaStructure(&root->simple,StructureChangeType_ChildrenInvalidated,id,3);}release(root);}
}
// Never call UI Automation while holding providerLock: an event client can
// synchronously read this provider. Retain the element across that callback.
static FyneElement* eventElement(uintptr_t handle,uint64_t id){
    FyneElement*e=NULL;AcquireSRWLockShared(&providerLock);FyneWindow*w=findWindow((HWND)handle);
    if(w)for(int i=0;i<w->count;i++)if(w->elements[i]->id==id){e=w->elements[i];retain(e);break;}
    ReleaseSRWLockShared(&providerLock);return e;
}
int WinAccessibilityListening(void){return uiaListening&&uiaListening();}
void WinAccessibilityEvent(uintptr_t handle,uint64_t id,int event){
    if(!uiaEvent)return;FyneElement*e=eventElement(handle,id);if(!e)return;
    uiaEvent(&e->simple,event);release(e);
}
static void propertyEvent(uintptr_t handle,uint64_t id,int property,VARIANT before,VARIANT after){
    if(!uiaProperty)return;FyneElement*e=eventElement(handle,id);if(!e)return;
    uiaProperty(&e->simple,property,before,after);release(e);
}
void WinAccessibilityBoolProperty(uintptr_t handle,uint64_t id,int property,int before,int after,int boolean){
    VARIANT a,b;VariantInit(&a);VariantInit(&b);a.vt=b.vt=boolean?VT_BOOL:VT_I4;
    if(boolean){a.boolVal=before?VARIANT_TRUE:VARIANT_FALSE;b.boolVal=after?VARIANT_TRUE:VARIANT_FALSE;}
    else {a.lVal=before;b.lVal=after;}propertyEvent(handle,id,property,a,b);
}
void WinAccessibilityNumberProperty(uintptr_t handle,uint64_t id,int property,double before,double after){
    VARIANT a,b;VariantInit(&a);VariantInit(&b);a.vt=b.vt=VT_R8;a.dblVal=before;b.dblVal=after;propertyEvent(handle,id,property,a,b);
}
void WinAccessibilityTextProperty(uintptr_t handle,uint64_t id,int property,const char* before,const char* after){
    WCHAR*x=wide(before),*y=wide(after);VARIANT a,b;VariantInit(&a);VariantInit(&b);
    a.vt=b.vt=VT_BSTR;a.bstrVal=x?SysAllocString(x):NULL;b.bstrVal=y?SysAllocString(y):NULL;
    free(x);free(y);if(a.bstrVal&&b.bstrVal)propertyEvent(handle,id,property,a,b);VariantClear(&a);VariantClear(&b);
}
int WinAccessibilityScrollTextRange(uintptr_t handle,uint64_t id,int start,int end,int top){
    FyneRange*r=NULL;AcquireSRWLockExclusive(&providerLock);FyneWindow*w=findWindow((HWND)handle);
    if(w)for(int i=0;i<w->count;i++)if(w->elements[i]->id==id){r=newRange(w->elements[i],start,end);break;}
    ReleaseSRWLockExclusive(&providerLock);if(!r)return 0;
    HRESULT hr=ITextRangeProvider_ScrollIntoView(&r->iface,top);TR_R(&r->iface);return SUCCEEDED(hr);
}
int WinAccessibilitySelectionContractTest(uintptr_t handle,uint64_t id,uint64_t childID){
    FyneElement*e=eventElement(handle,id);if(!e)return 0;SAFEARRAY*selected=NULL;
    HRESULT hr=ISelectionProvider_GetSelection(&e->selectionContainer,&selected);int ok=SUCCEEDED(hr)&&selected;
    if(selected){LONG low=0,high=-1;SafeArrayGetLBound(selected,1,&low);SafeArrayGetUBound(selected,1,&high);ok=ok&&high-low==0;
        IUnknown*value=NULL;if(high>=low&&SUCCEEDED(SafeArrayGetElement(selected,&low,&value))&&value){
            IRawElementProviderSimple*simple=NULL;hr=IUnknown_QueryInterface(value,&simpleIID,(void**)&simple);
            if(SUCCEEDED(hr)&&simple){FyneElement*child=ELEMENT(simple,simple);BOOL state=FALSE;IRawElementProviderSimple*parent=NULL;
                ok=ok&&child->id==childID&&SUCCEEDED(ISelectionItemProvider_get_IsSelected(&child->selection,&state))&&state;
                ok=ok&&SUCCEEDED(ISelectionItemProvider_get_SelectionContainer(&child->selection,&parent))&&parent==&e->simple;
                if(parent)IRawElementProviderSimple_Release(parent);IRawElementProviderSimple_Release(simple);
            }else ok=0;IUnknown_Release(value);
        }else ok=0;SafeArrayDestroy(selected);
    }release(e);return ok;
}
int WinAccessibilityVirtualContractTest(uintptr_t handle,uint64_t id){
    FyneElement*e=NULL;AcquireSRWLockShared(&providerLock);FyneWindow*w=findWindow((HWND)handle);if(w)for(int i=0;i<w->count;i++)if(w->elements[i]->id==id){e=w->elements[i];retain(e);break;}ReleaseSRWLockShared(&providerLock);if(!e)return 0;
    IUnknown*pattern=NULL;int ok=SUCCEEDED(IRawElementProviderSimple_GetPatternProvider(&e->simple,UIA_ItemContainerPatternId,&pattern))&&pattern;if(pattern)IUnknown_Release(pattern);
    // A virtual collection still has ordinary siblings in its parent layout.
    // Its logical row navigation must not swallow the neighbouring controls.
    IRawElementProviderFragment*neighbour=NULL;
    ok=ok&&SUCCEEDED(IRawElementProviderFragment_Navigate(&e->fragment,NavigateDirection_NextSibling,&neighbour))&&neighbour;
    if(neighbour){FyneElement*item=ELEMENT(neighbour,fragment);ok=ok&&wcscmp(item->name,L"After collection")==0;IRawElementProviderFragment_Release(neighbour);}
    neighbour=NULL;ok=ok&&SUCCEEDED(IRawElementProviderFragment_Navigate(&e->fragment,NavigateDirection_PreviousSibling,&neighbour))&&neighbour;
    if(neighbour){FyneElement*item=ELEMENT(neighbour,fragment);ok=ok&&wcscmp(item->name,L"Before collection")==0;IRawElementProviderFragment_Release(neighbour);}
    IRawElementProviderFragment*last=NULL;ok=ok&&SUCCEEDED(IRawElementProviderFragment_Navigate(&e->fragment,NavigateDirection_LastChild,&last))&&last;
    if(last){FyneElement*child=ELEMENT(last,fragment);VARIANT value;VariantInit(&value);ok=ok&&SUCCEEDED(IRawElementProviderSimple_GetPropertyValue(&child->simple,UIA_IsOffscreenPropertyId,&value))&&value.vt==VT_BOOL&&value.boolVal==VARIANT_TRUE;VariantClear(&value);
        IUnknown*virtual=NULL;ok=ok&&SUCCEEDED(IRawElementProviderSimple_GetPatternProvider(&child->simple,UIA_VirtualizedItemPatternId,&virtual))&&virtual;if(virtual)IUnknown_Release(virtual);
        IRawElementProviderFragment*parent=NULL;ok=ok&&SUCCEEDED(IRawElementProviderFragment_Navigate(last,NavigateDirection_Parent,&parent))&&parent==&e->fragment;if(parent)IRawElementProviderFragment_Release(parent);IRawElementProviderFragment_Release(last);
    }
    SAFEARRAY*selected=NULL;ok=ok&&SUCCEEDED(ISelectionProvider_GetSelection(&e->selectionContainer,&selected))&&selected;if(selected){LONG low=0,high=-1;SafeArrayGetLBound(selected,1,&low);SafeArrayGetUBound(selected,1,&high);ok=ok&&high-low+1==1;IUnknown*item=NULL;if(high>=low)SafeArrayGetElement(selected,&low,&item);if(item){FyneElement*child=ELEMENT((IRawElementProviderSimple*)item,simple);ok=ok&&child->name&&wcsstr(child->name,L"90000")&&(child->states&128);IUnknown_Release(item);}else ok=0;SafeArrayDestroy(selected);}release(e);return ok;
}
static int eventContractMask;
static HRESULT WINAPI eventContractProperty(IRawElementProviderSimple*p,PROPERTYID property,VARIANT before,VARIANT after){
    VARIANT name;HRESULT hr=S_Property(p,UIA_NamePropertyId,&name);if(FAILED(hr))return hr;VariantClear(&name);
    if(property==UIA_IsEnabledPropertyId&&before.vt==VT_BOOL&&before.boolVal==VARIANT_TRUE&&after.vt==VT_BOOL&&after.boolVal==VARIANT_FALSE)eventContractMask|=1;
    if(property==UIA_RangeValueValuePropertyId&&before.vt==VT_R8&&before.dblVal==2&&after.vt==VT_R8&&after.dblVal==7)eventContractMask|=2;
    if(property==UIA_NamePropertyId&&before.vt==VT_BSTR&&SysStringLen(before.bstrVal)==1&&after.vt==VT_BSTR&&SysStringLen(after.bstrVal)==2&&after.bstrVal[0]==0x65E5)eventContractMask|=4;
    return S_OK;
}
static HRESULT WINAPI eventContractEvent(IRawElementProviderSimple*p,EVENTID event){if(event==UIA_AutomationFocusChangedEventId)eventContractMask|=8;return S_OK;}
int WinAccessibilityEventsContractTest(void){
    HWND hwnd=CreateWindowW(L"STATIC",L"events ABI",0,0,0,100,100,NULL,NULL,NULL,NULL);if(!hwnd)return 0;
    WinAccessibilityBegin((uintptr_t)hwnd);WinAccessibilityAdd((uintptr_t)hwnd,30000001,0,"root","",50032,0,0,0,0,100,100);WinAccessibilityEnd((uintptr_t)hwnd);
    HRESULT(WINAPI*savedProperty)(IRawElementProviderSimple*,PROPERTYID,VARIANT,VARIANT)=uiaProperty;
    HRESULT(WINAPI*savedEvent)(IRawElementProviderSimple*,EVENTID)=uiaEvent;
    eventContractMask=0;uiaProperty=eventContractProperty;uiaEvent=eventContractEvent;
    WinAccessibilityBoolProperty((uintptr_t)hwnd,30000001,UIA_IsEnabledPropertyId,1,0,1);
    WinAccessibilityNumberProperty((uintptr_t)hwnd,30000001,UIA_RangeValueValuePropertyId,2,7);
    WinAccessibilityTextProperty((uintptr_t)hwnd,30000001,UIA_NamePropertyId,"a","\xe6\x97\xa5\xe6\x9c\xac");
    WinAccessibilityEvent((uintptr_t)hwnd,30000001,UIA_AutomationFocusChangedEventId);
    uiaProperty=savedProperty;uiaEvent=savedEvent;int ok=eventContractMask==15;
    WinAccessibilityCleanup((uintptr_t)hwnd);DestroyWindow(hwnd);return ok;
}
void WinAccessibilityCleanup(uintptr_t handle) {
    HWND hwnd=(HWND)handle;AcquireSRWLockExclusive(&providerLock);FyneWindow **link=&windows;while(*link&&(*link)->hwnd!=hwnd)link=&(*link)->next;FyneWindow *w=*link;if(!w){ReleaseSRWLockExclusive(&providerLock);return;}
    *link=w->next;RemoveWindowSubclass(hwnd,windowProc,0xF1A11);
    for(int i=0;i<w->count;i++){FyneElement *e=w->elements[i];e->detached=1;e->parent=NULL;e->count=0;}
    ReleaseSRWLockExclusive(&providerLock);
    for(int i=0;i<w->count;i++){if(uiaDisconnect)uiaDisconnect(&w->elements[i]->simple);release(w->elements[i]);}free(w->elements);free(w);
}
int WinAccessibilityTextContractTest(void){HWND hwnd=CreateWindowW(L"STATIC",L"text ABI",0,0,0,320,200,NULL,NULL,NULL,NULL);if(!hwnd)return 0;
    WinAccessibilityBegin((uintptr_t)hwnd);WinAccessibilityAdd((uintptr_t)hwnd,20000001,0,"root","",50032,0,0,0,0,320,200);WinAccessibilityAdd((uintptr_t)hwnd,20000002,20000001,"text","a\xcc\x81\xf0\x9f\x98\x80z",50004,0,2|64,0,0,200,40);WinAccessibilityDetails((uintptr_t)hwnd,20000002,0,0,0,0,0,0,0,0,0,0,0);WinAccessibilityEnd((uintptr_t)hwnd);
    AcquireSRWLockShared(&providerLock);FyneElement*e=findWindow(hwnd)->elements[1];retain(e);ReleaseSRWLockShared(&providerLock);IUnknown*pattern=NULL;int ok=SUCCEEDED(S_Pattern(&e->simple,UIA_TextPatternId,&pattern))&&pattern;ITextRangeProvider*range=NULL;BSTR text=NULL;
    if(pattern){ok=ok&&SUCCEEDED(ITextProvider_get_DocumentRange((ITextProvider*)pattern,&range));IUnknown_Release(pattern);}
    if(range){ok=ok&&SUCCEEDED(ITextRangeProvider_GetText(range,-1,&text))&&SysStringLen(text)==5;SysFreeString(text);int moved=0;ok=ok&&SUCCEEDED(ITextRangeProvider_MoveEndpointByUnit(range,TextPatternRangeEndpoint_Start,TextUnit_Character,1,&moved))&&moved==1;ok=ok&&SUCCEEDED(ITextRangeProvider_GetText(range,-1,&text))&&SysStringLen(text)==3&&text[0]==0xD83D;SysFreeString(text);text=NULL;ok=ok&&SUCCEEDED(ITextRangeProvider_GetText(range,1,&text))&&SysStringLen(text)==0;SysFreeString(text);}
    WinAccessibilityBegin((uintptr_t)hwnd);WinAccessibilityAdd((uintptr_t)hwnd,20000001,0,"root","",50032,0,0,0,0,320,200);WinAccessibilityEnd((uintptr_t)hwnd);if(range){text=NULL;ok=ok&&ITextRangeProvider_GetText(range,-1,&text)==UNAVAILABLE;ITextRangeProvider_Release(range);}release(e);WinAccessibilityCleanup((uintptr_t)hwnd);DestroyWindow(hwnd);return ok;}

// Exercises external COM ownership across snapshots and independent windows.
int WinAccessibilityContractTest(void) {
    HWND a=CreateWindowW(L"STATIC",L"a",0,0,0,100,100,NULL,NULL,NULL,NULL),b=CreateWindowW(L"STATIC",L"b",0,0,0,100,100,NULL,NULL,NULL,NULL);
    if(!a||!b)return 0;
    WinAccessibilityBegin((uintptr_t)a);WinAccessibilityAdd((uintptr_t)a,10000001,0,"a","",50032,0,0,0,0,100,100);WinAccessibilityAdd((uintptr_t)a,10000002,10000001,"input","hello",50004,0,2|32,0,0,50,20);WinAccessibilityEnd((uintptr_t)a);
    WinAccessibilityBegin((uintptr_t)b);WinAccessibilityAdd((uintptr_t)b,10000003,0,"b","",50032,0,0,0,0,100,100);WinAccessibilityEnd((uintptr_t)b);
    AcquireSRWLockShared(&providerLock);FyneElement *e=findWindow(a)->elements[1];retain(e);ReleaseSRWLockShared(&providerLock);
    IUnknown *pattern=NULL;int ok=SUCCEEDED(S_Pattern(&e->simple,UIA_ValuePatternId,&pattern))&&pattern;
    BSTR value=NULL;if(pattern){ok=ok&&SUCCEEDED(IValueProvider_get_Value((IValueProvider*)pattern,&value))&&wcscmp(value,L"hello")==0;SysFreeString(value);IUnknown_Release(pattern);}
    WinAccessibilityBegin((uintptr_t)a);WinAccessibilityAdd((uintptr_t)a,10000001,0,"a","",50032,0,0,0,0,100,100);WinAccessibilityEnd((uintptr_t)a);
    VARIANT property;ok=ok&&S_Property(&e->simple,UIA_NamePropertyId,&property)==UNAVAILABLE;release(e);
    WinAccessibilityCleanup((uintptr_t)a);AcquireSRWLockShared(&providerLock);FyneElement *root=findWindow(b)->root;retain(root);ReleaseSRWLockShared(&providerLock);
    ok=ok&&SUCCEEDED(S_Property(&root->simple,UIA_NamePropertyId,&property));VariantClear(&property);WinAccessibilityCleanup((uintptr_t)b);ok=ok&&S_Property(&root->simple,UIA_NamePropertyId,&property)==UNAVAILABLE;release(root);
    DestroyWindow(a);DestroyWindow(b);return ok;
}
