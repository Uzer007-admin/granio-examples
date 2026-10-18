//go:build windows
#define CINTERFACE
#define COBJMACROS
#include <windows.h>
#include <msctf.h>
#include <ole2.h>
#include "ime_tsf_windows.h"
static ITfInputProcessorProfiles*testProfiles;
static LANGID testPreviousLanguage;
static HKL testPreviousLayout;
static BOOL testPreviouslyEnabled;
static int testStep,testEnabledChanged,testOwnLayout;
static HKL testJapaneseLayout;
static const GUID japaneseClass={0x03b5835f,0xf03c,0x411b,{0x9c,0xe2,0xaa,0x23,0xe1,0x17,0x1e,0x36}};
static const GUID japaneseProfile={0xa76c93d9,0x5523,0x4e90,{0xaa,0xfa,0x4d,0xb1,0x12,0xf9,0xac,0x76}};
// The gated acceptance test restores the exact enabled flag and calling-thread
// language/layout. No FORSESSION activation or installed-language-list mutation.
long FyneRealJapaneseBegin(uintptr_t hwnd){if(testProfiles)return E_UNEXPECTED;testStep=1;testEnabledChanged=testOwnLayout=0;HRESULT hr=CoCreateInstance(&CLSID_TF_InputProcessorProfiles,NULL,CLSCTX_INPROC_SERVER,&IID_ITfInputProcessorProfiles,(void**)&testProfiles);if(FAILED(hr))return hr;testPreviousLayout=GetKeyboardLayout(0);ITfInputProcessorProfiles_GetCurrentLanguage(testProfiles,&testPreviousLanguage);testStep=2;hr=ITfInputProcessorProfiles_IsEnabledLanguageProfile(testProfiles,&japaneseClass,0x411,&japaneseProfile,&testPreviouslyEnabled);if(FAILED(hr))return hr;if(!testPreviouslyEnabled){testStep=3;hr=ITfInputProcessorProfiles_EnableLanguageProfile(testProfiles,&japaneseClass,0x411,&japaneseProfile,TRUE);if(FAILED(hr))return hr;testEnabledChanged=1;}testStep=4;HKL layouts[64];int count=GetKeyboardLayoutList(64,layouts);testJapaneseLayout=LoadKeyboardLayoutW(L"00000411",KLF_NOTELLSHELL);if(testJapaneseLayout){testOwnLayout=1;for(int i=0;i<count;i++)if(layouts[i]==testJapaneseLayout)testOwnLayout=0;ActivateKeyboardLayout(testJapaneseLayout,0);}hr=ITfInputProcessorProfiles_ChangeCurrentLanguage(testProfiles,0x411);if(SUCCEEDED(hr)){testStep=5;ITfInputProcessorProfileMgr*profiles=NULL;hr=ITfInputProcessorProfiles_QueryInterface(testProfiles,&IID_ITfInputProcessorProfileMgr,(void**)&profiles);if(SUCCEEDED(hr)){hr=ITfInputProcessorProfileMgr_ActivateProfile(profiles,TF_PROFILETYPE_INPUTPROCESSOR,0x411,&japaneseClass,&japaneseProfile,NULL,0);ITfInputProcessorProfileMgr_Release(profiles);}}if(SUCCEEDED(hr)){SetForegroundWindow((HWND)hwnd);SetFocus((HWND)hwnd);}return hr;}
void FyneRealJapaneseEnd(void){if(!testProfiles)return;ITfInputProcessorProfiles_ChangeCurrentLanguage(testProfiles,testPreviousLanguage);ActivateKeyboardLayout(testPreviousLayout,0);if(testEnabledChanged)ITfInputProcessorProfiles_EnableLanguageProfile(testProfiles,&japaneseClass,0x411,&japaneseProfile,FALSE);if(testOwnLayout)UnloadKeyboardLayout(testJapaneseLayout);ITfInputProcessorProfiles_Release(testProfiles);testProfiles=NULL;}
int FyneRealIMEKey(uintptr_t hwnd,int key){if(GetForegroundWindow()!=(HWND)hwnd)return 0;INPUT events[2]={{0},{0}};events[0].type=events[1].type=INPUT_KEYBOARD;events[0].ki.wVk=events[1].ki.wVk=key;events[1].ki.dwFlags=KEYEVENTF_KEYUP;return SendInput(2,events,sizeof(INPUT))==2;}

int FyneRealJapaneseStep(void){return testStep;}
