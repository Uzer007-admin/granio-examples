//go:build !android && !ios && !mobile && !wasm && !test_web_driver

package gl

func (n *nativeMaterialPipeline) ensureMask(p *painter) bool {
	if n.mask.ref != noProgram {
		return true
	}
	prefix := "#version 110\n"
	if n.es {
		prefix = "#version 100\nprecision highp float;\n"
	}
	program, err := p.createProgram(shaderVertTexturedPassthrough2D, []byte(prefix+materialCoverageShader))
	if err != nil {
		return false
	}
	n.mask = programState{ref: program, buff: p.createBuffer(coordinatesSizeRectangleWithTexture), uniforms: make(map[string]*uniformState), attributes: make(map[string]Attribute)}
	return true
}

// Match the material renderer's rounded mask and its bilinear upsampling. The
// shader has no sampler: this pass neither copies nor reads the framebuffer.
// Erase-by-coverage + add-premultiplied-result preserves the snapshot's alpha,
// including partially covered corners over transparent lower canvas layers.
const materialCoverageShader = `
uniform vec4 geometry; // logical extent xy, corner radius z, physical scale w
uniform vec2 resolution;
uniform sampler2D tex;
uniform float pass;
varying vec2 fragTexCoord;
float coverage(vec2 p) {
    p=clamp(p,0.5/resolution,1.0-0.5/resolution);
    float r=min(geometry.z,min(geometry.x,geometry.y)*0.5);
    vec2 q=abs((p-0.5)*geometry.xy)-geometry.xy*0.5+r;
    float d=length(max(q,vec2(0.0)))+min(max(q.x,q.y),0.0)-r;
    return clamp(0.5-d*geometry.w,0.0,1.0);
}
void main() {
    if (pass>0.5) {gl_FragColor=texture2D(tex,fragTexCoord);return;}
    vec2 pixel=fragTexCoord*resolution-0.5;
    vec2 base=(floor(pixel)+0.5)/resolution;
    vec2 step=1.0/resolution;
    vec2 f=fract(pixel);
    float a=mix(mix(coverage(base),coverage(base+vec2(step.x,0.0)),f.x),
                mix(coverage(base+vec2(0.0,step.y)),coverage(base+step),f.x),f.y);
    gl_FragColor=vec4(0.0,0.0,0.0,a);
}
`
