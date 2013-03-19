package gitpkg

import (
	"bytes"
	"fmt"
	. "github.com/theplant/pak/share"
	"os/exec"
	"strings"
)

// Notes:
// Containing branch named pak does not mean that pkg is managed by pak.
// Containing tag named _pak_latest_ means this pkg is managed by pak, but
// still can't make sure the pkg is on the pak branch or it's status is wanted
// by Pakfile or Pakfile.lock.
type GitPkgState struct {
	UnderGitControl		   bool
	ContainsBranchNamedPak bool
	ContainsPaktag         bool
	UnderPak               bool
	OnPakbranch            bool
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
	gitPkg.Path = getGitPkgPath(name)
	gitPkg.WorkTree = getGitWorkTreeOpt(gitPkg.Path)
	gitPkg.GitDir = getGitDirOpt(gitPkg.Path)

	return
}

func getGitPkgPath(pkg string) string {
	return fmt.Sprintf("%s/src/%s", Gopath, pkg)
}

func getGitDirOpt(pkgPath string) string {
	return fmt.Sprintf("--git-dir=%s/.git", pkgPath)
}

func getGitWorkTreeOpt(pkgPath string) string {
	return fmt.Sprintf("--work-tree=%s", pkgPath)
}

func RunCmd(cmd *exec.Cmd) (out bytes.Buffer, err error) {
	cmd.Stdout = &out
	err = cmd.Run()

	return
}

func (this *GitPkg) Sync() (err error) {
	// Should be Under the Control of Git
	var state bool
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

	// Pakbranch
	state = false
	if this.State.ContainsBranchNamedPak {
		if this.State.ContainsPaktag {
			state = this.PaktagChecksum == this.PakbranchChecksum
		}
	}
	this.State.UnderPak = state

	// on Pakbranch
	state = false
	if this.State.UnderPak {
		var pakbranchChecksum string
		pakbranchChecksum, err = this.GetChecksum(this.Pakbranch)
		if err != nil {
			return
		}

		state = pakbranchChecksum == this.HeadChecksum
	}
	this.State.OnPakbranch = state

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

func (this *GitPkg) IsUnderGitControl() (bool, error) {
	out, err := RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "rev-parse", "--is-inside-work-tree"))
	if err != nil {
	    return false, fmt.Errorf("Package: %s\n%s", this.Name, out.String())
	}

	return true, nil
}

// Not to check out the pakbranch, but just a branch named refs/heads/pak
func (this *GitPkg) ContainsPakbranch() (bool, error) {
	out, err := RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "show-ref"))
	if err != nil {
		return false, fmt.Errorf("git %s %s show-ref\n%s\n", this.GitDir, this.WorkTree, err.Error())
	}

	return strings.Contains(out.String(), " "+this.Pakbranch+"\n"), nil
}

func (this *GitPkg) ContainsPaktag() (bool, error) {
	out, err := RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "show-ref"))
	if err != nil {
		return false, fmt.Errorf("git %s %s show-ref\n%s\n", this.GitDir, this.WorkTree, err.Error())
	}

	return strings.Contains(out.String(), " "+this.Paktag+"\n"), nil
}

func (this *GitPkg) IsClean() (bool, error) {
	out, err := RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "status", "--porcelain", "--untracked-files=no"))
	if err != nil {
		return false, fmt.Errorf("git %s %s status --porcelain --untracked-files=no\n%s\n", this.GitDir, this.WorkTree, err.Error())
	}

	return out.String() == "", nil
}

func (this *GitPkg) GetChecksum(ref string) (string, error) {
	out, err := RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "show-ref", ref, "--hash"))
	if err != nil {
		return "", fmt.Errorf("git %s %s show-ref %s --hash\n%s\n", this.GitDir, this.WorkTree, ref, err.Error())
	}

	checksum := out.String()[:40]

	return checksum, nil
}

func (this *GitPkg) Fetch() error {
	_, err := RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "fetch", this.Remote, this.Branch))
	if err != nil {
		err = fmt.Errorf("git %s %s fetch %s %s\n%s\n", this.GitDir, this.WorkTree, this.Remote, this.Branch, err.Error())
	}

	return err
}

func (this *GitPkg) ContainsRemoteBranch() (bool, error) {
	out, err := RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "show-ref"))
	if err != nil {
		return false, fmt.Errorf("git %s %s show-ref\n%s\n", this.GitDir, this.WorkTree, err.Error())
	}

	return strings.Contains(out.String(), " "+this.RemoteBranch+"\n"), nil
}

func (this *GitPkg) GetHeadRefName() (string, error) {
	out, err := RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "symbolic-ref", "HEAD"))
	if err != nil {
		return "", fmt.Errorf("git %s %s symbolic-ref HEAD", this.GitDir, this.WorkTree)
	}

	refs := out.String()
	return refs[:len(refs)-1], err
}

func (this *GitPkg) GetHeadChecksum() (string, error) {
	headBranch, err := this.GetHeadRefName()
	if err != nil {
		return "", err
	}

	out, err := RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "show-ref", "--hash", headBranch))
	if err != nil {
		return "", fmt.Errorf("git %s %s symbolic-ref HEAD", this.GitDir, this.WorkTree)
	}

	refs := out.String()
	return refs[:len(refs)-1], err
}
