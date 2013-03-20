package pak

import (
	. "launchpad.net/gocheck"
	"os/exec"
	"github.com/theplant/pak/gitpkg"
	. "github.com/theplant/pak/share"
)

type GetSuite struct{}

var _ = Suite(&GetSuite{})

func mustRun(cmd *exec.Cmd) {
	err := cmd.Run()
	if err != nil {
	    panic(err)
	}
}

func (s *GetSuite) SetUpTest(c *C) {
	mustRun(exec.Command("git", "clone", "fixtures/package1", "../../package1"))
	mustRun(exec.Command("git", "clone", "fixtures/package2", "../../package2"))
	mustRun(exec.Command("git", "clone", "fixtures/package3", "../../package3"))
	mustRun(exec.Command("cp", "fixtures/Pakfile3", "Pakfile"))
}

func (s *GetSuite) TearDownTest(c *C) {
	mustRun(exec.Command("rm", "-rf", "../../package1"))
	mustRun(exec.Command("rm", "-rf", "../../package2"))
	mustRun(exec.Command("rm", "-rf", "../../package3"))
	mustRun(exec.Command("rm", "-rf", "Pakfile"))
	mustRun(exec.Command("rm", "-rf", "Pakfile.lock"))
}

func (s *GetSuite) TestGet(c *C) {
	pakPkgs := []PakPkg{}
	pkgsFixtures := [][]string{
		{"github.com/theplant/package1", "origin", "master"},
		{"github.com/theplant/package2", "origin", "dev"},
		{"github.com/theplant/package3", "origin", "master"},
	}
	for _, val := range pkgsFixtures {
		pakPkgs = append(pakPkgs, PakPkg{GitPkg: gitpkg.NewGitPkg(val[0], val[1], val[2])})
	}

	err := Get(PakOption{[]string{}, true, true, true})
	c.Check(err, Equals, nil)

	pakPkgs[0].Sync()
	pakPkgs[1].Sync()
	pakPkgs[2].Sync()
	c.Check(pakPkgs[0].HeadRefsName, Equals, "refs/heads/pak")
	c.Check(pakPkgs[1].HeadRefsName, Equals, "refs/heads/pak")
	c.Check(pakPkgs[2].HeadRefsName, Equals, "refs/heads/pak")
}

func (s *GetSuite) TestGetWithPakMeter(c *C) {
	pakPkgs := []PakPkg{}
	pkgsFixtures := [][]string{
		{"github.com/theplant/package1", "origin", "master"},
		{"github.com/theplant/package2", "origin", "dev"},
		{"github.com/theplant/package3", "origin", "master"},
	}
	for _, val := range pkgsFixtures {
		pakPkgs = append(pakPkgs, PakPkg{GitPkg: gitpkg.NewGitPkg(val[0], val[1], val[2])})
	}

	err := Get(PakOption{[]string{"github.com/theplant/package2"}, true, true, true})
	c.Check(err, Equals, nil)

	pakPkgs[0].Sync()
	pakPkgs[1].Sync()
	pakPkgs[2].Sync()
	c.Check(pakPkgs[0].HeadRefsName, Equals, "refs/heads/master")
	c.Check(pakPkgs[1].HeadRefsName, Equals, "refs/heads/pak")
	c.Check(pakPkgs[2].HeadRefsName, Equals, "refs/heads/master")
}
