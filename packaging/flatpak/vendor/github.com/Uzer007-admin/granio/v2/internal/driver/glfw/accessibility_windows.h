//go:build accessibility && windows
#ifndef ACCESSIBILITY_WINDOWS_H
#define ACCESSIBILITY_WINDOWS_H
#include <stdint.h>
void WinAccessibilityBegin(uintptr_t hwnd);
void WinAccessibilityKeep(uintptr_t hwnd,uint64_t id,uint64_t parent);
void WinAccessibilityAdd(uintptr_t hwnd, uint64_t id, uint64_t parent,
    const char* name, const char* value, int role, int states, int patterns,
    double x, double y, double width, double height);
void WinAccessibilityEnd(uintptr_t hwnd);
void WinAccessibilityDetails(uintptr_t hwnd,uint64_t id,int start,int end,int rows,int columns,int row,int column,int multiple,double value,double minimum,double maximum,double step);
void WinAccessibilityCleanup(uintptr_t hwnd);
int WinAccessibilityListening(void);
void WinAccessibilityEvent(uintptr_t hwnd,uint64_t id,int event);
void WinAccessibilityBoolProperty(uintptr_t hwnd,uint64_t id,int property,int before,int after,int boolean);
void WinAccessibilityNumberProperty(uintptr_t hwnd,uint64_t id,int property,double before,double after);
void WinAccessibilityTextProperty(uintptr_t hwnd,uint64_t id,int property,const char* before,const char* after);
int WinAccessibilityContractTest(void);
int WinAccessibilityTextContractTest(void);
int WinAccessibilityEventsContractTest(void);
int WinAccessibilityScrollTextRange(uintptr_t hwnd,uint64_t id,int start,int end,int top);
int WinAccessibilitySelectionContractTest(uintptr_t hwnd,uint64_t id,uint64_t child);
int WinAccessibilityVirtualContractTest(uintptr_t hwnd,uint64_t id);
#endif
