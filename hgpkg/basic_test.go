package hgpkg

import (
	"fmt"
	. "github.com/theplant/pak/share"
	. "launchpad.net/gocheck"
	"testing"
)

// Hook up gocheck into the gotest runner.
func Test(t *testing.T) { TestingT(t) }

var fp = fmt.Printf
var fl = fmt.Println

type HgPkgSuite struct {
	pkg *HgPkg
}

var _ = Suite(&HgPkgSuite{})

func (s *HgPkgSuite) SetUpTest(c *C) {
	MustRun("sh", "-c", "hg clone fixtures/package1 ../../hgpkg-package1")

	pkg := NewHgPkg("github.com/theplant/hgpkg-package1", "default", "default", Pakbranch)
	s.pkg = pkg.(*HgPkg)
}

func (s *HgPkgSuite) TearDownTest(c *C) {
	MustRun("sh", "-c", "rm -rf ../../hgpkg-package1")
}

func (s *HgPkgSuite) TestContainsPakbranch(c *C) {
	containing, err := s.pkg.ContainsPakbranch()
	c.Check(err, Equals, nil)
	c.Check(containing, Equals, false)

	MustRun("sh", "-c", "cd ../../hgpkg-package1; hg bookmarks pak")

	containing, err = s.pkg.ContainsPakbranch()
	c.Check(err, Equals, nil)
	c.Check(containing, Equals, true)
}

// func (s *HgPkgSuite) TestContainsPaktag(c *C) {
// 	containing, err := s.pkg.ContainsPaktag()
// 	c.Check(err, Equals, nil)
// 	c.Check(containing, Equals, false)

// 	MustRun("sh", "-c", "cd ../../hgpkg-package1; hg tag _pak_latest_")

// 	containing, err = s.pkg.ContainsPaktag()
// 	c.Check(err, Equals, nil)
// 	c.Check(containing, Equals, true)
// }

func (s *HgPkgSuite) TestIsClean(c *C) {
	clean, err := s.pkg.IsClean()
	c.Check(err, Equals, nil)
	c.Check(clean, Equals, true)

	MustRun("sh", "-c", "cd ../../hgpkg-package1; touch untracked-file")

	clean, err = s.pkg.IsClean()
	c.Check(err, Equals, nil)
	c.Check(clean, Equals, true)

	MustRun("sh", "-c", "cd ../../hgpkg-package1; rm untracked-file")
	MustRun("sh", "-c", "cd ../../hgpkg-package1; rm file1")

	clean, err = s.pkg.IsClean()
	c.Check(err, Equals, nil)
	c.Check(clean, Equals, false)
}

func (s *HgPkgSuite) TestGetChecksum(c *C) {
	checksum, err := s.pkg.GetChecksum("default")
	c.Check(err, Equals, nil)
	c.Check(checksum, Equals, "1eebd4597062386493ac83ac80b0b9e3d08f7af7")
}

func (s *HgPkgSuite) TestFetch(c *C) {
	MustRun("sh", "-c", "mv fixtures/package1 fixtures/package1-backup")
	MustRun("sh", "-c", "mv fixtures/package1-updated fixtures/package1")

	err := s.pkg.Fetch()
	c.Check(err, Equals, nil)

	checksum, err := s.pkg.GetChecksum("default")
	c.Check(err, Equals, nil)
	c.Check(checksum, Equals, "b095b01f34f066f667970a33e9e9166638038694")

	MustRun("sh", "-c", "mv fixtures/package1 fixtures/package1-updated")
	MustRun("sh", "-c", "mv fixtures/package1-backup fixtures/package1")
}

func (s *HgPkgSuite) TestGetHeadRefName(c *C) {
	name, err := s.pkg.GetHeadRefName()
	c.Check(err, Equals, nil)
	c.Check(name, Equals, "non-pak")

	MustRun("sh", "-c", "cd ../../hgpkg-package1; hg bookmark pak")

	name, err = s.pkg.GetHeadRefName()
	c.Check(err, Equals, nil)
	c.Check(name, Equals, Pakbranch)
}

func (s *HgPkgSuite) TestGetHeadChecksum(c *C) {
	checksum, err := s.pkg.GetHeadChecksum()
	c.Check(err, Equals, nil)
	c.Check(checksum, Equals, "1eebd4597062386493ac83ac80b0b9e3d08f7af7")

	MustRun("sh", "-c", "cd ../../hgpkg-package1; hg tag --local another-tag")
	MustRun("sh", "-c", "cd ../../hgpkg-package1; hg tag --local third-tag")

	checksum, err = s.pkg.GetHeadChecksum()
	c.Check(err, Equals, nil)
	c.Check(checksum, Equals, "1eebd4597062386493ac83ac80b0b9e3d08f7af7")
}

func (s *HgPkgSuite) TestPak(c *C) {
	checksum, err := s.pkg.Pak("default")
	c.Check(err, Equals, nil)
	c.Check(checksum, Equals, "1eebd4597062386493ac83ac80b0b9e3d08f7af7")

	containing, err := s.pkg.ContainsPakbranch()
	c.Check(err, Equals, nil)
	c.Check(containing, Equals, true)

	// containing, err = s.pkg.ContainsPaktag()
	// c.Check(err, Equals, nil)
	// c.Check(containing, Equals, true)
}

func (s *HgPkgSuite) TestUnpak(c *C) {
	checksum, err := s.pkg.Pak("default")
	c.Check(err, Equals, nil)
	c.Check(checksum, Equals, "1eebd4597062386493ac83ac80b0b9e3d08f7af7")

	err = s.pkg.Unpak()
	c.Check(err, Equals, nil)

	containing, err := s.pkg.ContainsPakbranch()
	c.Check(err, Equals, nil)
	c.Check(containing, Equals, false)

	// containing, err = s.pkg.ContainsPaktag()
	// c.Check(err, Equals, nil)
	// c.Check(containing, Equals, false)
}

//
// func (s *HgPkgSuite) Test(c *C) {
// }
