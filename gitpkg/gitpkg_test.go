package gitpkg

import (
	"fmt"
	. "github.com/theplant/pak/share"
	. "launchpad.net/gocheck"
	"os/exec"
)

type GitPkgSuite struct{}

var _ = Suite(&GitPkgSuite{})

var testGitPkg GitPkg
var testGpMasterChecksum = "11b174bd5acbf990687e6b068c97378d3219de04"

func (s *GitPkgSuite) SetUpTest(c *C) {
	testGitPkg = NewGitPkg("github.com/theplant/gitpkg-test-package", "origin", "master")
	var err error
	err = exec.Command("git", "clone", "fixtures/package1", "../../gitpkg-test-package").Run()
	if err != nil {
		panic(fmt.Errorf("GitPkgSuite.SetUpSuite: %s", err.Error()))
	}

	err = testGitPkg.Sync()
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
	exist, err := testGitPkg.IsPkgExist()
	c.Check(err, Equals, nil)
	c.Check(exist, Equals, true)

	exec.Command("rm", "-rf", "../../gitpkg-test-package").Run()

	exist, err = testGitPkg.IsPkgExist()
	c.Check(err, Equals, nil)
	c.Check(exist, Equals, false)
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

func (s *GitPkgSuite) TestSimplePak(c *C) {
	testGitPkg.Sync()

	c.Check(testGitPkg.HeadRefsName, Equals, "refs/heads/master")
	c.Check(testGitPkg.HeadChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.PakbranchChecksum, Equals, "")
	c.Check(testGitPkg.PaktagChecksum, Equals, "")
	c.Check(testGitPkg.State.ContainsBranchNamedPak, Equals, false)
	c.Check(testGitPkg.State.ContainsPaktag, Equals, false)
	// c.Check(testGitPkg.State.UnderPak, Equals, false)
	c.Check(testGitPkg.State.OnPakbranch, Equals, false)
	c.Check(testGitPkg.State.IsRemoteBranchExist, Equals, true)
	c.Check(testGitPkg.State.IsClean, Equals, true)

	testGitPkg.Pak(GetOption{true, ""})
	testGitPkg.Sync()

	c.Check(testGitPkg.HeadRefsName, Equals, "refs/heads/pak")
	c.Check(testGitPkg.HeadChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.PakbranchChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.PaktagChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.State.ContainsBranchNamedPak, Equals, true)
	c.Check(testGitPkg.State.ContainsPaktag, Equals, true)
	// c.Check(testGitPkg.State.UnderPak, Equals, true)
	c.Check(testGitPkg.State.OnPakbranch, Equals, true)
	c.Check(testGitPkg.State.IsRemoteBranchExist, Equals, true)
	c.Check(testGitPkg.State.IsClean, Equals, true)
}

func (s *GitPkgSuite) TestWeakPak(c *C) {
	exec.Command("git", testGitPkg.WorkTree, testGitPkg.GitDir, "branch", "pak").Run()
	testGitPkg.Sync()
	c.Check(testGitPkg.HeadRefsName, Equals, "refs/heads/master")
	c.Check(testGitPkg.HeadChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.PakbranchChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.PaktagChecksum, Equals, "")
	c.Check(testGitPkg.State.ContainsBranchNamedPak, Equals, true)
	c.Check(testGitPkg.State.ContainsPaktag, Equals, false)
	// c.Check(testGitPkg.State.UnderPak, Equals, false)
	c.Check(testGitPkg.State.OnPakbranch, Equals, false)
	c.Check(testGitPkg.State.IsRemoteBranchExist, Equals, true)
	c.Check(testGitPkg.State.IsClean, Equals, true)

	testGitPkg.Pak(GetOption{false, ""})
	testGitPkg.Sync()

	c.Check(testGitPkg.HeadRefsName, Equals, "refs/heads/master")
	c.Check(testGitPkg.HeadChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.PakbranchChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.PaktagChecksum, Equals, "")
	c.Check(testGitPkg.State.ContainsBranchNamedPak, Equals, true)
	c.Check(testGitPkg.State.ContainsPaktag, Equals, false)
	// c.Check(testGitPkg.State.UnderPak, Equals, false)
	c.Check(testGitPkg.State.OnPakbranch, Equals, false)
	c.Check(testGitPkg.State.IsRemoteBranchExist, Equals, true)
	c.Check(testGitPkg.State.IsClean, Equals, true)
}

func (s *GitPkgSuite) TestForcefulPak(c *C) {
	exec.Command("git", testGitPkg.WorkTree, testGitPkg.GitDir, "branch", "pak").Run()
	testGitPkg.Sync()

	testGitPkg.Pak(GetOption{true, ""})
	testGitPkg.Sync()

	c.Check(testGitPkg.HeadRefsName, Equals, "refs/heads/pak")
	c.Check(testGitPkg.HeadChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.PakbranchChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.PaktagChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.State.ContainsBranchNamedPak, Equals, true)
	c.Check(testGitPkg.State.ContainsPaktag, Equals, true)
	// c.Check(testGitPkg.State.UnderPak, Equals, true)
	c.Check(testGitPkg.State.OnPakbranch, Equals, true)
	c.Check(testGitPkg.State.IsRemoteBranchExist, Equals, true)
	c.Check(testGitPkg.State.IsClean, Equals, true)
}

func (s *GitPkgSuite) TestGetWithChecksum(c *C) {
	devChecksum := "711c1e206bca5ad99edf6da12074bbbe4a349932"
	testGitPkg.Sync()
	testGitPkg.Pak(GetOption{true, devChecksum})
	testGitPkg.Sync()

	c.Check(testGitPkg.HeadRefsName, Equals, "refs/heads/pak")
	c.Check(testGitPkg.HeadChecksum, Equals, devChecksum)
	c.Check(testGitPkg.PakbranchChecksum, Equals, devChecksum)
	c.Check(testGitPkg.PaktagChecksum, Equals, devChecksum)
	c.Check(testGitPkg.State.ContainsBranchNamedPak, Equals, true)
	c.Check(testGitPkg.State.ContainsPaktag, Equals, true)
	// c.Check(testGitPkg.State.UnderPak, Equals, true)
	c.Check(testGitPkg.State.OnPakbranch, Equals, true)
	c.Check(testGitPkg.State.IsRemoteBranchExist, Equals, true)
	c.Check(testGitPkg.State.IsClean, Equals, true)
}

func (s *GitPkgSuite) TestForcefulUnpak(c *C) {
	testGitPkg.Sync()
	testGitPkg.Pak(GetOption{true, testGpMasterChecksum})
	testGitPkg.Sync()
	testGitPkg.Unpak(true)
	testGitPkg.Sync()

	c.Check(testGitPkg.HeadRefsName, Equals, "refs/heads/master")
	c.Check(testGitPkg.HeadChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.PakbranchChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.PaktagChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.State.ContainsBranchNamedPak, Equals, false)
	c.Check(testGitPkg.State.ContainsPaktag, Equals, false)
	// c.Check(testGitPkg.State.UnderPak, Equals, false)
	c.Check(testGitPkg.State.OnPakbranch, Equals, false)
	c.Check(testGitPkg.State.IsRemoteBranchExist, Equals, true)
	c.Check(testGitPkg.State.IsClean, Equals, true)
}

func (s *GitPkgSuite) TestWeakUnpak(c *C) {
	exec.Command("git", testGitPkg.WorkTree, testGitPkg.GitDir, "branch", "pak").Run()
	testGitPkg.Sync()
	testGitPkg.Unpak(false)
	testGitPkg.Sync()

	c.Check(testGitPkg.HeadRefsName, Equals, "refs/heads/master")
	c.Check(testGitPkg.HeadChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.PakbranchChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.PaktagChecksum, Equals, "")
	c.Check(testGitPkg.State.ContainsBranchNamedPak, Equals, true)
	c.Check(testGitPkg.State.ContainsPaktag, Equals, false)
	// c.Check(testGitPkg.State.UnderPak, Equals, false)
	c.Check(testGitPkg.State.OnPakbranch, Equals, false)
	c.Check(testGitPkg.State.IsRemoteBranchExist, Equals, true)
	c.Check(testGitPkg.State.IsClean, Equals, true)
}

func (s *GitPkgSuite) TestPakBranchOwnership(c *C) {
	testGitPkg.Unpak(true)

	testGitPkg.Sync()
	c.Check(testGitPkg.State.ContainsBranchNamedPak, Equals, false)
	c.Check(testGitPkg.State.ContainsPaktag, Equals, false)
	c.Check(testGitPkg.State.OwnPakbranch, Equals, false)

	exec.Command("sh", "-c", "cd ../../gitpkg-test-package && git checkout -b pak && git checkout master").Run()
	testGitPkg.Sync()
	c.Check(testGitPkg.State.ContainsBranchNamedPak, Equals, true)
	c.Check(testGitPkg.State.ContainsPaktag, Equals, false)
	c.Check(testGitPkg.State.OwnPakbranch, Equals, false)

	// Although package has both pak branch and _pak_latest_ tag, but their commit hash are different, so
	// Pak wouldn't consider that pak branch is owned by it.
	exec.Command("sh", "-c", "cd ../../gitpkg-test-package && git tag _pak_latest_ HEAD^").Run()
	testGitPkg.Sync()
	c.Check(testGitPkg.State.ContainsBranchNamedPak, Equals, true)
	c.Check(testGitPkg.State.ContainsPaktag, Equals, true)
	c.Check(testGitPkg.State.OwnPakbranch, Equals, false)

	exec.Command("sh", "-c", "cd ../../gitpkg-test-package && git tag -d _pak_latest_ && git tag _pak_latest_").Run()
	testGitPkg.Sync()
	c.Check(testGitPkg.State.ContainsBranchNamedPak, Equals, true)
	c.Check(testGitPkg.State.ContainsPaktag, Equals, true)
	c.Check(testGitPkg.State.OwnPakbranch, Equals, true)
}
