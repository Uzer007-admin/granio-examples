//go:build (gles || arm || arm64) && !android && !ios && !mobile && !darwin && !wasm && !test_web_driver

package gl

import (
	"strings"
	"unsafe"

	gl "github.com/Uzer007-admin/gl/v2.0/gles2"

	"github.com/Uzer007-admin/granio/v2"
)

const (
	arrayBuffer           = gl.ARRAY_BUFFER
	bitColorBuffer        = gl.COLOR_BUFFER_BIT
	bitDepthBuffer        = gl.DEPTH_BUFFER_BIT
	clampToEdge           = gl.CLAMP_TO_EDGE
	colorFormatRGBA       = gl.RGBA
	compileStatus         = gl.COMPILE_STATUS
	constantAlpha         = gl.CONSTANT_ALPHA
	float                 = gl.FLOAT
	fragmentShader        = gl.FRAGMENT_SHADER
	front                 = gl.FRONT
	back                  = gl.BACK
	glFalse               = gl.FALSE
	linkStatus            = gl.LINK_STATUS
	maxTextureSizeParam   = gl.MAX_TEXTURE_SIZE
	one                   = gl.ONE
	zero                  = gl.ZERO
	oneMinusConstantAlpha = gl.ONE_MINUS_CONSTANT_ALPHA
	oneMinusSrcAlpha      = gl.ONE_MINUS_SRC_ALPHA
	scissorTest           = gl.SCISSOR_TEST
	srcAlpha              = gl.SRC_ALPHA
	staticDraw            = gl.STATIC_DRAW
	texture0              = gl.TEXTURE0
	texture1              = gl.TEXTURE1
	texture2D             = gl.TEXTURE_2D
	textureNearest        = gl.NEAREST
	textureMinFilter      = gl.TEXTURE_MIN_FILTER
	textureMagFilter      = gl.TEXTURE_MAG_FILTER
	textureWrapS          = gl.TEXTURE_WRAP_S
	textureWrapT          = gl.TEXTURE_WRAP_T
	triangles             = gl.TRIANGLES
	triangleStrip         = gl.TRIANGLE_STRIP
	unsignedByte          = gl.UNSIGNED_BYTE
	vertexShader          = gl.VERTEX_SHADER
)

const (
	noBuffer  = Buffer(0)
	noProgram = Program(0)
	noShader  = Shader(0)
)

type (
	// Attribute represents a GL attribute
	Attribute int32
	// Buffer represents a GL buffer
	Buffer uint32
	// Program represents a compiled GL program
	Program uint32
	// Shader represents a GL shader
	Shader uint32
	// Uniform represents a GL uniform
	Uniform int32
)

var textureFilterToGL = [...]int32{gl.LINEAR, gl.NEAREST, gl.LINEAR}

// glyphAtlasSupported enables batched atlas text; see atlas.go.
const glyphAtlasSupported = true

func (p *painter) Init() {
	p.ctx = nil
	procs, err := gl.Load()
	if err != nil {
		granio.LogError("failed to initialise OpenGL", err)
		return
	}
	p.ctx = &esContext{procs: procs}
	p.maxTextureSize = p.ctx.GetInteger(maxTextureSizeParam)

	procs.Disable(gl.DEPTH_TEST)
	procs.Enable(gl.BLEND)
	p.logError()
	p.programs = p.compilePrograms()
}

type esContext struct {
	procs         *gl.Functions
	premultiplied bool
}

var _ context = (*esContext)(nil)

func (c *esContext) ActiveTexture(textureUnit uint32) {
	c.procs.ActiveTexture(textureUnit)
}

func (c *esContext) AttachShader(program Program, shader Shader) {
	c.procs.AttachShader(uint32(program), uint32(shader))
}

func (c *esContext) BindBuffer(target uint32, buf Buffer) {
	c.procs.BindBuffer(target, uint32(buf))
}

func (c *esContext) BindTexture(target uint32, texture Texture) {
	c.procs.BindTexture(target, uint32(texture))
}

func (c *esContext) BlendColor(r, g, b, a float32) {
	c.procs.BlendColor(r, g, b, a)
}

func (c *esContext) BlendFunc(srcFactor, destFactor uint32) {
	if c.premultiplied {
		alphaSource := srcFactor
		if srcFactor == gl.SRC_ALPHA {
			alphaSource = gl.ONE
		}
		c.procs.BlendFuncSeparate(srcFactor, destFactor, alphaSource, destFactor)
		return
	}
	c.procs.BlendFunc(srcFactor, destFactor)
}

func (c *esContext) setPremultipliedAlpha(enabled bool) bool {
	if enabled && !c.procs.IsFunctionAvailable("glBlendFuncSeparate") {
		return false
	}
	c.premultiplied = enabled
	return true
}

func (c *esContext) BufferData(target uint32, points []float32, usage uint32) {
	c.procs.BufferData(target, 4*len(points), gl.SlicePtr(points), usage)
}

func (c *esContext) BufferSubData(target uint32, points []float32) {
	c.procs.BufferSubData(target, 0, 4*len(points), gl.SlicePtr(points))
}

func (c *esContext) Clear(mask uint32) {
	c.procs.Clear(mask)
}

func (c *esContext) ClearColor(r, g, b, a float32) {
	c.procs.ClearColor(r, g, b, a)
}

func (c *esContext) CompileShader(shader Shader) {
	c.procs.CompileShader(uint32(shader))
}

func (c *esContext) CreateBuffer() Buffer {
	var vbo uint32
	c.procs.GenBuffers(1, &vbo)
	return Buffer(vbo)
}

func (c *esContext) CreateProgram() Program {
	return Program(c.procs.CreateProgram())
}

func (c *esContext) CreateShader(typ uint32) Shader {
	return Shader(c.procs.CreateShader(typ))
}

func (c *esContext) CreateTexture() (texture Texture) {
	var tex uint32
	c.procs.GenTextures(1, &tex)
	return Texture(tex)
}

func (c *esContext) DeleteBuffer(buffer Buffer) {
	c.procs.DeleteBuffers(1, (*uint32)(&buffer))
}

func (c *esContext) DeleteProgram(program Program) {
	c.procs.DeleteProgram(uint32(program))
}

func (c *esContext) DeleteTexture(texture Texture) {
	tex := uint32(texture)
	c.procs.DeleteTextures(1, &tex)
}

func (c *esContext) Disable(capability uint32) {
	c.procs.Disable(capability)
}

func (c *esContext) DrawArrays(mode uint32, first, count int) {
	c.procs.DrawArrays(mode, int32(first), int32(count))
}

func (c *esContext) Enable(capability uint32) {
	c.procs.Enable(capability)
}

func (c *esContext) EnableVertexAttribArray(attribute Attribute) {
	c.procs.EnableVertexAttribArray(uint32(attribute))
}

func (c *esContext) GetAttribLocation(program Program, name string) Attribute {
	return Attribute(c.procs.GetAttribLocation(uint32(program), gl.Str(name+"\x00")))
}

func (c *esContext) GetError() uint32 {
	return c.procs.GetError()
}

func (c *esContext) GetInteger(pname uint32) int {
	var value int32
	c.procs.GetIntegerv(pname, &value)
	return int(value)
}

func (c *esContext) GetProgrami(program Program, param uint32) int {
	var value int32
	c.procs.GetProgramiv(uint32(program), param, &value)
	return int(value)
}

func (c *esContext) GetProgramInfoLog(program Program) string {
	var logLength int32
	c.procs.GetProgramiv(uint32(program), gl.INFO_LOG_LENGTH, &logLength)
	info := strings.Repeat("\x00", int(logLength+1))
	c.procs.GetProgramInfoLog(uint32(program), logLength, nil, gl.Str(info))
	return info
}

func (c *esContext) GetShaderi(shader Shader, param uint32) int {
	var value int32
	c.procs.GetShaderiv(uint32(shader), param, &value)
	return int(value)
}

func (c *esContext) GetShaderInfoLog(shader Shader) string {
	var logLength int32
	c.procs.GetShaderiv(uint32(shader), gl.INFO_LOG_LENGTH, &logLength)
	info := strings.Repeat("\x00", int(logLength+1))
	c.procs.GetShaderInfoLog(uint32(shader), logLength, nil, gl.Str(info))
	return info
}

func (c *esContext) GetUniformLocation(program Program, name string) Uniform {
	return Uniform(c.procs.GetUniformLocation(uint32(program), gl.Str(name+"\x00")))
}

func (c *esContext) LinkProgram(program Program) {
	c.procs.LinkProgram(uint32(program))
}

func (c *esContext) CopyTexSubImage2D(target uint32, level, xoffset, yoffset, x, y, width, height int) {
	c.procs.CopyTexSubImage2D(target, int32(level), int32(xoffset), int32(yoffset), int32(x), int32(y), int32(width), int32(height))
}

func (c *esContext) ReadBuffer(_ uint32) {
	// GLES 2 reads the current framebuffer; it has no ReadBuffer entrypoint or
	// desktop front/back buffer selection. Match the mobile/WebGL contract.
}

func (c *esContext) ReadPixels(x, y, width, height int, colorFormat, typ uint32, pixels []uint8) {
	c.procs.ReadPixels(int32(x), int32(y), int32(width), int32(height), colorFormat, typ, gl.SlicePtr(pixels))
}

func (c *esContext) Scissor(x, y, w, h int32) {
	c.procs.Scissor(x, y, w, h)
}

func (c *esContext) ShaderSource(shader Shader, source string) {
	if err := c.procs.ShaderSourceString(uint32(shader), source); err != nil {
		granio.LogError("failed to upload shader source", err)
	}
}

func (c *esContext) TexImage2D(target uint32, level, width, height int, colorFormat, typ uint32, data []uint8) {
	var ptr unsafe.Pointer
	if len(data) > 0 {
		ptr = gl.SlicePtr(data)
	}
	c.procs.TexImage2D(
		target,
		int32(level),
		int32(colorFormat),
		int32(width),
		int32(height),
		0,
		colorFormat,
		typ,
		ptr,
	)
}

func (c *esContext) TexSubImage2D(target uint32, level, x, y, width, height int, colorFormat, typ uint32, data []uint8) {
	c.procs.TexSubImage2D(
		target,
		int32(level),
		int32(x),
		int32(y),
		int32(width),
		int32(height),
		colorFormat,
		typ,
		gl.SlicePtr(data),
	)
}

func (c *esContext) TexParameteri(target, param uint32, value int32) {
	c.procs.TexParameteri(target, param, value)
}

func (c *esContext) Uniform1f(uniform Uniform, v float32) {
	c.procs.Uniform1f(int32(uniform), v)
}

func (c *esContext) Uniform1fv(uniform Uniform, v []float32) {
	c.procs.Uniform1fv(int32(uniform), int32(len(v)), &v[0])
}

func (c *esContext) Uniform1i(uniform Uniform, v int32) {
	c.procs.Uniform1i(int32(uniform), v)
}

func (c *esContext) Uniform2f(uniform Uniform, v0, v1 float32) {
	c.procs.Uniform2f(int32(uniform), v0, v1)
}

func (c *esContext) Uniform2fv(uniform Uniform, v []float32) {
	c.procs.Uniform2fv(int32(uniform), int32(len(v)/2), &v[0])
}

func (c *esContext) Uniform4f(uniform Uniform, v0, v1, v2, v3 float32) {
	c.procs.Uniform4f(int32(uniform), v0, v1, v2, v3)
}

func (c *esContext) UseProgram(program Program) {
	c.procs.UseProgram(uint32(program))
}

func (c *esContext) VertexAttribPointerWithOffset(attribute Attribute, size int, typ uint32, normalized bool, stride, offset int) {
	c.procs.VertexAttribPointerWithOffset(uint32(attribute), int32(size), typ, normalized, int32(stride), uintptr(offset))
}

func (c *esContext) Viewport(x, y, width, height int) {
	c.procs.Viewport(int32(x), int32(y), int32(width), int32(height))
}
