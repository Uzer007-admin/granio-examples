//go:build system_appearance_integration && windows && cgo && !ci && !wasm && !test_web_driver && !android && !ios && !mobile && !tinygo && !noos && !tamago

package app

/*
#include "system_appearance_windows.h"
*/
import "C"

import (
	"fmt"
	"github.com/Uzer007-admin/granio/v2"
)

type nativeAppearanceProbe struct{ native *C.FyneAppearance }

func openNativeAppearanceProbe() (*nativeAppearanceProbe, error) {
	var hr C.int32_t
	p := C.fyne_appearance_open(&hr)
	if p == nil {
		return nil, fmt.Errorf("HRESULT 0x%08x", uint32(hr))
	}
	return &nativeAppearanceProbe{p}, nil
}
func (p *nativeAppearanceProbe) read() (granio.SystemAppearance, uint32) {
	var raw C.FyneAppearanceSnapshot
	C.fyne_appearance_read(p.native, &raw)
	return windowsAppearance(raw), uint32(raw.subscriptions)
}
func (p *nativeAppearanceProbe) diagnostics() map[string]any {
	var raw C.FyneAppearanceSnapshot
	C.fyne_appearance_read(p.native, &raw)
	errors := make([]string, 5)
	for i, hr := range raw.event_errors {
		errors[i] = fmt.Sprintf("0x%08x", uint32(hr))
	}
	return map[string]any{"winrt_subscriptions": uint32(raw.winrt_subscriptions), "win32_subscriptions": uint32(raw.win32_subscriptions), "event_hresult": errors}
}
func (p *nativeAppearanceProbe) notify() { C.fyne_appearance_notify(p.native) }
func (p *nativeAppearanceProbe) postSettingsMessage() bool {
	return C.fyne_appearance_post_settings_message(p.native) != 0
}
func (p *nativeAppearanceProbe) wait(ms uint32) int {
	return int(C.fyne_appearance_wait(p.native, C.uint32_t(ms)))
}
func (p *nativeAppearanceProbe) cancel() { C.fyne_appearance_cancel(p.native) }
func (p *nativeAppearanceProbe) close()  { C.fyne_appearance_close(p.native); p.native = nil }
func nativeAppearanceResources() [3]int {
	var providers, delegates, wakes C.long
	C.fyne_appearance_resources(&providers, &delegates, &wakes)
	return [3]int{int(providers), int(delegates), int(wakes)}
}
