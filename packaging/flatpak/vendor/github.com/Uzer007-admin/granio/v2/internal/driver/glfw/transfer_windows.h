//go:build windows && !wasm && !test_web_driver
#ifndef FYNE_TRANSFER_WINDOWS_H
#define FYNE_TRANSFER_WINDOWS_H
#include <stdint.h>
int FyneDropBegin(uintptr_t hwnd);
void FyneDropAdd(uintptr_t hwnd,unsigned int format,const char* mime);
void FyneDropRemove(uintptr_t hwnd);
int FyneExternalDropTest(uintptr_t hwnd,unsigned int format,const void* data,int size);
#endif
