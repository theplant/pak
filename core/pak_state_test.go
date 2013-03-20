package pak

import (
	"fmt"
	"github.com/theplant/pak/gitpkg"
	. "github.com/theplant/pak/share"
	. "launchpad.net/gocheck"
	"os/exec"
)

type PakStateSuite struct{}

var _ = Suite(&PakStateSuite{})

func (s *PakStateSuite) SetUpSuite(c *C) {
	exec.Command("cp", "fixtures/Pakfile2", "./Pakfile").Run()
}

func (s *PakStateSuite) TearDownSuite(c *C) {
	exec.Command("rm", "-f", "Pakfile").Run()
}

func (s *PakStateSuite) TestParsePakfile(c *C) {
	pakPkgs, err := parsePakfile()
	expectedGitPkgs := []PakPkg{
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package1", "origin", "master")},
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package2", "origin", "dev")},
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package3", "origin", "dev")},
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package4", "nonorigin", "dev")},
	}

	c.Check(err, Equals, nil)
	c.Check(fmt.Sprintf("%s", pakPkgs), Equals, fmt.Sprintf("%s", expectedGitPkgs))
}

func (s *PakStateSuite) TestEqualPakfileNPakfileLock(c *C) {
	// Pakfile <===> Pakfile.lock
	sampleGitpkgs := []PakPkg{
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package1", "origin", "master")},
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package2", "origin", "dev")},
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package3", "origin", "dev")},
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package4", "nonorigin", "dev")},
	}
	samplePakLockInfo := PaklockInfo{
		"github.com/theplant/package1": "whoop-fake-checksum-hash",
		"github.com/theplant/package2": "whoop-fake-checksum-hash",
		"github.com/theplant/package3": "whoop-fake-checksum-hash",
		"github.com/theplant/package4": "whoop-fake-checksum-hash",
	}

	obtainedNewGps, obtainedToUpdateGps, obtainedToRemoteGps := ParsePakState(sampleGitpkgs, samplePakLockInfo)
	expectedToUpdateGps := []PakPkg{
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package1", "origin", "master")},
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package2", "origin", "dev")},
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package3", "origin", "dev")},
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package4", "nonorigin", "dev")},
	}
	for i := 0; i < 4; i++ {
		expectedToUpdateGps[i].Checksum = "whoop-fake-checksum-hash"
	}

	c.Check(fmt.Sprintf("%s", obtainedNewGps), Equals, fmt.Sprintf("%s", []PakPkg{}))
	c.Check(fmt.Sprintf("%s", obtainedToUpdateGps), Equals, fmt.Sprintf("%s", expectedToUpdateGps))
	c.Check(fmt.Sprintf("%s", obtainedToRemoteGps), Equals, fmt.Sprintf("%s", []PakPkg{}))
}

func (s *PakStateSuite) TestAddNewPkgs(c *C) {
	sampleGitpkgs := []PakPkg{
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package1", "origin", "master")},
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package2", "origin", "dev")},
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package3", "origin", "dev")},
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package4", "nonorigin", "dev")},
	}
	samplePakLockInfo := PaklockInfo{
		"github.com/theplant/package1": "whoop-fake-checksum-hash",
		"github.com/theplant/package2": "whoop-fake-checksum-hash",
		"github.com/theplant/package3": "whoop-fake-checksum-hash",
	}

	obtainedNewGps, obtainedToUpdateGps, obtainedToRemoteGps := ParsePakState(sampleGitpkgs, samplePakLockInfo)
	expectedToUpdateGps := []PakPkg{
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package1", "origin", "master")},
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package2", "origin", "dev")},
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package3", "origin", "dev")},
	}
	for i := 0; i < 3; i++ {
		expectedToUpdateGps[i].Checksum = "whoop-fake-checksum-hash"
	}
	expectedNewGps := []PakPkg{{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package4", "nonorigin", "dev")}}

	c.Check(fmt.Sprintf("%s", obtainedNewGps), Equals, fmt.Sprintf("%s", expectedNewGps))
	c.Check(fmt.Sprintf("%s", obtainedToUpdateGps), Equals, fmt.Sprintf("%s", expectedToUpdateGps))
	c.Check(fmt.Sprintf("%s", obtainedToRemoteGps), Equals, fmt.Sprintf("%s", []PakPkg{}))
}

func (s *PakStateSuite) TestRemovePkgs(c *C) {
	sampleGitpkgs := []PakPkg{
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package1", "origin", "master")},
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package2", "origin", "dev")},
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package3", "origin", "dev")},
	}
	samplePakLockInfo := PaklockInfo{
		"github.com/theplant/package1": "whoop-fake-checksum-hash",
		"github.com/theplant/package2": "whoop-fake-checksum-hash",
		"github.com/theplant/package3": "whoop-fake-checksum-hash",
		"github.com/theplant/package4": "whoop-fake-checksum-hash",
	}

	obtainedNewGps, obtainedToUpdateGps, obtainedToRemoteGps := ParsePakState(sampleGitpkgs, samplePakLockInfo)
	expectedToUpdateGps := []PakPkg{
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package1", "origin", "master")},
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package2", "origin", "dev")},
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package3", "origin", "dev")},
	}
	for i := 0; i < 3; i++ {
		expectedToUpdateGps[i].Checksum = "whoop-fake-checksum-hash"
	}
	expectedToRemoveGps := []PakPkg{{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package4", "", "")}}
	expectedToRemoveGps[0].Checksum = "whoop-fake-checksum-hash"

	c.Check(fmt.Sprintf("%s", obtainedNewGps), Equals, fmt.Sprintf("%s", []PakPkg{}))
	c.Check(fmt.Sprintf("%s", obtainedToUpdateGps), Equals, fmt.Sprintf("%s", expectedToUpdateGps))
	c.Check(fmt.Sprintf("%s", obtainedToRemoteGps), Equals, fmt.Sprintf("%s", expectedToRemoveGps))
}

func (s *PakStateSuite) TestNewUpdateRemovePkgs(c *C) {
	sampleGitpkgs := []PakPkg{
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package1", "origin", "master")},
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package2", "origin", "dev")},
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package3", "origin", "dev")},
	}
	samplePakLockInfo := PaklockInfo{
		"github.com/theplant/package1": "whoop-fake-checksum-hash",
		"github.com/theplant/package2": "whoop-fake-checksum-hash",
		"github.com/theplant/package4": "whoop-fake-checksum-hash",
	}

	obtainedNewGps, obtainedToUpdateGps, obtainedToRemoteGps := ParsePakState(sampleGitpkgs, samplePakLockInfo)
	expectedNewGps := []PakPkg{{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package3", "origin", "dev")}}
	expectedToUpdateGps := []PakPkg{
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package1", "origin", "master")},
		{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package2", "origin", "dev")},
	}
	for i := 0; i < 2; i++ {
		expectedToUpdateGps[i].Checksum = "whoop-fake-checksum-hash"
	}
	expectedToRemoveGps := []PakPkg{{GitPkg: gitpkg.NewGitPkg("github.com/theplant/package4", "", "")}}
	expectedToRemoveGps[0].Checksum = "whoop-fake-checksum-hash"

	c.Check(fmt.Sprintf("%s", obtainedNewGps), Equals, fmt.Sprintf("%s", expectedNewGps))
	c.Check(fmt.Sprintf("%s", obtainedToUpdateGps), Equals, fmt.Sprintf("%s", expectedToUpdateGps))
	c.Check(fmt.Sprintf("%s", obtainedToRemoteGps), Equals, fmt.Sprintf("%s", expectedToRemoveGps))
}
