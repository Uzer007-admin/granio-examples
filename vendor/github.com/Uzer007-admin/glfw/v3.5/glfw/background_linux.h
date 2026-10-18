#ifndef GLFW_GO_BACKGROUND_LINUX_H
#define GLFW_GO_BACKGROUND_LINUX_H
#define GLFW_INCLUDE_NONE
#include "glfw/include/GLFW/glfw3.h"
#include <stdint.h>

typedef struct GLFWgoBackgroundState {
    int valid, ownerThread, platform;
    int alphaSurface, transparent, composition, blur, requested;
    uint64_t revision, compositor;
} GLFWgoBackgroundState;

GLFWgoBackgroundState glfwQueryBackgroundBridge(GLFWwindow* window);
int glfwSetBackgroundBlurBridge(GLFWwindow* window, int enabled);
#endif
