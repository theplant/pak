package core

import (
	. "github.com/theplant/pak/share"
	. "launchpad.net/gocheck"
	"os/exec"
)

type PakPkgStateSuite struct {
	pkg            PakPkg
	masterChecksum string
}

var _ = Suite(&PakPkgStateSuite{
	pkg: NewPakPkg(PkgCfg{
		Name:                   "github.com/theplant/pakpkg-package",
		PakName:                "pak",
		TargetBranch:           "origin/master",
		AutoMatchingHostBranch: false,
	}),
})

func (s *PakPkgStateSuite) SetUpSuite(c *C) {
	MustRun("cp", "fixtures/Pakfile2", "./Pakfile")
	MustRun("git", "clone", "fixtures/package1", "../../pakpkg-package")

	if err := s.pkg.Dial(); err != nil {
		c.Fatal(err)
	}
}

func (s *PakPkgStateSuite) TearDownSuite(c *C) {
	MustRun("rm", "-f", "Pakfile")
	MustRun("rm", "-rf", "../../pakpkg-package")
}

func (s *PakPkgStateSuite) TestIsPkgExist(c *C) {
	exist, err := s.pkg.IsPkgExist()
	c.Check(err, Equals, nil)
	c.Check(exist, Equals, true)

	exec.Command("rm", "-rf", "../../pakpkg-package").Run()

	exist, err = s.pkg.IsPkgExist()
	c.Check(err, Equals, nil)
	c.Check(exist, Equals, false)
}

// func (s *PakPkgStateSuite) TestGoGet(c *C) {
// 	// TODO
// }

// TODO: add multiple svc tests
func (s *PakPkgStateSuite) TestDial(c *C) {
	c.Check(s.pkg.Dial(), Equals, nil)
}

// func (s *PakPkgStateSuite) TestParsePakfile(c *C) {
// 	pakPkgs, err := ParsePakfile()
// 	expectedGitPkgs := []PakPkg{
// 		NewPakPkg(PkgCfg{
// 			Name:                   "github.com/theplant/package1",
// 			PakName:                "pak",
// 			TargetBranch:           "origin/master",
// 			AutoMatchingHostBranch: false,
// 		}),
// 		NewPakPkg(PkgCfg{
// 			Name:                   "github.com/theplant/package2",
// 			PakName:                "pak",
// 			TargetBranch:           "origin/dev",
// 			AutoMatchingHostBranch: false,
// 		}),
// 		NewPakPkg(PkgCfg{
// 			Name:                   "github.com/theplant/package3",
// 			PakName:                "pak",
// 			TargetBranch:           "origin/dev",
// 			AutoMatchingHostBranch: false,
// 		}),
// 		NewPakPkg(PkgCfg{
// 			Name:                   "github.com/theplant/package4",
// 			PakName:                "pak",
// 			TargetBranch:           "nonorigin/dev",
// 			AutoMatchingHostBranch: false,
// 		}),
// 	}

// 	c.Check(err, Equals, nil)
// 	c.Check(pakPkgs, DeepEquals, expectedGitPkgs)
// }

// TODO: fix it
// func (s *PakPkgStateSuite) TestEqualPakfileNPakfileLock(c *C) {
// 	// Pakfile <===> Pakfile.lock
// 	sampleGitpkgs := []PakPkg{
// 		NewPakPkg("github.com/theplant/package1", "origin", "master"),
// 		NewPakPkg("github.com/theplant/package2", "origin", "dev"),
// 		NewPakPkg("github.com/theplant/package3", "origin", "dev"),
// 		NewPakPkg("github.com/theplant/package4", "nonorigin", "dev"),
// 	}
// 	samplePakLockInfo := PaklockInfo{
// 		"github.com/theplant/package1": "whoop-fake-checksum-hash",
// 		"github.com/theplant/package2": "whoop-fake-checksum-hash",
// 		"github.com/theplant/package3": "whoop-fake-checksum-hash",
// 		"github.com/theplant/package4": "whoop-fake-checksum-hash",
// 	}

// 	obtainedNewGps, obtainedToUpdateGps, obtainedToRemoteGps := CategorizePakPkgs(sampleGitpkgs, samplePakLockInfo)
// 	expectedToUpdateGps := []PakPkg{
// 		NewPakPkg("github.com/theplant/package1", "origin", "master"),
// 		NewPakPkg("github.com/theplant/package2", "origin", "dev"),
// 		NewPakPkg("github.com/theplant/package3", "origin", "dev"),
// 		NewPakPkg("github.com/theplant/package4", "nonorigin", "dev"),
// 	}
// 	for i := 0; i < 4; i++ {
// 		expectedToUpdateGps[i].Checksum = "whoop-fake-checksum-hash"
// 	}

// 	c.Check(obtainedNewGps, DeepEquals, []PakPkg(nil))
// 	c.Check(obtainedToUpdateGps, DeepEquals, expectedToUpdateGps)
// 	c.Check(obtainedToRemoteGps, DeepEquals, []PakPkg(nil))
// }

// func (s *PakPkgStateSuite) TestAddNewPkgs(c *C) {
// 	sampleGitpkgs := []PakPkg{
// 		NewPakPkg("github.com/theplant/package1", "origin", "master"),
// 		NewPakPkg("github.com/theplant/package2", "origin", "dev"),
// 		NewPakPkg("github.com/theplant/package3", "origin", "dev"),
// 		NewPakPkg("github.com/theplant/package4", "nonorigin", "dev"),
// 	}
// 	samplePakLockInfo := PaklockInfo{
// 		"github.com/theplant/package1": "whoop-fake-checksum-hash",
// 		"github.com/theplant/package2": "whoop-fake-checksum-hash",
// 		"github.com/theplant/package3": "whoop-fake-checksum-hash",
// 	}

// 	obtainedNewGps, obtainedToUpdateGps, obtainedToRemoteGps := CategorizePakPkgs(sampleGitpkgs, samplePakLockInfo)
// 	expectedToUpdateGps := []PakPkg{
// 		NewPakPkg("github.com/theplant/package1", "origin", "master"),
// 		NewPakPkg("github.com/theplant/package2", "origin", "dev"),
// 		NewPakPkg("github.com/theplant/package3", "origin", "dev"),
// 	}
// 	for i := 0; i < 3; i++ {
// 		expectedToUpdateGps[i].Checksum = "whoop-fake-checksum-hash"
// 	}
// 	expectedNewGps := []PakPkg{NewPakPkg("github.com/theplant/package4", "nonorigin", "dev")}

// 	c.Check(obtainedNewGps, DeepEquals, expectedNewGps)
// 	c.Check(obtainedToUpdateGps, DeepEquals, expectedToUpdateGps)
// 	c.Check(obtainedToRemoteGps, DeepEquals, []PakPkg(nil))
// }

// func (s *PakPkgStateSuite) TestRemovePkgs(c *C) {
// 	sampleGitpkgs := []PakPkg{
// 		NewPakPkg("github.com/theplant/package1", "origin", "master"),
// 		NewPakPkg("github.com/theplant/package2", "origin", "dev"),
// 		NewPakPkg("github.com/theplant/package3", "origin", "dev"),
// 	}
// 	samplePakLockInfo := PaklockInfo{
// 		"github.com/theplant/package1": "whoop-fake-checksum-hash",
// 		"github.com/theplant/package2": "whoop-fake-checksum-hash",
// 		"github.com/theplant/package3": "whoop-fake-checksum-hash",
// 		"github.com/theplant/package4": "whoop-fake-checksum-hash",
// 	}

// 	obtainedNewGps, obtainedToUpdateGps, obtainedToRemoteGps := CategorizePakPkgs(sampleGitpkgs, samplePakLockInfo)
// 	expectedToUpdateGps := []PakPkg{
// 		NewPakPkg("github.com/theplant/package1", "origin", "master"),
// 		NewPakPkg("github.com/theplant/package2", "origin", "dev"),
// 		NewPakPkg("github.com/theplant/package3", "origin", "dev"),
// 	}
// 	for i := 0; i < 3; i++ {
// 		expectedToUpdateGps[i].Checksum = "whoop-fake-checksum-hash"
// 	}
// 	expectedToRemoveGps := []PakPkg{NewPakPkg("github.com/theplant/package4", "", "")}
// 	expectedToRemoveGps[0].Checksum = "whoop-fake-checksum-hash"

// 	c.Check(obtainedNewGps, DeepEquals, []PakPkg(nil))
// 	c.Check(obtainedToUpdateGps, DeepEquals, expectedToUpdateGps)
// 	c.Check(obtainedToRemoteGps, DeepEquals, expectedToRemoveGps)
// }

// func (s *PakPkgStateSuite) TestNewUpdateRemovePkgs(c *C) {
// 	sampleGitpkgs := []PakPkg{
// 		NewPakPkg("github.com/theplant/package1", "origin", "master"),
// 		NewPakPkg("github.com/theplant/package2", "origin", "dev"),
// 		NewPakPkg("github.com/theplant/package3", "origin", "dev"),
// 	}
// 	samplePakLockInfo := PaklockInfo{
// 		"github.com/theplant/package1": "whoop-fake-checksum-hash",
// 		"github.com/theplant/package2": "whoop-fake-checksum-hash",
// 		"github.com/theplant/package4": "whoop-fake-checksum-hash",
// 	}

// 	obtainedNewGps, obtainedToUpdateGps, obtainedToRemoteGps := CategorizePakPkgs(sampleGitpkgs, samplePakLockInfo)
// 	expectedNewGps := []PakPkg{NewPakPkg("github.com/theplant/package3", "origin", "dev")}
// 	expectedToUpdateGps := []PakPkg{
// 		NewPakPkg("github.com/theplant/package1", "origin", "master"),
// 		NewPakPkg("github.com/theplant/package2", "origin", "dev"),
// 	}
// 	for i := 0; i < 2; i++ {
// 		expectedToUpdateGps[i].Checksum = "whoop-fake-checksum-hash"
// 	}
// 	expectedToRemoveGps := []PakPkg{NewPakPkg("github.com/theplant/package4", "", "")}
// 	expectedToRemoveGps[0].Checksum = "whoop-fake-checksum-hash"

// 	c.Check(obtainedNewGps, DeepEquals, expectedNewGps)
// 	c.Check(obtainedToUpdateGps, DeepEquals, expectedToUpdateGps)
// 	c.Check(obtainedToRemoteGps, DeepEquals, expectedToRemoveGps)
// }

type pakpkgTs struct {
	pkg              PakPkg
	pkgArgs          []string
	masterChecksum   string
	setUpFixtures    []string
	tearDownFixtures []string
}

type PakPkgActionSuite struct {
	pkgs []*pakpkgTs
}

var _ = Suite(&PakPkgActionSuite{
	pkgs: []*pakpkgTs{
		{
			pkgArgs:          []string{"github.com/theplant/pakpkg-action-git-package", "origin/master"},
			masterChecksum:   "11b174bd5acbf990687e6b068c97378d3219de04",
			setUpFixtures:    []string{"git", "clone", "fixtures/package1", "../../pakpkg-action-git-package"},
			tearDownFixtures: []string{"rm", "-rf", "../../pakpkg-action-git-package"},
		},
		{
			pkgArgs:          []string{"github.com/theplant/pakpkg-action-hg-package", "default/default"},
			masterChecksum:   "1eebd4597062386493ac83ac80b0b9e3d08f7af7",
			setUpFixtures:    []string{"hg", "clone", "fixtures/package1-for-hg", "../../pakpkg-action-hg-package"},
			tearDownFixtures: []string{"rm", "-rf", "../../pakpkg-action-hg-package"},
		},
	},
})

func (s *PakPkgActionSuite) SetUpTest(c *C) {
	for i, p := range s.pkgs {
		// p.pkg = NewPakPkg(p.pkgArgs[0], p.pkgArgs[1], p.pkgArgs[2])
		(*s.pkgs[i]).pkg = NewPakPkg(PkgCfg{Name: p.pkgArgs[0], TargetBranch: p.pkgArgs[1], PakName: "pak"})

		MustRun(p.setUpFixtures...)

		err := (*s.pkgs[i]).pkg.Dial()
		if err != nil {
			c.Fatal(err)
		}

	}
	GoGetImpl = func(name string) error { return nil }
}
func (s *PakPkgActionSuite) TearDownTest(c *C) {
	for i, p := range s.pkgs {
		_ = i
		MustRun(p.tearDownFixtures...)
	}
}

func (s *PakPkgActionSuite) TestSimplePak(c *C) {
	var simplePakHelper = [][]string{
		{"refs/heads/master", "refs/heads/pak"},
		{"non-pak", "pak"},
	}
	titles := []string{
		"Git Package",
		"Hg Package",
	}
	for i, p := range s.pkgs {
		c.Log(titles[i])

		err := p.pkg.Sync()
		c.Check(err, Equals, nil)

		c.Check(p.pkg.HeadRefName, Equals, simplePakHelper[i][0])
		c.Check(p.pkg.HeadChecksum, Equals, p.masterChecksum)
		c.Check(p.pkg.PakbranchChecksum, Equals, "")
		// c.Check(p.pkg.PaktagChecksum, Equals, "")
		c.Check(p.pkg.HasPakBranch, Equals, false)
		// c.Check(p.pkg.ContainsBranchNamedPak, Equals, false)
		// c.Check(p.pkg.ContainsPaktag, Equals, false)
		c.Check(p.pkg.OnPakbranch, Equals, false)
		c.Check(p.pkg.IsRemoteBranchExist, Equals, true)
		c.Check(p.pkg.IsClean, Equals, true)

		p.pkg.GetOption = GetOption{Force: true, Checksum: ""}
		p.pkg.Pak()
		p.pkg.Sync()

		c.Check(p.pkg.HeadRefName, Equals, simplePakHelper[i][1])
		c.Check(p.pkg.HeadChecksum, Equals, p.masterChecksum)
		c.Check(p.pkg.PakbranchChecksum, Equals, p.masterChecksum)
		// c.Check(p.pkg.PaktagChecksum, Equals, p.masterChecksum)
		c.Check(p.pkg.HasPakBranch, Equals, true)
		// c.Check(p.pkg.ContainsBranchNamedPak, Equals, true)
		// c.Check(p.pkg.ContainsPaktag, Equals, true)
		c.Check(p.pkg.OnPakbranch, Equals, true)
		c.Check(p.pkg.IsRemoteBranchExist, Equals, true)
		c.Check(p.pkg.IsClean, Equals, true)
	}
}

// TODO: is this test still necessary
// func (s *PakPkgActionSuite) TestWeakPak(c *C) {
// 	var weakPakHelper = [][]string{
// 		{"refs/heads/master", "refs/heads/master"},
// 		{"pak", "pak"},
// 	}
// 	for i, p := range s.pkgs {
// 		p.pkg.PkgProxy.NewBranch("pak")
// 		p.pkg.Sync()

// 		c.Check(p.pkg.HeadRefName, Equals, weakPakHelper[i][0])
// 		c.Check(p.pkg.HeadChecksum, Equals, p.masterChecksum)
// 		c.Check(p.pkg.PakbranchChecksum, Equals, p.masterChecksum)
// 		// c.Check(p.pkg.PaktagChecksum, Equals, "")
// 		c.Check(p.pkg.HasPakBranch, Equals, true)
// 		// c.Check(p.pkg.ContainsBranchNamedPak, Equals, true)
// 		// c.Check(p.pkg.ContainsPaktag, Equals, false)
// 		// c.Check(p.pkg.OwnPakbranch, Equals, false)
// 		c.Check(p.pkg.OnPakbranch, Equals, false)
// 		c.Check(p.pkg.IsRemoteBranchExist, Equals, true)
// 		c.Check(p.pkg.IsClean, Equals, true)

// 		p.pkg.GetOption = GetOption{Force: false, Checksum: ""}
// 		p.pkg.Pak()
// 		p.pkg.Sync()

// 		c.Check(p.pkg.HeadRefName, Equals, weakPakHelper[i][1])
// 		c.Check(p.pkg.HeadChecksum, Equals, p.masterChecksum)
// 		c.Check(p.pkg.PakbranchChecksum, Equals, p.masterChecksum)
// 		// c.Check(p.pkg.PaktagChecksum, Equals, "")
// 		c.Check(p.pkg.HasPakBranch, Equals, true)
// 		// c.Check(p.pkg.ContainsBranchNamedPak, Equals, true)
// 		// c.Check(p.pkg.ContainsPaktag, Equals, false)
// 		// c.Check(p.pkg.OwnPakbranch, Equals, false)
// 		c.Check(p.pkg.OnPakbranch, Equals, false)
// 		c.Check(p.pkg.IsRemoteBranchExist, Equals, true)
// 		c.Check(p.pkg.IsClean, Equals, true)
// 	}
// }

func (s *PakPkgActionSuite) TestForcefulPak(c *C) {
	var forcefulPakFixtures = [][]string{
		{"refs/heads/pak"},
		{"pak"},
	}
	for i, p := range s.pkgs {
		// p.pkg.PkgProxy.NewBranch("pak")

		p.pkg.Sync()
		p.pkg.GetOption = GetOption{Force: true, Checksum: ""}
		_, err := p.pkg.Pak()
		c.Check(err, Equals, nil)

		p.pkg.Sync()

		c.Check(p.pkg.HeadRefName, Equals, forcefulPakFixtures[i][0])
		c.Check(p.pkg.HeadChecksum, Equals, p.masterChecksum)
		c.Check(p.pkg.PakbranchChecksum, Equals, p.masterChecksum)
		// c.Check(p.pkg.PaktagChecksum, Equals, p.masterChecksum)
		c.Check(p.pkg.HasPakBranch, Equals, true)
		// c.Check(p.pkg.ContainsBranchNamedPak, Equals, true)
		// c.Check(p.pkg.ContainsPaktag, Equals, true)
		// c.Check(p.pkg.OwnPakbranch, Equals, true)
		c.Check(p.pkg.OnPakbranch, Equals, true)
		c.Check(p.pkg.IsRemoteBranchExist, Equals, true)
		c.Check(p.pkg.IsClean, Equals, true)
	}
}

func (s *PakPkgActionSuite) TestGetWithChecksum(c *C) {
	var helper = [][]string{
		{"711c1e206bca5ad99edf6da12074bbbe4a349932", "refs/heads/pak"},
		{"a8d6d716129a1bb5678e195828dfe4b53c078c2e", "pak"},
	}
	for i, p := range s.pkgs {
		devChecksum := helper[i][0]
		p.pkg.Sync()
		p.pkg.GetOption = GetOption{Force: true, Checksum: devChecksum}
		p.pkg.Pak()
		p.pkg.Sync()

		c.Check(p.pkg.HeadRefName, Equals, helper[i][1])
		c.Check(p.pkg.HeadChecksum, Equals, devChecksum)
		c.Check(p.pkg.PakbranchChecksum, Equals, devChecksum)
		// c.Check(p.pkg.PaktagChecksum, Equals, devChecksum)
		c.Check(p.pkg.HasPakBranch, Equals, true)
		// c.Check(p.pkg.ContainsBranchNamedPak, Equals, true)
		// c.Check(p.pkg.ContainsPaktag, Equals, true)
		// c.Check(p.pkg.OwnPakbranch, Equals, true)
		c.Check(p.pkg.OnPakbranch, Equals, true)
		c.Check(p.pkg.IsRemoteBranchExist, Equals, true)
		c.Check(p.pkg.IsClean, Equals, true)
	}
}

func (s *PakPkgActionSuite) TestForcefulUnpak(c *C) {
	helper := [][]string{
		{"refs/heads/master"},
		{"non-pak"},
	}
	titles := []string{
		"Git Package",
		"Hg Package",
	}
	for i, p := range s.pkgs {
		c.Log(titles[i])

		err := p.pkg.Sync()
		c.Check(err, Equals, nil)
		p.pkg.GetOption = GetOption{Force: true, Checksum: p.masterChecksum}
		_, err = p.pkg.Pak()
		c.Check(err, Equals, nil)
		err = p.pkg.Sync()
		c.Check(err, Equals, nil)
		err = p.pkg.Unpak(true)
		c.Check(err, Equals, nil)
		err = p.pkg.Sync()
		c.Check(err, Equals, nil)

		c.Check(p.pkg.HeadRefName, Equals, helper[i][0])
		c.Check(p.pkg.HeadChecksum, Equals, p.masterChecksum)
		c.Check(p.pkg.PakbranchChecksum, Equals, p.masterChecksum)
		// c.Check(p.pkg.PaktagChecksum, Equals, p.masterChecksum)
		c.Check(p.pkg.HasPakBranch, Equals, false)
		// c.Check(p.pkg.ContainsBranchNamedPak, Equals, false)
		// c.Check(p.pkg.ContainsPaktag, Equals, false)
		// c.Check(p.pkg.OwnPakbranch, Equals, false)
		c.Check(p.pkg.OnPakbranch, Equals, false)
		c.Check(p.pkg.IsRemoteBranchExist, Equals, true)
		c.Check(p.pkg.IsClean, Equals, true)
	}
}

// TODO: still necessary???
// func (s *PakPkgActionSuite) TestWeakUnpak(c *C) {
// 	helper := [][]string{
// 		{"refs/heads/master"},
// 		{"pak"},
// 	}
// 	titles := []string{
// 		"Git Package",
// 		"Hg Package",
// 	}
// 	for i, p := range s.pkgs {
// 		c.Log(titles[i])

// 		p.pkg.NewBranch("pak")
// 		p.pkg.Sync()
// 		p.pkg.Unpak(false)
// 		p.pkg.Sync()

// 		c.Check(p.pkg.HeadRefName, Equals, helper[i][0])
// 		c.Check(p.pkg.HeadChecksum, Equals, p.masterChecksum)
// 		c.Check(p.pkg.PakbranchChecksum, Equals, p.masterChecksum)
// 		// c.Check(p.pkg.PaktagChecksum, Equals, "")
// 		c.Check(p.pkg.HasPakBranch, Equals, true)
// 		// c.Check(p.pkg.ContainsBranchNamedPak, Equals, true)
// 		// c.Check(p.pkg.ContainsPaktag, Equals, false)
// 		// c.Check(p.pkg.OwnPakbranch, Equals, false)
// 		c.Check(p.pkg.OnPakbranch, Equals, false)
// 		c.Check(p.pkg.IsRemoteBranchExist, Equals, true)
// 		c.Check(p.pkg.IsClean, Equals, true)
// 	}
// }

func (s *PakPkgActionSuite) TestPakBranchOwnership(c *C) {
	helpers := [][]string{
		{"HEAD^", "refs/heads/pak"},
		{"ca48dd9fbded", "pak"},
	}
	titles := []string{
		"Git Package",
		"Hg Package",
	}
	for i, p := range s.pkgs {
		c.Log(titles[i])

		p.pkg.Unpak(true)

		p.pkg.Sync()
		c.Check(p.pkg.HasPakBranch, Equals, false)
		// c.Check(p.pkg.ContainsBranchNamedPak, Equals, false)
		// c.Check(p.pkg.ContainsPaktag, Equals, false)
		// c.Check(p.pkg.OwnPakbranch, Equals, false)

		p.pkg.NewBranch("pak")
		p.pkg.Sync()
		c.Check(p.pkg.HasPakBranch, Equals, true)
		// c.Check(p.pkg.ContainsBranchNamedPak, Equals, true)
		// c.Check(p.pkg.ContainsPaktag, Equals, false)
		// c.Check(p.pkg.OwnPakbranch, Equals, false)

		// Although package has both pak branch and _pak_latest_ tag, but their commit hash are different, so
		// Pak wouldn't consider that pak branch is owned by it.
		p.pkg.NewTag("_pak_latest_", helpers[i][0])
		p.pkg.Sync()
		c.Check(p.pkg.HasPakBranch, Equals, true)
		// c.Check(p.pkg.ContainsBranchNamedPak, Equals, true)
		// c.Check(p.pkg.ContainsPaktag, Equals, true)
		// c.Check(p.pkg.OwnPakbranch, Equals, false)

		err := p.pkg.RemoveTag("_pak_latest_")
		c.Check(err, Equals, nil)
		err = p.pkg.NewTag("_pak_latest_", helpers[i][1])
		c.Check(err, Equals, nil)
		p.pkg.Sync()
		c.Check(p.pkg.HasPakBranch, Equals, true)
		// c.Check(p.pkg.ContainsBranchNamedPak, Equals, true)
		// c.Check(p.pkg.ContainsPaktag, Equals, true)
		// c.Check(p.pkg.OwnPakbranch, Equals, true)
	}
}
