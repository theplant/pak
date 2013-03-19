package gitpkg

import(
	"os/exec"
	"fmt"
	. "launchpad.net/gocheck"
	. "github.com/theplant/pak/share"
)

type GitPkgSuite struct{}
var _ = Suite(&GitPkgSuite{})

var testGitPkg GitPkg
var testGpMasterChecksum = "11b174bd5acbf990687e6b068c97378d3219de04"

func (s *GitPkgSuite) SetUpTest(c *C) {
	testGitPkg = NewGitPkg("github.com/theplant/package1", "origin", "master")
	var err error
	err = exec.Command("git", "clone", "fixtures/package1", "../../package1").Run()
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
	err = (exec.Command("rm", "-rf", "../../package1").Run())
	if err != nil {
		panic(err)
	}
}

func (s *GitPkgSuite) TestClean(c *C) {
	// clean
	clean, _ := testGitPkg.IsClean()
	c.Check(clean, Equals, true)

	// not clean
	var err error
	err = exec.Command("mv", "../../package1/file3", "../../package1/file3m").Run()
	if err != nil {
	    panic(err)
	}

	clean, _ = testGitPkg.IsClean()
	c.Check(clean, Equals, false)
}

func (s *GitPkgSuite) TestGetChecksum(c *C) {
	checksum, _ := testGitPkg.GetChecksum("refs/heads/master")
	c.Check(checksum, Equals, testGpMasterChecksum)

	checksum, err := testGitPkg.GetChecksum("refs/heads/master_not_exist")
	c.Check(checksum, Equals, "")
	c.Check(err, Not(Equals), nil)
}

func (s *GitPkgSuite) TestFetch(c *C) {
	c.Check(testGitPkg.Fetch(), Equals, nil)
}

func (s *GitPkgSuite) TestSimpleGet(c *C) {
	testGitPkg.Sync()

	c.Check(testGitPkg.HeadRefsName, Equals, "refs/heads/master")
	c.Check(testGitPkg.HeadChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.PakbranchChecksum, Equals, "")
	c.Check(testGitPkg.PaktagChecksum, Equals, "")
	c.Check(testGitPkg.State.ContainsBranchNamedPak, Equals, false)
	c.Check(testGitPkg.State.ContainsPaktag, Equals, false)
	c.Check(testGitPkg.State.UnderPak, Equals, false)
	c.Check(testGitPkg.State.OnPakbranch, Equals, false)
	c.Check(testGitPkg.State.IsRemoteBranchExist, Equals, true)
	c.Check(testGitPkg.State.IsClean, Equals, true)

	testGitPkg.Get(GetOption{true, true, ""})
	testGitPkg.Sync()

	c.Check(testGitPkg.HeadRefsName, Equals, "refs/heads/pak")
	c.Check(testGitPkg.HeadChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.PakbranchChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.PaktagChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.State.ContainsBranchNamedPak, Equals, true)
	c.Check(testGitPkg.State.ContainsPaktag, Equals, true)
	c.Check(testGitPkg.State.UnderPak, Equals, true)
	c.Check(testGitPkg.State.OnPakbranch, Equals, true)
	c.Check(testGitPkg.State.IsRemoteBranchExist, Equals, true)
	c.Check(testGitPkg.State.IsClean, Equals, true)
}

func (s *GitPkgSuite) TestWeakGet(c *C) {
	exec.Command("git", testGitPkg.WorkTree, testGitPkg.GitDir, "branch", "pak").Run()
	testGitPkg.Sync()
	c.Check(testGitPkg.HeadRefsName, Equals, "refs/heads/master")
	c.Check(testGitPkg.HeadChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.PakbranchChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.PaktagChecksum, Equals, "")
	c.Check(testGitPkg.State.ContainsBranchNamedPak, Equals, true)
	c.Check(testGitPkg.State.ContainsPaktag, Equals, false)
	c.Check(testGitPkg.State.UnderPak, Equals, false)
	c.Check(testGitPkg.State.OnPakbranch, Equals, false)
	c.Check(testGitPkg.State.IsRemoteBranchExist, Equals, true)
	c.Check(testGitPkg.State.IsClean, Equals, true)

	testGitPkg.Get(GetOption{true, false, ""})
	testGitPkg.Sync()

	c.Check(testGitPkg.HeadRefsName, Equals, "refs/heads/master")
	c.Check(testGitPkg.HeadChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.PakbranchChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.PaktagChecksum, Equals, "")
	c.Check(testGitPkg.State.ContainsBranchNamedPak, Equals, true)
	c.Check(testGitPkg.State.ContainsPaktag, Equals, false)
	c.Check(testGitPkg.State.UnderPak, Equals, false)
	c.Check(testGitPkg.State.OnPakbranch, Equals, false)
	c.Check(testGitPkg.State.IsRemoteBranchExist, Equals, true)
	c.Check(testGitPkg.State.IsClean, Equals, true)
}


func (s *GitPkgSuite) TestForcefulGet(c *C) {
	exec.Command("git", testGitPkg.WorkTree, testGitPkg.GitDir, "branch", "pak").Run()
	testGitPkg.Sync()

	testGitPkg.Get(GetOption{true, true, ""})
	testGitPkg.Sync()

	c.Check(testGitPkg.HeadRefsName, Equals, "refs/heads/pak")
	c.Check(testGitPkg.HeadChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.PakbranchChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.PaktagChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.State.ContainsBranchNamedPak, Equals, true)
	c.Check(testGitPkg.State.ContainsPaktag, Equals, true)
	c.Check(testGitPkg.State.UnderPak, Equals, true)
	c.Check(testGitPkg.State.OnPakbranch, Equals, true)
	c.Check(testGitPkg.State.IsRemoteBranchExist, Equals, true)
	c.Check(testGitPkg.State.IsClean, Equals, true)
}

func (s *GitPkgSuite) TestGetWithChecksum(c *C) {
	devChecksum := "711c1e206bca5ad99edf6da12074bbbe4a349932"
	testGitPkg.Sync()
	testGitPkg.Get(GetOption{true, true, devChecksum})
	testGitPkg.Sync()

	c.Check(testGitPkg.HeadRefsName, Equals, "refs/heads/pak")
	c.Check(testGitPkg.HeadChecksum, Equals, devChecksum)
	c.Check(testGitPkg.PakbranchChecksum, Equals, devChecksum)
	c.Check(testGitPkg.PaktagChecksum, Equals, devChecksum)
	c.Check(testGitPkg.State.ContainsBranchNamedPak, Equals, true)
	c.Check(testGitPkg.State.ContainsPaktag, Equals, true)
	c.Check(testGitPkg.State.UnderPak, Equals, true)
	c.Check(testGitPkg.State.OnPakbranch, Equals, true)
	c.Check(testGitPkg.State.IsRemoteBranchExist, Equals, true)
	c.Check(testGitPkg.State.IsClean, Equals, true)
}

func (s *GitPkgSuite) TestForcefulUnpak(c *C) {
	testGitPkg.Sync()
	testGitPkg.Get(GetOption{true, true, testGpMasterChecksum})
	testGitPkg.Sync()
	testGitPkg.Unpak(true)
	testGitPkg.Sync()

	c.Check(testGitPkg.HeadRefsName, Equals, "refs/heads/master")
	c.Check(testGitPkg.HeadChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.PakbranchChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.PaktagChecksum, Equals, testGpMasterChecksum)
	c.Check(testGitPkg.State.ContainsBranchNamedPak, Equals, false)
	c.Check(testGitPkg.State.ContainsPaktag, Equals, false)
	c.Check(testGitPkg.State.UnderPak, Equals, false)
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
	c.Check(testGitPkg.State.UnderPak, Equals, false)
	c.Check(testGitPkg.State.OnPakbranch, Equals, false)
	c.Check(testGitPkg.State.IsRemoteBranchExist, Equals, true)
	c.Check(testGitPkg.State.IsClean, Equals, true)
}
