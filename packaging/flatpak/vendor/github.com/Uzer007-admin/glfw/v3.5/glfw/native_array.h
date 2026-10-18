#ifndef GO_GLFW_NATIVE_ARRAY_H
#define GO_GLFW_NATIVE_ARRAY_H
#include "glfw/include/GLFW/glfw3.h"

typedef struct { const float *data; int count; } GLFWgoAxes;
typedef struct { const unsigned char *data; int count; } GLFWgoBytes;
typedef struct { GLFWmonitor **data; int count; } GLFWgoMonitors;
typedef struct { const GLFWvidmode *data; int count; } GLFWgoVideoModes;
typedef struct { GLFWgamepadstate state; int present; } GLFWgoGamepad;

GLFWgoAxes glfwGetJoystickAxesBridge(int joystick);
GLFWgoBytes glfwGetJoystickButtonsBridge(int joystick);
GLFWgoBytes glfwGetJoystickHatsBridge(int joystick);
GLFWgoGamepad glfwGetGamepadStateBridge(int joystick);
GLFWgoMonitors glfwGetMonitorsBridge(void);
GLFWgoVideoModes glfwGetVideoModesBridge(GLFWmonitor *monitor);
typedef struct {
    int width, height, framebufferWidth, framebufferHeight;
    float scaleX, scaleY;
    int focused, iconified, visible, maximized, transparent;
} GLFWgoSurfaceMetrics;
GLFWgoSurfaceMetrics glfwGetSurfaceMetricsBridge(GLFWwindow *window);
typedef struct {
    int width, height, framebufferWidth, framebufferHeight;
    float scaleX, scaleY;
} GLFWgoSurfaceGeometry;
GLFWgoSurfaceGeometry glfwGetSurfaceGeometryBridge(GLFWwindow *window);
#endif
