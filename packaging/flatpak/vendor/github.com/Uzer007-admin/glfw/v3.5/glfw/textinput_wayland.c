#include "glfw/src/internal.h"
#include "_cgo_export.h"
#include <string.h>

#if defined(_GLFW_WAYLAND)
// Generated from the MIT-licensed Wayland text-input-v3 protocol in
// glfw/deps/wayland/text-input-unstable-v3.xml using wayland-scanner.
#include "text-input-v3-client.h"
#include "text-input-v3-code.h"

typedef struct FyneWaylandInput {
    _GLFWwindow* window;
    struct zwp_text_input_v3* input;
    int entered, wanted, enabled, sensitive;
    int x, y, cursor, anchor, begin, end;
    uint32_t sequence, before, after;
    char *surrounding, *preedit, *commit;
    struct FyneWaylandInput* next;
} FyneWaylandInput;
static struct wl_registry* inputRegistry;
static struct zwp_text_input_manager_v3* inputManager;
static FyneWaylandInput* inputs;
static FyneWaylandInput* findInput(GLFWwindow* handle) {
    for (FyneWaylandInput* p=inputs;p;p=p->next) if (p->window==(_GLFWwindow*)handle) return p;
    return NULL;
}
static void publish(FyneWaylandInput* p) {
    if (!p->entered) return;
    int enabled=p->wanted;
    if (enabled && !p->enabled) zwp_text_input_v3_enable(p->input);
    else if (!enabled && p->enabled) zwp_text_input_v3_disable(p->input);
    p->enabled=enabled;
    if (enabled) {
        zwp_text_input_v3_set_surrounding_text(p->input,p->surrounding?p->surrounding:"",p->cursor,p->anchor);
        zwp_text_input_v3_set_text_change_cause(p->input,ZWP_TEXT_INPUT_V3_CHANGE_CAUSE_OTHER);
        zwp_text_input_v3_set_content_type(p->input,p->sensitive?ZWP_TEXT_INPUT_V3_CONTENT_HINT_HIDDEN_TEXT|ZWP_TEXT_INPUT_V3_CONTENT_HINT_SENSITIVE_DATA:ZWP_TEXT_INPUT_V3_CONTENT_HINT_NONE,p->sensitive?ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_PASSWORD:ZWP_TEXT_INPUT_V3_CONTENT_PURPOSE_NORMAL);
        zwp_text_input_v3_set_cursor_rectangle(p->input,p->x,p->y,1,1);
    }
    zwp_text_input_v3_commit(p->input);p->sequence++;
}
static void entered(void* data,struct zwp_text_input_v3* input,struct wl_surface* surface) {
    FyneWaylandInput* p=data;p->entered=surface==p->window->wl.surface;if(p->entered)publish(p);
}
static void left(void* data,struct zwp_text_input_v3* input,struct wl_surface* surface) {
    FyneWaylandInput* p=data;p->entered=p->enabled=0;goFynePreeditDone((GLFWwindow*)p->window);
}
static void preedit(void* data,struct zwp_text_input_v3* input,const char* text,int32_t begin,int32_t end) {
    FyneWaylandInput* p=data;free(p->preedit);p->preedit=strdup(text?text:"");p->begin=begin;p->end=end;
}
static void committed(void* data,struct zwp_text_input_v3* input,const char* text) {FyneWaylandInput* p=data;free(p->commit);p->commit=text?strdup(text):NULL;}
static void deleted(void* data,struct zwp_text_input_v3* input,uint32_t before,uint32_t after) {FyneWaylandInput* p=data;p->before=before;p->after=after;}
static void done(void* data,struct zwp_text_input_v3* input,uint32_t serial) {
    FyneWaylandInput* p=data;
    if (serial==p->sequence && p->entered && p->enabled)
        goFyneWaylandBatch((GLFWwindow*)p->window,p->preedit,p->begin,p->end,p->commit,p->before,p->after);
    free(p->preedit);free(p->commit);p->preedit=p->commit=NULL;p->before=p->after=0;
}
static const struct zwp_text_input_v3_listener inputListener={entered,left,preedit,committed,deleted,done};
static void global(void* data,struct wl_registry* registry,uint32_t name,const char* interface,uint32_t version) {
    if (!strcmp(interface,"zwp_text_input_manager_v3")) inputManager=wl_registry_bind(registry,name,&zwp_text_input_manager_v3_interface,1);
}
static void globalRemoved(void* data,struct wl_registry* registry,uint32_t name) {}
static const struct wl_registry_listener registryListener={global,globalRemoved};
#endif

// -1 selects the X11 path; zero retains GLFW's compositor fallback.
int fyneEnableWaylandInput(GLFWwindow* handle) {
#if defined(_GLFW_WAYLAND)
    if (_glfw.platform.platformID!=GLFW_PLATFORM_WAYLAND) return -1;
    if (!inputRegistry) {
        inputRegistry=wl_display_get_registry(_glfw.wl.display);
        wl_registry_add_listener(inputRegistry,&registryListener,NULL);
        wl_display_roundtrip(_glfw.wl.display);
    }
    if (!inputManager || !_glfw.wl.seat) return 0;
    FyneWaylandInput* p=calloc(1,sizeof(*p));if(!p)return 0;
    p->window=(_GLFWwindow*)handle;p->input=zwp_text_input_manager_v3_get_text_input(inputManager,_glfw.wl.seat);
    zwp_text_input_v3_add_listener(p->input,&inputListener,p);p->next=inputs;inputs=p;return 1;
#else
    return -1;
#endif
}
void fyneDisableWaylandInput(GLFWwindow* handle) {
#if defined(_GLFW_WAYLAND)
    FyneWaylandInput** link=&inputs;while(*link&&(*link)->window!=(_GLFWwindow*)handle)link=&(*link)->next;
    FyneWaylandInput* p=*link;if(p){*link=p->next;zwp_text_input_v3_destroy(p->input);free(p->surrounding);free(p->preedit);free(p->commit);free(p);}
    if(!inputs){if(inputManager)zwp_text_input_manager_v3_destroy(inputManager);if(inputRegistry)wl_registry_destroy(inputRegistry);inputManager=NULL;inputRegistry=NULL;}
#endif
}
void fyneResetWaylandInput(GLFWwindow* handle) {
#if defined(_GLFW_WAYLAND)
    FyneWaylandInput* p=findInput(handle);if(!p)return;
    if(p->enabled){zwp_text_input_v3_disable(p->input);zwp_text_input_v3_commit(p->input);p->sequence++;p->enabled=0;publish(p);}
#endif
}
void fyneWaylandInputCursor(GLFWwindow* handle,int x,int y) {
#if defined(_GLFW_WAYLAND)
    FyneWaylandInput* p=findInput(handle);if(p&&(p->x!=x||p->y!=y)){p->x=x;p->y=y;publish(p);}
#endif
}
void fyneTextInputSurrounding(GLFWwindow* handle,const char* text,int cursor,int anchor,int sensitive,int enabled) {
#if defined(_GLFW_WAYLAND)
    FyneWaylandInput* p=findInput(handle);if(!p)return;
    free(p->surrounding);p->surrounding=strdup(text);p->cursor=cursor;p->anchor=anchor;p->sensitive=sensitive;p->wanted=enabled;publish(p);
#endif
}
