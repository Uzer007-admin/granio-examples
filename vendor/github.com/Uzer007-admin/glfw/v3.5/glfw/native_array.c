#define GLFW_INCLUDE_NONE
#include "native_array.h"

// Output pointers stay on the C stack even when native getters invoke Go callbacks.
// Only the resulting values cross cgo; no Go count/state pointer is passed to C.
GLFWgoAxes glfwGetJoystickAxesBridge(int joystick) {
    GLFWgoAxes result = {0};
    result.data = glfwGetJoystickAxes(joystick, &result.count);
    return result;
}
GLFWgoBytes glfwGetJoystickButtonsBridge(int joystick) {
    GLFWgoBytes result = {0};
    result.data = glfwGetJoystickButtons(joystick, &result.count);
    return result;
}
GLFWgoBytes glfwGetJoystickHatsBridge(int joystick) {
    GLFWgoBytes result = {0};
    result.data = glfwGetJoystickHats(joystick, &result.count);
    return result;
}
GLFWgoGamepad glfwGetGamepadStateBridge(int joystick) {
    GLFWgoGamepad result = {0};
    result.present = glfwGetGamepadState(joystick, &result.state);
    return result;
}
GLFWgoMonitors glfwGetMonitorsBridge(void) {
    GLFWgoMonitors result = {0};
    result.data = glfwGetMonitors(&result.count);
    return result;
}
GLFWgoVideoModes glfwGetVideoModesBridge(GLFWmonitor *monitor) {
    GLFWgoVideoModes result = {0};
    result.data = glfwGetVideoModes(monitor, &result.count);
    return result;
}

GLFWgoSurfaceGeometry glfwGetSurfaceGeometryBridge(GLFWwindow *window) {
    GLFWgoSurfaceGeometry result = {0};
    glfwGetWindowSize(window, &result.width, &result.height);
    glfwGetFramebufferSize(window, &result.framebufferWidth, &result.framebufferHeight);
    glfwGetWindowContentScale(window, &result.scaleX, &result.scaleY);
    return result;
}

GLFWgoSurfaceMetrics glfwGetSurfaceMetricsBridge(GLFWwindow *window) {
    GLFWgoSurfaceMetrics result = {0};
    GLFWgoSurfaceGeometry geometry = glfwGetSurfaceGeometryBridge(window);
    result.width = geometry.width;
    result.height = geometry.height;
    result.framebufferWidth = geometry.framebufferWidth;
    result.framebufferHeight = geometry.framebufferHeight;
    result.scaleX = geometry.scaleX;
    result.scaleY = geometry.scaleY;
    result.focused = glfwGetWindowAttrib(window, GLFW_FOCUSED);
    result.iconified = glfwGetWindowAttrib(window, GLFW_ICONIFIED);
    result.visible = glfwGetWindowAttrib(window, GLFW_VISIBLE);
    result.maximized = glfwGetWindowAttrib(window, GLFW_MAXIMIZED);
    result.transparent = glfwGetWindowAttrib(window, GLFW_TRANSPARENT_FRAMEBUFFER);
    return result;
}
