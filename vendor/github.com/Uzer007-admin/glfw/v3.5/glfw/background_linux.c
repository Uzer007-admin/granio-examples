/* Local fork: optional compositor backgrounds. Uses GLFW's connection and
 * proxies, with no event pumps, second connection or per-frame readback. */
#include "glfw/src/internal.h"
#include "background_linux.h"
#include <limits.h>

#if defined(_GLFW_WAYLAND)
#include "glfw/include/wayland-client-protocol.h"
#include "glfw/include/ext-background-effect-v1-client-protocol.h"
#endif

GLFWgoBackgroundState glfwQueryBackgroundBridge(GLFWwindow* handle)
{
    GLFWgoBackgroundState s = {0};
    if (!_glfw.initialized) return s;
    s.valid = 1;
    s.ownerThread = _glfwIsMainThread();
    if (!s.ownerThread) return s;
    _GLFWwindow* window = (_GLFWwindow*) handle;
#if defined(_GLFW_X11)
    if (_glfw.platform.platformID == GLFW_PLATFORM_X11)
    {
        s.platform = 1;
        s.alphaSurface = window->x11.transparent;
        s.compositor = XGetSelectionOwner(_glfw.x11.display, _glfw.x11.NET_WM_CM_Sx);
        s.composition = s.compositor != None;
        s.transparent = s.alphaSurface && s.composition;
        Atom actualType;
        int actualFormat;
        unsigned long count, after;
        unsigned char* data = NULL;
        int result = XGetWindowProperty(_glfw.x11.display, _glfw.x11.root,
                                       _glfw.x11.backgroundBlurAtom, 0, 0,
                                       False, AnyPropertyType, &actualType,
                                       &actualFormat, &count, &after, &data);
        s.blur = s.composition && result == Success && actualType != None;
        if (data) XFree(data);
        s.requested = window->x11.backgroundBlur;
        return s;
    }
#endif
#if defined(_GLFW_WAYLAND)
    if (_glfw.platform.platformID == GLFW_PLATFORM_WAYLAND)
    {
        s.platform = 2;
        s.alphaSurface = window->wl.transparent;
        if (s.alphaSurface && window->context.client != GLFW_NO_API)
        {
            EGLint alpha = 0;
            /* Wayland's NATIVE_CONTEXT_API also creates an EGL context. */
            s.alphaSurface = window->context.egl.config && _glfw.egl.GetConfigAttrib &&
                _glfw.egl.GetConfigAttrib(_glfw.egl.display, window->context.egl.config,
                                         EGL_ALPHA_SIZE, &alpha) && alpha > 0;
        }
        s.composition = 1;
        s.transparent = s.alphaSurface;
        s.blur = _glfw.wl.backgroundManager &&
                 (_glfw.wl.backgroundCapabilities & EXT_BACKGROUND_EFFECT_MANAGER_V1_CAPABILITY_BLUR);
        s.requested = window->wl.backgroundBlur;
        s.revision = _glfw.wl.backgroundRevision;
        s.compositor = _glfw.wl.backgroundName;
        return s;
    }
#endif
    s.valid = 0;
    return s;
}

int glfwSetBackgroundBlurBridge(GLFWwindow* handle, int enabled)
{
    GLFWgoBackgroundState s = glfwQueryBackgroundBridge(handle);
    if (!s.valid) return 1;
    if (!s.ownerThread) return 2;
    if (enabled && !s.alphaSurface) return 3;
    if (enabled && !s.composition) return 4;
    if (enabled && !s.blur) return 5;
    _GLFWwindow* window = (_GLFWwindow*) handle;
#if defined(_GLFW_X11)
    if (s.platform == 1)
    {
        if (enabled)
        {
            /* KWin: an empty CARDINAL/32 region requests the entire window.
             * No geometry query or property update is required on resize. */
            XChangeProperty(_glfw.x11.display, window->x11.handle,
                            _glfw.x11.backgroundBlurAtom, XA_CARDINAL, 32,
                            PropModeReplace, NULL, 0);
        }
        else
            XDeleteProperty(_glfw.x11.display, window->x11.handle, _glfw.x11.backgroundBlurAtom);
        window->x11.backgroundBlur = enabled;
        XFlush(_glfw.x11.display);
        return 0;
    }
#endif
#if defined(_GLFW_WAYLAND)
    if (s.platform == 2)
    {
        if (enabled)
        {
            if (!window->wl.backgroundEffect)
                window->wl.backgroundEffect = ext_background_effect_manager_v1_get_background_effect(
                    _glfw.wl.backgroundManager, window->wl.surface);
            if (!window->wl.backgroundEffect) return 6;
            struct wl_region* region = wl_compositor_create_region(_glfw.wl.compositor);
            if (!region) return 6;
            /* Surface-local coordinates, clipped by the compositor. Covers
             * subsequent resize/fractional-scale changes without new requests. */
            wl_region_add(region, 0, 0, INT_MAX, INT_MAX);
            ext_background_effect_surface_v1_set_blur_region(window->wl.backgroundEffect, region);
            wl_region_destroy(region);
        }
        else if (window->wl.backgroundEffect)
            ext_background_effect_surface_v1_set_blur_region(window->wl.backgroundEffect, NULL);
        window->wl.backgroundBlur = enabled;
        return 0;
    }
#endif
    return 1;
}
