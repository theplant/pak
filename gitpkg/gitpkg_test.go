package gitpkg

import (
	"fmt"
	. "github.com/theplant/pak/share"
	. "launchpad.net/gocheck"
	"os/exec"
)

type GitPkgSuite struct{}

var _ = Suite(&GitPkgSuite{})

var testGitPkg VCSPkg
var testGpMasterChecksum = "11b174bd5acbf990687e6b068c97378d3219de04"

func (s *GitPkgSuite) SetUpTest(c *C) {
	testGitPkg = NewGitPkg("github.com/theplant/gitpkg-test-package", "origin", "master")
	var err error
	err = exec.Command("git", "clone", "fixtures/package1", "../../gitpkg-test-package").Run()
	if err != nil {
		panic(fmt.Errorf("GitPkgSuite.SetUpSuite: %s", err.Error()))
	}

	err = testGitPkg.(*GitPkg).Sync()
	if err != nil {
		panic(err)
	}
}

func (s *GitPkgSuite) TearDownTest(c *C) {
	var err error
	err = exec.Command("rm", "-rf", "../../gitpkg-test-package").Run()
	if err != nil {
		panic(err)
	}
}

func (s *GitPkgSuite) TestIsPkgExist(c *C) {
	exist, err := testGitPkg.(*GitPkg).IsPkgExist()
	c.Check(err, Equals, nil)
	c.Check(exist, Equals, true)

	exec.Command("rm", "-rf", "../../gitpkg-test-package").Run()

	exist, err = testGitPkg.(*GitPkg).IsPkgExist()
	c.Check(err, Equals, nil)
	c.Check(exist, Equals, false)
}

func (s *GitPkgSuite) TestClean(c *C) {
	// Clean
	clean, _ := testGitPkg.(*GitPkg).IsClean()
	c.Check(clean, Equals, true)

	// Not Clean
	var err error
	err = exec.Command("mv", "../../gitpkg-test-package/file3", "../../gitpkg-test-package/file3m").Run()
	if err != nil {
		panic(err)
	}

	clean, _ = testGitPkg.(*GitPkg).IsClean()
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
	name, err := testGitPkg.(*GitPkg).GetHeadRefName()
	c.Check(err, Equals, nil)
	c.Check(name, Equals, "refs/heads/master")

	testGitPkg.Git("checkout", "refs/remotes/origin/master")
	name, err = testGitPkg.GetHeadRefName()
	c.Check(err, Equals, nil)
	c.Check(name, Equals, "no branch")
}

func (s *GitPkgSuite) TestGetChecksum(c *C) {
	checksum, _ := testGitPkg.(*GitPkg).GetChecksum("refs/heads/master")
	c.Check(checksum, Equals, testGpMasterChecksum)

	checksum, err := testGitPkg.(*GitPkg).GetChecksum("refs/heads/master_not_exist")
	c.Check(checksum, Equals, "")
	c.Check(err, Not(Equals), nil)
}

func (s *GitPkgSuite) TestFetch(c *C) {
	headChecksum, _ := testGitPkg.(*GitPkg).GetHeadChecksum()
	c.Check(headChecksum, Equals, testGpMasterChecksum)

	err := exec.Command("cp", "-r", "fixtures/package1", "fixtures/package1-backup").Run()
	if err != nil {
		c.Error(err)
	}
	err = exec.Command("git", "--git-dir=fixtures/updated-package1/.git", "--work-tree=fixtures/updated-package1", "push", "origin", "master").Run()
	if err != nil {
		c.Error(err)
	}

	c.Check(testGitPkg.(*GitPkg).Fetch(), Equals, nil)

	headChecksum, _ = testGitPkg.(*GitPkg).GetChecksum("refs/remotes/origin/master")
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

func (s *GitPkgSuite) TestSimplePak(c *C) {
	testGitPkg.(*GitPkg).Sync()

	c.Check(testGitPkg.(*GitPkg).HeadRefsName, Equals, "refs/heads/master")
	c.Check(testGitPkg.(*GitPkg).HeadChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.(*GitPkg).PakbranchChecksum, Equals, "")
	c.Check(testGitPkg.(*GitPkg).PaktagChecksum, Equals, "")
	c.Check(testGitPkg.(*GitPkg).State.ContainsBranchNamedPak, Equals, false)
	c.Check(testGitPkg.(*GitPkg).State.ContainsPaktag, Equals, false)
	// c.Check(testGitPkg.(*GitPkg).State.UnderPak, Equals, false)
	c.Check(testGitPkg.(*GitPkg).State.OnPakbranch, Equals, false)
	c.Check(testGitPkg.(*GitPkg).State.IsRemoteBranchExist, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.IsClean, Equals, true)

	testGitPkg.(*GitPkg).Pak(GetOption{true, ""})
	testGitPkg.(*GitPkg).Sync()

	c.Check(testGitPkg.(*GitPkg).HeadRefsName, Equals, "refs/heads/pak")
	c.Check(testGitPkg.(*GitPkg).HeadChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.(*GitPkg).PakbranchChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.(*GitPkg).PaktagChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.(*GitPkg).State.ContainsBranchNamedPak, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.ContainsPaktag, Equals, true)
	// c.Check(testGitPkg.(*GitPkg).State.UnderPak, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.OnPakbranch, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.IsRemoteBranchExist, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.IsClean, Equals, true)
}

func (s *GitPkgSuite) TestWeakPak(c *C) {
	exec.Command("git", testGitPkg.(*GitPkg).WorkTree, testGitPkg.(*GitPkg).GitDir, "branch", "pak").Run()
	testGitPkg.(*GitPkg).Sync()
	c.Check(testGitPkg.(*GitPkg).HeadRefsName, Equals, "refs/heads/master")
	c.Check(testGitPkg.(*GitPkg).HeadChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.(*GitPkg).PakbranchChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.(*GitPkg).PaktagChecksum, Equals, "")
	c.Check(testGitPkg.(*GitPkg).State.ContainsBranchNamedPak, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.ContainsPaktag, Equals, false)
	// c.Check(testGitPkg.(*GitPkg).State.UnderPak, Equals, false)
	c.Check(testGitPkg.(*GitPkg).State.OnPakbranch, Equals, false)
	c.Check(testGitPkg.(*GitPkg).State.IsRemoteBranchExist, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.IsClean, Equals, true)

	testGitPkg.(*GitPkg).Pak(GetOption{false, ""})
	testGitPkg.(*GitPkg).Sync()

	c.Check(testGitPkg.(*GitPkg).HeadRefsName, Equals, "refs/heads/master")
	c.Check(testGitPkg.(*GitPkg).HeadChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.(*GitPkg).PakbranchChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.(*GitPkg).PaktagChecksum, Equals, "")
	c.Check(testGitPkg.(*GitPkg).State.ContainsBranchNamedPak, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.ContainsPaktag, Equals, false)
	// c.Check(testGitPkg.(*GitPkg).State.UnderPak, Equals, false)
	c.Check(testGitPkg.(*GitPkg).State.OnPakbranch, Equals, false)
	c.Check(testGitPkg.(*GitPkg).State.IsRemoteBranchExist, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.IsClean, Equals, true)
}

func (s *GitPkgSuite) TestForcefulPak(c *C) {
	exec.Command("git", testGitPkg.(*GitPkg).WorkTree, testGitPkg.(*GitPkg).GitDir, "branch", "pak").Run()
	testGitPkg.(*GitPkg).Sync()

	testGitPkg.(*GitPkg).Pak(GetOption{true, ""})
	testGitPkg.(*GitPkg).Sync()

	c.Check(testGitPkg.(*GitPkg).HeadRefsName, Equals, "refs/heads/pak")
	c.Check(testGitPkg.(*GitPkg).HeadChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.(*GitPkg).PakbranchChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.(*GitPkg).PaktagChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.(*GitPkg).State.ContainsBranchNamedPak, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.ContainsPaktag, Equals, true)
	// c.Check(testGitPkg.(*GitPkg).State.UnderPak, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.OnPakbranch, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.IsRemoteBranchExist, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.IsClean, Equals, true)
}

func (s *GitPkgSuite) TestGetWithChecksum(c *C) {
	devChecksum := "711c1e206bca5ad99edf6da12074bbbe4a349932"
	testGitPkg.(*GitPkg).Sync()
	testGitPkg.(*GitPkg).Pak(GetOption{true, devChecksum})
	testGitPkg.(*GitPkg).Sync()

	c.Check(testGitPkg.(*GitPkg).HeadRefsName, Equals, "refs/heads/pak")
	c.Check(testGitPkg.(*GitPkg).HeadChecksum, Equals, devChecksum)
	c.Check(testGitPkg.(*GitPkg).PakbranchChecksum, Equals, devChecksum)
	c.Check(testGitPkg.(*GitPkg).PaktagChecksum, Equals, devChecksum)
	c.Check(testGitPkg.(*GitPkg).State.ContainsBranchNamedPak, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.ContainsPaktag, Equals, true)
	// c.Check(testGitPkg.(*GitPkg).State.UnderPak, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.OnPakbranch, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.IsRemoteBranchExist, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.IsClean, Equals, true)
}

func (s *GitPkgSuite) TestForcefulUnpak(c *C) {
	testGitPkg.(*GitPkg).Sync()
	testGitPkg.(*GitPkg).Pak(GetOption{true, testGpMasterChecksum})
	testGitPkg.(*GitPkg).Sync()
	testGitPkg.(*GitPkg).Unpak(true)
	testGitPkg.(*GitPkg).Sync()

	c.Check(testGitPkg.(*GitPkg).HeadRefsName, Equals, "refs/heads/master")
	c.Check(testGitPkg.(*GitPkg).HeadChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.(*GitPkg).PakbranchChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.(*GitPkg).PaktagChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.(*GitPkg).State.ContainsBranchNamedPak, Equals, false)
	c.Check(testGitPkg.(*GitPkg).State.ContainsPaktag, Equals, false)
	// c.Check(testGitPkg.(*GitPkg).State.UnderPak, Equals, false)
	c.Check(testGitPkg.(*GitPkg).State.OnPakbranch, Equals, false)
	c.Check(testGitPkg.(*GitPkg).State.IsRemoteBranchExist, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.IsClean, Equals, true)
}

func (s *GitPkgSuite) TestWeakUnpak(c *C) {
	exec.Command("git", testGitPkg.(*GitPkg).WorkTree, testGitPkg.(*GitPkg).GitDir, "branch", "pak").Run()
	testGitPkg.(*GitPkg).Sync()
	testGitPkg.(*GitPkg).Unpak(false)
	testGitPkg.(*GitPkg).Sync()

	c.Check(testGitPkg.(*GitPkg).HeadRefsName, Equals, "refs/heads/master")
	c.Check(testGitPkg.(*GitPkg).HeadChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.(*GitPkg).PakbranchChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.(*GitPkg).PaktagChecksum, Equals, "")
	c.Check(testGitPkg.(*GitPkg).State.ContainsBranchNamedPak, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.ContainsPaktag, Equals, false)
	// c.Check(testGitPkg.(*GitPkg).State.UnderPak, Equals, false)
	c.Check(testGitPkg.(*GitPkg).State.OnPakbranch, Equals, false)
	c.Check(testGitPkg.(*GitPkg).State.IsRemoteBranchExist, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.IsClean, Equals, true)
}

func (s *GitPkgSuite) TestPakBranchOwnership(c *C) {
	testGitPkg.(*GitPkg).Unpak(true)

	testGitPkg.(*GitPkg).Sync()
	c.Check(testGitPkg.(*GitPkg).State.ContainsBranchNamedPak, Equals, false)
	c.Check(testGitPkg.(*GitPkg).State.ContainsPaktag, Equals, false)
	c.Check(testGitPkg.(*GitPkg).State.OwnPakbranch, Equals, false)

	exec.Command("sh", "-c", "cd ../../gitpkg-test-package && git checkout -b pak && git checkout master").Run()
	testGitPkg.(*GitPkg).Sync()
	c.Check(testGitPkg.(*GitPkg).State.ContainsBranchNamedPak, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.ContainsPaktag, Equals, false)
	c.Check(testGitPkg.(*GitPkg).State.OwnPakbranch, Equals, false)

	// Although package has both pak branch and _pak_latest_ tag, but their commit hash are different, so
	// Pak wouldn't consider that pak branch is owned by it.
	exec.Command("sh", "-c", "cd ../../gitpkg-test-package && git tag _pak_latest_ HEAD^").Run()
	testGitPkg.(*GitPkg).Sync()
	c.Check(testGitPkg.(*GitPkg).State.ContainsBranchNamedPak, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.ContainsPaktag, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.OwnPakbranch, Equals, false)

	exec.Command("sh", "-c", "cd ../../gitpkg-test-package && git tag -d _pak_latest_ && git tag _pak_latest_").Run()
	testGitPkg.(*GitPkg).Sync()
	c.Check(testGitPkg.(*GitPkg).State.ContainsBranchNamedPak, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.ContainsPaktag, Equals, true)
	c.Check(testGitPkg.(*GitPkg).State.OwnPakbranch, Equals, true)
}
