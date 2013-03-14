package pak

import (
	. "launchpad.net/gocheck"
	"launchpad.net/goyaml"
	"testing"
	"io/ioutil"
	"os"
)

// Hook up gocheck into the gotest runner.
func Test(t *testing.T) { TestingT(t) }

type PakSuite struct{}
var _ = Suite(&PakSuite{})

func (s *PakSuite) TestInitPakFile(c *C) {
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
	msg string
}{
	{pakfile, "Can read Pakfile in curreint Folder"},
	{"../" + pakfile, "Can read Pakfile in parent Folder"},
	{os.Getenv("GOPATH") + "/../Pakfile", "Won't go beyond GOPATH to find Pakfile"},
}

func (s *PakSuite) TestReadPakfile(c *C) {
	for _, pakfilePath := range pakfilePaths {
		pakInfo := PakInfo{Packages: []string{"github.com/test", "gihub.com/test2"}}
		pakInfoBytes, _ := goyaml.Marshal(&pakInfo)
		ioutil.WriteFile(pakfilePath.path, pakInfoBytes, os.FileMode(0644))

		pakInfo2 := readPakfile()
		c.Log(pakfilePath.msg)
		c.Check(SamePakInfo(pakInfo, pakInfo2), Equals, true)

		os.Remove(pakfilePath.path)
	}
}