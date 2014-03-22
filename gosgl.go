/*Package gosgl is package gosgl
 */
package gosgl

import (
	"container/list"
	"io/ioutil"
	"math"
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
	gQuadraticDrawer          *Drawer
	gTriangleDrawer           *Drawer
	gFillDrawer               *Drawer
	gQuadraticApproxPrecision float64 = 10
)

type DrawConfig interface {
	Apply(*Drawer)
}

type Point struct {
	X float64
	Y float64
}

func (pt Point) Add(pt2 Point) Point {
	return ToPoint(pt.Mathgl().Add(pt2.Mathgl()))
}

func (pt Point) Sub(pt2 Point) Point {
	vt := pt.Mathgl().Sub(pt2.Mathgl())
	return Point{vt[0], vt[1]}
}

type TransFunc func(Point) Point

//DrawOp is Draw Operation
type DrawOp struct {
	Canvas    *Canvas
	transform TransFunc //Transformation function
}

func defaultTransFunc(pt Point) Point { return pt }

func NewDrawOp(canv *Canvas) *DrawOp {
	return &DrawOp{canv, defaultTransFunc}
}

func (op *DrawOp) SetTransformationFunc(f TransFunc) {
	op.transform = f
}

func (op *DrawOp) transformAll(pts []Point) []Point {
	r := make([]Point, len(pts))
	for i, pt := range pts {
		r[i] = op.transform(pt)
	}
	return r
}

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

//Drawer represents a program and its buffers
type Drawer struct {
	program gl.Program
	vao     gl.VertexArray
	vbo     gl.Buffer
	ebo     gl.Buffer

	configs DrawConfig
}

type QuadraticDrawConfig struct {
	excludeTrans bool
}

func (conf *QuadraticDrawConfig) Apply(dr *Drawer) {
	loc := dr.program.GetUniformLocation("excludeTrans")
	if loc != -1 {
		if !conf.excludeTrans {
			loc.Uniform1i(0)
		} else {
			loc.Uniform1i(1)
		}
	}
}

type Canvas struct {
	W, H int
}

func MakeCanvas(w, h int) *Canvas {
	return &Canvas{w, h}
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

func NewDrawer(vshader, fshader string) *Drawer {
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

	return &Drawer{
		program: program,
		vao:     vao,
		vbo:     vbo,
		ebo:     ebo,
	}
}

func newQuadraticDrawer() *Drawer {
	dr := newTexDrawer("vshader.glsl", "quadratic_fshader.glsl")

	dr.configs = &QuadraticDrawConfig{false}
	return dr
}

func newTexDrawer(vshader, fshader string) *Drawer {
	dr := NewDrawer(vshader, fshader)
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

func newTriangleDrawer() *Drawer {
	dr := NewDrawer("vshader.glsl", "triangle_fshader.glsl")
	program := dr.program
	posAttr := program.GetAttribLocation("position")
	posAttr.AttribPointer(2, gl.FLOAT, false, 2*4, uintptr(0))
	posAttr.EnableArray()

	return dr
}

func newFillDrawer() *Drawer {
	return newTexDrawer("vshader.glsl", "fill_fshader.glsl")
}

func (dr *Drawer) activate() {
	dr.vao.Bind()
	dr.vbo.Bind(gl.ARRAY_BUFFER)
	dr.program.Use()
}

func Init() {
	gQuadraticDrawer = newQuadraticDrawer()
	gTriangleDrawer = newTriangleDrawer()
	gFillDrawer = newFillDrawer()

	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Enable(gl.BLEND)
	gl.Enable(gl.STENCIL_TEST)
}

type QuadraticCurve struct {
	points [3]Point
}

type BezierCurve struct {
	points [4]Point
	repr   *Path //Quadratics representation
}

type PathSegment interface {
	Draw(*DrawOp)
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

func fill(op *DrawOp, alphaTex *glh.Texture) {
	gFillDrawer.activate()
	gl.ColorMask(true, true, true, true)
	gl.StencilMask(0x3)
	gl.StencilFunc(gl.LESS, 0, 0xff)
	w, h := op.Canvas.W, op.Canvas.H
	p := op.Canvas.toGLPoints([]Point{
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

func (p *Path) draw(op *DrawOp) {
	alphaBuffer := new(glh.Framebuffer)
	alphaBuffer.Texture = glh.NewTexture(op.Canvas.W, op.Canvas.H)
	alphaBuffer.Texture.Init()
	glh.With(alphaBuffer, func() {
		p.glDraw(op)
	})
	gl.ColorMask(false, false, false, false)
	quadConf := gQuadraticDrawer.configs.(*QuadraticDrawConfig)
	gl.ClearStencil(0)
	gl.Clear(gl.STENCIL_BUFFER_BIT)
	gl.StencilMask(0x3)
	gl.StencilFunc(gl.ALWAYS, 0, 0xff)
	gl.StencilOp(gl.KEEP, gl.KEEP, gl.INVERT)
	quadConf.excludeTrans = false
	p.glDraw(op)
	gl.StencilMask(0x1)
	quadConf.excludeTrans = true
	p.glDraw(op)

	fill(op, alphaBuffer.Texture)
}

func (p *Path) DrawFill(canv *Canvas) {
	op := NewDrawOp(canv)
	op.SetTransformationFunc(defaultTransFunc)
	p.draw(op)
}

func (p *Path) DrawStroke(canv *Canvas) {
	op := NewDrawOp(canv)
	op.SetTransformationFunc(func(pt Point) Point {
		bb := p.BoundingBox()
		pivot := bb.Min.Add(Pt(bb.Dx()/2, bb.Dy()/2))
		strokeWidth := 30
		v := pt.Sub(pivot)
		k := float64(strokeWidth) / math.Sqrt(math.Pow(float64(v.X), 2)+math.Pow(float64(v.Y), 2))
		offset := Point{k * float64(v.X), k * float64(v.Y)}
		pt = pt.Add(offset)
		return pt
		//return Pt(pt.X*2, pt.Y*2)
	})
	p.draw(op)
}

func (p *Path) BoundingBox() Rectangle {
	var minX, minY, maxX, maxY float64 = oo, oo, 0, 0
	for e := p.Segs.Front(); e != nil; e = e.Next() {
		for _, pt := range e.Value.(PathSegment).Points() {
			minX = math.Min(minX, pt.X)
			minY = math.Min(minY, pt.Y)
			maxX = math.Max(maxX, pt.X)
			maxY = math.Max(maxY, pt.Y)
		}
	}

	return Rectangle{Pt(minX, minY), Pt(maxX, maxY)}
}

func GetPt(pt *list.Element) Point {
	return pt.Value.(Point)
}

func (p *Path) glDraw(op *DrawOp) {
	for e := p.Segs.Front(); e != nil; e = e.Next() {
		e.Value.(PathSegment).Draw(op)
	}

	if GetPt(p.endPoints.Back()).Mathgl().ApproxEqual(
		GetPt(p.endPoints.Front()).Mathgl()) {
		p.endPoints.Remove(p.endPoints.Back())
	}

	if p.endPoints.Len() < 3 {
		return
	}

	gTriangleDrawer.activate()

	gl.BlendFunc(gl.ONE_MINUS_DST_ALPHA, gl.ZERO)
	defer gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	pa := make(p2t.PointArray, 0, p.endPoints.Len())
	for e := p.endPoints.Front(); e != nil; e = e.Next() {
		p := op.transform(GetPt(e))
		pa = append(pa, &p2t.Point{X: float64(p.X), Y: float64(p.Y)})
	}
	p2t.Init(pa)
	triArr := p2t.Triangulate()
	vertices := make([]float32, 6, 6)
	for _, tri := range triArr {
		for i, triPt := range tri.Point {
			pt := op.Canvas.toGLPoint(Pt(triPt.X, triPt.Y))
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

func (c *QuadraticCurve) draw(op *DrawOp) {
	p := op.Canvas.toGLPoints(op.transformAll(c.Points()))
	//fmt.Printf("%v\n", op.transformAll(c.Points()))
	vertices := []float32{
		p[0].X, p[0].Y, 0.0, 0.0,
		p[1].X, p[1].Y, 0.5, 0.0,
		p[2].X, p[2].Y, 1.0, 1.0,
	}
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, vertices, gl.STATIC_DRAW)
	gl.DrawArrays(gl.TRIANGLES, 0, 3)
}

func (c *QuadraticCurve) Draw(op *DrawOp) {
	gQuadraticDrawer.activate()
	gQuadraticDrawer.configs.Apply(gQuadraticDrawer)
	c.draw(op)
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

func (c *BezierCurve) Draw(op *DrawOp) {
	gQuadraticDrawer.activate()
	c.repr.glDraw(op)
}

func ShaderFromFile(stype gl.GLenum, filename string) (shader glh.Shader) {
	_, f, _, _ := runtime.Caller(0)
	dir := path.Dir(f)
	fcont, _ := ioutil.ReadFile(path.Join(dir, filename))
	shader = glh.Shader{stype, string(fcont[:])}
	shader.Compile()
	return shader
}
