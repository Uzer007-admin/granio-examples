//go:build !android && !ios && !mobile && !wasm && !test_web_driver

package gl

import (
	"math"

	"github.com/Uzer007-admin/gl/capabilities"
	capnative "github.com/Uzer007-admin/gl/capabilities/native"
	"github.com/Uzer007-admin/gl/material"
	es "github.com/Uzer007-admin/gl/v2.0/gles2"
	core "github.com/Uzer007-admin/gl/v2.1/material"
	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/internal"
	"github.com/Uzer007-admin/granio/v2/theme"
)

type nativeMaterialPipeline struct {
	ctx      context
	es       bool
	mask     programState
	device   *material.GLDevice
	renderer *material.Renderer
	source   material.Target
	slot     int
	keys     [material.MaxLayers]material.Key
	owners   [material.MaxLayers]*canvas.Material
	valid    [material.MaxLayers]bool
	quality  [material.MaxLayers]uint8
	serial   uint64
}

func newMaterialPipeline(p *painter) materialPipeline {
	context, err := capnative.Query(glfw.GetProcAddress)
	if err != nil {
		return nil
	}
	var api material.DrawProcedures
	if context.API == capabilities.OpenGLES {
		api, err = es.LoadWithProcAddrFunc(glfw.GetProcAddress)
	} else {
		api, err = core.LoadWithProcAddrFunc(glfw.GetProcAddress)
	}
	if err != nil {
		return nil
	}
	device, err := material.NewGLDevice(context, api)
	if err != nil {
		return nil
	}
	// 16 MiB is reserved for the GPU-only lower-layer capture, including its
	// transactional resize. Pool texels and capture together never exceed 64 MiB.
	renderer, err := material.NewRenderer(context, api, 48<<20)
	if err != nil {
		return nil
	}
	return &nativeMaterialPipeline{ctx: p.ctx, es: context.API == capabilities.OpenGLES, device: device, renderer: renderer}
}
func (n *nativeMaterialPipeline) ResetFrame() { n.slot = 0 }
func (n *nativeMaterialPipeline) Close() {
	if n.mask.ref != noProgram {
		n.ctx.DeleteProgram(n.mask.ref)
		n.ctx.DeleteBuffer(n.mask.buff)
	}
	n.renderer.Close()
	if n.source.Texture != 0 {
		n.device.DeleteTarget(n.source.Handle)
	}
	n.source = material.Target{}
}
func (n *nativeMaterialPipeline) Statistics() map[string]uint64 {
	s := n.renderer.Stats()
	source := uint64(n.source.Width) * uint64(n.source.Height) * 4
	return map[string]uint64{"pool_bytes": s.UsedBytes, "capture_bytes": source, "cache_hits": s.CacheHits, "target_allocations": s.TargetAllocations, "budget_bytes": material.DefaultBudget}
}
func (n *nativeMaterialPipeline) Draw(p *painter, m *canvas.Material, pos granio.Position, frame granio.Size, clip *internal.ClipItem) bool {
	if p.windowSurface == WindowSurfaceBackdrop && !n.ensureMask(p) {
		return false
	}
	slot := n.slot
	n.slot++
	if slot >= material.MaxLayers {
		return false
	}
	x, y := int(math.Round(float64(pos.X*p.pixScale))), int(math.Round(float64(pos.Y*p.pixScale)))
	w, h := int(math.Round(float64(m.Size().Width*p.pixScale))), int(math.Round(float64(m.Size().Height*p.pixScale)))
	fw := int(math.Round(float64(frame.Width * p.pixScale)))
	if w <= 0 || h <= 0 || x < 0 || y < 0 || x > fw-w || y > p.fbHeight-h || w > n.device.MaxTextureSize() || h > n.device.MaxTextureSize() {
		return false
	}
	bytes := uint64(w) * uint64(h) * 4
	if bytes > 16<<20 {
		return false
	}
	if n.source.Width < w || n.source.Height < h {
		width, height := max(w, n.source.Width), max(h, n.source.Height)
		bytes = uint64(width) * uint64(height) * 4
		oldBytes := uint64(n.source.Width) * uint64(n.source.Height) * 4
		if oldBytes+bytes > 16<<20 {
			return false
		}
		handle, err := n.device.CreateTarget(width, height)
		if err != nil {
			return false
		}
		if n.source.Texture != 0 {
			n.device.DeleteTarget(n.source.Handle)
		}
		n.source = material.Target{Handle: handle, Width: width, Height: height}
		n.valid = [material.MaxLayers]bool{}
	}
	quality := m.Downsample
	if quality == 0 {
		quality = 2
	}
	bg := theme.Color(theme.ColorNameBackground)
	r, g, b, a := bg.RGBA()
	key := material.Key{Surface: uint64(uint32(fw))<<32 | uint64(uint32(p.fbHeight)), Content: p.materialGeneration, Style: uint64(uint32(x))<<32 | uint64(uint32(y)), ScaleX: p.pixScale, ScaleY: p.pixScale, Radius: m.BlurRadius, Corners: m.CornerRadius, Noise: m.Noise, Tint: [4]float32{float32(m.Tint.R) / 255, float32(m.Tint.G) / 255, float32(m.Tint.B) / 255, float32(m.Tint.A) / 255}, Linear: m.Linear}
	key.Content ^= uint64(r)<<48 | uint64(g)<<32 | uint64(b)<<16 | uint64(a)
	if n.owners[slot] != m {
		n.serial++
		n.owners[slot] = m
		n.valid[slot] = false
	}
	key.Surface ^= n.serial
	p.FlushGlyphs()
	if !n.valid[slot] || n.keys[slot] != key || n.quality[slot] != quality {
		capture := n.source
		capture.Width, capture.Height = w, h
		if err := n.device.CopySnapshot(capture, x, p.fbHeight-y-h); err != nil {
			return false
		}
	}
	result, err := n.renderer.Render(slot, n.source, material.Region{Width: w, Height: h}, quality, key)
	p.ctx.Enable(0x0BE2)
	if clip != nil {
		p.StartClipping(clip.Rect())
	} else {
		p.StopClipping()
	}
	if err != nil {
		n.valid[slot] = false
		return false
	}
	n.keys[slot], n.valid[slot], n.quality[slot] = key, true, quality
	// Material passes own blend/scissor/program/vertex state; restore the Fyne
	// clipping contract and use its premultiplied image program for the result.
	p.ctx.Enable(0x0BE2)
	if clip != nil {
		p.StartClipping(clip.Rect())
	} else {
		p.StopClipping()
	}
	points, insets, inner := p.rectCoords(m.Size(), pos, frame, canvas.ImageFillStretch, 1, 0)
	for i := 3; i < len(points); i += 4 {
		points[i] = 1 - points[i]
	} // framebuffer textures have a bottom-left origin
	program := p.programs.simple
	if p.windowSurface == WindowSurfaceBackdrop {
		// The result already contains the lower layers. Erase their covered
		// contribution once before adding it, rather than alpha-over twice.
		mask := n.mask
		p.useProgram(mask.ref)
		p.updateBuffer(mask.buff, points[:])
		p.UpdateVertexArray(mask, attrVertex, 2, 4, 0)
		p.UpdateVertexArray(mask, attrVertexTextureCoordinates, 2, 4, 2)
		p.SetUniform4f(mask, "geometry", float32(w)/key.ScaleX, float32(h)/key.ScaleY, key.Corners, min(key.ScaleX, key.ScaleY))
		p.SetUniform2f(mask, "resolution", float32(result.Target.Width), float32(result.Target.Height))
		p.SetUniform1f(mask, "pass", 0)
		p.ctx.BlendFunc(zero, oneMinusSrcAlpha)
		p.ctx.DrawArrays(triangleStrip, 0, 4)
		p.SetUniform1f(mask, "pass", 1)
		p.ctx.BlendFunc(one, one)
		p.ctx.ActiveTexture(texture0)
		p.ctx.BindTexture(texture2D, Texture(result.Target.Texture))
		p.ctx.DrawArrays(triangleStrip, 0, 4)
		return true
	}
	p.useProgram(program.ref)
	p.updateBuffer(program.buff, points[:])
	p.UpdateVertexArray(program, attrVertex, 2, 4, 0)
	p.UpdateVertexArray(program, attrVertexTextureCoordinates, 2, 4, 2)
	p.SetUniform1f(program, attrRadiusCorner, 0)
	p.SetUniform2f(program, attrSize, inner.Width*p.pixScale, inner.Height*p.pixScale)
	p.SetUniform4f(program, attrInset, insets[0], insets[1], insets[2], insets[3])
	p.SetUniform1f(program, attrAlpha, 1)
	p.ctx.BlendFunc(one, oneMinusSrcAlpha)
	p.ctx.ActiveTexture(texture0)
	p.ctx.BindTexture(texture2D, Texture(result.Target.Texture))
	p.ctx.DrawArrays(triangleStrip, 0, 4)
	return true
}
