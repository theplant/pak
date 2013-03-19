package gitpkg

import (
	. "launchpad.net/gocheck"
	"testing"
	"fmt"
)

// Hook up gocheck into the gotest runner.
func Test(t *testing.T) { TestingT(t) }

var fp = fmt.Printf
var fl = fmt.Println
