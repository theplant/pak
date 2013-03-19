package pak

import (
	// "io/ioutil"
	. "launchpad.net/gocheck"
	// "launchpad.net/goyaml"
	// "os"
	"os/exec"
	// "testing"
	"fmt"
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
	gitPkgs, err := parsePakfile()
	expectedGitPkgs := []GitPkg{
		NewGitPkg("github.com/theplant/package1", "origin", "master"),
		NewGitPkg("github.com/theplant/package2", "origin", "dev"),
		NewGitPkg("github.com/theplant/package3", "origin", "dev"),
		NewGitPkg("github.com/theplant/package4", "nonorigin", "dev"),
	}

	c.Check(err, Equals, nil)
	c.Check(fmt.Sprintf("%s", gitPkgs), Equals, fmt.Sprintf("%s", expectedGitPkgs))
}

func (s *PakStateSuite) TestEqualPakfileNPakfileLock(c *C) {
	// Pakfile <===> Pakfile.lock
	sampleGitpkgs := []GitPkg{
		NewGitPkg("github.com/theplant/package1", "origin", "master"),
		NewGitPkg("github.com/theplant/package2", "origin", "dev"),
		NewGitPkg("github.com/theplant/package3", "origin", "dev"),
		NewGitPkg("github.com/theplant/package4", "nonorigin", "dev"),
	}
	samplePakLockInfo := PaklockInfo{
		"github.com/theplant/package1": "whoop-fake-checksum-hash",
		"github.com/theplant/package2": "whoop-fake-checksum-hash",
		"github.com/theplant/package3": "whoop-fake-checksum-hash",
		"github.com/theplant/package4": "whoop-fake-checksum-hash",
	}

	obtainedNewGps, obtainedToUpdateGps, obtainedToRemoteGps := ParsePakState(sampleGitpkgs, samplePakLockInfo)
	expectedToUpdateGps := []GitPkg{
		NewGitPkg("github.com/theplant/package1", "origin", "master"),
		NewGitPkg("github.com/theplant/package2", "origin", "dev"),
		NewGitPkg("github.com/theplant/package3", "origin", "dev"),
		NewGitPkg("github.com/theplant/package4", "nonorigin", "dev"),
	}
	for i := 0; i < 4; i++ {
		expectedToUpdateGps[i].Checksum = "whoop-fake-checksum-hash"
	}

	c.Check(fmt.Sprintf("%s", obtainedNewGps), Equals, fmt.Sprintf("%s", []GitPkg{}))
	c.Check(fmt.Sprintf("%s", obtainedToUpdateGps), Equals, fmt.Sprintf("%s", expectedToUpdateGps))
	c.Check(fmt.Sprintf("%s", obtainedToRemoteGps), Equals, fmt.Sprintf("%s", []GitPkg{}))
}

func (s *PakStateSuite) TestAddNewPkgs(c *C) {
	sampleGitpkgs := []GitPkg{
		NewGitPkg("github.com/theplant/package1", "origin", "master"),
		NewGitPkg("github.com/theplant/package2", "origin", "dev"),
		NewGitPkg("github.com/theplant/package3", "origin", "dev"),
		NewGitPkg("github.com/theplant/package4", "nonorigin", "dev"),
	}
	samplePakLockInfo := PaklockInfo{
		"github.com/theplant/package1": "whoop-fake-checksum-hash",
		"github.com/theplant/package2": "whoop-fake-checksum-hash",
		"github.com/theplant/package3": "whoop-fake-checksum-hash",
	}

	obtainedNewGps, obtainedToUpdateGps, obtainedToRemoteGps := ParsePakState(sampleGitpkgs, samplePakLockInfo)
	expectedToUpdateGps := []GitPkg{
		NewGitPkg("github.com/theplant/package1", "origin", "master"),
		NewGitPkg("github.com/theplant/package2", "origin", "dev"),
		NewGitPkg("github.com/theplant/package3", "origin", "dev"),
	}
	for i := 0; i < 3; i++ {
		expectedToUpdateGps[i].Checksum = "whoop-fake-checksum-hash"
	}
	expectedNewGps := []GitPkg{NewGitPkg("github.com/theplant/package4", "nonorigin", "dev")}

	c.Check(fmt.Sprintf("%s", obtainedNewGps), Equals, fmt.Sprintf("%s", expectedNewGps))
	c.Check(fmt.Sprintf("%s", obtainedToUpdateGps), Equals, fmt.Sprintf("%s", expectedToUpdateGps))
	c.Check(fmt.Sprintf("%s", obtainedToRemoteGps), Equals, fmt.Sprintf("%s", []GitPkg{}))
}

func (s *PakStateSuite) TestRemovePkgs(c *C) {
	sampleGitpkgs := []GitPkg{
		NewGitPkg("github.com/theplant/package1", "origin", "master"),
		NewGitPkg("github.com/theplant/package2", "origin", "dev"),
		NewGitPkg("github.com/theplant/package3", "origin", "dev"),
	}
	samplePakLockInfo := PaklockInfo{
		"github.com/theplant/package1": "whoop-fake-checksum-hash",
		"github.com/theplant/package2": "whoop-fake-checksum-hash",
		"github.com/theplant/package3": "whoop-fake-checksum-hash",
		"github.com/theplant/package4": "whoop-fake-checksum-hash",
	}

	obtainedNewGps, obtainedToUpdateGps, obtainedToRemoteGps := ParsePakState(sampleGitpkgs, samplePakLockInfo)
	expectedToUpdateGps := []GitPkg{
		NewGitPkg("github.com/theplant/package1", "origin", "master"),
		NewGitPkg("github.com/theplant/package2", "origin", "dev"),
		NewGitPkg("github.com/theplant/package3", "origin", "dev"),
	}
	for i := 0; i < 3; i++ {
		expectedToUpdateGps[i].Checksum = "whoop-fake-checksum-hash"
	}
	expectedToRemoveGps := []GitPkg{NewGitPkg("github.com/theplant/package4", "", "")}
	expectedToRemoveGps[0].Checksum = "whoop-fake-checksum-hash"

	c.Check(fmt.Sprintf("%s", obtainedNewGps), Equals, fmt.Sprintf("%s", []GitPkg{}))
	c.Check(fmt.Sprintf("%s", obtainedToUpdateGps), Equals, fmt.Sprintf("%s", expectedToUpdateGps))
	c.Check(fmt.Sprintf("%s", obtainedToRemoteGps), Equals, fmt.Sprintf("%s", expectedToRemoveGps))
}

func (s *PakStateSuite) TestNewUpdateRemovePkgs(c *C) {
	sampleGitpkgs := []GitPkg{
		NewGitPkg("github.com/theplant/package1", "origin", "master"),
		NewGitPkg("github.com/theplant/package2", "origin", "dev"),
		NewGitPkg("github.com/theplant/package3", "origin", "dev"),
	}
	samplePakLockInfo := PaklockInfo{
		"github.com/theplant/package1": "whoop-fake-checksum-hash",
		"github.com/theplant/package2": "whoop-fake-checksum-hash",
		"github.com/theplant/package4": "whoop-fake-checksum-hash",
	}

	obtainedNewGps, obtainedToUpdateGps, obtainedToRemoteGps := ParsePakState(sampleGitpkgs, samplePakLockInfo)
	expectedNewGps := []GitPkg{NewGitPkg("github.com/theplant/package3", "origin", "dev")}
	expectedToUpdateGps := []GitPkg{
		NewGitPkg("github.com/theplant/package1", "origin", "master"),
		NewGitPkg("github.com/theplant/package2", "origin", "dev"),
	}
	for i := 0; i < 2; i++ {
		expectedToUpdateGps[i].Checksum = "whoop-fake-checksum-hash"
	}
	expectedToRemoveGps := []GitPkg{NewGitPkg("github.com/theplant/package4", "", "")}
	expectedToRemoveGps[0].Checksum = "whoop-fake-checksum-hash"

	c.Check(fmt.Sprintf("%s", obtainedNewGps), Equals, fmt.Sprintf("%s", expectedNewGps))
	c.Check(fmt.Sprintf("%s", obtainedToUpdateGps), Equals, fmt.Sprintf("%s", expectedToUpdateGps))
	c.Check(fmt.Sprintf("%s", obtainedToRemoteGps), Equals, fmt.Sprintf("%s", expectedToRemoveGps))
}
