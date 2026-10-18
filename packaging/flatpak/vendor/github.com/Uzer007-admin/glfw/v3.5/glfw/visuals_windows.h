#ifndef GLFW_GO_VISUALS_WINDOWS_H
#define GLFW_GO_VISUALS_WINDOWS_H
#define GLFW_INCLUDE_NONE
#include "glfw/include/GLFW/glfw3.h"
#include <stdint.h>

typedef struct GLFWgoVisualState {
    uint32_t build, supported;
    int valid, ownerThread, composition, remote, contrast, contrastKnown;
    int transparent, pixelFormat, alphaBits, backdrop, backdropKnown;
} GLFWgoVisualState;
typedef struct GLFWgoVisualResult {
    uint32_t hr[4];
    int valid, ownerThread;
} GLFWgoVisualResult;
GLFWgoVisualState glfwQueryVisualsBridge(GLFWwindow*);
GLFWgoVisualResult glfwApplyVisualsBridge(GLFWwindow*, uint32_t, int, int, int, int);
#endif
