package hgpkg

import (
	"bytes"
	"fmt"
	. "github.com/theplant/pak/share"
	// "os"
	"os/exec"
	"regexp"
	"strings"
)

type HgPkg struct {
	Name string // github.com/theplant/pak
	// Remote       string // origin, etc
	Branch string // master, dev, etc
	// Bookmark string
	Path         string // $GOPATH/src/:Name
	RemoteBranch string // refs/remotes/:Remote/:Branch
	PakbranchRef string // pak; not longer a branch but a bookmark in mercurial
	PaktagRef    string // _pak_latest_
}

func NewHgPkg(name, remote, branch string) PkgProxy {
	hgPkg := HgPkg{}
	hgPkg.Name = name
	hgPkg.Branch = branch
	hgPkg.RemoteBranch = remote
	hgPkg.PakbranchRef = Pakbranch
	hgPkg.PaktagRef = Paktag

	hgPkg.Path = fmt.Sprintf("%s/src/%s", Gopath, name)

	return &hgPkg
}

func GetPkgRoot(pkg string) (string, error) {
	return GetPkgRootImpl(pkg, "hg")
}

// Hg is a simple command wrapper for hg.
func (this *HgPkg) Hg(params ...string) (*exec.Cmd, error) {
	cmd := exec.Command("hg", params...)
	cmd.Dir = this.Path

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("Error\n%s: hg %s => %s", this.Name, strings.Join(params, " "), stderr.String())
		return cmd, err
	}

	return cmd, nil
}

func (this *HgPkg) GetPakbranchRef() string {
	return this.PakbranchRef
}

func (this *HgPkg) GetPaktagRef() string {
	return this.PaktagRef
}

func (this *HgPkg) GetRemoteBranch() string {
	return this.RemoteBranch
}

// ContainsPakbranch will tell the existence of a branch named refs/heads/pak
func (this *HgPkg) ContainsPakbranch() (bool, error) {
	cmd, err := this.Hg("bookmarks")
	if err != nil {
		return false, err
	}

	return strings.Contains(cmd.Stdout.(*bytes.Buffer).String(), " "+this.PakbranchRef+" "), nil
}

func (this *HgPkg) ContainsPaktag() (bool, error) {
	cmd, err := this.Hg("tags")
	if err != nil {
		return false, err
	}

	return strings.Contains(cmd.Stdout.(*bytes.Buffer).String(), this.PaktagRef+" "), nil
}

// IsClean will check whether the package contains modified file, but it allows the existence of untracked file.
func (this *HgPkg) IsClean() (bool, error) {
	cmd, err := this.Hg("status", "-mard")
	if err != nil {
		return false, err
	}

	return cmd.Stdout.(*bytes.Buffer).String() == "", nil
}

func (this *HgPkg) GetChecksum(ref string) (string, error) {
	cmd, err := this.Hg("log", "-r", ref, "--template", "{node}")
	if err != nil {
		return "", err
	}

	return cmd.Stdout.(*bytes.Buffer).String()[:40], nil
}

func (this *HgPkg) Fetch() error {
	_, err := this.Hg("pull", "-u", this.RemoteBranch)

	return err
}

func (this *HgPkg) ContainsRemoteBranch() (bool, error) {
	cmd, err := this.Hg("paths")
	if err != nil {
		return false, err
	}

	return strings.Contains(cmd.Stdout.(*bytes.Buffer).String(), this.RemoteBranch+" = "), nil
}

// GetHeadRefName will return the pak bookmark name if it contains
var RefsRegexp = regexp.MustCompile("\nbookmarks:.*pak.*\n")

func (this *HgPkg) GetHeadRefName() (string, error) {
	cmd, err := this.Hg("summary")
	if err != nil {
		return "", err
	}

	if RefsRegexp.MatchString(cmd.Stdout.(*bytes.Buffer).String()) {
		return Pakbranch, nil
	}

	return "non-pak", err
}

var ChecksumRegexp = regexp.MustCompile(`parent: \d:(.*) .*\n`)

func (this *HgPkg) GetHeadChecksum() (string, error) {
	cmd, err := this.Hg("summary")
	if err != nil {
		return "", err
	}

	matchStr := ChecksumRegexp.FindStringSubmatch(cmd.Stdout.(*bytes.Buffer).String())

	checksum, err := this.GetChecksum(matchStr[1])
	if err != nil {
		return "", err
	}

	return checksum, err
}

// Unpak will create a branch named pak and a tag named _pak_latest_.
// Caution: Make sure you the branch and tag is not exist.
func (this *HgPkg) Pak(ref string) (string, error) {
	// Create Pakbranch
	_, err := this.Hg("update", ref)
	if err != nil {
		return "", err
	}
	_, err = this.Hg("bookmark", this.PakbranchRef)
	if err != nil {
		return "", err
	}

	// Create Paktag
	checksum, err := this.GetChecksum(this.PakbranchRef)
	if err != nil {
		return "", err
	}
	_, err = this.Hg("tag", "-r", checksum, Paktag)
	if err != nil {
		return "", err
	}

	return checksum, err
}

// Unpak will remove a branch named pak and a tag named _pak_latest_ in a hg repo.
func (this *HgPkg) Unpak() error {
	// Delete Pakbranch
	if contained, err := this.ContainsPakbranch(); err == nil {
		if contained {
			_, err = this.Hg("bookmark", "-d", Pakbranch)
			if err != nil {
				return err
			}
		}
	} else {
		return err
	}

	// Delete Paktag
	if contained, err := this.ContainsPaktag(); err == nil {
		if contained {
			_, err = this.Hg("tag", "--remove", Paktag)
			if err != nil {
				return err
			}
		}
	} else {
		return err
	}

	return nil
}

func (this *HgPkg) NewBranch(name string) error {
	_, err := this.Hg("branch", name)
	return err
}

func (this *HgPkg) CheckOut(ref string) error {
	_, err := this.Hg("update", ref)
	return err
}

func (this *HgPkg) NewTag(tag, object string) error {
	_, err := this.Hg("tag", tag, object)
	return err
}

func (this *HgPkg) RemoveTag(tag string) error {
	_, err := this.Hg("tag", "--remove", tag)
	return err
}
