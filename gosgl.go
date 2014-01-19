/*Package gosgl is package gosgl
 */
package gosgl

import (
	"image"
	"io/ioutil"

	"github.com/go-gl/gl"
	"github.com/go-gl/glh"
)

type Drawer struct {
	program gl.Program
	vao     gl.VertexArray
}

type canvas struct {
	W, H int
}

func MakeCanvas(w, h int) Canvas {
	return Canvas{w, h}
}

var (
	g_QuadraticDrawer *Drawer
)

func NewDrawer(vshader, fshader string) *Drawer {
	vao := gl.GenVertexArray()
	vao.Bind()

	vsh := sgl.ShaderFromFile(gl.VERTEX_SHADER, vshader)
	fsh := sgl.ShaderFromFile(gl.FRAGMENT_SHADER, fshader)

	program := glh.NewProgram(vsh, fsh)
	program.BindFragDataLocation(0, "outColor")

	return &Drawer{
		program: program,
		vao:     vao,
	}
}

func NewQuadraticDrawer() *Drawer {
	dr := NewDrawer("vshader.glsl", "quadratic_fshader.glsl")
	program := dr.program
	posAttr := program.GetAttribLocation("position")
	posAttr.AttribPointer(2, gl.FLOAT, false, 4*4, uintptr(0))
	posAttr.EnableArray()

	texAttr := program.GetAttribLocation("texcoord")
	texAttr.AttribPointer(2, gl.FLOAT, false, 4*4, uintptr(8))
	texAttr.EnableArray()

	return dr
}

func Init() {
	g_QuadraticDrawer = NewQuadraticDrawer()

	vbo := gl.GenBuffer()
	vbo.Bind(gl.ARRAY_BUFFER)
}

func Point(x, y) image.Point {
	return image.Point{x, y}
}

type QuadraticCurve struct {
	Points [3]image.Point
}

func MakeQuadraticCurve(x1, y1, x2, y2, x3, y3 int) {
	return QuadraticCurve{image.Point{x1, y1}, image.Point{x2, y2}, image.Point{x3, y3}}
}

type GLPoint struct {
	X, Y float32
}

func (canv canvas) toGLPoints(points []image.Point) []GLPoint {
	ps := make([]GLPoint, len(points))
	for i, p := range points {
		ps[i] = GLPoint{float32(p.X) / canv.W, float32(p.Y) / canv.H}
	}
	return ps
}

func (c QuadraticCurve) Draw(canv canvas) {
	g_QuadraticDrawer.vao.Bind()
	g_QuadraticDrawer.program.Use()
	p := canv.toGLPoints(c.Points)
	vertices := []float32{
		p[1].X, p[1].Y, 0.0, 0.0,
		p[2].X, p[2].Y, 0.5, 0.0,
		p[3].X, p[3].Y, 1.0, 1.0,
	}
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, vertices, gl.STATIC_DRAW)
	gl.DrawArrays(gl.TRIANGLES, 0, 3)
}

func ShaderFromFile(stype gl.GLenum, filename string) (shader glh.Shader) {
	fcont, _ := ioutil.ReadFile(filename)
	shader = glh.Shader{stype, string(fcont[:])}
	shader.Compile()
	return shader
}
