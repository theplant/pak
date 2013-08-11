package core

import (
	. "github.com/theplant/pak/share"
	. "launchpad.net/gocheck"
)

type GetSuite struct {
	pakPkgs           []*PakPkg
	originalGoGetImpl func(name string) error
}

var _ = Suite(&GetSuite{})

func (s *GetSuite) SetUpSuite(c *C) {
	s.originalGoGetImpl = GoGetImpl
	GoGetImpl = func(name string) error { return nil }
}

func (s *GetSuite) TearDownSuite(c *C) {
	GoGetImpl = s.originalGoGetImpl
}

func (s *GetSuite) SetUpTest(c *C) {
	MustRun("git", "clone", "fixtures/package1", "../../package1")
	MustRun("git", "clone", "fixtures/package2", "../../package2")
	MustRun("git", "clone", "fixtures/package3", "../../package3")
	MustRun("hg", "clone", "fixtures/package1-for-hg", "../../package1-for-hg-get-test")
	MustRun("cp", "fixtures/Pakfile3", "Pakfile")

	// Now this assupmption that a pakName might be used by user is no longer exist.
	// // make package1 a pakcage managed by pak
	// MustRun("sh", "-c", "cd ../../package1 && git checkout -b pak && git checkout master && git tag _pak_latest_")

	s.pakPkgs = []*PakPkg{}
	pkgsFixtures := [][]string{
		{"github.com/theplant/package1", "origin/master"},
		{"github.com/theplant/package2", "origin/dev"},
		{"github.com/theplant/package3", "origin/master"},
		{"github.com/theplant/package1-for-hg-get-test", "default/default"},
	}
	for _, vals := range pkgsFixtures {
		// pkg := NewPakPkg(vals[0], vals[1], vals[2])
		pkg := NewPakPkg(PkgCfg{Name: vals[0], PakName: "pak", TargetBranch: vals[1]})
		err := pkg.Dial()
		if err != nil {
			c.Fatal(err, Equals, nil)
		}
		s.pakPkgs = append(s.pakPkgs, &pkg)
	}
}

func (s *GetSuite) TearDownTest(c *C) {
	MustRun("rm", "-rf", "../../package1")
	MustRun("rm", "-rf", "../../package2")
	MustRun("rm", "-rf", "../../package3")
	MustRun("rm", "-rf", "../../package1-for-hg-get-test")
	MustRun("rm", "-rf", "Pakfile")
	MustRun("rm", "-rf", "Pakfile.lock")
}

func (s *GetSuite) TestSimpleGet(c *C) {
	err := Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	c.Log("Git Packages")
	s.pakPkgs[0].Sync()
	s.pakPkgs[1].Sync()
	s.pakPkgs[2].Sync()
	c.Check(s.pakPkgs[0].HeadRefName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[1].HeadRefName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[2].HeadRefName, Equals, "refs/heads/pak")

	// c.Log("Hg Packages")
	s.pakPkgs[3].Sync()
	c.Check(s.pakPkgs[3].HeadRefName, Equals, "pak")
}

func (s *GetSuite) TestCanGetPackageWithoutRemoteBranch(c *C) {
	MustRun("git", "--git-dir=../../package1/.git", "--work-tree=../../package1", "branch", "-dr", "origin/dev")
	MustRun("git", "--git-dir=../../package1/.git", "--work-tree=../../package1", "branch", "-dr", "origin/master")

	err := Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	c.Log("Git Packages")
	s.pakPkgs[0].Sync()
	s.pakPkgs[1].Sync()
	s.pakPkgs[2].Sync()
	c.Check(s.pakPkgs[0].HeadRefName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[1].HeadRefName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[2].HeadRefName, Equals, "refs/heads/pak")

	c.Log("Hg Packages")
	s.pakPkgs[3].Sync()
	c.Check(s.pakPkgs[3].HeadRefName, Equals, "pak")
}

func (s *GetSuite) TestGetWithPakMeterForUpdating(c *C) {
	err := Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	MustRun("git", "--git-dir=../../package1/.git", "--work-tree=../../package1", "checkout", "master")
	MustRun("git", "--git-dir=../../package3/.git", "--work-tree=../../package3", "checkout", "master")

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

func (s *GetSuite) TestPaklockInfoShouldBeUpdatedAfterGet(c *C) {
	expectedPaklockInfo := PaklockInfo{
		"github.com/theplant/package3":                 "d5f51ca77f5d4f37a8105a74b67d2f1aefea939c",
		"github.com/theplant/package1":                 "11b174bd5acbf990687e6b068c97378d3219de04",
		"github.com/theplant/package2":                 "941af3b182a1d0a5859fd451a8b5a633f479d7bc",
		"github.com/theplant/package1-for-hg-get-test": "1eebd4597062386493ac83ac80b0b9e3d08f7af7",
	}

	err := Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: false,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	paklockInfo, _ := GetPaklockInfo("")
	c.Check(len(paklockInfo), Equals, 4)
	c.Check(paklockInfo, DeepEquals, expectedPaklockInfo)

	err = Get(PakOption{
		PakMeter:       []string{"github.com/theplant/package2"},
		UsePakfileLock: false,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	paklockInfo, _ = GetPaklockInfo("")
	c.Check(len(paklockInfo), Equals, 4)
	c.Check(paklockInfo, DeepEquals, expectedPaklockInfo)
}

func (s *GetSuite) TestGoGetBeforePak(c *C) {
	originalGoGetImpl := GoGetImpl
	gogetPkg1Count := 0
	GoGetImpl = func(name string) error {
		if name == "github.com/theplant/package1" {
			gogetPkg1Count++
			if gogetPkg1Count == 1 {
				MustRun("git", "clone", "fixtures/package1", "../../package1")
			}
		}

		return nil
	}
	defer func() { GoGetImpl = originalGoGetImpl }()

	MustRun("rm", "-rf", "../../package1")

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

	MustRun("rm", "-rf", "../../package1")

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
			MustRun("git", "clone", "fixtures/package1", "../../package1")
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

	paklockInfo, _ := GetPaklockInfo("")
	c.Check(len(paklockInfo), Equals, 4)

	MustRun("rm", "-rf", "Pakfile")
	MustRun("cp", "fixtures/Pakfile3-updated", "Pakfile")

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

	paklockInfo, _ = GetPaklockInfo("")
	c.Check(len(paklockInfo), Equals, 2)
}

func (s *GetSuite) TestCanGetPackageOnNoBranch(c *C) {
	MustRun("git", "--git-dir=../../package1/.git", "--work-tree=../../package1", "checkout", "origin/master")

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
	MustRun("sh", "-c", "rm -rf ../../package1/file1")

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
		PakMeter:        []string{},
		UsePakfileLock:  true,
		Force:           false,
		SkipUncleanPkgs: true,
	})
	// c.Log("Should not succeed when running [pak -s get] without the existence of Pakfile.lock")
	// c.Check(err, Not(Equals), nil)

	err = Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	s.pakPkgs[0].Sync()
	c.Check(s.pakPkgs[0].HeadRefName, Equals, "refs/heads/pak")

	MustRun("sh", "-c", "cd ../../package1; git checkout master")
	MustRun("sh", "-c", "cd ../../package2; git checkout master")
	s.pakPkgs[0].Sync()
	c.Check(s.pakPkgs[0].HeadRefName, Equals, "refs/heads/master")
	MustRun("sh", "-c", "rm -rf ../../package1/file1")

	err = Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Not(Equals), nil)

	s.pakPkgs[0].Sync()
	s.pakPkgs[1].Sync()
	c.Check(s.pakPkgs[0].HeadRefName, Equals, "refs/heads/master")
	// c.Check(s.pakPkgs[1].HeadRefName, Equals, "refs/heads/master")
	c.Check(s.pakPkgs[1].HeadRefName, Equals, "refs/heads/pak")

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

func (s *GetSuite) TestGetWithSkipUncleanPkgsOptionAfterUntrackingUncleanPkgs(c *C) {
	err := Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	s.pakPkgs[1].Sync()
	c.Check(s.pakPkgs[1].HeadRefName, Equals, "refs/heads/pak")
	c.Check(s.pakPkgs[1].IsClean, Equals, true)

	MustRun("sh", "-c", "cp fixtures/Pakfile3-for-skip-unclean-pkgs Pakfile")
	MustRun("sh", "-c", "git --git-dir=../../package2/.git --work-tree=../../package2 checkout master")
	MustRun("sh", "-c", "mv ../../package2/file1 ../../package2/file1-1")

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
	// c.Check(s.pakPkgs[1].ContainsBranchNamedPak, Equals, true)
	c.Check(s.pakPkgs[1].HasPakBranch, Equals, true)
	// c.Check(s.pakPkgs[1].ContainsPaktag, Equals, true)
	c.Check(s.pakPkgs[1].IsClean, Equals, false)
	paklockInfo, _ := GetPaklockInfo("")

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

	paklockInfo, _ := GetPaklockInfo("")
	c.Check(len(paklockInfo), Equals, 4)

	MustRun("sh", "-c", "mv ../../package2/file1 ../../package2/file1-1")

	err = Get(PakOption{
		PakMeter:        []string{},
		UsePakfileLock:  true,
		SkipUncleanPkgs: true,
		Force:           false,
	})
	c.Check(err, Equals, nil)

	s.pakPkgs[1].Sync()
	c.Check(s.pakPkgs[1].IsClean, Equals, false)

	paklockInfo, _ = GetPaklockInfo("")
	c.Check(len(paklockInfo), Equals, 4)
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

	paklockInfo, _ := GetPaklockInfo("")
	c.Check(len(paklockInfo), Equals, 4)

	MustRun("sh", "-c", "mv ../../package2/file1 ../../package2/file1-1")

	err = Get(PakOption{
		PakMeter:        []string{"package2"},
		UsePakfileLock:  true,
		SkipUncleanPkgs: true,
		Force:           false,
	})
	c.Check(err, Not(Equals), nil)

	s.pakPkgs[1].Sync()
	c.Check(s.pakPkgs[1].IsClean, Equals, false)

	paklockInfo, _ = GetPaklockInfo("")
	c.Check(len(paklockInfo), Equals, 4)
	c.Check(paklockInfo["github.com/theplant/package1"], Not(Equals), "")
	c.Check(paklockInfo["github.com/theplant/package2"], Not(Equals), "")
	c.Check(paklockInfo["github.com/theplant/package3"], Not(Equals), "")
}

func (s *GetSuite) TestGetWithPakMeterForGetting(c *C) {
	MustRun("sh", "-c", "cp fixtures/Pakfile3-for-get-args Pakfile")

	err := Get(PakOption{
		PakMeter:       []string{},
		UsePakfileLock: true,
		Force:          false,
	})
	c.Check(err, Equals, nil)

	paklockInfo, _ := GetPaklockInfo("")
	c.Check(len(paklockInfo), Equals, 1)

	MustRun("sh", "-c", "cp fixtures/Pakfile3 Pakfile")
	MustRun("sh", "-c", "rm ../../package1/file1")

	err = Get(PakOption{
		PakMeter:        []string{"package2"},
		UsePakfileLock:  true,
		Force:           false,
		SkipUncleanPkgs: true,
	})
	c.Check(err, Equals, nil)

	paklockInfo, _ = GetPaklockInfo("")
	c.Check(len(paklockInfo), Equals, 2)

	s.pakPkgs[0].Sync()
	// s.pakPkgs[1].Sync()
	// s.pakPkgs[2].Sync()
	c.Check(s.pakPkgs[0].HeadRefName, Equals, "refs/heads/pak")
	// c.Check(s.pakPkgs[1].HeadRefName, Equals, "refs/heads/pak")
	// c.Check(s.pakPkgs[2].HeadRefName, Equals, "refs/heads/pak")
}
