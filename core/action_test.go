package core

import (
	"github.com/theplant/pak/gitpkg"
	. "github.com/theplant/pak/share"
	. "launchpad.net/gocheck"
	"os/exec"
	"bytes"
)

type GetSuite struct{
	pakPkgs []PakPkg
}

var _ = Suite(&GetSuite{})

func mustRun(cmd *exec.Cmd) {
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		fp(stderr.String())
		panic(err)
	}
}

func (s *GetSuite) SetUpTest(c *C) {
	mustRun(exec.Command("git", "clone", "fixtures/package1", "../../package1"))
	mustRun(exec.Command("git", "clone", "fixtures/package2", "../../package2"))
	mustRun(exec.Command("sh", "-c", "cd ../../package1 && git checkout -b pak && git checkout master && git tag _pak_latest_"))
	mustRun(exec.Command("git", "clone", "fixtures/package3", "../../package3"))
	mustRun(exec.Command("cp", "fixtures/Pakfile3", "Pakfile"))

	s.pakPkgs = []PakPkg{}
	pkgsFixtures := [][]string{
		{"github.com/theplant/package1", "origin", "master"},
		{"github.com/theplant/package2", "origin", "dev"},
		{"github.com/theplant/package3", "origin", "master"},
	}
	for _, val := range pkgsFixtures {
		s.pakPkgs = append(s.pakPkgs, PakPkg{GitPkg: gitpkg.NewGitPkg(val[0], val[1], val[2])})
	}
}

func (s *GetSuite) TearDownTest(c *C) {
	mustRun(exec.Command("rm", "-rf", "../../package1"))
	mustRun(exec.Command("rm", "-rf", "../../package2"))
	mustRun(exec.Command("rm", "-rf", "../../package3"))
	mustRun(exec.Command("rm", "-rf", "Pakfile"))
	mustRun(exec.Command("rm", "-rf", "Pakfile.lock"))
}

func (s *GetSuite) TestGet(c *C) {
	err := Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Fetch:          true,
		Force:          false,
		NotGet:         true,
	})
	c.Check(err, Equals, nil)

	s.pakPkgs[0].Sync()
	s.pakPkgs[1].Sync()
	s.pakPkgs[2].Sync()
	c.Check(s.pakPkgs[0].HeadRefsName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[1].HeadRefsName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[2].HeadRefsName, Equals, "refs/heads/pak")
}

// func (s *GetSuite) TestCanGetPackageWithoutRemoteBranch(c *C) {
// 	pakPkgs[1].
//
// 	err := Get(PakOption{
// 		PakMeter:       []string{},
// 		UsePakfileLock: true,
// 		Fetch:          true,
// 		Force:          false,
// 		NotGet:         true,
// 	})
// 	c.Check(err, Equals, nil)
//
// 	pakPkgs[0].Sync()
// 	pakPkgs[1].Sync()
// 	pakPkgs[2].Sync()
// 	c.Check(pakPkgs[0].HeadRefsName, Equals, "refs/heads/pak")
// 	c.Check(pakPkgs[1].HeadRefsName, Equals, "refs/heads/pak")
// 	c.Check(pakPkgs[2].HeadRefsName, Equals, "refs/heads/pak")
// }

func (s *GetSuite) TestGetWithPakMeter(c *C) {
	err := Get(PakOption{
		PakMeter:       []string{"github.com/theplant/package2"},
		UsePakfileLock: true,
		Fetch:          true,
		Force:          false,
		NotGet:         true,
	})
	c.Check(err, Equals, nil)

	s.pakPkgs[0].Sync()
	s.pakPkgs[1].Sync()
	s.pakPkgs[2].Sync()
	c.Check(s.pakPkgs[0].HeadRefsName, Equals, "refs/heads/master")
	c.Check(s.pakPkgs[1].HeadRefsName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[2].HeadRefsName, Equals, "refs/heads/master")
}
