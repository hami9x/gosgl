package gosgl

import (
	"testing"

	chk "launchpad.net/gocheck"
)

func Test(t *testing.T) { chk.TestingT(t) }

type MySuite struct{}

var _ = chk.Suite(&MySuite{})

func (s *MySuite) TestBezierToQuadratic(c *chk.C) {
	c.Check(len(MakeBezierCurve(0, 0, 100, 100, 200, 100, 300, 0).ToQuadratics()),
		chk.Equals, 1)
	c.Check(len(MakeBezierCurve(0, 0, 100, 100, 200, -100, 300, 0).ToQuadratics()),
		chk.Equals, 8)
}
