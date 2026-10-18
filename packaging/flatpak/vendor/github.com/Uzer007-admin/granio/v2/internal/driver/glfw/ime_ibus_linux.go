//go:build linux && !wasm && !test_web_driver

package glfw

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/Uzer007-admin/dbus/v5"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/internal/build"
)

const ibusContext = "org.freedesktop.IBus.InputContext"

type ibusInput struct {
	conn           *dbus.Conn
	path           dbus.ObjectPath
	window         *window
	signals        chan *dbus.Signal
	done           chan struct{}
	closed         atomic.Bool
	focusEpoch     atomic.Uint64
	focusBarrier   atomic.Uint64
	receiver       desktop.TextInput // UI thread only
	text           string
	cursor, anchor int
	focused        bool
	caret          [4]int32
	sensitive      bool
}

var ibusWindows sync.Map // *window -> *ibusInput
func (w *window) initIBus() *ibusInput {
	if build.IsWayland {
		return nil
	}
	address := os.Getenv("IBUS_ADDRESS")
	if address == "" && strings.Contains(os.Getenv("XMODIFIERS"), "ibus") {
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		output, err := exec.CommandContext(ctx, "ibus", "address").Output()
		cancel()
		if err == nil && len(output) < 4096 {
			address = strings.TrimSpace(string(output))
		}
	}
	// The selected local input-method session supplies this address. Do not
	// guess a TCP endpoint or mutate the process/user's input configuration.
	if !strings.HasPrefix(address, "unix:") {
		return nil
	}
	conn, err := dbus.Connect(address)
	if err != nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var path dbus.ObjectPath
	err = conn.Object("org.freedesktop.IBus", "/org/freedesktop/IBus").CallWithContext(ctx, "org.freedesktop.IBus.CreateInputContext", 0, "Fyne").Store(&path)
	if err != nil || !path.IsValid() {
		conn.Close()
		return nil
	}
	input := &ibusInput{conn: conn, path: path, window: w, signals: make(chan *dbus.Signal, 128), done: make(chan struct{})}
	conn.Signal(input.signals)
	if conn.AddMatchSignal(dbus.WithMatchObjectPath(path), dbus.WithMatchInterface(ibusContext)) != nil {
		conn.Close()
		return nil
	}
	// Fyne draws inline preedit; the session's IBus panel owns candidates.
	if input.call("SetCapabilities", uint32(1|8|32)) != nil {
		conn.Close()
		return nil
	}
	ibusWindows.Store(w, input)
	go input.listen()
	return input
}
func (i *ibusInput) call(method string, args ...interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	return i.conn.Object("org.freedesktop.IBus", i.path).CallWithContext(ctx, ibusContext+"."+method, 0, args...).Err
}
func (i *ibusInput) focusCall(method string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	call := i.conn.Object("org.freedesktop.IBus", i.path).CallWithContext(ctx, ibusContext+"."+method, 0)
	if call.Err == nil {
		i.focusBarrier.Store(uint64(call.ResponseSequence))
	}
	return call.Err
}
func (i *ibusInput) key(sym, code, state uint32, release bool) bool {
	if i.closed.Load() || !i.focused {
		return false
	}
	if release {
		state |= 1 << 30
	}
	if code >= 8 {
		code -= 8
	}
	var handled bool
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	err := i.conn.Object("org.freedesktop.IBus", i.path).CallWithContext(ctx, ibusContext+".ProcessKeyEvent", 0, sym, code, state).Store(&handled)
	if err != nil {
		i.close()
		if i.receiver != nil {
			i.receiver.CancelComposition()
		}
		return false
	}
	return handled
}
func ibusText(value dbus.Variant) (string, bool) {
	fields, ok := value.Value().([]interface{})
	if !ok || len(fields) < 3 || fields[0] != "IBusText" {
		return "", false
	}
	text, ok := fields[2].(string)
	return text, ok && len(text) <= 1<<20 && utf8.ValidString(text)
}

type ibusAttributes struct {
	Name        string
	Attachments map[string]dbus.Variant
	Attributes  []dbus.Variant
}
type ibusSerializedText struct {
	Name        string
	Attachments map[string]dbus.Variant
	Text        string
	Attributes  dbus.Variant
}

func (i *ibusInput) update() {
	if i.closed.Load() {
		return
	}
	input, _ := i.window.canvas.Focused().(desktop.TextInput)
	if curWindow != i.window {
		input = nil
	}
	if disabled, ok := input.(granio.Disableable); ok && disabled.Disabled() {
		input = nil
	}
	if input != i.receiver {
		i.focusEpoch.Add(1)
		if i.focused {
			_ = i.focusCall("Reset")
			_ = i.focusCall("FocusOut")
		}
		i.receiver = input
		i.focused = input != nil
		i.text = ""
		i.cursor, i.anchor = -1, -1
		if i.focused {
			_ = i.focusCall("FocusIn")
		}
	}
	if input == nil {
		return
	}
	state := input.TextInputState()
	if state.Sensitive != i.sensitive {
		purpose, hints := uint32(0), uint32(0)
		if state.Sensitive {
			purpose, hints = 8, 1<<11
		}
		_ = i.call("SetContentType", purpose, hints)
		i.sensitive = state.Sensitive
	}
	text, cursor, anchor := state.Text, state.Cursor, state.SelectionStart
	if state.Sensitive {
		text = ""
		cursor, anchor = 0, 0
	}
	if text != i.text || cursor != i.cursor || anchor != i.anchor {
		attributes := dbus.MakeVariant(ibusAttributes{"IBusAttrList", map[string]dbus.Variant{}, []dbus.Variant{}})
		value := dbus.MakeVariant(ibusSerializedText{"IBusText", map[string]dbus.Variant{}, text, attributes})
		_ = i.call("SetSurroundingText", value, uint32(cursor), uint32(anchor))
		i.text, i.cursor, i.anchor = text, cursor, anchor
	}
	if object, ok := input.(granio.CanvasObject); ok {
		position := i.window.driver.AbsolutePositionForObject(object).Add(state.CaretPosition)
		x, y := i.window.viewport.GetPos()
		scale := i.window.canvas.Scale()
		caret := [4]int32{int32(x) + int32(position.X*scale), int32(y) + int32((position.Y+state.CaretSize.Height)*scale), int32(max(1, state.CaretSize.Width*scale)), int32(max(1, state.CaretSize.Height*scale))}
		if caret != i.caret {
			_ = i.call("SetCursorLocation", caret[0], caret[1], caret[2], caret[3])
			i.caret = caret
		}
	}
}
func (i *ibusInput) listen() {
	for {
		var signal *dbus.Signal
		select {
		case <-i.done:
			return
		case signal = <-i.signals:
		}
		if signal == nil || i.closed.Load() {
			return
		}
		if signal.Path != i.path || !strings.HasPrefix(signal.Name, ibusContext+".") {
			continue
		}
		name := strings.TrimPrefix(signal.Name, ibusContext+".")
		body := signal.Body
		epoch := i.focusEpoch.Load()
		var text string
		var cursor uint32
		var visible bool
		if name == "CommitText" || name == "UpdatePreeditText" || name == "UpdatePreeditTextWithMode" {
			if len(body) == 0 {
				continue
			}
			value, ok := body[0].(dbus.Variant)
			if !ok {
				continue
			}
			text, ok = ibusText(value)
			if !ok {
				continue
			}
			if name != "CommitText" {
				if len(body) < 3 {
					continue
				}
				cursor, _ = body[1].(uint32)
				visible, _ = body[2].(bool)
			}
		}
		runOnMainWithWait(func() {
			if i.closed.Load() || i.window.closing || !i.focused || i.receiver == nil || epoch != i.focusEpoch.Load() || uint64(signal.Sequence) <= i.focusBarrier.Load() {
				return
			}
			switch name {
			case "CommitText":
				i.receiver.CommitComposition(text)
			case "UpdatePreeditText", "UpdatePreeditTextWithMode":
				if visible {
					i.receiver.UpdateComposition(desktop.CompositionEvent{Text: text, Cursor: min(int(cursor), utf8.RuneCountInString(text))})
				} else {
					i.receiver.CancelComposition()
				}
			case "HidePreeditText":
				i.receiver.CancelComposition()
			case "RequireSurroundingText":
				i.cursor, i.anchor = -1, -1
			case "DeleteSurroundingText":
				if len(body) < 2 {
					return
				}
				offset, ok := body[0].(int32)
				count, valid := body[1].(uint32)
				state := i.receiver.TextInputState()
				if !ok || !valid || state.Sensitive || count > 1<<20 {
					return
				}
				start := int64(state.Cursor) + int64(offset)
				end := start + int64(count)
				if start < 0 || end > int64(utf8.RuneCountInString(state.Text)) {
					return
				}
				if editor, ok := i.receiver.(desktop.TextInputRangeEditor); ok {
					editor.ReplaceTextRange("", int(start), int(end))
				}
			}
			i.update()
		}, false)
	}
}
func (i *ibusInput) close() {
	if i.closed.Swap(true) {
		return
	}
	close(i.done)
	i.conn.RemoveSignal(i.signals)
	_ = i.call("Destroy")
	_ = i.conn.Close()
	ibusWindows.Delete(i.window)
}
func (w *window) cleanupIBus() {
	if stored, ok := ibusWindows.Load(w); ok {
		stored.(*ibusInput).close()
	}
}
