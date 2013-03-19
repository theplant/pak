package pak

import (
	. "launchpad.net/gocheck"
	"os/exec"
)

type GetSuite struct{}

var _ = Suite(&GetSuite{})

func (s *GetSuite) dSetUpSuite(c *C) {
	var err error
	err = exec.Command("git", "clone", "fixtures/package1", "../package1").Run()
	if err != nil {
		panic(err)
	}
	err = exec.Command("git", "clone", "fixtures/package2", "../package2").Run()
	if err != nil {
		panic(err)
	}
	err = exec.Command("git", "clone", "fixtures/package3", "../package3").Run()
	if err != nil {
		panic(err)
	}
}

func (s *GetSuite) dTearDownSuite(c *C) {
	exec.Command("rm", "-rf", "../package1").Run()
	exec.Command("rm", "-rf", "../package2").Run()
	exec.Command("rm", "-rf", "../package3").Run()
}

func (s *GetSuite) TestGet(c *C) {
	gitPkgs := []GitPkg{}
	pkgsFixtures := map[string][]string{
		"github.com/theplant/package1": {"origin", "master"},
		"github.com/theplant/package2": {"origin", "dev"},
		"github.com/theplant/package3": {"origin", "master"},
	}
	for key, val := range pkgsFixtures {
		gitPkgs = append(gitPkgs, NewGitPkg(key, val[0], val[1]))
	}
	getOptions := GetOptions{gitPkgs, true, true, false}
	gotExpected := true
	expectedPaklockInfo := PaklockInfo{
		"github.com/theplant/package3": "d5f51ca77f5d4f37a8105a74b67d2f1aefea939c",
		"github.com/theplant/package1": "11b174bd5acbf990687e6b068c97378d3219de04",
		"github.com/theplant/package2": "941af3b182a1d0a5859fd451a8b5a633f479d7bc",
	}

	paklockInfo, err := Get(getOptions)

	for key, val := range expectedPaklockInfo {
		if val != paklockInfo[key] {
			gotExpected = false
			break
		}
	}

	c.Check(err, Equals, nil)
	c.Check(gotExpected, Equals, true)

	// gitPkg1 := getOptions.GitPkgs[0]
	// gitPkg1.GetPak()

	paklockInfo, err = Get(getOptions)

	c.Check(err, Equals, nil)

	// exec.Command("cp", "fixtures/Pakfile", ".").Run()
	// exec.Command("rm", "Pakfile").Run()

	// getOptions2 := GetOptions{gitPkgs, true, false}
	// fl(Get(getOptions2))
	// c.Check(Get(getOptions2), Not(Equals), nil)
}
