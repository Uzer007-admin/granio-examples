//go:build windows && cgo && !ci && !wasm && !test_web_driver && !android && !ios && !mobile && !tinygo && !noos && !tamago

package app

/*
#cgo windows LDFLAGS: -lole32 -luuid -luser32
#include "system_appearance_windows.h"
*/
import "C"

import (
	"context"
	"fmt"
	"image/color"
	"runtime"

	"github.com/Uzer007-admin/granio/v2"
)

func windowsAppearance(raw C.FyneAppearanceSnapshot) granio.SystemAppearance {
	known := func(mask uint32) bool { return uint32(raw.available)&mask != 0 }
	argb := func(v C.uint32_t) color.NRGBA {
		return color.NRGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 255}
	}
	snapshot := granio.SystemAppearance{Available: raw.available != 0}
	if !snapshot.Available {
		snapshot.UnavailableReason = "Windows appearance properties unavailable"
		return snapshot
	}
	if known(C.FYNE_APPEARANCE_COLORS) {
		background := argb(raw.background)
		variant := granio.SystemColorLight
		if int(background.R)*299+int(background.G)*587+int(background.B)*114 < 128000 {
			variant = granio.SystemColorDark
		}
		snapshot.ColorScheme = granio.SystemPreference[granio.SystemColorScheme]{Value: variant, Available: true}
		snapshot.Accent = granio.SystemPreference[color.NRGBA]{Value: argb(raw.accent), Available: true}
	}
	snapshot.HighContrast = granio.SystemPreference[bool]{Value: raw.high_contrast != 0, Available: known(C.FYNE_APPEARANCE_CONTRAST)}
	snapshot.ReducedMotion = granio.SystemPreference[bool]{Value: raw.animations == 0, Available: known(C.FYNE_APPEARANCE_MOTION)}
	snapshot.Transparency = granio.SystemPreference[bool]{Value: raw.transparency != 0, Available: known(C.FYNE_APPEARANCE_EFFECTS)}
	snapshot.TextScale = granio.SystemPreference[float32]{Value: float32(raw.text_scale), Available: known(C.FYNE_APPEARANCE_TEXT)}
	snapshot.Colors = granio.SystemPreference[granio.SystemColors]{Available: true, Value: granio.SystemColors{
		Window: argb(raw.colors[0]), WindowText: argb(raw.colors[1]), ButtonFace: argb(raw.colors[2]), ButtonText: argb(raw.colors[3]),
		Highlight: argb(raw.colors[4]), HighlightText: argb(raw.colors[5]), GrayText: argb(raw.colors[6]), Hotlight: argb(raw.colors[7]),
	}}
	if uint32(raw.available)&^uint32(raw.subscriptions) != 0 {
		snapshot.UnavailableReason = "some Windows appearance events unavailable; using 500ms fallback reads"
	}
	return snapshot
}

func (s *settings) cacheSystemVariant(granio.ThemeVariant) {}

func (s *settings) startSystemAppearance() {
	s.appearanceMu.Lock()
	if s.appearanceCancel != nil || s.appearanceStopped {
		s.appearanceMu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	initial := make(chan granio.SystemAppearance, 1)
	ready := make(chan struct{})
	s.appearanceCancel, s.appearanceDone = cancel, done
	s.appearanceMu.Unlock()
	go s.runWindowsAppearance(ctx, done, initial, ready)
	// Run calls this before starting the driver. Do not enqueue a UI callback and
	// wait for a loop which has not started yet. The first frame sees this value.
	snapshot := <-initial
	if ctx.Err() == nil {
		s.applySystemAppearance(snapshot)
	}
	close(ready)
}

func (s *settings) runWindowsAppearance(ctx context.Context, done chan<- struct{}, initial chan<- granio.SystemAppearance, ready <-chan struct{}) {
	defer close(done)
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	var hr C.int32_t
	native := C.fyne_appearance_open(&hr)
	if native == nil {
		initial <- granio.SystemAppearance{Generation: 1, UnavailableReason: fmt.Sprintf("Windows appearance activation failed: HRESULT 0x%08x", uint32(hr))}
		return
	}
	defer C.fyne_appearance_close(native)
	// Synchronize an in-flight cancellation before releasing native memory.
	canceled := make(chan struct{})
	stopCancel := context.AfterFunc(ctx, func() { C.fyne_appearance_cancel(native); close(canceled) })
	defer func() {
		if !stopCancel() {
			<-canceled
		}
	}()
	var raw C.FyneAppearanceSnapshot
	C.fyne_appearance_read(native, &raw)
	previous := windowsAppearance(raw)
	previous.Generation = 1
	initial <- previous
	select {
	case <-ready:
	case <-ctx.Done():
		s.appearance.Store(&granio.SystemAppearance{Generation: 2, UnavailableReason: "Windows appearance provider closed"})
		return
	}
	for {
		timeout := C.uint32_t(0xffffffff) // no periodic work when all OS events are supported
		if uint32(raw.available)&^uint32(raw.subscriptions) != 0 {
			timeout = 500
		}
		result := C.fyne_appearance_wait(native, timeout)
		if result <= 0 || ctx.Err() != nil {
			closed := granio.SystemAppearance{Generation: previous.Generation + 1, UnavailableReason: "Windows appearance provider closed"}
			if ctx.Err() == nil {
				s.queueSystemAppearance(ctx, closed)
			} else {
				s.appearance.Store(&closed)
			}
			return
		}
		C.fyne_appearance_read(native, &raw)
		next := windowsAppearance(raw)
		next.Generation = previous.Generation
		if next != previous {
			next.Generation++
			previous = next
			s.queueSystemAppearance(ctx, next)
		}
	}
}
