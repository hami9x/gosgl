/*Package gosgl is package gosgl
 */
package gosgl

import (
	"container/list"
	"image/color"
	"io/ioutil"
	"path"
	"runtime"

	"github.com/Jragonmiris/mathgl"
	"github.com/go-gl/gl"
	"github.com/go-gl/glh"
	"github.com/phaikawl/poly2tri-go/p2t"
)

const (
	oo = 32767 //Infinity
)

var (
	gQuadraticApproxPrecision float64 = 10
	gGl                       *OpenGL
	Black                     = color.RGBA{0, 0, 0, 255}
)

//Describe opengl state object
type OpenGL struct {
	QuadraticDrawer *GlDrawer
	TriangleDrawer  *GlDrawer
	FillDrawer      *GlDrawer
	currentDrawer   *GlDrawer
	configs         []drawConfig

	QuadraticDrawConfig *QuadraticDrawConfig
	GlColorConfig       *GlColorConfig
}

func (g *OpenGL) Activate(dr *GlDrawer) {
	dr.Activate()
	g.currentDrawer = dr

	for _, config := range g.configs {
		config.SetProgram(g.currentDrawer.program)
		config.Apply()
	}
}

func Init() {
	gGl = OpenGLInit()
}

type GlColorConfig struct {
	glProgramInfo
	color color.Color
}

func (conf *GlColorConfig) SetColor(color color.Color) {
	conf.color = color
}

func (conf *GlColorConfig) Reset() {
	conf.color = Black
}

func (conf *GlColorConfig) Apply() {
	r, g, b, a := conf.color.RGBA()
	rf, gf, bf, af := float32(r)/65535., float32(g)/65535., float32(b)/65535., float32(a)/65535.
	if conf.Program() == 0 {
		panic("Program is nil!")
	}

	loc := conf.Program().GetUniformLocation("color")
	loc.Uniform4f(rf, gf, bf, af)
}

func OpenGLInit() *OpenGL {
	g := &OpenGL{
		QuadraticDrawer: newQuadraticDrawer(),
		TriangleDrawer:  newTriangleDrawer(),
		FillDrawer:      newFillDrawer(),
	}
	g.QuadraticDrawConfig = &QuadraticDrawConfig{glProgramInfo{g.QuadraticDrawer.program}, false}
	g.QuadraticDrawer.AddConfig(g.QuadraticDrawConfig)
	g.GlColorConfig = &GlColorConfig{glProgramInfo{g.QuadraticDrawer.program}, color.RGBA{0, 0, 0, 1}}
	g.configs = append(g.configs, g.GlColorConfig)
	g.GlColorConfig.SetColor(Black)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Enable(gl.BLEND)
	gl.Enable(gl.STENCIL_TEST)
	return g
}

type Paint struct {
	fillColor color.Color
}

func NewPaint() *Paint {
	return new(Paint)
}

func (p *Paint) SetFill(color color.Color) *Paint {
	p.fillColor = color
	return p
}

type Config interface {
	Apply()
}

type drawConfig interface {
	Config
	Program() gl.Program
	SetProgram(gl.Program)
}

type glProgramInfo struct {
	program gl.Program
}

func (pi *glProgramInfo) SetProgram(p gl.Program) {
	pi.program = p
}

func (pi *glProgramInfo) Program() gl.Program {
	return pi.program
}

type Point struct {
	X float64
	Y float64
}

func (pt Point) Add(pt2 Point) Point {
	return ToPoint(pt.Mathgl().Add(pt2.Mathgl()))
}

func (pt Point) Mul(ratio float64) Point {
	return ToPoint(pt.Mathgl().Mul(ratio))
}

func (pt Point) Sub(pt2 Point) Point {
	vt := pt.Mathgl().Sub(pt2.Mathgl())
	return Point{vt[0], vt[1]}
}

//type TransFunc func(Point) Point

////DrawOp is Draw Operation
//type DrawOp struct {
//	Canvas    *Canvas
//	transform TransFunc //Transformation function
//}

//func defaultTransFunc(pt Point) Point { return pt }

//func NewDrawOp(canv *Canvas) *DrawOp {
//	return &DrawOp{canv, defaultTransFunc}
//}

//func (op *DrawOp) SetTransformationFunc(f TransFunc) {
//	op.transform = f
//}

//func (op *DrawOp) transformAll(pts []Point) []Point {
//	r := make([]Point, len(pts))
//	for i, pt := range pts {
//		r[i] = op.transform(pt)
//	}
//	return r
//}

func (p Point) Mathgl() mathgl.Vec2d {
	return mathgl.Vec2d{p.X, p.Y}
}

type Rectangle struct {
	Min Point
	Max Point
}

func (r Rectangle) Dx() float64 {
	return r.Max.X - r.Min.X
}

func (r Rectangle) Dy() float64 {
	return r.Max.Y - r.Min.Y
}

//GlDrawer represents an OpenGl program and shader, like an opengl draw mode
type GlDrawer struct {
	program gl.Program
	vao     gl.VertexArray
	vbo     gl.Buffer
	ebo     gl.Buffer

	configs []drawConfig
}

type QuadraticDrawConfig struct {
	glProgramInfo
	excludeTransluFrags bool //Exclude translucent (alpha != 0 && alpha != 1) fragments
}

func (conf *QuadraticDrawConfig) SetExcludeTransluFrags(v bool) {
	conf.excludeTransluFrags = v
}

func (conf *QuadraticDrawConfig) Apply() {
	loc := conf.Program().GetUniformLocation("excludeTrans")
	if loc != -1 {
		if !conf.excludeTransluFrags {
			loc.Uniform1i(0)
		} else {
			loc.Uniform1i(1)
		}
	}
}

type Canvas struct {
	W, H   int
	buffer *glh.Framebuffer
}

func NewCanvas(w, h int) *Canvas {
	buffer := new(glh.Framebuffer)
	buffer.Texture = glh.NewTexture(w, h)
	buffer.Texture.Init()
	return &Canvas{w, h, buffer}
}

func lastPt(l []Point) Point {
	return l[len(l)-1]
}

func Pt(x, y float64) Point {
	return Point{x, y}
}

func iPt(x, y int) Point {
	return Point{float64(x), float64(y)}
}

func NewGlDrawer(vshader, fshader string) *GlDrawer {
	vao := gl.GenVertexArray()
	vao.Bind()

	vbo := gl.GenBuffer()
	vbo.Bind(gl.ARRAY_BUFFER)

	ebo := gl.GenBuffer()
	ebo.Bind(gl.ELEMENT_ARRAY_BUFFER)

	vsh := ShaderFromFile(gl.VERTEX_SHADER, vshader)
	fsh := ShaderFromFile(gl.FRAGMENT_SHADER, fshader)

	program := glh.NewProgram(vsh, fsh)
	program.BindFragDataLocation(0, "outColor")
	program.Use()

	return &GlDrawer{
		program: program,
		vao:     vao,
		vbo:     vbo,
		ebo:     ebo,
	}
}

func newQuadraticDrawer() *GlDrawer {
	dr := newTexDrawer("vshader.glsl", "quadratic_fshader.glsl")
	return dr
}

func newTexDrawer(vshader, fshader string) *GlDrawer {
	dr := NewGlDrawer(vshader, fshader)
	program := dr.program
	posAttr := program.GetAttribLocation("position")
	posAttr.AttribPointer(2, gl.FLOAT, false, 4*4, uintptr(0))
	posAttr.EnableArray()

	texAttr := program.GetAttribLocation("texcoord")
	texAttr.AttribPointer(2, gl.FLOAT, false, 4*4, uintptr(8))
	glh.OpenGLSentinel()
	texAttr.EnableArray()

	return dr
}

func newTriangleDrawer() *GlDrawer {
	dr := NewGlDrawer("vshader.glsl", "triangle_fshader.glsl")
	program := dr.program
	posAttr := program.GetAttribLocation("position")
	posAttr.AttribPointer(2, gl.FLOAT, false, 2*4, uintptr(0))
	posAttr.EnableArray()

	return dr
}

func newFillDrawer() *GlDrawer {
	return newTexDrawer("vshader.glsl", "fill_fshader.glsl")
}

func (dr *GlDrawer) AddConfig(conf drawConfig) {
	dr.configs = append(dr.configs, conf)
}

func (dr *GlDrawer) Activate() {
	dr.vao.Bind()
	dr.vbo.Bind(gl.ARRAY_BUFFER)
	dr.program.Use()
	for _, config := range dr.configs {
		config.Apply()
	}
}

type QuadraticCurve struct {
	points [3]Point
}

type BezierCurve struct {
	points [4]Point
	repr   *Path //Quadratics representation
}

type PathSegment interface {
	Draw(*Canvas)
	Points() []Point //May be unnecessary
}

type Path struct {
	Segs      *list.List
	endPoints *list.List
}

func NewPath() *Path {
	p := new(Path)
	p.Segs = new(list.List)
	return p
}

func (p *Path) EndPoint() Point {
	return p.endPoints.Back().Value.(Point)
}

func (p *Path) NewEnd(pt Point) {
	if p.endPoints == nil {
		p.endPoints = new(list.List)
	}
	p.endPoints.PushBack(pt)
}

func (p *Path) StartAt(pt Point) *Path {
	p.NewEnd(pt)
	return p
}

func (p *Path) QuadraticTo(p2, c Point) *Path {
	p.Segs.PushBack(NewQuadraticCurve(
		p.EndPoint(),
		c, p2))
	p.NewEnd(p2)
	return p
}

func (p *Path) BezierTo(p2, c1, c2 Point) *Path {
	p.Segs.PushBack(NewBezierCurve(
		p.EndPoint(),
		c1, c2, p2))
	p.NewEnd(p2)
	return p
}

func fill(canv *Canvas, alphaTex *glh.Texture, paint *Paint) {
	gGl.GlColorConfig.SetColor(paint.fillColor)
	defer gGl.GlColorConfig.Reset()
	gGl.Activate(gGl.FillDrawer)
	gl.ColorMask(true, true, true, true)
	gl.StencilMask(0x3)
	gl.StencilFunc(gl.LESS, 0, 0xff)
	w, h := canv.W, canv.H
	p := canv.toGLPoints([]Point{
		iPt(0, 0),
		iPt(w, 0),
		iPt(w, h),
		iPt(0, h),
	})
	vertices := []float32{
		p[0].X, p[0].Y, 0, 1,
		p[1].X, p[1].Y, 1, 1,
		p[2].X, p[2].Y, 1, 0,
		p[3].X, p[3].Y, 0, 0,
	}
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, vertices, gl.STATIC_DRAW)

	elements := []uint32{
		0, 1, 2,
		2, 3, 0,
	}
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(elements)*4, elements, gl.STATIC_DRAW)

	glh.With(alphaTex, func() {
		gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)
	})
}

func (p *Path) draw(canv *Canvas, alphaBuffer *glh.Framebuffer, clrStencil bool) {
	gGl.QuadraticDrawConfig.SetExcludeTransluFrags(false)
	glh.With(alphaBuffer, func() {
		gl.ClearColor(0, 0, 0, 0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		p.glDraw(canv)
	})
	if clrStencil {
		gl.StencilMask(0x3)
		gl.ClearStencil(0x0)
		gl.Clear(gl.STENCIL_BUFFER_BIT)
	}
	gl.ColorMask(false, false, false, false)
	gl.StencilMask(0x3)
	gl.StencilFunc(gl.ALWAYS, 0, 0xff)
	gl.StencilOp(gl.KEEP, gl.KEEP, gl.INVERT)
	p.glDraw(canv)
	gl.StencilMask(0x1)
	gGl.QuadraticDrawConfig.SetExcludeTransluFrags(true)
	p.glDraw(canv)
}

func (p *Path) DrawFill(canv *Canvas, paint *Paint) {
	alphaBuffer := canv.buffer
	p.draw(canv, alphaBuffer, true)
	fill(canv, alphaBuffer.Texture, paint)
}

func GetPt(pt *list.Element) Point {
	return pt.Value.(Point)
}

func (p *Path) glDraw(canv *Canvas) {
	for e := p.Segs.Front(); e != nil; e = e.Next() {
		e.Value.(PathSegment).Draw(canv)
	}

	if GetPt(p.endPoints.Back()).Mathgl().ApproxEqual(
		GetPt(p.endPoints.Front()).Mathgl()) {
		p.endPoints.Remove(p.endPoints.Back())
	}

	if p.endPoints.Len() < 3 {
		return
	}

	gGl.Activate(gGl.TriangleDrawer)

	gl.BlendFunc(gl.ONE_MINUS_DST_ALPHA, gl.ZERO)
	defer gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	pa := make(p2t.PointArray, 0, p.endPoints.Len())
	for e := p.endPoints.Front(); e != nil; e = e.Next() {
		p := GetPt(e)
		pa = append(pa, &p2t.Point{X: float64(p.X), Y: float64(p.Y)})
	}
	p2t.Init(pa)
	triArr := p2t.Triangulate()
	vertices := make([]float32, 6, 6)
	for _, tri := range triArr {
		for i, triPt := range tri.Point {
			pt := canv.toGLPoint(Pt(triPt.X, triPt.Y))
			vertices[i*2] = pt.X
			vertices[i*2+1] = pt.Y
		}
		gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, vertices, gl.STATIC_DRAW)
		gl.DrawArrays(gl.TRIANGLES, 0, 3)
	}
}

func NewQuadraticCurve(p1, c, p2 Point) *QuadraticCurve {
	return &QuadraticCurve{
		points: [3]Point{p1, c, p2},
	}
}

func ToPoint(v mathgl.Vec2d) Point {
	return Point{v[0], v[1]}
}

func makeQuadraticCurve(p1, c, p2 mathgl.Vec2d) QuadraticCurve {
	return QuadraticCurve{
		points: [3]Point{
			ToPoint(p1),
			ToPoint(c),
			ToPoint(p2),
		},
	}
}

func NewBezierCurve(p1, c1, c2, p2 Point) (bc *BezierCurve) {
	bc = &BezierCurve{
		points: [4]Point{p1, c1, c2, p2},
	}
	quads := bc.ToQuadratics()
	if len(quads) < 1 {
		panic("Something's wrong.")
	}
	path := NewPath().StartAt(quads[0].points[0])
	for _, quadc := range quads {
		path.QuadraticTo(quadc.points[2], quadc.points[1])
	}

	bc.repr = path
	return
}

type GLPoint struct {
	X, Y float32
}

func (canv *Canvas) toGLPoint(p Point) GLPoint {
	return GLPoint{float32(p.X) / float32(canv.W), float32(p.Y) / float32(canv.H)}
}

func (canv *Canvas) toGLPoints(points []Point) []GLPoint {
	ps := make([]GLPoint, len(points))
	for i, p := range points {
		ps[i] = canv.toGLPoint(p)
	}
	return ps
}

func (c *QuadraticCurve) draw(canv *Canvas) {
	p := canv.toGLPoints(c.Points())
	vertices := []float32{
		p[0].X, p[0].Y, 0.0, 0.0,
		p[1].X, p[1].Y, 0.5, 0.0,
		p[2].X, p[2].Y, 1.0, 1.0,
	}
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, vertices, gl.STATIC_DRAW)
	gl.DrawArrays(gl.TRIANGLES, 0, 3)
}

func (c *QuadraticCurve) Draw(canv *Canvas) {
	gGl.Activate(gGl.QuadraticDrawer)
	c.draw(canv)
}

func (c *QuadraticCurve) Points() []Point {
	return c.points[:]
}

func (c *QuadraticCurve) SetPoints(pts [3]Point) {
	c.points = pts
}

func Vector(p Point) mathgl.Vec2d {
	return mathgl.Vec2d{float64(p.X), float64(p.Y)}
}

func (c *BezierCurve) quadApprox(p1, c1, c2, p2 mathgl.Vec2d) (v mathgl.Vec2d, ok bool) {
	//P2 - 3·C2 + 3·C1 - P1
	d01 := p2.Sub(c2.Mul(3)).Add(c1.Mul(3)).Sub(p1).Len() / 2
	if d01 <= gQuadraticApproxPrecision {
		// (3·C2 - P2 + 3·C1 - P1)/4
		return c2.Mul(3).Sub(p2).Add(c1.Mul(3)).Sub(p1).Mul(1 / 4.), true
	}
	return v, false
}

func mid(v1 mathgl.Vec2d, v2 mathgl.Vec2d) mathgl.Vec2d {
	return v1.Add(v2).Mul(1 / 2.)
}

func (c *BezierCurve) toQuadratics(p1, c1, c2, p2 mathgl.Vec2d) []QuadraticCurve {
	if newcp, ok := c.quadApprox(p1, c1, c2, p2); ok {
		return []QuadraticCurve{makeQuadraticCurve(p1, newcp, p2)}
	}

	p4, p6 := mid(p1, c1), mid(p2, c2)
	p5 := mid(c1, c2)
	p7, p8 := mid(p4, p5), mid(p5, p6)
	p9 := mid(p7, p8)

	return append(c.toQuadratics(p1, p4, p7, p9), c.toQuadratics(p9, p8, p6, p2)...)
}

//ToQuadratics approximates a cubic bezier curve with quadratics.
//Algorithm by Adrian Colomitchi at
//http://www.caffeineowl.com/graphics/2d/vectorial/cubic2quad01.html
func (c *BezierCurve) ToQuadratics() []QuadraticCurve {
	p1, c1 := Vector(c.points[0]), Vector(c.points[1])
	c2, p2 := Vector(c.points[2]), Vector(c.points[3])
	return c.toQuadratics(p1, c1, c2, p2)
}

func (c *BezierCurve) Points() (l []Point) {
	l = make([]Point, 0)
	p := c.repr
	for e := p.Segs.Front(); e != nil; e = e.Next() {
		l = append(l, e.Value.(*QuadraticCurve).Points()...)
	}
	return
}

func (c *BezierCurve) Draw(canv *Canvas) {
	gGl.Activate(gGl.QuadraticDrawer)
	c.repr.glDraw(canv)
}

func ShaderFromFile(stype gl.GLenum, filename string) (shader glh.Shader) {
	_, f, _, _ := runtime.Caller(0)
	dir := path.Dir(f)
	fcont, _ := ioutil.ReadFile(path.Join(dir, filename))
	shader = glh.Shader{stype, string(fcont[:])}
	shader.Compile()
	return shader
}
