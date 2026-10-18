package glfw

import "unsafe"

// These helpers copy native arrays before their GLFW-owned storage can change.
// No C calls are needed for individual elements.
func copyNativeSlice[T ~float32 | ~uint16](dst []T, source unsafe.Pointer, count int) []T {
	if source == nil {
		return nil
	}
	if dst == nil || cap(dst) < count {
		dst = make([]T, count)
	} else {
		dst = dst[:count]
	}
	copy(dst, unsafe.Slice((*T)(source), count))
	return dst
}

func copyNativeBytes[T ~int](dst []T, source unsafe.Pointer, count int) []T {
	if source == nil {
		return nil
	}
	if dst == nil || cap(dst) < count {
		dst = make([]T, count)
	} else {
		dst = dst[:count]
	}
	for i, value := range unsafe.Slice((*byte)(source), count) {
		dst[i] = T(value)
	}
	return dst
}

func validateGammaRamp(ramp *GammaRamp) int {
	if ramp == nil || len(ramp.Red) == 0 || len(ramp.Red) != len(ramp.Green) || len(ramp.Red) != len(ramp.Blue) || uint64(len(ramp.Red)) > uint64(^uint32(0)) {
		panic(&Error{Code: invalidValue, Desc: "gamma ramp channels must have equal, nonzero lengths representable by unsigned int"})
	}
	return len(ramp.Red)
}
