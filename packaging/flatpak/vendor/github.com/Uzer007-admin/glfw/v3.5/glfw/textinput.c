#include "glfw/src/internal.h"
#include "_cgo_export.h"
#include <stdio.h>
int fyneEnableWaylandInput(GLFWwindow* handle);
void fyneDisableWaylandInput(GLFWwindow* handle);
void fyneResetWaylandInput(GLFWwindow* handle);
void fyneWaylandInputCursor(GLFWwindow* handle,int x,int y);

#if defined(_GLFW_X11)
#include <wchar.h>

static int fynePreeditStart(XIC ic, XPointer client, XPointer unused)
{ if(getenv("FYNE_LOG_IME"))fprintf(stderr,"FYNE-XIM: start\n");goFynePreeditStart((GLFWwindow*) client); return 1048576; }
static void fynePreeditDone(XIC ic, XPointer client, XPointer unused)
{ goFynePreeditDone((GLFWwindow*) client); }
static void fynePreeditDraw(XIC ic, XPointer client, XIMPreeditDrawCallbackStruct* draw)
{
    if (!draw) return;
    if(getenv("FYNE_LOG_IME"))fprintf(stderr,"FYNE-XIM: draw %d\n",draw->text?draw->text->length:0);
    if (draw->text && !draw->text->string.multi_byte)
    {
        goFynePreeditDraw((GLFWwindow*) client, draw->chg_first, draw->chg_length,
                         "", 0, draw->caret, 1);
        return;
    }
    int characters = draw->text ? draw->text->length : 0;
    char* utf8 = calloc((size_t) characters * 4 + 1, 1);
    if (!utf8) return;
    int bytes = 0;
    if (draw->text && draw->text->encoding_is_wchar)
    {
        for (int i = 0; i < characters; i++)
            bytes += _glfwEncodeUTF8(utf8 + bytes, draw->text->string.wide_char[i]);
    }
    else if (draw->text)
    {
        mbstate_t conversion = {0};
        const char* source = draw->text->string.multi_byte;
        for (int i = 0; i < characters; i++)
        {
            wchar_t rune;
            size_t width = mbrtowc(&rune, source, MB_CUR_MAX, &conversion);
            if (!width || width == (size_t)-1 || width == (size_t)-2) break;
            bytes += _glfwEncodeUTF8(utf8 + bytes, (uint32_t) rune);
            source += width;
        }
    }
    goFynePreeditDraw((GLFWwindow*) client, draw->chg_first, draw->chg_length,
                     utf8, bytes, draw->caret, 0);
    free(utf8);
}
static void fynePreeditCaret(XIC ic, XPointer client, XIMPreeditCaretCallbackStruct* caret)
{
    if (caret) caret->position = goFynePreeditCaret((GLFWwindow*) client, caret->position, caret->direction);
}
static void fyneInputContextDestroyed(XIC ic, XPointer client, XPointer unused)
{ ((_GLFWwindow*) client)->x11.ic = NULL; goFynePreeditDone((GLFWwindow*) client); }
#endif

int fyneEnableTextInput(GLFWwindow* handle)
{
    int wayland=fyneEnableWaylandInput(handle);
    if (wayland>=0) return wayland;
#if defined(_GLFW_X11)
    if (_glfw.platform.platformID != GLFW_PLATFORM_X11 || !_glfw.x11.im) {if(getenv("FYNE_LOG_IME"))fprintf(stderr,"FYNE-XIM: input method unavailable\n");return 0;}
    _GLFWwindow* window = (_GLFWwindow*) handle;
    XIMStyles* styles = NULL;
    if (XGetIMValues(_glfw.x11.im, XNQueryInputStyle, &styles, NULL) || !styles) return 0;
    const XIMStyle desired = XIMPreeditCallbacks | XIMStatusNothing;
    int supported = 0;
    for (int i = 0; i < styles->count_styles; i++) if (styles->supported_styles[i] == desired) supported = 1;
    XFree(styles);
    if (!supported) {if(getenv("FYNE_LOG_IME"))fprintf(stderr,"FYNE-XIM: callback style unavailable\n");return 0;}
    XIMCallback start = {(XPointer) handle, (XIMProc) fynePreeditStart};
    XIMCallback done = {(XPointer) handle, (XIMProc) fynePreeditDone};
    XIMCallback draw = {(XPointer) handle, (XIMProc) fynePreeditDraw};
    XIMCallback caret = {(XPointer) handle, (XIMProc) fynePreeditCaret};
    XIMCallback destroyed = {(XPointer) handle, (XIMProc) fyneInputContextDestroyed};
    XVaNestedList attributes = XVaCreateNestedList(0, XNPreeditStartCallback, &start,
        XNPreeditDoneCallback, &done, XNPreeditDrawCallback, &draw, XNPreeditCaretCallback, &caret, NULL);
    XIC context = XCreateIC(_glfw.x11.im, XNInputStyle, desired, XNClientWindow, window->x11.handle,
        XNFocusWindow, window->x11.handle, XNDestroyCallback, &destroyed, XNPreeditAttributes, attributes, NULL);
    XFree(attributes);
    if (!context) {if(getenv("FYNE_LOG_IME"))fprintf(stderr,"FYNE-XIM: callback context creation failed\n");return 0;}
    if (window->x11.ic) XDestroyIC(window->x11.ic);
    window->x11.ic = context;
    XWindowAttributes current;
    XGetWindowAttributes(_glfw.x11.display, window->x11.handle, &current);
    unsigned long filter = 0;
    if (!XGetICValues(context, XNFilterEvents, &filter, NULL))
        XSelectInput(_glfw.x11.display, window->x11.handle, current.your_event_mask | filter);
    Window focused; int revert;
    XGetInputFocus(_glfw.x11.display, &focused, &revert);
    if (focused == window->x11.handle) XSetICFocus(context);
    if(getenv("FYNE_LOG_IME"))fprintf(stderr,"FYNE-XIM: callback context enabled\n");
    return 1;
#else
    return 0;
#endif
}

void fyneDisableTextInput(GLFWwindow* handle) { fyneDisableWaylandInput(handle); /* GLFW owns XIC. */ }
void fyneResetTextInput(GLFWwindow* handle)
{
    fyneResetWaylandInput(handle);
#if defined(_GLFW_X11)
    if (_glfw.platform.platformID != GLFW_PLATFORM_X11) return;
    _GLFWwindow* window = (_GLFWwindow*) handle;
    if (window->x11.ic) { char* pending = Xutf8ResetIC(window->x11.ic); if (pending) XFree(pending); }
#endif
}
void fyneTextInputCursor(GLFWwindow* handle, int x, int y)
{
    fyneWaylandInputCursor(handle,x,y);
#if defined(_GLFW_X11)
    if (_glfw.platform.platformID != GLFW_PLATFORM_X11) return;
    _GLFWwindow* window = (_GLFWwindow*) handle;
    if (!window->x11.ic) return;
    XPoint spot = {(short) x, (short) y};
    XVaNestedList attributes = XVaCreateNestedList(0, XNSpotLocation, &spot, NULL);
    XSetICValues(window->x11.ic, XNPreeditAttributes, attributes, NULL);
    XFree(attributes);
#endif
}

int fyneDispatchCommit(_GLFWwindow* window, const char* text, int length)
{ return goFyneCommitText((GLFWwindow*) window, (char*) text, length); }

int fyneTextInputActive(_GLFWwindow* window)
{ return goFyneTextInputActive((GLFWwindow*) window); }
int fyneFilterTextKey(_GLFWwindow*window,unsigned long sym,unsigned int code,unsigned int state,int release){return goFyneTextInputKey((GLFWwindow*)window,(unsigned int)sym,code,state,release);}

void fyneRestoreInputContext(_GLFWwindow* window)
{ if(goFyneTextInputRegistered((GLFWwindow*)window))fyneEnableTextInput((GLFWwindow*)window); }
