package gitpkg

import (
	"bytes"
	"fmt"
	. "github.com/theplant/pak/share"
	"os"
	"os/exec"
	"regexp"
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

func NewGitPkg(name, remote, branch string) VCSPkg {
	gitPkg := GitPkg{}
	gitPkg.Name = name
	gitPkg.Remote = remote
	gitPkg.Branch = branch
	gitPkg.RemoteBranch = fmt.Sprintf("refs/remotes/%s/%s", remote, branch)
	gitPkg.Pakbranch = "refs/heads/" + Pakbranch
	gitPkg.Paktag = "refs/tags/" + Paktag
	gitPkg.Path = fmt.Sprintf("%s/src/%s", Gopath, name)
	gitPkg.WorkTree = fmt.Sprintf("--work-tree=%s", gitPkg.Path)
	gitPkg.GitDir = fmt.Sprintf("--git-dir=%s/.git", gitPkg.Path)

	return &gitPkg
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

// TODO: should make it a core package function
func (this *GitPkg) IsPkgExist() (bool, error) {
	_, err := os.Stat(this.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

// TODO: should make it a core package function
func (this *GitPkg) IsUnderGitControl() (bool, error) {
	_, err := this.Git("rev-parse", "--is-inside-work-tree")
	if err != nil {
		return false, fmt.Errorf("Package %s Is Not Git Tracked\n", this.Name)
	}

	return true, nil
}

// Not to check out the pakbranch, but just a branch named refs/heads/pak
func (this *GitPkg) ContainsPakbranch() (bool, error) {
	cmd, err := this.Git("show-ref")
	if err != nil {
		return false, err
	}

	return strings.Contains(cmd.Stdout.(*bytes.Buffer).String(), " "+this.Pakbranch+"\n"), nil
}

func (this *GitPkg) ContainsPaktag() (bool, error) {
	cmd, err := this.Git("show-ref")
	if err != nil {
		return false, err
	}

	return strings.Contains(cmd.Stdout.(*bytes.Buffer).String(), " "+this.Paktag+"\n"), nil
}

func (this *GitPkg) IsClean() (bool, error) {
	cmd, err := this.Git("status", "--porcelain", "--untracked-files=no")
	if err != nil {
		return false, err
	}

	return cmd.Stdout.(*bytes.Buffer).String() == "", nil
}

func (this *GitPkg) GetChecksum(ref string) (string, error) {
	cmd, err := this.Git("show-ref", ref, "--hash")
	if err != nil {
		return "", err
	}

	checksum := cmd.Stdout.(*bytes.Buffer).String()[:40]

	return checksum, nil
}

func (this *GitPkg) Fetch() error {
	// TODO: add tests for => 1. remote is changed; 2. remote is no exist
	_, err := this.Git("fetch", this.Remote, this.Branch + ":" + this.RemoteBranch)

	return err
}

func (this *GitPkg) ContainsRemoteBranch() (bool, error) {
	cmd, err := this.Git("show-ref")
	if err != nil {
		return false, err
	}

	return strings.Contains(cmd.Stdout.(*bytes.Buffer).String(), " "+this.RemoteBranch+"\n"), nil
}

var RefsRegexp = regexp.MustCompile("^ref: (.+)\n")
func (this *GitPkg) GetHeadRefName() (string, error) {
	cmd := exec.Command("cat", this.Path + "/.git/HEAD")
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &bytes.Buffer{}
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%s: cat .git/HEAD => %s", this.Name, cmd.Stderr.(*bytes.Buffer).String())
	}

	refline := cmd.Stdout.(*bytes.Buffer).String()
	ref := ""
	if RefsRegexp.MatchString(refline) {
		ref = RefsRegexp.FindStringSubmatch(refline)[1]
	} else {
		ref = "no branch"
	}

	return ref, nil
}

func (this *GitPkg) GetHeadChecksum() (string, error) {
	headBranch, err := this.GetHeadRefName()
	if err != nil {
		return "", err
	}

	if headBranch == "no branch" {
		cmd := exec.Command("cat", this.Path + "/.git/HEAD")
		cmd.Stdout = &bytes.Buffer{}
		cmd.Stderr = &bytes.Buffer{}
		err := cmd.Run()
		if err != nil {
		    return "", fmt.Errorf("%s: cat .git/HEAD => %s", this.Name, cmd.Stderr.(*bytes.Buffer).String())
		}

		checksum := cmd.Stdout.(*bytes.Buffer).String()
		return checksum[:len(checksum)-1], nil
	}

	cmd, err := this.Git("show-ref", "--hash", headBranch)
	if err != nil {
		return "", err
	}

	refs := cmd.Stdout.(*bytes.Buffer).String()

	return refs[:len(refs)-1], err
}

// TODO: should make it a core package function
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

func (this *GitPkg) Git(params ...string) (*exec.Cmd, error) {
	fullParams := append([]string{this.GitDir, this.WorkTree}, params...)
	cmd := exec.Command("git", fullParams...)

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("Error\n%s: git %s => %s", this.Name, strings.Join(params, " "), stderr.String())
		return cmd, err
	}

	return cmd, nil
}
