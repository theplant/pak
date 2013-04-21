package core

import (
	"bytes"
	// "github.com/theplant/pak/gitpkg"
	. "github.com/theplant/pak/share"
	. "launchpad.net/gocheck"
	"os/exec"
)

type GetSuite struct {
	pakPkgs           []*PakPkg
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
	s.originalGoGetImpl = GoGetImpl
	GoGetImpl = func(name string) error { return nil }
}

func (s *GetSuite) TearDownSuite(c *C) {
	GoGetImpl = s.originalGoGetImpl
}

func (s *GetSuite) SetUpTest(c *C) {
	mustRun("git", "clone", "fixtures/package1", "../../package1")
	mustRun("git", "clone", "fixtures/package2", "../../package2")
	mustRun("sh", "-c", "cd ../../package1 && git checkout -b pak && git checkout master && git tag _pak_latest_")
	mustRun("git", "clone", "fixtures/package3", "../../package3")
	mustRun("cp", "fixtures/Pakfile3", "Pakfile")

	s.pakPkgs = []*PakPkg{}
	pkgsFixtures := [][]string{
		{"github.com/theplant/package1", "origin", "master"},
		{"github.com/theplant/package2", "origin", "dev"},
		{"github.com/theplant/package3", "origin", "master"},
	}
	for _, vals := range pkgsFixtures {
		pkg := NewPakPkg(vals[0], vals[1], vals[2])
		err := pkg.Dial()
		if err != nil {
			c.Fatal(err, Equals, nil)
		}
		s.pakPkgs = append(s.pakPkgs, &pkg)
	}
}

func (s *GetSuite) TearDownTest(c *C) {
	mustRun("rm", "-rf", "../../package1")
	mustRun("rm", "-rf", "../../package2")
	mustRun("rm", "-rf", "../../package3")
	mustRun("rm", "-rf", "Pakfile")
	mustRun("rm", "-rf", "Pakfile.lock")
}

func (s *GetSuite) TestSimpleGet(c *C) {
	err := Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	s.pakPkgs[0].Sync()
	s.pakPkgs[1].Sync()
	s.pakPkgs[2].Sync()
	c.Check(s.pakPkgs[0].HeadRefName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[1].HeadRefName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[2].HeadRefName, Equals, "refs/heads/pak")
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
	c.Check(s.pakPkgs[0].HeadRefName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[1].HeadRefName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[2].HeadRefName, Equals, "refs/heads/pak")
}

func (s *GetSuite) TestGetWithPakMeter(c *C) {
	err := Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	mustRun("git", "--git-dir=../../package1/.git", "--work-tree=../../package1", "checkout", "master")
	mustRun("git", "--git-dir=../../package3/.git", "--work-tree=../../package3", "checkout", "master")

	err = Get(PakOption{
		PakMeter:       []string{"github.com/theplant/package2"},
		UsePakfileLock: false,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	s.pakPkgs[0].Sync()
	s.pakPkgs[1].Sync()
	s.pakPkgs[2].Sync()
	c.Check(s.pakPkgs[0].HeadRefName, Equals, "refs/heads/master")
	c.Check(s.pakPkgs[1].HeadRefName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[2].HeadRefName, Equals, "refs/heads/master")
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

func (s *GetSuite) TestGoGetBeforePak(c *C) {
	originalGoGetImpl := GoGetImpl
	gogetPkg1Count := 0
	GoGetImpl = func(name string) error {
		if name == "github.com/theplant/package1" {
			gogetPkg1Count++
			if gogetPkg1Count == 1 {
				mustRun("git", "clone", "fixtures/package1", "../../package1")
			}
		}

		return nil
	}
	defer func() { GoGetImpl = originalGoGetImpl }()

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
	c.Check(s.pakPkgs[0].HeadRefName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[1].HeadRefName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[2].HeadRefName, Equals, "refs/heads/pak")
}

func (s *GetSuite) TestGoGetWhenPkgIsOnPakBranch(c *C) {
	originalGoGetImpl := GoGetImpl
	gogetPkg1Count := 0
	GoGetImpl = func(name string) error {
		if name == "github.com/theplant/package1" {
			gogetPkg1Count++
		}

		return nil
	}
	defer func() { GoGetImpl = originalGoGetImpl }()

	err := Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Equals, nil)
	err = Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	c.Check(gogetPkg1Count, Equals, 2)

	s.pakPkgs[0].Sync()
	s.pakPkgs[1].Sync()
	s.pakPkgs[2].Sync()
	c.Check(s.pakPkgs[0].HeadRefName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[1].HeadRefName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[2].HeadRefName, Equals, "refs/heads/pak")
}

func (s *GetSuite) TestDoNotCheckOtherPkgsWhenGettingWithPakMeter(c *C) {
	err := Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	mustRun("rm", "-rf", "../../package1")

	err = Get(PakOption{
		PakMeter:       []string{"github.com/theplant/package2"},
		UsePakfileLock: false,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	s.pakPkgs[0].Sync()
	s.pakPkgs[1].Sync()
	exist, _ := s.pakPkgs[0].IsPkgExist()
	c.Check(exist, Equals, false)
	c.Check(s.pakPkgs[1].HeadRefName, Equals, "refs/heads/pak")

	originalGoGetImpl := GoGetImpl
	GoGetImpl = func(name string) error {
		if name == "github.com/theplant/package1" {
			mustRun("git", "clone", "fixtures/package1", "../../package1")
			GoGetImpl = originalGoGetImpl
		}
		return nil
	}

	err = Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	s.pakPkgs[0].Sync()
	exist, _ = s.pakPkgs[0].IsPkgExist()
	c.Check(exist, Equals, true)
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
	c.Check(s.pakPkgs[0].HeadRefName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[1].HeadRefName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[2].HeadRefName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[1].PakbranchChecksum, Equals, "941af3b182a1d0a5859fd451a8b5a633f479d7bc")

	paklockInfo, _ := GetPaklockInfo()
	c.Check(len(paklockInfo), Equals, 3)

	mustRun("rm", "-rf", "Pakfile")
	mustRun("cp", "fixtures/Pakfile3-updated", "Pakfile")

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
	c.Check(s.pakPkgs[0].HeadRefName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[1].HeadRefName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[1].PakbranchChecksum, Equals, "e373579a64e367338ff09b5143e312c81204c074")

	paklockInfo, _ = GetPaklockInfo()
	c.Check(len(paklockInfo), Equals, 2)
}

func (s *GetSuite) TestCanGetPackageOnNoBranch(c *C) {
	mustRun("git", "--git-dir=../../package1/.git", "--work-tree=../../package1", "checkout", "origin/master")

	err := s.pakPkgs[0].Sync()
	c.Check(err, Equals, nil)
	c.Check(s.pakPkgs[0].HeadRefName, Equals, "no branch")
	c.Check(s.pakPkgs[0].HeadChecksum, Equals, "11b174bd5acbf990687e6b068c97378d3219de04")

	err = Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	s.pakPkgs[0].Sync()
	c.Check(s.pakPkgs[0].HeadRefName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[0].HeadChecksum, Equals, "11b174bd5acbf990687e6b068c97378d3219de04")
}

func (s *GetSuite) TestNonLockedGetWithSkipUncleanPkgsOption(c *C) {
	mustRun("sh", "-c", "rm -rf ../../package1/file1")

	err := Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Not(Equals), nil)

	err = Get(PakOption{
		PakMeter:        []string{},
		UsePakfileLock:  true,
		SkipUncleanPkgs: true,
		Force:           false,
	})
	c.Check(err, Not(Equals), nil)
}

func (s *GetSuite) TestLockedGetWithSkipUncleanPkgsOption(c *C) {
	err := Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
		SkipUncleanPkgs: true,
	})
	c.Log("Should not succeed when running [pak -s get] without the existence of Pakfile.lock")
	c.Check(err, Not(Equals), nil)

	err = Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	s.pakPkgs[0].Sync()
	c.Check(s.pakPkgs[0].HeadRefName, Equals, "refs/heads/pak")

	mustRun("sh", "-c", "cd ../../package1; git checkout master")
	mustRun("sh", "-c", "cd ../../package2; git checkout master")
	s.pakPkgs[0].Sync()
	c.Check(s.pakPkgs[0].HeadRefName, Equals, "refs/heads/master")
	mustRun("sh", "-c", "rm -rf ../../package1/file1")

	err = Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Not(Equals), nil)

	s.pakPkgs[0].Sync()
	s.pakPkgs[1].Sync()
	c.Check(s.pakPkgs[0].HeadRefName, Equals, "refs/heads/master")
	c.Check(s.pakPkgs[1].HeadRefName, Equals, "refs/heads/master")

	err = Get(PakOption{
		PakMeter:        []string{},
		UsePakfileLock:  true,
		SkipUncleanPkgs: true,
		Force:           false,
	})
	c.Check(err, Equals, nil)

	s.pakPkgs[0].Sync()
	s.pakPkgs[1].Sync()
	c.Check(s.pakPkgs[0].HeadRefName, Equals, "refs/heads/master")
	c.Check(s.pakPkgs[1].HeadRefName, Equals, "refs/heads/pak")
}

func (s *GetSuite) TestGetWithSkipUncleanPkgsOptionAfterUntrackUncleanPkgs(c *C) {
	err := Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	s.pakPkgs[1].Sync()
	c.Check(s.pakPkgs[1].HeadRefName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[1].IsClean, Equals, true)

	mustRun("sh", "-c", "cp fixtures/Pakfile3-for-skip-unclean-pkgs Pakfile")
	mustRun("sh", "-c", "git --git-dir=../../package2/.git --work-tree=../../package2 checkout master")
	mustRun("sh", "-c", "mv ../../package2/file1 ../../package2/file1-1")

	s.pakPkgs[1].Sync()
	c.Check(s.pakPkgs[1].IsClean, Equals, false)

	err = Get(PakOption{
		PakMeter:        []string{},
		UsePakfileLock:  true,
		SkipUncleanPkgs: true,
		Force:           false,
	})
	c.Check(err, Equals, nil)

	s.pakPkgs[1].Sync()
	c.Check(s.pakPkgs[1].HeadRefName, Equals, "refs/heads/master")
	c.Check(s.pakPkgs[1].ContainsBranchNamedPak, Equals, true)
	c.Check(s.pakPkgs[1].ContainsPaktag, Equals, true)
	c.Check(s.pakPkgs[1].IsClean, Equals, false)
	paklockInfo, _ := GetPaklockInfo()
	c.Check(len(paklockInfo), Equals, 2)
	c.Check(paklockInfo["github.com/theplant/package1"], Not(Equals), "")
	c.Check(paklockInfo["github.com/theplant/package3"], Not(Equals), "")
}

func (s *GetSuite) TestShouldNotRemoveUncleanPkgChecksumInPakfileLockWithSkipOption(c *C) {
	err := Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	paklockInfo, _ := GetPaklockInfo()
	c.Check(len(paklockInfo), Equals, 3)

	mustRun("sh", "-c", "mv ../../package2/file1 ../../package2/file1-1")

	err = Get(PakOption{
		PakMeter:        []string{},
		UsePakfileLock:  true,
		SkipUncleanPkgs: true,
		Force:           false,
	})
	c.Check(err, Equals, nil)

	s.pakPkgs[1].Sync()
	c.Check(s.pakPkgs[1].IsClean, Equals, false)

	paklockInfo, _ = GetPaklockInfo()
	c.Check(len(paklockInfo), Equals, 3)
	c.Check(paklockInfo["github.com/theplant/package1"], Not(Equals), "")
	c.Check(paklockInfo["github.com/theplant/package2"], Not(Equals), "")
	c.Check(paklockInfo["github.com/theplant/package3"], Not(Equals), "")
}

func (s *GetSuite) TestShouldNotModifyPakfileLockWhenGetUncleanPkgWithSkipOption(c *C) {
	err := Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	paklockInfo, _ := GetPaklockInfo()
	c.Check(len(paklockInfo), Equals, 3)

	mustRun("sh", "-c", "mv ../../package2/file1 ../../package2/file1-1")

	err = Get(PakOption{
		PakMeter:        []string{"package2"},
		UsePakfileLock:  true,
		SkipUncleanPkgs: true,
		Force:           false,
	})
	c.Check(err, Not(Equals), nil)

	s.pakPkgs[1].Sync()
	c.Check(s.pakPkgs[1].IsClean, Equals, false)

	paklockInfo, _ = GetPaklockInfo()
	c.Check(len(paklockInfo), Equals, 3)
	c.Check(paklockInfo["github.com/theplant/package1"], Not(Equals), "")
	c.Check(paklockInfo["github.com/theplant/package2"], Not(Equals), "")
	c.Check(paklockInfo["github.com/theplant/package3"], Not(Equals), "")
}

func (s *GetSuite) TestGetArguments(c *C) {
	mustRun("sh", "-c", "cp fixtures/Pakfile3-for-get-args Pakfile")

	err := Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	paklockInfo, _ := GetPaklockInfo()
	c.Check(len(paklockInfo), Equals, 1)

	mustRun("sh", "-c", "cp fixtures/Pakfile3 Pakfile")
	mustRun("sh", "-c", "rm ../../package1/file1")

	err = Get(PakOption{
		PakMeter:        []string{"package2"},
		UsePakfileLock:  true,
		Force:           false,
		SkipUncleanPkgs: true,
	})
	c.Check(err, Equals, nil)

	paklockInfo, _ = GetPaklockInfo()
	c.Check(len(paklockInfo), Equals, 2)

	s.pakPkgs[0].Sync()
	// s.pakPkgs[1].Sync()
	// s.pakPkgs[2].Sync()
	c.Check(s.pakPkgs[0].HeadRefName, Equals, "refs/heads/pak")
	// c.Check(s.pakPkgs[1].HeadRefName, Equals, "refs/heads/pak")
	// c.Check(s.pakPkgs[2].HeadRefName, Equals, "refs/heads/pak")
}