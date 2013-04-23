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
	pkg: NewPakPkg("github.com/theplant/pakpkg-package", "origin", "master"),
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

func (s *PakPkgStateSuite) TestGoGet(c *C) {
	// TODO
}

// TODO: add multiple svc tests
func (s *PakPkgStateSuite) TestDial(c *C) {
	c.Check(s.pkg.Dial(), Equals, nil)
}

func (s *PakPkgStateSuite) TestParsePakfile(c *C) {
	pakPkgs, err := ParsePakfile()
	expectedGitPkgs := []PakPkg{
		NewPakPkg("github.com/theplant/package1", "origin", "master"),
		NewPakPkg("github.com/theplant/package2", "origin", "dev"),
		NewPakPkg("github.com/theplant/package3", "origin", "dev"),
		NewPakPkg("github.com/theplant/package4", "nonorigin", "dev"),
	}

	c.Check(err, Equals, nil)
	c.Check(pakPkgs, DeepEquals, expectedGitPkgs)
}

func (s *PakPkgStateSuite) TestEqualPakfileNPakfileLock(c *C) {
	// Pakfile <===> Pakfile.lock
	sampleGitpkgs := []PakPkg{
		NewPakPkg("github.com/theplant/package1", "origin", "master"),
		NewPakPkg("github.com/theplant/package2", "origin", "dev"),
		NewPakPkg("github.com/theplant/package3", "origin", "dev"),
		NewPakPkg("github.com/theplant/package4", "nonorigin", "dev"),
	}
	samplePakLockInfo := PaklockInfo{
		"github.com/theplant/package1": "whoop-fake-checksum-hash",
		"github.com/theplant/package2": "whoop-fake-checksum-hash",
		"github.com/theplant/package3": "whoop-fake-checksum-hash",
		"github.com/theplant/package4": "whoop-fake-checksum-hash",
	}

	obtainedNewGps, obtainedToUpdateGps, obtainedToRemoteGps := CategorizePakPkgs(sampleGitpkgs, samplePakLockInfo)
	expectedToUpdateGps := []PakPkg{
		NewPakPkg("github.com/theplant/package1", "origin", "master"),
		NewPakPkg("github.com/theplant/package2", "origin", "dev"),
		NewPakPkg("github.com/theplant/package3", "origin", "dev"),
		NewPakPkg("github.com/theplant/package4", "nonorigin", "dev"),
	}
	for i := 0; i < 4; i++ {
		expectedToUpdateGps[i].Checksum = "whoop-fake-checksum-hash"
	}

	c.Check(obtainedNewGps, DeepEquals, []PakPkg(nil))
	c.Check(obtainedToUpdateGps, DeepEquals, expectedToUpdateGps)
	c.Check(obtainedToRemoteGps, DeepEquals, []PakPkg(nil))
}

func (s *PakPkgStateSuite) TestAddNewPkgs(c *C) {
	sampleGitpkgs := []PakPkg{
		NewPakPkg("github.com/theplant/package1", "origin", "master"),
		NewPakPkg("github.com/theplant/package2", "origin", "dev"),
		NewPakPkg("github.com/theplant/package3", "origin", "dev"),
		NewPakPkg("github.com/theplant/package4", "nonorigin", "dev"),
	}
	samplePakLockInfo := PaklockInfo{
		"github.com/theplant/package1": "whoop-fake-checksum-hash",
		"github.com/theplant/package2": "whoop-fake-checksum-hash",
		"github.com/theplant/package3": "whoop-fake-checksum-hash",
	}

	obtainedNewGps, obtainedToUpdateGps, obtainedToRemoteGps := CategorizePakPkgs(sampleGitpkgs, samplePakLockInfo)
	expectedToUpdateGps := []PakPkg{
		NewPakPkg("github.com/theplant/package1", "origin", "master"),
		NewPakPkg("github.com/theplant/package2", "origin", "dev"),
		NewPakPkg("github.com/theplant/package3", "origin", "dev"),
	}
	for i := 0; i < 3; i++ {
		expectedToUpdateGps[i].Checksum = "whoop-fake-checksum-hash"
	}
	expectedNewGps := []PakPkg{NewPakPkg("github.com/theplant/package4", "nonorigin", "dev")}

	c.Check(obtainedNewGps, DeepEquals, expectedNewGps)
	c.Check(obtainedToUpdateGps, DeepEquals, expectedToUpdateGps)
	c.Check(obtainedToRemoteGps, DeepEquals, []PakPkg(nil))
}

func (s *PakPkgStateSuite) TestRemovePkgs(c *C) {
	sampleGitpkgs := []PakPkg{
		NewPakPkg("github.com/theplant/package1", "origin", "master"),
		NewPakPkg("github.com/theplant/package2", "origin", "dev"),
		NewPakPkg("github.com/theplant/package3", "origin", "dev"),
	}
	samplePakLockInfo := PaklockInfo{
		"github.com/theplant/package1": "whoop-fake-checksum-hash",
		"github.com/theplant/package2": "whoop-fake-checksum-hash",
		"github.com/theplant/package3": "whoop-fake-checksum-hash",
		"github.com/theplant/package4": "whoop-fake-checksum-hash",
	}

	obtainedNewGps, obtainedToUpdateGps, obtainedToRemoteGps := CategorizePakPkgs(sampleGitpkgs, samplePakLockInfo)
	expectedToUpdateGps := []PakPkg{
		NewPakPkg("github.com/theplant/package1", "origin", "master"),
		NewPakPkg("github.com/theplant/package2", "origin", "dev"),
		NewPakPkg("github.com/theplant/package3", "origin", "dev"),
	}
	for i := 0; i < 3; i++ {
		expectedToUpdateGps[i].Checksum = "whoop-fake-checksum-hash"
	}
	expectedToRemoveGps := []PakPkg{NewPakPkg("github.com/theplant/package4", "", "")}
	expectedToRemoveGps[0].Checksum = "whoop-fake-checksum-hash"

	c.Check(obtainedNewGps, DeepEquals, []PakPkg(nil))
	c.Check(obtainedToUpdateGps, DeepEquals, expectedToUpdateGps)
	c.Check(obtainedToRemoteGps, DeepEquals, expectedToRemoveGps)
}

func (s *PakPkgStateSuite) TestNewUpdateRemovePkgs(c *C) {
	sampleGitpkgs := []PakPkg{
		NewPakPkg("github.com/theplant/package1", "origin", "master"),
		NewPakPkg("github.com/theplant/package2", "origin", "dev"),
		NewPakPkg("github.com/theplant/package3", "origin", "dev"),
	}
	samplePakLockInfo := PaklockInfo{
		"github.com/theplant/package1": "whoop-fake-checksum-hash",
		"github.com/theplant/package2": "whoop-fake-checksum-hash",
		"github.com/theplant/package4": "whoop-fake-checksum-hash",
	}

	obtainedNewGps, obtainedToUpdateGps, obtainedToRemoteGps := CategorizePakPkgs(sampleGitpkgs, samplePakLockInfo)
	expectedNewGps := []PakPkg{NewPakPkg("github.com/theplant/package3", "origin", "dev")}
	expectedToUpdateGps := []PakPkg{
		NewPakPkg("github.com/theplant/package1", "origin", "master"),
		NewPakPkg("github.com/theplant/package2", "origin", "dev"),
	}
	for i := 0; i < 2; i++ {
		expectedToUpdateGps[i].Checksum = "whoop-fake-checksum-hash"
	}
	expectedToRemoveGps := []PakPkg{NewPakPkg("github.com/theplant/package4", "", "")}
	expectedToRemoveGps[0].Checksum = "whoop-fake-checksum-hash"

	c.Check(obtainedNewGps, DeepEquals, expectedNewGps)
	c.Check(obtainedToUpdateGps, DeepEquals, expectedToUpdateGps)
	c.Check(obtainedToRemoteGps, DeepEquals, expectedToRemoveGps)
}

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
			pkgArgs:          []string{"github.com/theplant/pakpkg-action-git-package", "origin", "master"},
			masterChecksum:   "11b174bd5acbf990687e6b068c97378d3219de04",
			setUpFixtures:    []string{"git", "clone", "fixtures/package1", "../../pakpkg-action-git-package"},
			tearDownFixtures: []string{"rm", "-rf", "../../pakpkg-action-git-package"},
		},
	},
})

func (s *PakPkgActionSuite) SetUpTest(c *C) {
	for _, p := range s.pkgs {
		p.pkg = NewPakPkg(p.pkgArgs[0], p.pkgArgs[1], p.pkgArgs[2])

		MustRun(p.setUpFixtures...)

		err := p.pkg.Dial()
		if err != nil {
			c.Fatal(err)
		}
	}
}
func (s *PakPkgActionSuite) TearDownTest(c *C) {
	for _, p := range s.pkgs {
		MustRun(p.tearDownFixtures...)
	}
}

func (s *PakPkgActionSuite) TestSimplePak(c *C) {
	for _, p := range s.pkgs {
		err := p.pkg.Sync()
		c.Check(err, Equals, nil)

		c.Check(p.pkg.HeadRefName, Equals, "refs/heads/master")
		c.Check(p.pkg.HeadChecksum, Equals, p.masterChecksum)
		c.Check(p.pkg.PakbranchChecksum, Equals, "")
		c.Check(p.pkg.PaktagChecksum, Equals, "")
		c.Check(p.pkg.ContainsBranchNamedPak, Equals, false)
		c.Check(p.pkg.ContainsPaktag, Equals, false)
		c.Check(p.pkg.OnPakbranch, Equals, false)
		c.Check(p.pkg.IsRemoteBranchExist, Equals, true)
		c.Check(p.pkg.IsClean, Equals, true)

		p.pkg.Pak(GetOption{Force: true, Checksum: ""})
		p.pkg.Sync()

		c.Check(p.pkg.HeadRefName, Equals, "refs/heads/pak")
		c.Check(p.pkg.HeadChecksum, Equals, p.masterChecksum)
		c.Check(p.pkg.PakbranchChecksum, Equals, p.masterChecksum)
		c.Check(p.pkg.PaktagChecksum, Equals, p.masterChecksum)
		c.Check(p.pkg.ContainsBranchNamedPak, Equals, true)
		c.Check(p.pkg.ContainsPaktag, Equals, true)
		c.Check(p.pkg.OnPakbranch, Equals, true)
		c.Check(p.pkg.IsRemoteBranchExist, Equals, true)
		c.Check(p.pkg.IsClean, Equals, true)
	}
}

func (s *PakPkgActionSuite) TestWeakPak(c *C) {
	for _, p := range s.pkgs {
		p.pkg.PkgProxy.NewBranch("pak")
		p.pkg.Sync()

		c.Check(p.pkg.HeadRefName, Equals, "refs/heads/master")
		c.Check(p.pkg.HeadChecksum, Equals, p.masterChecksum)
		c.Check(p.pkg.PakbranchChecksum, Equals, p.masterChecksum)
		c.Check(p.pkg.PaktagChecksum, Equals, "")
		c.Check(p.pkg.ContainsBranchNamedPak, Equals, true)
		c.Check(p.pkg.ContainsPaktag, Equals, false)
		c.Check(p.pkg.OwnPakbranch, Equals, false)
		c.Check(p.pkg.OnPakbranch, Equals, false)
		c.Check(p.pkg.IsRemoteBranchExist, Equals, true)
		c.Check(p.pkg.IsClean, Equals, true)

		p.pkg.Pak(GetOption{Force: false, Checksum: ""})
		p.pkg.Sync()

		c.Check(p.pkg.HeadRefName, Equals, "refs/heads/master")
		c.Check(p.pkg.HeadChecksum, Equals, p.masterChecksum)
		c.Check(p.pkg.PakbranchChecksum, Equals, p.masterChecksum)
		c.Check(p.pkg.PaktagChecksum, Equals, "")
		c.Check(p.pkg.ContainsBranchNamedPak, Equals, true)
		c.Check(p.pkg.ContainsPaktag, Equals, false)
		c.Check(p.pkg.OwnPakbranch, Equals, false)
		c.Check(p.pkg.OnPakbranch, Equals, false)
		c.Check(p.pkg.IsRemoteBranchExist, Equals, true)
		c.Check(p.pkg.IsClean, Equals, true)
	}
}

func (s *PakPkgActionSuite) TestForcefulPak(c *C) {
	for _, p := range s.pkgs {
		p.pkg.PkgProxy.NewBranch("pak")
		p.pkg.Sync()

		p.pkg.Pak(GetOption{Force: true, Checksum: ""})
		p.pkg.Sync()

		c.Check(p.pkg.HeadRefName, Equals, "refs/heads/pak")
		c.Check(p.pkg.HeadChecksum, Equals, p.masterChecksum)
		c.Check(p.pkg.PakbranchChecksum, Equals, p.masterChecksum)
		c.Check(p.pkg.PaktagChecksum, Equals, p.masterChecksum)
		c.Check(p.pkg.ContainsBranchNamedPak, Equals, true)
		c.Check(p.pkg.ContainsPaktag, Equals, true)
		c.Check(p.pkg.OwnPakbranch, Equals, true)
		c.Check(p.pkg.OnPakbranch, Equals, true)
		c.Check(p.pkg.IsRemoteBranchExist, Equals, true)
		c.Check(p.pkg.IsClean, Equals, true)
	}
}

func (s *PakPkgActionSuite) TestGetWithChecksum(c *C) {
	for _, p := range s.pkgs {
		devChecksum := "711c1e206bca5ad99edf6da12074bbbe4a349932"
		p.pkg.Sync()
		p.pkg.Pak(GetOption{Force: true, Checksum: devChecksum})
		p.pkg.Sync()

		c.Check(p.pkg.HeadRefName, Equals, "refs/heads/pak")
		c.Check(p.pkg.HeadChecksum, Equals, devChecksum)
		c.Check(p.pkg.PakbranchChecksum, Equals, devChecksum)
		c.Check(p.pkg.PaktagChecksum, Equals, devChecksum)
		c.Check(p.pkg.ContainsBranchNamedPak, Equals, true)
		c.Check(p.pkg.ContainsPaktag, Equals, true)
		c.Check(p.pkg.OwnPakbranch, Equals, true)
		c.Check(p.pkg.OnPakbranch, Equals, true)
		c.Check(p.pkg.IsRemoteBranchExist, Equals, true)
		c.Check(p.pkg.IsClean, Equals, true)
	}
}

func (s *PakPkgActionSuite) TestForcefulUnpak(c *C) {
	for _, p := range s.pkgs {
		p.pkg.Sync()
		p.pkg.Pak(GetOption{Force: true, Checksum: p.masterChecksum})
		p.pkg.Sync()
		p.pkg.Unpak(true)
		p.pkg.Sync()

		c.Check(p.pkg.HeadRefName, Equals, "refs/heads/master")
		c.Check(p.pkg.HeadChecksum, Equals, p.masterChecksum)
		c.Check(p.pkg.PakbranchChecksum, Equals, p.masterChecksum)
		c.Check(p.pkg.PaktagChecksum, Equals, p.masterChecksum)
		c.Check(p.pkg.ContainsBranchNamedPak, Equals, false)
		c.Check(p.pkg.ContainsPaktag, Equals, false)
		c.Check(p.pkg.OwnPakbranch, Equals, false)
		c.Check(p.pkg.OnPakbranch, Equals, false)
		c.Check(p.pkg.IsRemoteBranchExist, Equals, true)
		c.Check(p.pkg.IsClean, Equals, true)
	}
}

func (s *PakPkgActionSuite) TestWeakUnpak(c *C) {
	for _, p := range s.pkgs {
		p.pkg.NewBranch("pak")
		p.pkg.Sync()
		p.pkg.Unpak(false)
		p.pkg.Sync()

		c.Check(p.pkg.HeadRefName, Equals, "refs/heads/master")
		c.Check(p.pkg.HeadChecksum, Equals, p.masterChecksum)
		c.Check(p.pkg.PakbranchChecksum, Equals, p.masterChecksum)
		c.Check(p.pkg.PaktagChecksum, Equals, "")
		c.Check(p.pkg.ContainsBranchNamedPak, Equals, true)
		c.Check(p.pkg.ContainsPaktag, Equals, false)
		c.Check(p.pkg.OwnPakbranch, Equals, false)
		c.Check(p.pkg.OnPakbranch, Equals, false)
		c.Check(p.pkg.IsRemoteBranchExist, Equals, true)
		c.Check(p.pkg.IsClean, Equals, true)
	}
}

func (s *PakPkgActionSuite) TestPakBranchOwnership(c *C) {
	for _, p := range s.pkgs {
		p.pkg.Unpak(true)

		p.pkg.Sync()
		c.Check(p.pkg.ContainsBranchNamedPak, Equals, false)
		c.Check(p.pkg.ContainsPaktag, Equals, false)
		c.Check(p.pkg.OwnPakbranch, Equals, false)

		p.pkg.NewBranch("pak")
		p.pkg.Sync()
		c.Check(p.pkg.ContainsBranchNamedPak, Equals, true)
		c.Check(p.pkg.ContainsPaktag, Equals, false)
		c.Check(p.pkg.OwnPakbranch, Equals, false)

		// Although package has both pak branch and _pak_latest_ tag, but their commit hash are different, so
		// Pak wouldn't consider that pak branch is owned by it.
		p.pkg.NewTag("_pak_latest_", "HEAD^")
		p.pkg.Sync()
		c.Check(p.pkg.ContainsBranchNamedPak, Equals, true)
		c.Check(p.pkg.ContainsPaktag, Equals, true)
		c.Check(p.pkg.OwnPakbranch, Equals, false)

		err := p.pkg.RemoveTag("_pak_latest_")
		c.Check(err, Equals, nil)
		err = p.pkg.NewTag("_pak_latest_", "refs/heads/pak")
		c.Check(err, Equals, nil)
		p.pkg.Sync()
		c.Check(p.pkg.ContainsBranchNamedPak, Equals, true)
		c.Check(p.pkg.ContainsPaktag, Equals, true)
		c.Check(p.pkg.OwnPakbranch, Equals, true)
	}
}
