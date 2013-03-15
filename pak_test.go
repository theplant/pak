package pak

import (
	"io/ioutil"
	. "launchpad.net/gocheck"
	"launchpad.net/goyaml"
	"os"
	"os/exec"
	"testing"
	"fmt"
)

// Hook up gocheck into the gotest runner.
func Test(t *testing.T) { TestingT(t) }

var fp = fmt.Printf

type PakSuite struct{}

var _ = Suite(&PakSuite{})

func (s *PakSuite) SetUpSuite(c *C) {
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

func (s *PakSuite) TearDownSuite(c *C) {
	var err error
	err = (exec.Command("rm", "-rf", "../package1").Run())
	if err != nil {
		panic(err)
	}
	err = (exec.Command("rm", "-rf", "../package2").Run())
	if err != nil {
		panic(err)
	}
	err = (exec.Command("rm", "-rf", "../package3").Run())
	if err != nil {
		panic(err)
	}
}

func (s *PakSuite) TestInit(c *C) {
	os.Remove(pakfile)
	Init()
	_, err := os.Stat(pakfile)
	c.Check(err, Equals, nil)

	tmpPakfile, _ := os.Create(pakfile)
	tmpPakfileInfo, _ := tmpPakfile.Stat()
	Init()
	tmpPakfile2, _ := os.Create(pakfile)
	tmpPakfileInfo2, _ := tmpPakfile2.Stat()
	c.Log("Should not create Pakfile if it already existed.")
	c.Check(os.SameFile(tmpPakfileInfo, tmpPakfileInfo2), Equals, true)
	os.Remove(pakfile)
}

var pakfilePaths = []struct {
	path string
	msg  string
	pakfileState bool
}{
	{pakfile, "Can read Pakfile in curreint Folder", true},
	{"../" + pakfile, "Can read Pakfile in parent Folder", true},
	{gopath + "/../Pakfile", "Won't go beyond GOPATH to find Pakfile", false},
}

func (s *PakSuite) TestReadPakfile(c *C) {
	for _, pakfilePath := range pakfilePaths {
		pakInfo := PakInfo{Packages: []string{"github.com/test", "gihub.com/test2"}}
		pakInfoBytes, _ := goyaml.Marshal(&pakInfo)
		ioutil.WriteFile(pakfilePath.path, pakInfoBytes, os.FileMode(0644))

		pakfileExist, pakInfo2 := readPakfile()
		c.Log(pakfilePath.msg)
		c.Check(pakfileExist, Equals, pakfilePath.pakfileState)
		c.Check(SamePakInfo(pakInfo, pakInfo2), Equals, pakfilePath.pakfileState)

		os.Remove(pakfilePath.path)
	}
}

func (s *PakSuite) TestUpdate(c *C) {
	pakInfo, _ := goyaml.Marshal(&PakInfo{[]string{"github.com/theplant/pak"}})
	ioutil.WriteFile(pakfile, pakInfo, os.FileMode(0644))

	Update()

	os.Remove(pakfile)
	os.Remove(paklock)
}

func (s *PakSuite) TestIsPackageClean(c *C) {
	c.Check(isPackageClean("github.com/sunfmin/batchbuy"), Equals, true)
	c.Check(isPackageClean("github.com/theplant/pak"), Equals, false)
}

func (s *PakSuite) TestCheckoutPakbranch(c *C) {
	c.Check(checkoutPakbranch("github.com/sunfmin/batchbuy", "3b61e71b65325275d1d043d4c558e674b2d2862f"), Equals, true)
	c.Check(checkoutPakbranch("github.com/theplant/batchbuy", "3b61e71b65325275d1d043d4c558e674b2d2862f"), Equals, false)
}