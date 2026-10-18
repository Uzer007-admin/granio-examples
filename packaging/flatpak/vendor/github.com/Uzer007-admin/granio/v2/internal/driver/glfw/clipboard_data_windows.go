//go:build windows && !wasm && !test_web_driver

package glfw

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf16"
	"unsafe"

	"github.com/Uzer007-admin/granio/v2"
	"golang.org/x/sys/windows"
)

var nativeClipboardMu sync.Mutex

const clipboardLengthsMIME = "application/x-fyne-clipboard-lengths"

var (
	cbUser32     = windows.NewLazySystemDLL("user32.dll")
	cbOpen       = cbUser32.NewProc("OpenClipboard")
	cbClose      = cbUser32.NewProc("CloseClipboard")
	cbEmpty      = cbUser32.NewProc("EmptyClipboard")
	cbGet        = cbUser32.NewProc("GetClipboardData")
	cbSet        = cbUser32.NewProc("SetClipboardData")
	cbEnum       = cbUser32.NewProc("EnumClipboardFormats")
	cbRegister   = cbUser32.NewProc("RegisterClipboardFormatW")
	cbFormatName = cbUser32.NewProc("GetClipboardFormatNameW")
	cbKernel     = windows.NewLazySystemDLL("kernel32.dll")
	cbAlloc      = cbKernel.NewProc("GlobalAlloc")
	cbFree       = cbKernel.NewProc("GlobalFree")
	cbLock       = cbKernel.NewProc("GlobalLock")
	cbUnlock     = cbKernel.NewProc("GlobalUnlock")
	cbSize       = cbKernel.NewProc("GlobalSize")
	cbShell      = windows.NewLazySystemDLL("shell32.dll")
	cbQueryFile  = cbShell.NewProc("DragQueryFileW")
)

func openClipboard() error {
	for attempt := 0; attempt < 4; attempt++ {
		if result, _, _ := cbOpen.Call(0); result != 0 {
			return nil
		}
		if attempt < 3 {
			time.Sleep(4 * time.Millisecond)
		}
	}
	return errors.New("clipboard is busy")
}
func winClipboardFormat(format string) (uint32, error) {
	if format == granio.MIMETextPlain {
		return 13, nil
	}
	name := format
	if format == granio.MIMETextHTML {
		name = "HTML Format"
	}
	if format == granio.MIMEImagePNG {
		name = "PNG"
	}
	if strings.HasPrefix(format, "application/x-windows-clipboard-format;") {
		_, params, err := mime.ParseMediaType(format)
		if err != nil {
			return 0, err
		}
		number, err := strconv.ParseUint(params["id"], 10, 32)
		return uint32(number), err
	}
	text, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return 0, err
	}
	id, _, err := cbRegister.Call(uintptr(unsafe.Pointer(text)))
	if id == 0 {
		return 0, err
	}
	return uint32(id), nil
}
func clipboardFormats() ([]string, error) {
	nativeClipboardMu.Lock()
	defer nativeClipboardMu.Unlock()
	if err := openClipboard(); err != nil {
		return nil, err
	}
	defer cbClose.Call()
	formats := make(map[string][]byte)
	var previous uintptr
	for {
		next, _, _ := cbEnum.Call(previous)
		if next == 0 {
			break
		}
		previous = next
		format := ""
		switch next {
		case 13:
			format = granio.MIMETextPlain
		case 15:
			format = granio.MIMETextURIList
		default:
			name := make([]uint16, 256)
			count, _, _ := cbFormatName.Call(next, uintptr(unsafe.Pointer(&name[0])), uintptr(len(name)))
			if count > 0 {
				format = string(utf16.Decode(name[:min(int(count), len(name))]))
				if format == "PNG" {
					format = granio.MIMEImagePNG
				}
				if format == "HTML Format" {
					format = granio.MIMETextHTML
				}
			}
			if validateClipboardFormat(format) != nil {
				format = fmt.Sprintf("application/x-windows-clipboard-format;id=%d", next)
			}
		}
		if format != clipboardLengthsMIME {
			formats[format] = nil
		}
	}
	return sortedClipboardFormats(formats), nil
}
func readClipboardData(format string) ([]byte, error) {
	nativeClipboardMu.Lock()
	defer nativeClipboardMu.Unlock()
	id, err := winClipboardFormat(format)
	if err != nil {
		return nil, err
	}
	if err = openClipboard(); err != nil {
		return nil, err
	}
	defer cbClose.Call()
	handle, _, _ := cbGet.Call(uintptr(id))
	if handle == 0 && format == granio.MIMETextURIList {
		handle, _, _ = cbGet.Call(15)
		if handle != 0 {
			return winClipboardFileList(handle)
		}
	}
	if handle == 0 {
		return nil, errClipboardFormat
	}
	size, _, _ := cbSize.Call(handle)
	if size > maxClipboardBytes {
		return nil, errors.New("clipboard representation exceeds 64 MiB")
	}
	if size == 0 {
		return []byte{}, nil
	}
	pointer, _, err := cbLock.Call(handle)
	if pointer == 0 {
		return nil, err
	}
	defer cbUnlock.Call(handle)
	data := append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(pointer)), int(size))...)
	// GlobalSize may include allocator padding. A private companion representation
	// preserves exact lengths without changing interoperable native payloads.
	metaID, _ := winClipboardFormat(clipboardLengthsMIME)
	if metaID != id {
		if metaHandle, _, _ := cbGet.Call(uintptr(metaID)); metaHandle != 0 {
			metaSize, _, _ := cbSize.Call(metaHandle)
			if metaSize > 0 && metaSize <= 32768 {
				metaPointer, _, _ := cbLock.Call(metaHandle)
				if metaPointer != 0 {
					raw := append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(metaPointer)), int(metaSize))...)
					cbUnlock.Call(metaHandle)
					raw = bytes.TrimRight(raw, "\x00")
					var lengths map[uint32]int
					if json.Unmarshal(raw, &lengths) == nil {
						if length, ok := lengths[id]; ok && length >= 0 && length <= len(data) {
							data = data[:length]
						}
					}
				}
			}
		}
	}
	if id == 13 {
		runes := make([]uint16, 0, len(data)/2)
		for i := 0; i+1 < len(data); i += 2 {
			unit := binary.LittleEndian.Uint16(data[i:])
			if unit == 0 {
				break
			}
			runes = append(runes, unit)
		}
		return []byte(string(utf16.Decode(runes))), nil
	}
	if format == granio.MIMETextHTML {
		return decodeClipboardHTML(data), nil
	}
	return data, nil
}
func winClipboardFileList(handle uintptr) ([]byte, error) {
	count, _, _ := cbQueryFile.Call(handle, 0xffffffff, 0, 0)
	if count > 1<<16 {
		return nil, errors.New("clipboard file count exceeds limit")
	}
	var result strings.Builder
	for i := uintptr(0); i < count; i++ {
		size, _, _ := cbQueryFile.Call(handle, i, 0, 0)
		if size > 32768 {
			return nil, errors.New("clipboard file name exceeds limit")
		}
		path := make([]uint16, size+1)
		cbQueryFile.Call(handle, i, uintptr(unsafe.Pointer(&path[0])), size+1)
		value := filepath.ToSlash(string(utf16.Decode(path[:size])))
		uri := url.URL{Scheme: "file", Path: "/" + value}
		if strings.HasPrefix(value, "//") {
			parts := strings.SplitN(value[2:], "/", 2)
			uri.Host = parts[0]
			uri.Path = "/"
			if len(parts) > 1 {
				uri.Path += parts[1]
			}
		}
		result.WriteString(uri.String() + "\r\n")
		if result.Len() > maxClipboardBytes {
			return nil, errors.New("clipboard file list exceeds limit")
		}
	}
	return []byte(result.String()), nil
}
func writeClipboardData(data map[string][]byte) error {
	nativeClipboardMu.Lock()
	defer nativeClipboardMu.Unlock()
	prepared := make(map[uint32]uintptr, len(data)+1)
	lengths := make(map[uint32]int, len(data)+1)
	defer func() {
		for _, handle := range prepared {
			if handle != 0 {
				cbFree.Call(handle)
			}
		}
	}()
	for _, format := range sortedClipboardFormats(data) {
		if format == clipboardLengthsMIME {
			return errors.New("reserved clipboard representation")
		}
		id, err := winClipboardFormat(format)
		if err != nil {
			return err
		}
		value := data[format]
		if format == granio.MIMETextPlain {
			units := utf16.Encode([]rune(string(value)))
			value = make([]byte, (len(units)+1)*2)
			for i, unit := range units {
				binary.LittleEndian.PutUint16(value[i*2:], unit)
			}
		} else if format == granio.MIMETextHTML {
			value = encodeClipboardHTML(value)
		}
		handle, err := allocateClipboardData(value)
		if err != nil {
			return err
		}
		if previous := prepared[id]; previous != 0 {
			cbFree.Call(previous)
		}
		prepared[id] = handle
		lengths[id] = len(value)
		if format == granio.MIMETextURIList {
			if files := encodeClipboardFiles(data[format]); files != nil {
				handle, err = allocateClipboardData(files)
				if err != nil {
					return err
				}
				prepared[15] = handle
				lengths[15] = len(files)
			}
		}
	}
	metadata, _ := json.Marshal(lengths)
	metaID, err := winClipboardFormat(clipboardLengthsMIME)
	if err != nil {
		return err
	}
	metaHandle, err := allocateClipboardData(metadata)
	if err != nil {
		return err
	}
	prepared[metaID] = metaHandle
	if err := openClipboard(); err != nil {
		return err
	}
	defer cbClose.Call()
	if result, _, err := cbEmpty.Call(); result == 0 {
		return err
	}
	for id, handle := range prepared {
		if result, _, err := cbSet.Call(uintptr(id), handle); result == 0 {
			return err
		}
		prepared[id] = 0
	}
	return nil
}
func allocateClipboardData(data []byte) (uintptr, error) {
	if len(data) > maxClipboardBytes {
		return 0, errors.New("clipboard representation exceeds 64 MiB")
	}
	handle, _, err := cbAlloc.Call(2, uintptr(max(1, len(data))))
	if handle == 0 {
		return 0, err
	}
	pointer, _, err := cbLock.Call(handle)
	if pointer == 0 {
		cbFree.Call(handle)
		return 0, err
	}
	copy(unsafe.Slice((*byte)(unsafe.Pointer(pointer)), max(1, len(data))), data)
	cbUnlock.Call(handle)
	return handle, nil
}
func encodeClipboardFiles(data []byte) []byte {
	var names []uint16
	for _, line := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		uri, err := url.Parse(line)
		if err != nil || uri.Scheme != "file" {
			return nil
		}
		path := filepath.FromSlash(strings.TrimPrefix(uri.Path, "/"))
		if uri.Host != "" {
			path = "\\\\" + uri.Host + filepath.FromSlash(uri.Path)
		}
		units, err := windows.UTF16FromString(path)
		if err != nil {
			return nil
		}
		names = append(names, units...)
	}
	if len(names) == 0 {
		return nil
	}
	names = append(names, 0)
	out := make([]byte, 20+len(names)*2)
	binary.LittleEndian.PutUint32(out, 20)
	binary.LittleEndian.PutUint32(out[16:], 1)
	for i, unit := range names {
		binary.LittleEndian.PutUint16(out[20+i*2:], unit)
	}
	return out
}
func encodeClipboardHTML(fragment []byte) []byte {
	header := "Version:0.9\r\nStartHTML:%010d\r\nEndHTML:%010d\r\nStartFragment:%010d\r\nEndFragment:%010d\r\n"
	prefix, suffix := "<html><body><!--StartFragment-->", "<!--EndFragment--></body></html>"
	start := len(fmt.Sprintf(header, 0, 0, 0, 0))
	value := fmt.Sprintf(header, start, start+len(prefix)+len(fragment)+len(suffix), start+len(prefix), start+len(prefix)+len(fragment))
	return []byte(value + prefix + string(fragment) + suffix)
}
func decodeClipboardHTML(data []byte) []byte {
	header := string(data[:min(4096, len(data))])
	offset := func(name string) int {
		start := strings.Index(header, name+":")
		if start < 0 {
			return -1
		}
		value := header[start+len(name)+1:]
		end := strings.IndexAny(value, "\r\n")
		if end >= 0 {
			value = value[:end]
		}
		number, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return -1
		}
		return number
	}
	start, end := offset("StartFragment"), offset("EndFragment")
	if start >= 0 && end >= start && end <= len(data) {
		return append([]byte(nil), data[start:end]...)
	}
	return data
}
