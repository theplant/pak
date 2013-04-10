package gitpkg

import (
	"bytes"
	"fmt"
	. "github.com/theplant/pak/share"
	"os"
	"os/exec"
	"strings"
)

// Notes:
// Containing branch named pak does not mean that pkg is managed by pak.
// Containing tag named _pak_latest_ means this pkg is managed by pak, but
// still can't make sure the pkg is on the pak branch or it's status is wanted
// by Pakfile or Pakfile.lock.
type GitPkgState struct {
	IsPkgExist             bool
	UnderGitControl        bool
	ContainsBranchNamedPak bool
	ContainsPaktag         bool
	OnPakbranch            bool
	OwnPakbranch           bool
	IsRemoteBranchExist    bool
	IsClean                bool
}

type GitPkg struct {
	Name              string // github.com/theplant/pak
	Remote            string // origin, etc
	Branch            string // master, dev, etc
	Path              string // $GOPATH/src/:Name
	RemoteBranch      string // refs/remotes/:Remote/:Branch
	Pakbranch         string // refs/heads/pak
	Paktag            string // refs/tags/_pak_latest_
	WorkTree          string
	GitDir            string
	HeadRefsName      string
	HeadChecksum      string
	PakbranchChecksum string
	PaktagChecksum    string

	State GitPkgState
}

func NewGitPkg(name, remote, branch string) (gitPkg GitPkg) {
	gitPkg.Name = name
	gitPkg.Remote = remote
	gitPkg.Branch = branch
	gitPkg.RemoteBranch = fmt.Sprintf("refs/remotes/%s/%s", remote, branch)
	gitPkg.Pakbranch = "refs/heads/" + Pakbranch
	gitPkg.Paktag = "refs/tags/" + Paktag
	gitPkg.Path = fmt.Sprintf("%s/src/%s", Gopath, name)
	gitPkg.WorkTree = fmt.Sprintf("--work-tree=%s", gitPkg.Path)
	gitPkg.GitDir = fmt.Sprintf("--git-dir=%s/.git", gitPkg.Path)

	return
}

func (this *GitPkg) Sync() (err error) {
	// Should be Under the Control of Git
	var state bool
	state, err = this.IsPkgExist()
	if err != nil {
		return err
	}
	this.State.IsPkgExist = state
	if !this.State.IsPkgExist {
		return fmt.Errorf("Package %s Is Not Exist.", this.Name)
	}

	state, err = this.IsUnderGitControl()
	if err != nil {
		return err
	}
	this.State.UnderGitControl = state

	var info string
	info, err = this.GetHeadRefName()
	if err != nil {
		return
	}
	this.HeadRefsName = info

	info, err = this.GetHeadChecksum()
	if err != nil {
		return err
	}
	this.HeadChecksum = info

	// Retrieve State

	// Branch Named Pak
	state, err = this.ContainsPakbranch()
	if err != nil {
		return
	}

	this.State.ContainsBranchNamedPak = state
	if this.State.ContainsBranchNamedPak {
		var checksum string
		checksum, err = this.GetChecksum(this.Pakbranch)
		if err != nil {
			return
		}

		this.PakbranchChecksum = checksum
	}

	// Paktag _pak_latest_
	state, err = this.ContainsPaktag()
	if err != nil {
		return
	}
	this.State.ContainsPaktag = state
	if this.State.ContainsPaktag {
		var checksum string
		checksum, err = this.GetChecksum(this.Paktag)
		if err != nil {
			return
		}

		this.PaktagChecksum = checksum
	}

	this.State.OwnPakbranch = this.State.ContainsBranchNamedPak &&
		this.State.ContainsPaktag &&
		this.PaktagChecksum == this.PakbranchChecksum

	// on Pakbranch
	// TODO: add OnPakbranch Test(same checksum but different refs)
	this.State.OnPakbranch = this.State.ContainsBranchNamedPak &&
		this.State.ContainsPaktag &&
		this.PaktagChecksum == this.PakbranchChecksum &&
		this.PakbranchChecksum == this.HeadChecksum &&
		this.Pakbranch == this.HeadRefsName

	state, err = this.IsClean()
	if err != nil {
		return
	}
	this.State.IsClean = state

	state, err = this.ContainsRemoteBranch()
	if err != nil {
		return
	}
	this.State.IsRemoteBranchExist = state

	return
}

func (this *GitPkg) IsPkgExist() (bool, error) {
	_, err := os.Stat(this.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, err
		}
	}

	return true, nil
}

func (this *GitPkg) IsUnderGitControl() (bool, error) {
	_, err := this.Run("rev-parse", "--is-inside-work-tree")
	if err != nil {
		return false, fmt.Errorf("Package %s Is Not Git Tracked\n", this.Name)
	}

	return true, nil
}

// Not to check out the pakbranch, but just a branch named refs/heads/pak
func (this *GitPkg) ContainsPakbranch() (bool, error) {
	cmd, err := this.Run("show-ref")
	if err != nil {
		return false, err
	}

	return strings.Contains(cmd.Stdout.(*bytes.Buffer).String(), " "+this.Pakbranch+"\n"), nil
}

func (this *GitPkg) ContainsPaktag() (bool, error) {
	cmd, err := this.Run("show-ref")
	if err != nil {
		return false, err
	}

	return strings.Contains(cmd.Stdout.(*bytes.Buffer).String(), " "+this.Paktag+"\n"), nil
}

func (this *GitPkg) IsClean() (bool, error) {
	cmd, err := this.Run("status", "--porcelain", "--untracked-files=no")
	if err != nil {
		return false, err
	}

	return cmd.Stdout.(*bytes.Buffer).String() == "", nil
}

func (this *GitPkg) GetChecksum(ref string) (string, error) {
	cmd, err := this.Run("show-ref", ref, "--hash")
	if err != nil {
		return "", err
	}

	checksum := cmd.Stdout.(*bytes.Buffer).String()[:40]

	return checksum, nil
}

func (this *GitPkg) Fetch() error {
	_, err := this.Run("fetch")

	return err
}

func (this *GitPkg) ContainsRemoteBranch() (bool, error) {
	cmd, err := this.Run("show-ref")
	if err != nil {
		return false, err
	}

	return strings.Contains(cmd.Stdout.(*bytes.Buffer).String(), " "+this.RemoteBranch+"\n"), nil
}

func (this *GitPkg) GetHeadRefName() (string, error) {
	cmd, err := this.Run("symbolic-ref", "HEAD")

	refs := ""
	// TODO: add tests
	if err != nil {
		// TODO: find other way to tell out no branch
		if cmd.Stderr.(*bytes.Buffer).String() == "fatal: ref HEAD is not a symbolic ref\n" {
			refs = "no branch"
			err = nil
		} else {
			return "", err
		}
	} else {
		refs = cmd.Stdout.(*bytes.Buffer).String()
	}

	return refs[:len(refs)-1], nil
}

func (this *GitPkg) GetHeadChecksum() (string, error) {
	headBranch, err := this.GetHeadRefName()
	if err != nil {
		return "", err
	}

	cmd, err := this.Run("show-ref", "--hash", headBranch)
	if err != nil {
		return "", err
	}

	refs := cmd.Stdout.(*bytes.Buffer).String()

	return refs[:len(refs)-1], err
}

func (this *GitPkg) GoGet() error {
	return GoGetImpl(this.Name)
}

var GoGetImpl = func(name string) error {
	cmd := exec.Command("go", "get", name)
	cmd.Stderr = &bytes.Buffer{}
	err := cmd.Run()
	if err != nil {
		getErr := cmd.Stderr.(*bytes.Buffer).String() // for removing the end-of-line
		err = fmt.Errorf("go get %s:\n%s", name, getErr[:len(getErr)-1])
	}

	return err
}

func (this *GitPkg) Run(params ...string) (*exec.Cmd, error) {
	fullParams := append([]string{this.GitDir, this.WorkTree}, params...)
	cmd := exec.Command("git", fullParams...)

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("Error\n%s: %s => %s", this.Name, params, err.Error())
		return cmd, err
	}

	return cmd, nil
}
