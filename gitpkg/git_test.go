package gitpkg

import(
	"os/exec"
	"fmt"
	. "launchpad.net/gocheck"
)

type GitPkgSuite struct{}
var _ = Suite(&GitPkgSuite{})

var testGitPkg = NewGitPkg("github.com/theplant/package1", "origin", "master")

func (s *GitPkgSuite) sSetUpSuite(c *C) {
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

func (s *GitPkgSuite) sTearDownSuite(c *C) {
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

	err = exec.Command("mv", "../../package1/file3m", "../../package1/file3").Run()
	if err != nil {
	    panic(err)
	}
}

func (s *GitPkgSuite) TestGetChecksum(c *C) {
	checksum, _ := testGitPkg.GetChecksum("refs/heads/master")
	c.Check(checksum, Equals, "11b174bd5acbf990687e6b068c97378d3219de04")

	checksum, err := testGitPkg.GetChecksum("refs/heads/master_not_exist")
	c.Check(checksum, Equals, "")
	c.Check(err, Not(Equals), nil)
}

func (s *GitPkgSuite) TestFetch(c *C) {
	c.Check(testGitPkg.Fetch(), Equals, nil)
}

func (s *GitPkgSuite) TestRemovable(c *C) {
    testGitPkg.RemovePak()
    testGitPkg.GetPak()
    removable, _ := testGitPkg.PakbranchRemovable()
    c.Check(removable, Equals, true)
    onPak, _ := testGitPkg.EqualPakBranchAndTag()
    c.Check(onPak, Equals, true)
}


func (s *GitPkgSuite) TestSimpleGet(c *C) {
	testGitPkg.Get(true, true, false)

	refs, err := testGitPkg.GetHeadRefName()
	c.Check(err, Equals, nil)
	c.Check(refs, Equals, "refs/heads/pak")
}

func (s *GitPkgSuite) TestGetWithChecksum(c *C) {
	testGitPkg.Checksum = "11b174bd5acbf990687e6b068c97378d3219de04"
	testGitPkg.Get(true, true, true)

	refs, _ := testGitPkg.GetHeadRefName()
	c.Check(refs, Equals, "refs/heads/pak")
	checksum, _ := testGitPkg.GetHeadChecksum()
	c.Check(checksum, Equals, "11b174bd5acbf990687e6b068c97378d3219de04")
}