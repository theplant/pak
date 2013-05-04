package gitpkg

import (
	"fmt"
	. "github.com/theplant/pak/share"
	. "launchpad.net/gocheck"
	"os/exec"
	"testing"
)

// Hook up gocheck into the gotest runner.
func Test(t *testing.T) { TestingT(t) }

var fp = fmt.Printf
var fl = fmt.Println

type GitPkgSuite struct{}

var _ = Suite(&GitPkgSuite{})

var testGitPkg *GitPkg
var testGpMasterChecksum = "11b174bd5acbf990687e6b068c97378d3219de04"

func (s *GitPkgSuite) SetUpTest(c *C) {
	err := exec.Command("git", "clone", "fixtures/package1", "../../gitpkg-test-package").Run()
	if err != nil {
		panic(fmt.Errorf("GitPkgSuite.SetUpSuite: %s", err.Error()))
	}

	testGitPkg = NewGitPkg("github.com/theplant/gitpkg-test-package", "origin", "master").(*GitPkg)
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
	headChecksum, err := testGitPkg.GetHeadChecksum()
	c.Check(err, Equals, nil)
	c.Check(headChecksum, Equals, testGpMasterChecksum)

	MustRun("cp", "-r", "fixtures/package1", "fixtures/package1-backup")
	MustRun("sh", "-c", "git clone fixtures/package1 ../../package1-for-fetch; cd ../../package1-for-fetch; touch file4; git add file4; git commit -m 'file4'; git push")
	// MustRun("sh", "-c", "cd fixtures/updated-package1; git push")
	defer func() {
		MustRun("rm", "-rf", "fixtures/package1")
		MustRun("rm", "-rf", "../../package1-for-fetch")
		MustRun("mv", "fixtures/package1-backup", "fixtures/package1")
	}()

	err = testGitPkg.Fetch()
	c.Check(err, Equals, nil)

	headChecksum, err = testGitPkg.GetChecksum("refs/remotes/origin/master")
	c.Check(err, Equals, nil)
	c.Check(headChecksum, Equals, "149b1e18aa3e118a804ccddeefbb6e64b4a5807e")
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

	MustRun("mkdir", "-p", "../../git-istracking-test/sub-package/.git")
	MustRun("mkdir", "-p", "../../git-istracking-test2/package")
	defer func() {
		MustRun("rm", "-r", "../../git-istracking-test")
		MustRun("rm", "-r", "../../git-istracking-test2")
	}()

	tracking, err = IsTracking("github.com/theplant/git-istracking-test/sub-package")
	c.Check(err, Equals, nil)
	c.Check(tracking, Equals, true)

	tracking, err = IsTracking("github.com/theplant/git-istracking-test2/package")
	c.Check(err, Equals, nil)
	c.Check(tracking, Equals, false)
}

func (s *GitPkgSuite) TestGetPkgRoot(c *C) {
	tracking, err2 := IsTracking("github.com/theplant/gitpkg-test-package")
	c.Check(err2, Equals, nil)
	c.Check(tracking, Equals, true)

	pkgName, err := GetPkgRoot("github.com/theplant/gitpkg-test-package")
	c.Check(err, Equals, nil)
	c.Check(pkgName, Equals, "github.com/theplant/gitpkg-test-package")

	MustRun("mkdir", "-p", "../../git-istracking-test/.git/ll")
	MustRun("mkdir", "-p", "../../git-istracking-test/sub-package/ll")
	MustRun("mkdir", "-p", "../../git-istracking-test2/package")
	defer func() {
		MustRun("rm", "-r", "../../git-istracking-test")
		MustRun("rm", "-r", "../../git-istracking-test2")
	}()

	pkgName, err = GetPkgRoot("github.com/theplant/git-istracking-test/sub-package")
	c.Check(err, Equals, nil)
	c.Check(pkgName, Equals, "github.com/theplant/git-istracking-test")

	pkgName, err = GetPkgRoot("github.com/theplant/git-istracking-test2/package")
	c.Check(err, Not(Equals), nil)
	c.Check(pkgName, Equals, "")
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
