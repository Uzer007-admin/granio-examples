package desktop

// FrameRateDriver is an optional capability for desktop frame scheduling.
// A zero rate selects the highest refresh rate reported by connected monitors.
// The target is an upper bound; expensive frames and display synchronization
// can limit the delivered rate.
type FrameRateDriver interface {
	SetFrameRate(framesPerSecond int) error
	FrameRate() int
}
