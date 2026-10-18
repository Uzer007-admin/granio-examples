package gl

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/internal"
)

type materialPipeline interface {
	Draw(*painter, *canvas.Material, granio.Position, granio.Size, *internal.ClipItem) bool
	ResetFrame()
	Close()
	Statistics() map[string]uint64
}

func (p *painter) drawMaterial(m *canvas.Material, pos granio.Position, frame granio.Size, clip *internal.ClipItem) {
	disabled := !m.Enabled || !m.Valid()
	if theme, ok := granio.CurrentApp().Settings().Theme().(interface{ HighContrast() bool }); ok {
		disabled = disabled || theme.HighContrast()
	}
	if settings, ok := granio.CurrentApp().Settings().(granio.SystemAppearanceProvider); ok {
		appearance := settings.SystemAppearance()
		disabled = disabled || appearance.HighContrast.Available && appearance.HighContrast.Value || appearance.Transparency.Available && !appearance.Transparency.Value
	}
	if !disabled {
		if p.materials == nil && !p.materialUnavailable {
			p.materials = newMaterialPipeline(p)
			p.materialUnavailable = p.materials == nil
		}
		if p.materials != nil && p.materials.Draw(p, m, pos, frame, clip) {
			return
		}
	}
	p.drawRectangle(m.FallbackRectangle(), pos, frame)
}

// ReleaseMaterials is called with the owning context current before Destroy.
func (p *painter) ReleaseMaterials() {
	if p.materials != nil {
		p.materials.Close()
		p.materials = nil
	}
}
func (p *painter) MaterialStatistics() map[string]uint64 {
	if p.materials == nil {
		return map[string]uint64{}
	}
	return p.materials.Statistics()
}
