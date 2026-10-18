# Local gl-js correction

Public upstream: https://github.com/fyne-io/gl-js

Snapshot: `v0.2.1-0.20260315212741-029c47fd27e8` (Go module archive).
The original BSD licence and source notices are preserved.

The WebAssembly `TexSubImage2D` binding converts its named `Enum` format to an
ordinary `int` before `syscall/js.Value.Call`. Named numeric Go types are not
accepted by `syscall/js.ValueOf`; glyph-atlas uploads otherwise panic at runtime.
The existing `SliceToTypedArray` conversion already handles pixel byte slices.
`ReadPixels` also uses a JavaScript typed array and copies the result back into
the supplied Go destination. Passing raw Go slices directly to `Value.Call`
previously caused the same runtime panic when capturing a canvas.

`go.work` selects this local snapshot. This is a workspace correction, with no
published upstream fork or changes to the global Go module cache.
