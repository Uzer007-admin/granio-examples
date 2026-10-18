package glfw

import (
	"math"
	"os"
	"strconv"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
)

const (
	baselineDPI = 120.0
	maxDPI      = 1000
	minDPI      = 10
	mmPerInch   = 25.4
	scaleEnvKey = "FYNE_SCALE"
	scaleAuto   = float32(-1.0) // some platforms allow setting auto-scale (linux/BSD)
)

func calculateDetectedScale(widthMm, widthPx int) float32 {
	dpi := float32(widthPx) / (float32(widthMm) / mmPerInch)
	if dpi > maxDPI || dpi < minDPI {
		dpi = baselineDPI
	}

	scale := float32(float64(dpi) / baselineDPI)
	if scale < 1.0 {
		return 1.0
	}
	return scale
}

func calculateScale(user, system, detected float32) float32 {
	return calculateScaleWithPolicy(desktop.ScaleLegacyRounded, user, system, detected)
}

func validScale(value float32) bool {
	return value > 0 && !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0)
}

func calculateScaleWithPolicy(policy desktop.ScalePolicy, user, system, detected float32) float32 {
	if user < 0 {
		user = 1.0
	}

	if system == scaleAuto {
		system = detected
	}
	if !validScale(user) {
		user = 1
	}
	if !validScale(system) {
		system = 1
	}

	raw := system * user
	if !validScale(raw) {
		return 1
	}
	if policy == desktop.ScaleExact {
		return raw
	}
	// A tiny positive preference must not become a zero divisor in pixel/DIP
	// conversions. Multiply in float64 so valid float32 scales cannot overflow.
	rounded := float32(math.Round(float64(raw)*10.0) / 10.0)
	return max(0.1, rounded)
}

func userScale() float32 {
	env := os.Getenv(scaleEnvKey)

	if env != "" && env != "auto" {
		scale, err := strconv.ParseFloat(env, 32)
		if err == nil && scale != 0 {
			return float32(scale)
		}
		granio.LogError("Error reading scale", err)
	}

	if env != "auto" {
		if setting := granio.CurrentApp().Settings().Scale(); setting > 0 {
			return setting
		}
	}

	return 1.0 // user preference for auto is now passed as 1 so the system auto is picked up
}
