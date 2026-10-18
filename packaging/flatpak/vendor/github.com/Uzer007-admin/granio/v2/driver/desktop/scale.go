package desktop

// ScalePolicy controls geometry scaling without changing framebuffer density.
type ScalePolicy uint8

const (
	// ScaleLegacyRounded preserves the historic rounding to tenths.
	ScaleLegacyRounded ScalePolicy = iota
	// ScaleExact preserves fractional system and user scales.
	ScaleExact
)

// WindowScaleInfo separates logical geometry from framebuffer pixel density.
// EffectiveScale converts Fyne coordinates to native window coordinates;
// multiplying by FramebufferScale converts them to framebuffer pixels.
type WindowScaleInfo struct {
	Policy                                                   ScalePolicy
	SystemScale, UserScale, EffectiveScale, FramebufferScale float32
}

// ScalableWindow is optional. Like other Window mutations, SetScalePolicy must
// be called on the UI goroutine. It can be used before Show or on a live window.
type ScalableWindow interface {
	SetScalePolicy(ScalePolicy) error
	ScaleInfo() WindowScaleInfo
}
