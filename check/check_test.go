package check

import (
	"fmt"
	. "launchpad.net/gocheck"
	"testing"
)

// Hook up gocheck into the gotest runner.
func Test(t *testing.T) { TestingT(t) }

var fp = fmt.Printf
var fl = fmt.Println

type CheckSuite struct{}

var _ = Suite(&CheckSuite{})

func (s *CheckSuite) TestParsePkgName(c *C) {
	name, count, _, _ := parsePkg()

	c.Check(name, Equals, "check")
	c.Check(count, Equals, 2)
}
