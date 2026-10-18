#define WIN32_LEAN_AND_MEAN
#define GLFW_EXPOSE_NATIVE_WIN32
#include <windows.h>
#include <string.h>
#include "visuals_windows.h"
#include "glfw/include/GLFW/glfw3native.h"

typedef HRESULT (WINAPI *SetAttribute)(HWND, DWORD, LPCVOID, DWORD);
typedef HRESULT (WINAPI *GetAttribute)(HWND, DWORD, PVOID, DWORD);
typedef HRESULT (WINAPI *IsCompositionEnabled)(BOOL*);
typedef LONG (WINAPI *GetRealVersion)(OSVERSIONINFOW*);

/* Numeric values are from the public Windows SDK enum. Older MinGW headers
 * do not declare REDIRECTIONBITMAP_ALPHA (39). No undocumented DWM API is used. */
static const DWORD attributes[4] = {38, 20, 33, 39};

static HMODULE loadDwm(void) {
    WCHAR path[MAX_PATH + 16];
    UINT length = GetSystemDirectoryW(path, MAX_PATH);
    if (!length || length >= MAX_PATH) return NULL;
    memcpy(path + length, L"\\dwmapi.dll", sizeof(L"\\dwmapi.dll"));
    return LoadLibraryW(path); /* absolute system-directory path */
}

GLFWgoVisualState glfwQueryVisualsBridge(GLFWwindow* window) {
    GLFWgoVisualState state = {0};
    HWND handle = glfwGetWin32Window(window);
    if (!handle || !IsWindow(handle)) return state;
    state.valid = 1;
    state.ownerThread = GetWindowThreadProcessId(handle, NULL) == GetCurrentThreadId();
    if (!state.ownerThread) return state;
    OSVERSIONINFOW version = {0};
    version.dwOSVersionInfoSize = sizeof(version);
    GetRealVersion realVersion = (GetRealVersion) GetProcAddress(GetModuleHandleW(L"ntdll.dll"), "RtlGetVersion");
    if (realVersion && realVersion(&version) == 0) state.build = version.dwBuildNumber;
    state.remote = GetSystemMetrics(SM_REMOTESESSION) != 0;
    HIGHCONTRASTW contrast = {0};
    contrast.cbSize = sizeof(contrast);
    state.contrastKnown = SystemParametersInfoW(SPI_GETHIGHCONTRAST, sizeof(contrast), &contrast, 0) != 0;
    state.contrast = state.contrastKnown && (contrast.dwFlags & HCF_HIGHCONTRASTON) != 0;
    state.transparent = glfwGetWindowAttrib(window, GLFW_TRANSPARENT_FRAMEBUFFER);
    HDC dc = GetDC(handle);
    if (dc) {
        PIXELFORMATDESCRIPTOR format = {0};
        state.pixelFormat = GetPixelFormat(dc);
        if (state.pixelFormat && DescribePixelFormat(dc, state.pixelFormat, sizeof(format), &format)) state.alphaBits = format.cAlphaBits;
        ReleaseDC(handle, dc);
    }
    HMODULE dwm = loadDwm();
    if (!dwm) return state;
    SetAttribute set = (SetAttribute) GetProcAddress(dwm, "DwmSetWindowAttribute");
    GetAttribute get = (GetAttribute) GetProcAddress(dwm, "DwmGetWindowAttribute");
    IsCompositionEnabled composition = (IsCompositionEnabled) GetProcAddress(dwm, "DwmIsCompositionEnabled");
    BOOL enabled = FALSE;
    if (composition && composition(&enabled) == S_OK) state.composition = enabled != FALSE;
    if (set) {
        if (state.build >= 22000) state.supported |= 2 | 4;
        if (state.build >= 22621) state.supported |= 1;
        if (state.build >= 26100) state.supported |= 8;
    }
    if (state.build >= 22621 && get) state.backdropKnown = get(handle, attributes[0], &state.backdrop, sizeof(state.backdrop)) == S_OK;
    FreeLibrary(dwm);
    return state;
}

GLFWgoVisualResult glfwApplyVisualsBridge(GLFWwindow* window, uint32_t mask, int backdrop, int dark, int corners, int alpha) {
    GLFWgoVisualResult result = {{0}, 0, 0};
    HWND handle = glfwGetWin32Window(window);
    if (!handle || !IsWindow(handle)) return result;
    result.valid = 1;
    result.ownerThread = GetWindowThreadProcessId(handle, NULL) == GetCurrentThreadId();
    if (!result.ownerThread) return result;
    HMODULE dwm = loadDwm();
    SetAttribute set = dwm ? (SetAttribute) GetProcAddress(dwm, "DwmSetWindowAttribute") : NULL;
    int values[4] = {backdrop, dark, corners, alpha}; /* BOOL and enum: 32 bits */
    for (int i = 0; i < 4; ++i) {
        if (!(mask & (1u << i))) continue;
        result.hr[i] = (uint32_t) (set ? set(handle, attributes[i], &values[i], sizeof(values[i])) : E_NOTIMPL);
    }
    if (dwm) FreeLibrary(dwm);
    return result;
}
