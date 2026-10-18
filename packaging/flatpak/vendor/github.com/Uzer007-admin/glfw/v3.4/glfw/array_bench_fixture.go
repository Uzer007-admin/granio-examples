//go:build glfw_bench

package glfw

//float GetAxisAtIndex(float *axis, int i);
import "C"

import "unsafe"

// Retain the previous per-element cgo conversion for a reproducible comparison.
func legacyAxisCopy(dst []float32, source unsafe.Pointer) {
	for i := range dst {
		dst[i] = float32(C.GetAxisAtIndex((*C.float)(source), C.int(i)))
	}
}
