//go:build windows
#include "glfw/src/internal.h"
#include "_cgo_export.h"
int fynePreTranslateWin32(MSG*message){if(message->message!=WM_KEYDOWN&&message->message!=WM_KEYUP&&message->message!=WM_SYSKEYDOWN&&message->message!=WM_SYSKEYUP)return 0;return goFynePreTranslateWin32((uintptr_t)message->hwnd,message->message,message->wParam,message->lParam);}
