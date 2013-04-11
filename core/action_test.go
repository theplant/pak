package core

import (
	"bytes"
	"github.com/theplant/pak/gitpkg"
	. "github.com/theplant/pak/share"
	. "launchpad.net/gocheck"
	"os/exec"
)

type GetSuite struct {
	pakPkgs           []PakPkg
	originalGoGetImpl func(name string) error
}

var _ = Suite(&GetSuite{})

func mustRun(params ...string) (cmd *exec.Cmd) {
	cmd = exec.Command(params[0], params[1:]...)

	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		fl(stderr.String())

		panic(err)
	}

	return
}

func (s *GetSuite) SetUpSuite(c *C) {
	s.originalGoGetImpl = gitpkg.GoGetImpl
	gitpkg.GoGetImpl = func(name string) error { return nil }
}

func (s *GetSuite) TearDownSuite(c *C) {
	gitpkg.GoGetImpl = s.originalGoGetImpl
}

func (s *GetSuite) SetUpTest(c *C) {
	mustRun("git", "clone", "fixtures/package1", "../../package1")
	mustRun("git", "clone", "fixtures/package2", "../../package2")
	mustRun("sh", "-c", "cd ../../package1 && git checkout -b pak && git checkout master && git tag _pak_latest_")
	mustRun("git", "clone", "fixtures/package3", "../../package3")
	mustRun("cp", "fixtures/Pakfile3", "Pakfile")

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
	mustRun("rm", "-rf", "../../package1")
	mustRun("rm", "-rf", "../../package2")
	mustRun("rm", "-rf", "../../package3")
	mustRun("rm", "-rf", "Pakfile")
	mustRun("rm", "-rf", "Pakfile.lock")
}

func (s *GetSuite) TestGet(c *C) {
	err := Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	s.pakPkgs[0].Sync()
	s.pakPkgs[1].Sync()
	s.pakPkgs[2].Sync()
	c.Check(s.pakPkgs[0].HeadRefsName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[1].HeadRefsName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[2].HeadRefsName, Equals, "refs/heads/pak")
}

func (s *GetSuite) TestCanGetPackageWithoutRemoteBranch(c *C) {
	mustRun("git", "--git-dir=../../package1/.git", "--work-tree=../../package1", "branch", "-dr", "origin/dev")
	mustRun("git", "--git-dir=../../package1/.git", "--work-tree=../../package1", "branch", "-dr", "origin/master")

	err := Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	s.pakPkgs[0].Sync()
	s.pakPkgs[1].Sync()
	s.pakPkgs[2].Sync()
	c.Check(s.pakPkgs[0].HeadRefsName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[1].HeadRefsName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[2].HeadRefsName, Equals, "refs/heads/pak")
}

func (s *GetSuite) TestGetWithPakMeter(c *C) {
	err := Get(PakOption{
		PakMeter:       []string{"github.com/theplant/package2"},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	s.pakPkgs[0].Sync()
	s.pakPkgs[1].Sync()
	s.pakPkgs[2].Sync()
	c.Check(s.pakPkgs[0].HeadRefsName, Equals, "refs/heads/master")
	c.Check(s.pakPkgs[1].HeadRefsName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[2].HeadRefsName, Equals, "refs/heads/master")
}

func (s *GetSuite) TestPaklockInfoShouldUpdateAfterGet(c *C) {
	expectedPaklockInfo := PaklockInfo{
		"github.com/theplant/package3": "d5f51ca77f5d4f37a8105a74b67d2f1aefea939c",
		"github.com/theplant/package1": "11b174bd5acbf990687e6b068c97378d3219de04",
		"github.com/theplant/package2": "941af3b182a1d0a5859fd451a8b5a633f479d7bc",
	}

	err := Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: false,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	paklockInfo, _ := GetPaklockInfo()
	c.Check(len(paklockInfo), Equals, 3)
	c.Check(paklockInfo, DeepEquals, expectedPaklockInfo)

	err = Get(PakOption{
		PakMeter:       []string{"github.com/theplant/package2"},
		UsePakfileLock: false,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	paklockInfo, _ = GetPaklockInfo()
	c.Check(len(paklockInfo), Equals, 3)
	c.Check(paklockInfo, DeepEquals, expectedPaklockInfo)
}

func (s *GetSuite) TestGoGetAndFetchBeforePak(c *C) {
	originalGoGetImpl := gitpkg.GoGetImpl
	gogetPkg1Count := 0
	gitpkg.GoGetImpl = func(name string) error {
		if name == "github.com/theplant/package1" {
			gogetPkg1Count++
			if gogetPkg1Count == 1 {
				mustRun("git", "clone", "fixtures/package1", "../../package1")
			}
		}

		return nil
	}
	defer func() { gitpkg.GoGetImpl = originalGoGetImpl }()

	mustRun("rm", "-rf", "../../package1")

	err := Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	c.Check(gogetPkg1Count, Equals, 2)

	s.pakPkgs[0].Sync()
	s.pakPkgs[1].Sync()
	s.pakPkgs[2].Sync()
	c.Check(s.pakPkgs[0].HeadRefsName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[1].HeadRefsName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[2].HeadRefsName, Equals, "refs/heads/pak")
}

func (s *GetSuite) TestDoNotCheckOtherPkgWhenGettingWithPakMeter(c *C) {
	mustRun("rm", "-rf", "../../package1")

	err := Get(PakOption{
		PakMeter:       []string{"github.com/theplant/package2"},
		UsePakfileLock: false,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	s.pakPkgs[1].Sync()
	c.Check(s.pakPkgs[1].HeadRefsName, Equals, "refs/heads/pak")
}

func (s *GetSuite) TestComplainUncompilablePartialMatching(c *C) {
	err := Get(PakOption{
		PakMeter:       []string{"...***"},
		UsePakfileLock: false,
		Force:          false,
	})
	c.Check(err, Not(Equals), nil)
}

func (s *GetSuite) TestGetWithUpdatedPakfile(c *C) {
	err := Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	s.pakPkgs[0].Sync()
	s.pakPkgs[1].Sync()
	s.pakPkgs[2].Sync()
	c.Check(s.pakPkgs[0].HeadRefsName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[1].HeadRefsName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[2].HeadRefsName, Equals, "refs/heads/pak")

	c.Check(s.pakPkgs[1].PakbranchChecksum, Equals, "941af3b182a1d0a5859fd451a8b5a633f479d7bc")

	paklockInfo, _ := GetPaklockInfo()
	c.Check(len(paklockInfo), Equals, 3)

	mustRun("rm", "-rf", "Pakfile")
	mustRun("cp", "fixtures/Pakfile3-updated", "Pakfile")
	// mustRun("rm", "-rf", "Pakfile.lock")

	err = Get(PakOption{
		PakMeter:       []string{"github.com/theplant/package2"},
		UsePakfileLock: false,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	s.pakPkgs[1].Sync()
	c.Check(s.pakPkgs[1].PakbranchChecksum, Equals, "e373579a64e367338ff09b5143e312c81204c074")

	err = Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	s.pakPkgs[0].Sync()
	s.pakPkgs[1].Sync()
	c.Check(s.pakPkgs[0].HeadRefsName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[1].HeadRefsName, Equals, "refs/heads/pak")

	c.Check(s.pakPkgs[1].PakbranchChecksum, Equals, "e373579a64e367338ff09b5143e312c81204c074")

	paklockInfo, _ = GetPaklockInfo()
	c.Check(len(paklockInfo), Equals, 2)
}
