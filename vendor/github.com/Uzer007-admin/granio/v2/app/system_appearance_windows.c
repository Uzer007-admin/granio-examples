//go:build windows && cgo && !ci && !wasm && !test_web_driver && !android && !ios && !mobile && !tinygo && !noos && !tamago

#define COBJMACROS
#define INITGUID
#include <windows.h>
#include <inspectable.h>
#include <windows.ui.viewmanagement.h>
#include <stdlib.h>
#include <wchar.h>
#include <math.h>
#include "system_appearance_windows.h"

typedef __x_ABI_CWindows_CUI_CViewManagement_CIUISettings UI1;
typedef __x_ABI_CWindows_CUI_CViewManagement_CIUISettings2 UI2;
typedef __x_ABI_CWindows_CUI_CViewManagement_CIUISettings3 UI3;
typedef __FITypedEventHandler_2_Windows__CUI__CViewManagement__CUISettings_IInspectable UIHandler;

// MinGW ships IUISettings/2/3. These small ABI declarations for newer optional
// interfaces match Windows SDK 10.0.26100.0 windows.ui.viewmanagement.h. Query-
// Interface, not the OS build number, decides availability. All WinRT booleans
// are one byte; EventRegistrationToken is a signed 64-bit value, even on x86.
typedef struct Effects Effects;
typedef struct {
    IInspectableVtbl base;
    HRESULT (STDMETHODCALLTYPE *get_Enabled)(Effects *, boolean *);
    HRESULT (STDMETHODCALLTYPE *add_Changed)(Effects *, void *, EventRegistrationToken *);
    HRESULT (STDMETHODCALLTYPE *remove_Changed)(Effects *, EventRegistrationToken);
} EffectsVtbl;
struct Effects { const EffectsVtbl *lpVtbl; };
typedef struct Motion Motion;
typedef struct {
    IInspectableVtbl base;
    HRESULT (STDMETHODCALLTYPE *add_Changed)(Motion *, void *, EventRegistrationToken *);
    HRESULT (STDMETHODCALLTYPE *remove_Changed)(Motion *, EventRegistrationToken);
    HRESULT (STDMETHODCALLTYPE *add_MessageChanged)(Motion *, void *, EventRegistrationToken *);
    HRESULT (STDMETHODCALLTYPE *remove_MessageChanged)(Motion *, EventRegistrationToken);
} MotionVtbl;
struct Motion { const MotionVtbl *lpVtbl; };
typedef struct Contrast Contrast;
typedef struct {
    IInspectableVtbl base;
    HRESULT (STDMETHODCALLTYPE *get_Enabled)(Contrast *, boolean *);
    HRESULT (STDMETHODCALLTYPE *get_Scheme)(Contrast *, HSTRING *);
    HRESULT (STDMETHODCALLTYPE *add_Changed)(Contrast *, void *, EventRegistrationToken *);
    HRESULT (STDMETHODCALLTYPE *remove_Changed)(Contrast *, EventRegistrationToken);
} ContrastVtbl;
struct Contrast { const ContrastVtbl *lpVtbl; };

static const IID effectsIID = {0x52bb3002,0x919b,0x4d6b,{0x9b,0x78,0x8d,0xd6,0x6f,0xf4,0xb9,0x3b}};
static const IID motionIID = {0xaef19bd7,0xfe31,0x5a04,{0xad,0xa4,0x46,0x9a,0xae,0xc6,0xdf,0xa9}};
static const IID contrastIID = {0xfe0e8147,0xc4c0,0x4562,{0xb9,0x62,0x13,0x27,0xb5,0x2a,0xd5,0xb9}};
static const IID agileIID = {0x94ea2b94,0xe9cc,0x49e0,{0xc0,0xff,0xee,0x64,0xca,0x8f,0x5b,0x90}};
static const IID contrastHandlerIID = {0xf5917e6f,0x5abf,0x5e65,{0xb5,0xb4,0x1b,0x9c,0x8d,0x94,0xe7,0x88}};
static const IID motionHandlerIID = {0xdeff0f90,0x59e1,0x5859,{0xa2,0x45,0x3d,0x27,0x50,0x81,0xc2,0xad}};

typedef struct {
    volatile LONG refs, active;
    HANDLE event;
} Wake;
typedef struct Delegate Delegate;
typedef struct {
    HRESULT (STDMETHODCALLTYPE *QueryInterface)(Delegate *, REFIID, void **);
    ULONG (STDMETHODCALLTYPE *AddRef)(Delegate *);
    ULONG (STDMETHODCALLTYPE *Release)(Delegate *);
    HRESULT (STDMETHODCALLTYPE *Invoke)(Delegate *, void *, void *);
} DelegateVtbl;
struct Delegate {
    const DelegateVtbl *lpVtbl;
    volatile LONG refs;
    IID iid;
    Wake *wake;
};

static volatile LONG liveProviders, liveDelegates, liveWakes;
static void wake_release(Wake *wake) {
    if (!InterlockedDecrement(&wake->refs)) {
        CloseHandle(wake->event);
        InterlockedDecrement(&liveWakes);
        free(wake);
    }
}
static ULONG STDMETHODCALLTYPE delegate_addref(Delegate *self) {
    return (ULONG)InterlockedIncrement(&self->refs);
}
static ULONG STDMETHODCALLTYPE delegate_release(Delegate *self) {
    LONG refs = InterlockedDecrement(&self->refs);
    if (!refs) {
        wake_release(self->wake);
        InterlockedDecrement(&liveDelegates);
        free(self);
    }
    return (ULONG)refs;
}
static HRESULT STDMETHODCALLTYPE delegate_query(Delegate *self, REFIID iid, void **out) {
    if (!out) return E_POINTER;
    *out = NULL;
    if (!IsEqualIID(iid, &IID_IUnknown) && !IsEqualIID(iid, &agileIID) && !IsEqualIID(iid, &self->iid))
        return E_NOINTERFACE;
    *out = self;
    delegate_addref(self);
    return S_OK;
}
static HRESULT STDMETHODCALLTYPE delegate_invoke(Delegate *self, void *sender, void *args) {
    (void)sender; (void)args;
    // This callback never dereferences Go memory, calls GLFW, or reads settings.
    // Each delegate owns its wake reference, including during late callbacks.
    if (InterlockedCompareExchange(&self->wake->active, 0, 0)) SetEvent(self->wake->event);
    return S_OK;
}
static const DelegateVtbl delegateVtbl = {delegate_query, delegate_addref, delegate_release, delegate_invoke};
static Delegate *delegate_new(Wake *wake, const IID *iid) {
    Delegate *d = (Delegate *)calloc(1, sizeof(*d));
    if (!d) return NULL;
    d->lpVtbl = &delegateVtbl;
    d->refs = 1;
    d->iid = *iid;
    d->wake = wake;
    InterlockedIncrement(&wake->refs);
    InterlockedIncrement(&liveDelegates);
    return d;
}

typedef HRESULT (WINAPI *RoInit)(unsigned);
typedef void (WINAPI *RoUninit)(void);
typedef HRESULT (WINAPI *CreateString)(PCNZWCH, UINT32, HSTRING *);
typedef HRESULT (WINAPI *DeleteString)(HSTRING);
typedef HRESULT (WINAPI *Activate)(HSTRING, IInspectable **);
struct FyneAppearance {
    HMODULE combase;
    RoUninit uninitialize;
    CreateString createString;
    DeleteString deleteString;
    Activate activate;
    int initialized;
    Wake *wake;
    UI1 *ui;
    UI2 *text;
    UI3 *color;
    Effects *effects;
    Motion *motion;
    Contrast *contrast;
    Delegate *handlers[5];
    EventRegistrationToken tokens[5];
    uint32_t subscriptions;
    int32_t event_errors[5];
    HWND window;
    WCHAR window_class[80];
    ATOM window_atom;
};
static LRESULT CALLBACK appearance_window(HWND window, UINT message, WPARAM wparam, LPARAM lparam) {
    FyneAppearance *p = (FyneAppearance *)GetWindowLongPtrW(window, GWLP_USERDATA);
    if (message == WM_NCCREATE) {
        p = (FyneAppearance *)((CREATESTRUCTW *)lparam)->lpCreateParams;
        SetWindowLongPtrW(window, GWLP_USERDATA, (LONG_PTR)p);
    }
    if (p && (message == WM_SETTINGCHANGE || message == WM_SYSCOLORCHANGE || message == WM_THEMECHANGED))
        fyne_appearance_notify(p);
    if (message == WM_NCDESTROY) SetWindowLongPtrW(window, GWLP_USERDATA, 0);
    return DefWindowProcW(window, message, wparam, lparam);
}
static void create_settings_window(FyneAppearance *p) {
    // A message-only HWND does not receive OS broadcasts. This owned top-level
    // window is never shown and has no GL context, taskbar entry or DPI handler.
    swprintf(p->window_class, 80, L"Fyne.SystemAppearance.%p", (void *)p);
    WNDCLASSEXW wc = {0};
    wc.cbSize = sizeof(wc);
    wc.lpfnWndProc = appearance_window;
    wc.hInstance = GetModuleHandleW(NULL);
    wc.lpszClassName = p->window_class;
    p->window_atom = RegisterClassExW(&wc);
    if (p->window_atom)
        p->window = CreateWindowExW(WS_EX_TOOLWINDOW | WS_EX_NOACTIVATE, p->window_class, L"", WS_POPUP, 0, 0, 0, 0, NULL, NULL, wc.hInstance, p);
}
static HRESULT activate_class(FyneAppearance *p, const WCHAR *name, const IID *iid, void **out) {
    HSTRING id = NULL;
    IInspectable *object = NULL;
    HRESULT hr = p->createString(name, (UINT32)wcslen(name), &id);
    if (SUCCEEDED(hr)) {
        hr = p->activate(id, &object);
        p->deleteString(id);
    }
    if (SUCCEEDED(hr)) {
        hr = IInspectable_QueryInterface(object, iid, out);
        IInspectable_Release(object);
    }
    return hr;
}
static void query(void *object, const IID *iid, void **out) {
    if (object) IUnknown_QueryInterface((IUnknown *)object, iid, out);
}
static void release(void *object) {
    if (object) IUnknown_Release((IUnknown *)object);
}

FyneAppearance *fyne_appearance_open(int32_t *error) {
    HRESULT hr = E_OUTOFMEMORY;
    FyneAppearance *p = (FyneAppearance *)calloc(1, sizeof(*p));
    if (!p) { *error = hr; return NULL; }
    InterlockedIncrement(&liveProviders);
    p->combase = LoadLibraryExW(L"combase.dll", NULL, LOAD_LIBRARY_SEARCH_SYSTEM32);
    if (!p->combase) { hr = HRESULT_FROM_WIN32(GetLastError()); goto failed; }
    RoInit initialize = (RoInit)GetProcAddress(p->combase, "RoInitialize");
    p->uninitialize = (RoUninit)GetProcAddress(p->combase, "RoUninitialize");
    p->createString = (CreateString)GetProcAddress(p->combase, "WindowsCreateString");
    p->deleteString = (DeleteString)GetProcAddress(p->combase, "WindowsDeleteString");
    p->activate = (Activate)GetProcAddress(p->combase, "RoActivateInstance");
    if (!initialize || !p->uninitialize || !p->createString || !p->deleteString || !p->activate) {
        hr = E_NOTIMPL; goto failed;
    }
    hr = initialize(1); // RO_INIT_MULTITHREADED; dedicated Go OS thread
    if (FAILED(hr)) goto failed;
    p->initialized = 1;
    p->wake = (Wake *)calloc(1, sizeof(*p->wake));
    if (!p->wake) { hr = E_OUTOFMEMORY; goto failed; }
    p->wake->refs = p->wake->active = 1;
    InterlockedIncrement(&liveWakes);
    p->wake->event = CreateEventW(NULL, FALSE, FALSE, NULL);
    if (!p->wake->event) { hr = HRESULT_FROM_WIN32(GetLastError()); goto failed; }
    HRESULT uihr = activate_class(p, L"Windows.UI.ViewManagement.UISettings", &IID___x_ABI_CWindows_CUI_CViewManagement_CIUISettings, (void **)&p->ui);
    HRESULT achr = activate_class(p, L"Windows.UI.ViewManagement.AccessibilitySettings", &contrastIID, (void **)&p->contrast);
    if (FAILED(uihr) && FAILED(achr)) { hr = uihr; goto failed; }
    query(p->ui, &IID___x_ABI_CWindows_CUI_CViewManagement_CIUISettings2, (void **)&p->text);
    query(p->ui, &IID___x_ABI_CWindows_CUI_CViewManagement_CIUISettings3, (void **)&p->color);
    query(p->ui, &effectsIID, (void **)&p->effects);
    query(p->ui, &motionIID, (void **)&p->motion);
    // Register before the first read, so a setting change during startup is not
    // lost. Subscription failure remains explicit and enables bounded fallback.
    for (unsigned i = 0; i < 5; ++i) {
        p->event_errors[i] = E_NOINTERFACE;
        int present = i == 0 ? p->color != NULL : i == 1 ? p->text != NULL : i == 2 ? p->contrast != NULL : i == 3 ? p->motion != NULL : p->effects != NULL;
        if (!present) continue;
        const IID *iid = i == 2 ? &contrastHandlerIID : i == 3 ? &motionHandlerIID : &IID___FITypedEventHandler_2_Windows__CUI__CViewManagement__CUISettings_IInspectable;
        Delegate *d = delegate_new(p->wake, iid);
        if (!d) { p->event_errors[i] = E_OUTOFMEMORY; continue; }
        p->handlers[i] = d;
        if (i == 0) hr = p->color->lpVtbl->add_ColorValuesChanged(p->color, (UIHandler *)d, &p->tokens[i]);
        if (i == 1) hr = p->text->lpVtbl->add_TextScaleFactorChanged(p->text, (UIHandler *)d, &p->tokens[i]);
        if (i == 2) hr = p->contrast->lpVtbl->add_Changed(p->contrast, d, &p->tokens[i]);
        if (i == 3) hr = p->motion->lpVtbl->add_Changed(p->motion, d, &p->tokens[i]);
        if (i == 4) hr = p->effects->lpVtbl->add_Changed(p->effects, d, &p->tokens[i]);
        p->event_errors[i] = hr;
        if (SUCCEEDED(hr)) p->subscriptions |= 1u << i;
    }
    create_settings_window(p);
    *error = S_OK;
    return p;
failed:
    *error = hr;
    fyne_appearance_close(p);
    return NULL;
}

static uint32_t color_argb(struct __x_ABI_CWindows_CUI_CColor c) {
    return 0xff000000u | ((uint32_t)c.R << 16) | ((uint32_t)c.G << 8) | c.B;
}
void fyne_appearance_read(FyneAppearance *p, FyneAppearanceSnapshot *out) {
    memset(out, 0, sizeof(*out));
    out->winrt_subscriptions = p->subscriptions;
    out->win32_subscriptions = p->window ? FYNE_APPEARANCE_CONTRAST : 0;
    out->subscriptions = out->winrt_subscriptions | out->win32_subscriptions;
    memcpy(out->event_errors, p->event_errors, sizeof(out->event_errors));
    boolean enabled = 0;
    if (p->ui && SUCCEEDED(p->ui->lpVtbl->get_AnimationsEnabled(p->ui, &enabled))) {
        out->available |= FYNE_APPEARANCE_MOTION; out->animations = enabled;
    }
    if (p->contrast && SUCCEEDED(p->contrast->lpVtbl->get_Enabled(p->contrast, &enabled))) {
        out->available |= FYNE_APPEARANCE_CONTRAST; out->high_contrast = enabled;
    } else {
        HIGHCONTRASTW contrast = {0};
        contrast.cbSize = sizeof(contrast);
        if (SystemParametersInfoW(SPI_GETHIGHCONTRAST, sizeof(contrast), &contrast, 0)) {
            out->available |= FYNE_APPEARANCE_CONTRAST;
            out->high_contrast = (contrast.dwFlags & HCF_HIGHCONTRASTON) != 0;
        }
    }
    if (p->effects && SUCCEEDED(p->effects->lpVtbl->get_Enabled(p->effects, &enabled))) {
        out->available |= FYNE_APPEARANCE_EFFECTS; out->transparency = enabled;
    }
    double scale = 0;
    if (p->text && SUCCEEDED(p->text->lpVtbl->get_TextScaleFactor(p->text, &scale)) && isfinite(scale) && scale >= 1 && scale <= 2.25) {
        out->available |= FYNE_APPEARANCE_TEXT; out->text_scale = scale;
    }
    struct __x_ABI_CWindows_CUI_CColor accent, background;
    if (p->color && SUCCEEDED(p->color->lpVtbl->GetColorValue(p->color, UIColorType_Accent, &accent)) &&
        SUCCEEDED(p->color->lpVtbl->GetColorValue(p->color, UIColorType_Background, &background))) {
        out->available |= FYNE_APPEARANCE_COLORS;
        out->accent = color_argb(accent); out->background = color_argb(background);
    }
    const int roles[8] = {COLOR_WINDOW, COLOR_WINDOWTEXT, COLOR_BTNFACE, COLOR_BTNTEXT, COLOR_HIGHLIGHT, COLOR_HIGHLIGHTTEXT, COLOR_GRAYTEXT, COLOR_HOTLIGHT};
    for (unsigned i = 0; i < 8; ++i) {
        COLORREF c = GetSysColor(roles[i]);
        out->colors[i] = 0xff000000u | ((uint32_t)GetRValue(c) << 16) | ((uint32_t)GetGValue(c) << 8) | GetBValue(c);
    }
}
int fyne_appearance_wait(FyneAppearance *p, uint32_t timeout_ms) {
    if (!InterlockedCompareExchange(&p->wake->active, 0, 0)) return 0;
    DWORD result = MsgWaitForMultipleObjectsEx(1, &p->wake->event, timeout_ms, QS_ALLINPUT, MWMO_INPUTAVAILABLE);
    if (result == WAIT_OBJECT_0 + 1) {
        MSG msg;
        while (PeekMessageW(&msg, NULL, 0, 0, PM_REMOVE)) {
            TranslateMessage(&msg);
            DispatchMessageW(&msg);
        }
    }
    if (!InterlockedCompareExchange(&p->wake->active, 0, 0)) return 0;
    return result == WAIT_OBJECT_0 || result == WAIT_OBJECT_0 + 1 || result == WAIT_TIMEOUT ? 1 : -1;
}
void fyne_appearance_notify(FyneAppearance *p) {
    if (InterlockedCompareExchange(&p->wake->active, 0, 0)) SetEvent(p->wake->event);
}
void fyne_appearance_cancel(FyneAppearance *p) {
    InterlockedExchange(&p->wake->active, 0);
    SetEvent(p->wake->event);
}
int fyne_appearance_post_settings_message(FyneAppearance *p) {
    // Used by the integration probe to exercise the actual owned message path.
    return p->window && PostMessageW(p->window, WM_SYSCOLORCHANGE, 0, 0);
}
void fyne_appearance_close(FyneAppearance *p) {
    if (!p) return;
    if (p->wake) {
        InterlockedExchange(&p->wake->active, 0);
        if (p->subscriptions & 1) p->color->lpVtbl->remove_ColorValuesChanged(p->color, p->tokens[0]);
        if (p->subscriptions & 2) p->text->lpVtbl->remove_TextScaleFactorChanged(p->text, p->tokens[1]);
        if (p->subscriptions & 4) p->contrast->lpVtbl->remove_Changed(p->contrast, p->tokens[2]);
        if (p->subscriptions & 8) p->motion->lpVtbl->remove_Changed(p->motion, p->tokens[3]);
        if (p->subscriptions & 16) p->effects->lpVtbl->remove_Changed(p->effects, p->tokens[4]);
        for (unsigned i = 0; i < 5; ++i) if (p->handlers[i]) delegate_release(p->handlers[i]);
    }
    if (p->window) DestroyWindow(p->window);
    if (p->window_atom) UnregisterClassW(p->window_class, GetModuleHandleW(NULL));
    release(p->text); release(p->color); release(p->effects); release(p->motion); release(p->contrast); release(p->ui);
    if (p->wake) wake_release(p->wake);
    if (p->initialized) p->uninitialize();
    if (p->combase) FreeLibrary(p->combase);
    InterlockedDecrement(&liveProviders);
    free(p);
}
void fyne_appearance_resources(long *providers, long *delegates, long *wakes) {
    *providers = InterlockedCompareExchange(&liveProviders, 0, 0);
    *delegates = InterlockedCompareExchange(&liveDelegates, 0, 0);
    *wakes = InterlockedCompareExchange(&liveWakes, 0, 0);
}
