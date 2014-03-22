package gosgl

import (
	"testing"

	chk "launchpad.net/gocheck"
)

func Test(t *testing.T) { chk.TestingT(t) }

type MySuite struct{}

var _ = chk.Suite(&MySuite{})

func (s *MySuite) TestBezierToQuadratic(c *chk.C) {
	c.Check(len(NewBezierCurve(Pt(0, 0), Pt(100, 100), Pt(200, 100), Pt(300, 0)).ToQuadratics()),
		chk.Equals, 1)
	c.Check(len(NewBezierCurve(Pt(0, 0), Pt(100, 100), Pt(200, -100), Pt(300, 0)).ToQuadratics()),
		chk.Equals, 4)
}

func (s *MySuite) TestTransformation(c *chk.C) {
	op := NewDrawOp(MakeCanvas(100, 100))
	op.SetTransformationFunc(func(pt Point) Point {
		return Pt(pt.X*2, pt.Y*2)
	})
	c.Check(op.transform(Pt(10, 10)), chk.Equals, Pt(20, 20))
	c.Check(op.transformAll([]Point{Pt(10, 10), Pt(5, 5)}), chk.DeepEquals, []Point{Pt(20, 20), Pt(10, 10)})
}
