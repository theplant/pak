package gitpkg

import (
	"fmt"
	// . "github.com/theplant/pak/share"
	. "launchpad.net/gocheck"
	"os/exec"
)

type GitPkgSuite struct{}

var _ = Suite(&GitPkgSuite{})

var testGitPkg *GitPkg
var testGpMasterChecksum = "11b174bd5acbf990687e6b068c97378d3219de04"

func (s *GitPkgSuite) SetUpTest(c *C) {
	testGitPkg = NewGitPkg("github.com/theplant/gitpkg-test-package", "origin", "master").(*GitPkg)
	var err error
	err = exec.Command("git", "clone", "fixtures/package1", "../../gitpkg-test-package").Run()
	if err != nil {
		panic(fmt.Errorf("GitPkgSuite.SetUpSuite: %s", err.Error()))
	}
}

func (s *GitPkgSuite) TearDownTest(c *C) {
	var err error
	err = exec.Command("rm", "-rf", "../../gitpkg-test-package").Run()
	if err != nil {
		panic(err)
	}
}

func (s *GitPkgSuite) TestClean(c *C) {
	// Clean
	clean, _ := testGitPkg.IsClean()
	c.Check(clean, Equals, true)

	// Not Clean
	var err error
	err = exec.Command("mv", "../../gitpkg-test-package/file3", "../../gitpkg-test-package/file3m").Run()
	if err != nil {
		panic(err)
	}

	clean, _ = testGitPkg.IsClean()
	c.Check(clean, Equals, false)
}

func (s *GitPkgSuite) TestCleanOnNoBranch(c *C) {
	// Clean
	clean, _ := testGitPkg.IsClean()
	c.Check(clean, Equals, true)

	// Not Clean
	_, err := testGitPkg.Git("checkout", "origin/master")
	if err != nil {
		panic(err)
	}

	clean, err = testGitPkg.IsClean()
	c.Check(err, Equals, nil)
	c.Check(clean, Equals, true)
}

func (s *GitPkgSuite) TestGetHeadRefName(c *C) {
	name, err := testGitPkg.GetHeadRefName()
	c.Check(err, Equals, nil)
	c.Check(name, Equals, "refs/heads/master")

	testGitPkg.Git("checkout", "refs/remotes/origin/master")
	name, err = testGitPkg.GetHeadRefName()
	c.Check(err, Equals, nil)
	c.Check(name, Equals, "no branch")
}

func (s *GitPkgSuite) TestGetChecksum(c *C) {
	checksum, _ := testGitPkg.GetChecksum("refs/heads/master")
	c.Check(checksum, Equals, testGpMasterChecksum)

	checksum, err := testGitPkg.GetChecksum("refs/heads/master_not_exist")
	c.Check(checksum, Equals, "")
	c.Check(err, Not(Equals), nil)
}

func (s *GitPkgSuite) TestFetch(c *C) {
	headChecksum, _ := testGitPkg.GetHeadChecksum()
	c.Check(headChecksum, Equals, testGpMasterChecksum)

	err := exec.Command("cp", "-r", "fixtures/package1", "fixtures/package1-backup").Run()
	if err != nil {
		c.Error(err)
	}
	err = exec.Command("git", "--git-dir=fixtures/updated-package1/.git", "--work-tree=fixtures/updated-package1", "push", "origin", "master").Run()
	if err != nil {
		c.Error(err)
	}

	c.Check(testGitPkg.Fetch(), Equals, nil)

	headChecksum, _ = testGitPkg.GetChecksum("refs/remotes/origin/master")
	c.Check(headChecksum, Equals, "149b1e18aa3e118a804ccddeefbb6e64b4a5807e")

	err = exec.Command("rm", "-rf", "fixtures/package1").Run()
	if err != nil {
		c.Error(err)
	}
	err = exec.Command("mv", "fixtures/package1-backup", "fixtures/package1").Run()
	if err != nil {
		c.Error(err)
	}
}

func (s *GitPkgSuite) TestContainsRemotebranch(c *C) {
	contained, err := testGitPkg.ContainsRemoteBranch()
	c.Check(err, Equals, nil)
	c.Check(contained, Equals, true)
}

func (s *GitPkgSuite) TestIsTracking(c *C) {
	tracking, err := IsTracking("github.com/theplant/gitpkg-test-package")
	c.Check(err, Equals, nil)
	c.Check(tracking, Equals, true)
}

func (s *GitPkgSuite) TestPak(c *C) {
	checksum, err := testGitPkg.Pak("refs/remotes/origin/master")
	c.Check(err, Equals, nil)
	c.Check(checksum, Equals, testGpMasterChecksum)

	contained, err := testGitPkg.ContainsPakbranch()
	c.Check(err, Equals, nil)
	c.Check(contained, Equals, true)
}

func (s *GitPkgSuite) TestUnpak(c *C) {
	_, err := testGitPkg.Pak("refs/remotes/origin/master")
	c.Check(err, Equals, nil)

	contained, err := testGitPkg.ContainsPakbranch()
	c.Check(err, Equals, nil)
	c.Check(contained, Equals, true)

	err = testGitPkg.Unpak()
	c.Check(err, Equals, nil)

	contained, err = testGitPkg.ContainsPakbranch()
	c.Check(err, Equals, nil)
	c.Check(contained, Equals, false)
}
