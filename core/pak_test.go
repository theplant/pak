package core

import (
	"fmt"
	. "github.com/theplant/pak/share"
	"io/ioutil"
	. "launchpad.net/gocheck"
	"launchpad.net/goyaml"
	"os"
	"reflect"
	"testing"
)

// Hook up gocheck into the gotest runner.
func Test(t *testing.T) { TestingT(t) }

var fp = fmt.Printf
var fl = fmt.Println

type PakSuite struct{}

var _ = Suite(&PakSuite{})

func (s *PakSuite) TestInit(c *C) {
	os.Remove(Pakfile)
	Init()
	_, err := os.Stat(Pakfile)
	c.Check(err, Equals, nil)

	tmpPakfile, _ := os.Create(Pakfile)
	tmpPakfileInfo, _ := tmpPakfile.Stat()
	Init()
	tmpPakfile2, _ := os.Create(Pakfile)
	tmpPakfileInfo2, _ := tmpPakfile2.Stat()
	c.Log("Should not create Pakfile if it already existed.")
	c.Check(os.SameFile(tmpPakfileInfo, tmpPakfileInfo2), Equals, true)
	os.Remove(Pakfile)
}

var pakfilePaths = []struct {
	path         string
	msg          string
	pakfileState bool
}{
	{Pakfile, "Can read Pakfile in curreint Folder", true},
	{"../" + Pakfile, "Can read Pakfile in parent Folder", true},
	{Gopath + "/../" + Pakfile, "Won't go beyond GOPATH to find Pakfile", false},
}

func (s *PakSuite) TestReadPakfile(c *C) {
	pakInfo := PakInfo{Packages: []string{"github.com/test", "gihub.com/test2"}}
	pakInfoBytes, _ := goyaml.Marshal(&pakInfo)
	for _, pakfilePath := range pakfilePaths {
		ioutil.WriteFile(pakfilePath.path, pakInfoBytes, os.FileMode(0644))

		pakInfo2, err := GetPakInfo()
		c.Log(pakfilePath.msg)
		c.Check(err == nil, Equals, pakfilePath.pakfileState)
		c.Check(reflect.DeepEqual(pakInfo, pakInfo2), Equals, pakfilePath.pakfileState)

		os.Remove(pakfilePath.path)
	}
}

var paklockPaths = []struct {
	path         string
	msg          string
	paklockState bool
}{
	{Paklock, "Can read Pakfile in curreint Folder", true},
	{"../" + Paklock, "Can read Pakfile in parent Folder", true},
	{Gopath + "/../" + Paklock, "Won't go beyond GOPATH to find Pakfile", false},
}

func (s *PakSuite) TestReadPakfileLock(c *C) {
	paklockInfo := PaklockInfo{
		"github.com/theplant/package3": "d5f51ca77f5d4f37a8105a74b67d2f1aefea939c",
		"github.com/theplant/package1": "11b174bd5acbf990687e6b068c97378d3219de04",
		"github.com/theplant/package2": "941af3b182a1d0a5859fd451a8b5a633f479d7bc",
	}
	paklockInfoBytes, _ := goyaml.Marshal(&paklockInfo)
	for _, paklockPath := range paklockPaths {
		ioutil.WriteFile(paklockPath.path, paklockInfoBytes, os.FileMode(0644))

		paklockInfo2, err := GetPaklockInfo()
		c.Log(paklockPath.msg)
		c.Check(err == nil, Equals, paklockPath.paklockState)
		c.Check(reflect.DeepEqual(paklockInfo2, paklockInfo), Equals, paklockPath.paklockState)

		os.Remove(paklockPath.path)
	}
}

// func (s *PakSuite) TestUpdate(c *C) {
// 	pakInfo, _ := goyaml.Marshal(&PakInfo{[]string{"github.com/theplant/pak"}})
// 	ioutil.WriteFile(pakfile, pakInfo, os.FileMode(0644))
//
// 	Update()
//
// 	os.Remove(pakfile)
// 	os.Remove(paklock)
// }
//
// func (s *PakSuite) TestIsPackageClean(c *C) {
// 	c.Check(isPackageClean("github.com/sunfmin/batchbuy"), Equals, true)
// 	c.Check(isPackageClean("github.com/theplant/pak"), Equals, false)
// }
//
// func (s *PakSuite) TestCheckoutPakbranch(c *C) {
// 	c.Check(checkoutPakbranch("github.com/sunfmin/batchbuy", "3b61e71b65325275d1d043d4c558e674b2d2862f"), Equals, true)
// 	c.Check(checkoutPakbranch("github.com/theplant/batchbuy", "3b61e71b65325275d1d043d4c558e674b2d2862f"), Equals, false)
// }
