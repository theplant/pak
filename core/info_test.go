package core

import (
	"fmt"
	. "github.com/theplant/pak/share"
	"io/ioutil"
	. "launchpad.net/gocheck"
	"launchpad.net/goyaml"
	"os"
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
	path           string
	msg            string
	checker        Checker
	pakInfoChecker Checker
}{
	{Pakfile, "Can read Pakfile in current Folder", Equals, DeepEquals},
	{"../" + Pakfile, "Can read Pakfile in parent Folder", Equals, DeepEquals},
	{Gopath + "/../" + Pakfile, "Won't go beyond GOPATH to find Pakfile", Equals, Not(DeepEquals)},
}

func (s *PakSuite) SetUpSuite(c *C) {
	MustRun("sh", "-c", "mkdir -p ../../../test")
	MustRun("sh", "-c", "mkdir -p ../../../test2")
	MustRun("sh", "-c", "cp fixtures/Pakfile-for-pakread ../../../test/Pakfile")
	MustRun("sh", "-c", "cp fixtures/Pakfile-for-pakread2 ../../../test2/Pakfile")
	MustRun("sh", "-c", "cp fixtures/TestReadPaklock-pakfile.lock ../../../test/Pakfile.lock")
	MustRun("sh", "-c", "cp fixtures/TestReadPaklock2-pakfile.lock ../../../test2/Pakfile.lock")
}

func (s *PakSuite) TearDownSuite(c *C) {
	MustRun("sh", "-c", "rm -rf ../../../test")
	MustRun("sh", "-c", "rm -rf ../../../test2")
}

// TODO: to test sub package info
func (s *PakSuite) TestReadPakfile(c *C) {
	pakInfo := PakInfo{Packages: []PkgCfg{
		{
			Name:                   "github.com/test",
			PakName:                "pak",
			TargetBranch:           "origin/master",
			AutoMatchingHostBranch: false,
		},
		{
			Name:                   "github.com/test2",
			PakName:                "pak",
			TargetBranch:           "origin/master",
			AutoMatchingHostBranch: false,
		},
	}}
	pakInfoBytes, _ := goyaml.Marshal(&pakInfo)

	expectingPakInfo := PakInfo{
		Packages: []PkgCfg{
			{
				Name:                   "github.com/theplant/package1",
				PakName:                "pak",
				TargetBranch:           "origin/master",
				AutoMatchingHostBranch: false,
			},
			{
				Name:                   "github.com/theplant/package2",
				PakName:                "pak",
				TargetBranch:           "origin/dev",
				AutoMatchingHostBranch: false,
			},
			{
				Name:                   "github.com/theplant/package3",
				PakName:                "pak",
				TargetBranch:           "origin/master",
				AutoMatchingHostBranch: false,
			},
			{
				Name:                   "github.com/theplant/package1-2",
				PakName:                "pak",
				TargetBranch:           "origin/master",
				AutoMatchingHostBranch: false,
			},
			{
				Name:                   "github.com/theplant/package2-2",
				PakName:                "pak",
				TargetBranch:           "origin/dev",
				AutoMatchingHostBranch: false,
			},
			{
				Name:                   "github.com/theplant/package3-2",
				PakName:                "pak",
				TargetBranch:           "origin/master",
				AutoMatchingHostBranch: false,
			},
			{
				Name:                   "github.com/test",
				PakName:                "pak",
				TargetBranch:           "origin/master",
				AutoMatchingHostBranch: false,
			},
			{
				Name:                   "github.com/test2",
				PakName:                "pak",
				TargetBranch:           "origin/master",
				AutoMatchingHostBranch: false,
			},
		},
		BasicDependences: []string{"github.com/test", "github.com/test2"},
	}
	for _, pakfilePath := range pakfilePaths {
		ioutil.WriteFile(pakfilePath.path, pakInfoBytes, os.FileMode(0644))

		pakInfo2, _, err := GetPakInfo(GpiParams{
			Type:                 "10",
			Path:                 "",
			DeepParse:            true,
			WithBasicDependences: true,
		})
		c.Log(pakfilePath.msg)
		c.Check(err, pakfilePath.checker, nil)
		c.Check(pakInfo2, pakfilePath.pakInfoChecker, expectingPakInfo)

		os.Remove(pakfilePath.path)
	}
}

var paklockPaths = []struct {
	path        string
	pakfilePath string
	msg         string
	errChecker  Checker
	lockChecker Checker
}{
	{Paklock, Pakfile, "Can read Pakfile.lock in current Folder", Equals, DeepEquals},
	{"../" + Paklock, "../" + Pakfile, "Can read Pakfile.lock in parent Folder", Equals, DeepEquals},
	{Gopath + "/../" + Paklock, Gopath + "/../" + Pakfile, "Won't go beyond GOPATH to find Pakfile.lock", Equals, Not(DeepEquals)},
}

func (s *PakSuite) TestReadPaklock(c *C) {
	pakInfo := PakInfo{Packages: []PkgCfg{
		{
			Name:                   "github.com/test",
			PakName:                "pak",
			TargetBranch:           "origin/master",
			AutoMatchingHostBranch: false,
		},
		{
			Name:                   "github.com/test2",
			PakName:                "pak",
			TargetBranch:           "origin/master",
			AutoMatchingHostBranch: false,
		},
	}}
	pakInfoBytes, _ := goyaml.Marshal(&pakInfo)

	paklockInfo := PaklockInfo{
		"github.com/theplant/package5": "11b174bd5acbf990687e6b068c97378d3219de04",
		"github.com/theplant/package4": "d5f51ca77f5d4f37a8105a74b67d2f1aefea939c",
		"github.com/theplant/package6": "941af3b182a1d0a5859fd451a8b5a633f479d7bc",
		"github.com/theplant/package7": "d5f51ca77f5d4f37a8105a74b67d2f1aefea939c",
		"github.com/theplant/package8": "11b174bd5acbf990687e6b068c97378d3219de04",
		"github.com/theplant/package9": "941af3b182a1d0a5859fd451a8b5a633f479d7bc",
		"github.com/theplant/package3": "d5f51ca77f5d4f37a8105a74b67d2f1aefea939c",
		"github.com/theplant/package1": "11b174bd5acbf990687e6b068c97378d3219de04",
		"github.com/theplant/package2": "941af3b182a1d0a5859fd451a8b5a633f479d7bc",
	}
	paklockInfoBytes, _ := goyaml.Marshal(&paklockInfo)
	for _, paklockPath := range paklockPaths {
		ioutil.WriteFile(paklockPath.pakfilePath, pakInfoBytes, os.FileMode(0644))
		ioutil.WriteFile(paklockPath.path, paklockInfoBytes, os.FileMode(0644))

		_, paklockInfo2, err := GetPakInfo(GpiParams{
			Type:                 "11",
			Path:                 "",
			DeepParse:            true,
			WithBasicDependences: false,
		})
		c.Log(paklockPath.msg)
		c.Check(err, paklockPath.errChecker, nil)
		c.Check(paklockInfo2, paklockPath.lockChecker, paklockInfo)

		os.Remove(paklockPath.path)
		os.Remove(paklockPath.pakfilePath)
	}
}
