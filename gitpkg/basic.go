package gitpkg

import (
	"bytes"
	"fmt"
	. "github.com/theplant/pak/share"
	"os/exec"
	"regexp"
	"strings"
)

type GitPkg struct {
	PakName      string
	Name         string // github.com/theplant/pak
	Remote       string // origin, etc
	Branch       string // master, dev, etc
	Path         string // $GOPATH/src/:Name
	RemoteBranch string // refs/remotes/:Remote/:Branch
	PakbranchRef string // refs/heads/pak
	// PaktagRef    string // refs/tags/_pak_latest_

	PkgRoot  string
	WorkTree string
	GitDir   string
}

// NewGitPkg now will panic, make sure to be invoked after IsTracking
func NewGitPkg(name, remote, branch, pakName string) PkgProxy {
	if branch == "" {
		branch = "master"
	}
	if remote == "" {
		remote = "origin"
	}

	gitPkg := GitPkg{}
	gitPkg.Name = name
	gitPkg.Remote = remote
	gitPkg.Branch = branch
	gitPkg.RemoteBranch = fmt.Sprintf("refs/remotes/%s/%s", remote, branch)
	gitPkg.PakName = pakName
	gitPkg.PakbranchRef = "refs/heads/" + gitPkg.PakName
	// gitPkg.PakbranchRef = "refs/heads/" + Pakbranch
	// gitPkg.PaktagRef = "refs/tags/" + Paktag
	gitPkg.Path = fmt.Sprintf("%s/src/%s", Gopath, name)

	root, err := GetPkgRoot(gitPkg.Name)
	if err != nil {
		panic(err) // TODO: would be nice if don't panic
	}

	gitPkg.PkgRoot = fmt.Sprintf("%s/src/%s", Gopath, root)
	gitPkg.WorkTree = fmt.Sprintf("--work-tree=%s", gitPkg.PkgRoot)
	gitPkg.GitDir = fmt.Sprintf("--git-dir=%s/.git", gitPkg.PkgRoot)

	return &gitPkg
}

func GetPkgRoot(pkg string) (string, error) {
	return GetPkgRootImpl(pkg, "git")
}

// Git is simple git command wrapper.
func (this *GitPkg) Git(params ...string) (*exec.Cmd, error) {
	fullParams := append([]string{this.GitDir, this.WorkTree}, params...)
	cmd := exec.Command("git", fullParams...)

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("%s: git %s => %s", this.Name, strings.Join(params, " "), stderr.String())
		return cmd, err
	}

	return cmd, nil
}

func (this *GitPkg) GetPakbranchRef() string {
	return this.PakbranchRef
}

// func (this *GitPkg) GetPaktagRef() string {
// 	return this.PaktagRef
// }

func (this *GitPkg) GetRemoteBranch() string {
	return this.RemoteBranch
}

// ContainsPakbranch will tell the existence of a branch named refs/heads/pak
func (this *GitPkg) ContainsPakbranch() (bool, error) {
	cmd, err := this.Git("show-ref")
	if err != nil {
		return false, err
	}

	return strings.Contains(cmd.Stdout.(*bytes.Buffer).String(), " "+this.PakbranchRef+"\n"), nil
}

// func (this *GitPkg) ContainsPaktag() (bool, error) {
// 	cmd, err := this.Git("show-ref")
// 	if err != nil {
// 		return false, err
// 	}

// 	return strings.Contains(cmd.Stdout.(*bytes.Buffer).String(), " "+this.PaktagRef+"\n"), nil
// }

// IsClean will check whether the package contains modified file, but it allows the existence of untracked file.
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
	_, err := this.Git("fetch", this.Remote, this.Branch+":"+this.RemoteBranch)

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

// TODO: add tests
func (this *GitPkg) GetHeadRefName() (string, error) {
	cmd := exec.Command("cat", this.PkgRoot+"/.git/HEAD")
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &bytes.Buffer{}
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Package %s: cat .git/HEAD => %s", this.Name, cmd.Stderr.(*bytes.Buffer).String())
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

// TODO: add tests
func (this *GitPkg) GetHeadChecksum() (string, error) {
	headBranch, err := this.GetHeadRefName()
	if err != nil {
		return "", err
	}

	if headBranch == "no branch" {
		cmd := exec.Command("cat", this.PkgRoot+"/.git/HEAD")
		cmd.Stdout = &bytes.Buffer{}
		cmd.Stderr = &bytes.Buffer{}
		err := cmd.Run()
		if err != nil {
			return "", fmt.Errorf("Package %s: cat .git/HEAD => %s", this.Name, cmd.Stderr.(*bytes.Buffer).String())
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

// Unpak will create a branch named pak and a tag named _pak_latest_.
// Caution: Make sure you the branch and tag is not exist.
func (this *GitPkg) Pak(ref string) (string, error) {
	// Create Pakbranch
	// _, err := this.Git("checkout", "-b", Pakbranch, ref)
	_, err := this.Git("checkout", "-b", this.PakName, ref)
	if err != nil {
		return "", err
	}

	// Create Paktag
	checksum, err := this.GetChecksum(this.PakbranchRef)
	if err != nil {
		return "", err
	}
	// _, err = this.Git("tag", Paktag, checksum)
	// if err != nil {
	// 	return "", err
	// }

	return checksum, err
}

// Unpak will remove a branch named pak and a tag named _pak_latest_ in a git repo.
func (this *GitPkg) Unpak() error {
	// Move to Master Branch
	if _, err := this.Git("checkout", "master"); err != nil {
		return err
	}

	// Delete Pakbranch
	if contained, err := this.ContainsPakbranch(); err == nil {
		if contained {
			// _, err = this.Git("branch", "-D", Pakbranch)
			_, err = this.Git("branch", "-D", this.PakName)
			if err != nil {
				return err
			}
		}
	} else {
		return err
	}

	// // Delete Paktag
	// if contained, err := this.ContainsPaktag(); err == nil {
	// 	if contained {
	// 		_, err = this.Git("tag", "-d", Paktag)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// } else {
	// 	return err
	// }

	return nil
}

func (this *GitPkg) NewBranch(name string) error {
	_, err := this.Git("branch", name)
	return err
}

func (this *GitPkg) CheckOut(ref string) error {
	_, err := this.Git("checkout", ref)
	return err
}

func (this *GitPkg) NewTag(tag, object string) error {
	_, err := this.Git("tag", tag, object)
	return err
}

func (this *GitPkg) RemoveTag(tag string) error {
	_, err := this.Git("tag", "-d", tag)
	return err
}
