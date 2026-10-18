//go:build windows && !wasm && !test_web_driver
#ifndef FYNE_TSF_H
#define FYNE_TSF_H
#include <stdint.h>
long FyneRealJapaneseBegin(uintptr_t hwnd);
int FyneRealJapaneseStep(void);
long FyneTSFOpenForTest(uintptr_t hwnd);
void FyneRealJapaneseEnd(void);
int FyneRealIMEKey(uintptr_t hwnd,int key);
int FyneTSFCreate(uintptr_t hwnd);
void FyneTSFUpdate(uintptr_t hwnd, const unsigned short* text, int length,
    int cursor, int start, int end, int enabled, int x, int y, int width, int height);
void FyneTSFReset(uintptr_t hwnd);
void FyneTSFDestroy(uintptr_t hwnd);
int FyneTSFKey(uintptr_t hwnd, unsigned int message, uintptr_t key, intptr_t flags);
int FyneTSFContractTest(void);
int FyneTSFEditContractTest(uintptr_t hwnd);
#endif
