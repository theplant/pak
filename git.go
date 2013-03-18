package pak

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type GitPkg struct {
	Name         string // github.com/theplant/pak
	Remote       string // origin, etc
	Branch       string // master, dev, etc
    Checksum     string // 0218f9eda0a90203265bc1482c6afec538a83576, etch
	Path         string // $GOPATH/src/:Name
	RemoteBranch string // refs/remotes/:Remote/:Branch
	Pakbranch    string // refs/heads/pak
	Paktag       string // refs/tags/_pak_latest_
	WorkTree     string
	GitDir       string
}

func NewGitPkg(name, remote, branch string) (this GitPkg) {
	this.Name = name
	this.Remote = remote
	this.Branch = branch
    // this.Checksum = checksum
	this.RemoteBranch = fmt.Sprintf("refs/remotes/%s/%s", remote, branch)
	this.Pakbranch = "refs/heads/" + pakbranch
	this.Paktag = "refs/tags/" + paktag
	this.Path = getGitPkgPath(name)
	this.WorkTree = getGitWorkTreeOpt(this.Path)
	this.GitDir = getGitDirOpt(this.Path)

	return
}

func getGitPkgPath(pkg string) string {
	return fmt.Sprintf("%s/src/%s", gopath, pkg)
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

func (this GitPkg) Get(fetchLatest bool, force bool, useChecksum bool) (string, error) {
	// Won't process non-clean package
	clean, err := this.IsClean()
	if err != nil {
		return "", err
	}
	if !clean {
		return "", fmt.Errorf("Package %s is not clean. Please clean it and re-start pak.", this.Name)
	}

	// remote branch should exist
	remoteExist, err := this.ContainsRemoteBranch()
	if err != nil {
		return "", err
	}
	if !remoteExist {
		return "", fmt.Errorf("`%s` does not contain reference `%s`", this.Name, this.RemoteBranch)
	}

	// fetch pkg before check out
	if fetchLatest {
		err := this.Fetch()
		if err != nil {
			return "", err
		}
	}

	containPakbranch, err := this.ContainsPakbranch()
	if err != nil {
		return "", err
	}

	onPakbranch := false

	// Remove pakbranch
	if containPakbranch {
		removable, err := this.PakbranchRemovable();
		if err != nil {
			return "", err
		}

		if !removable && !force {
			return "", fmt.Errorf(unrecognizedPakbranch)
		}

		// using checksum means pak is getting pkgs by Pakfile.lock
		if useChecksum {
			var err error
			onPakbranch, err = this.EqualChecksumAndPakbranch()
			if err != nil {
				return "", err
			}
		}

		if !onPakbranch {
			err := this.RemovePak()
			if err != nil {
				return "", err
			}
		}
	}

	// create pakbranch
	var checksum string
	// var err error
	if onPakbranch {
		checksum, err = this.GetChecksumHash(this.Pakbranch)
	} else {
		if useChecksum {
			checksum, err = this.GetPakByChecksum()
		} else {
			checksum, err = this.GetPak()
		}
	}
	if err != nil {
		return "", err
	}

    return checksum, err
}

// NOTE: will also try to delete pak tag
func (this GitPkg) RemovePak() (err error) {
    // move to master branch
    _, err = RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "checkout", "master"))
    if err != nil {
        return fmt.Errorf("git %s %s checkout master\n%s", this.GitDir, this.WorkTree, err.Error())
    }

	// delete pakbranch
	_, err = RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "branch", "-d", pakbranch))
	if err != nil {
		return fmt.Errorf("git %s %s branch -d %s\n%s\n", this.GitDir, this.WorkTree, pakbranch, err.Error())
	}

	// delete paktag if contains
	var containTag bool
	containTag, err = this.ContainsPaktag()
	if err != nil {
		return err
	}
	if !containTag {
		return
	}
	err = this.removePaktag()
	if err != nil {
		return err
	}

	return
}

func (this GitPkg) removePaktag() (err error) {
	_, err = RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "tag", "-d", paktag))
	if err != nil {
		err = fmt.Errorf("git %s %s tag -d %s\n%s\n", this.GitDir, this.WorkTree, paktag, err.Error())
	}

	return
}

func (this GitPkg) GetPak() (string, error) {
	return this.metaGetPak(this.Remote+"/"+this.Branch)
}

func (this GitPkg) GetPakByChecksum() (string, error) {
	return this.metaGetPak(this.Checksum)
}

// try to figure out the plumbing command for git-checkout
func (this GitPkg) metaGetPak(checksumOrRemoteBranchRef string) (checksum string, err error) {
	_, err = RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "checkout", "-b", pakbranch, checksumOrRemoteBranchRef))
	if err != nil {
		err = fmt.Errorf("git %s %s checkout -b %s %s\n%s\n", this.GitDir, this.WorkTree, pakbranch, checksumOrRemoteBranchRef, err.Error())
		return
	}

	checksum, err = this.GetChecksumHash(this.Pakbranch)
	if err != nil {
		return
	}
	err = this.tagPakbranch(checksum)

	return
}

func (this GitPkg) tagPakbranch(checksumHash string) (err error) {
	_, err = RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "tag", paktag, checksumHash))
	if err != nil {
		err = fmt.Errorf("git %s %s tag %s %s\n%s\n", this.GitDir, this.WorkTree, paktag, checksumHash, err.Error())
	}

	return
}

func (this GitPkg) EqualPakBranchAndTag() (bool, error) {
	pakbranchHash, err := this.GetChecksumHash(this.Pakbranch)
	if err != nil {
		return false, err
	}
	paktagHash, err := this.GetChecksumHash(this.Paktag)
	if err != nil {
		return false, err
	}

	if paktagHash != pakbranchHash {
		return false, err
	}

	return true, err
}

func (this GitPkg) ContainsPakbranch() (bool, error) {
	out, err := RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "show-ref"))
	if err != nil {
		return false, fmt.Errorf("git %s %s show-ref\n%s\n", this.GitDir, this.WorkTree, err.Error())
	}

	return strings.Contains(out.String(), " "+this.Pakbranch+"\n"), nil
}

func (this GitPkg) ContainsPaktag() (bool, error) {
	out, err := RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "show-ref"))
	if err != nil {
		return false, fmt.Errorf("git %s %s show-ref\n%s\n", this.GitDir, this.WorkTree, err.Error())
	}

	return strings.Contains(out.String(), " "+this.Paktag+"\n"), nil
}

func (this GitPkg) IsClean() (bool, error) {
	out, err := RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "status", "--porcelain", "--untracked-files=no"))
	if err != nil {
		return false, fmt.Errorf("git %s %s status --porcelain --untracked-files=no\n%s\n", this.GitDir, this.WorkTree, err.Error())
	}

	return out.String() == "", nil
}

func (this GitPkg) GetChecksumHash(target string) (string, error) {
	out, err := RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "show-ref", target, "--hash"))
	if err != nil {
		return "", fmt.Errorf("git %s %s show-ref %s --hash\n%s\n", this.GitDir, this.WorkTree, target, err.Error())
	}

	checksum := out.String()[:40]

	return checksum, nil
}

func (this GitPkg) Fetch() error {
	_, err := RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "fetch", this.Remote, this.Branch))
	if err != nil {
		err = fmt.Errorf("git %s %s fetch %s %s\n%s\n", this.GitDir, this.WorkTree, this.Remote, this.Branch, err.Error())
	}

	return err
}

var unrecognizedPakbranch = "Found unrecognized Pakbranch in %s.\n Please rename your branch or use -f to force Override it."
func (this GitPkg) PakbranchRemovable() (bool, error) {
	containPakbranch, err := this.ContainsPakbranch()
	if err != nil {
		return false, err
	}

	if containPakbranch {
		containPaktag, err := this.ContainsPaktag()
		if err != nil {
			return false, err
		}
		if !containPaktag {
			return false, fmt.Errorf(unrecognizedPakbranch, this.Name)
		}

		equal, err := this.EqualPakBranchAndTag()
		if err != nil {
			return false, err
		}

		if !equal {
			return false, fmt.Errorf(unrecognizedPakbranch, this.Name)
		}
	} else {
		return false, nil
	}

	return true, nil
}

func (this GitPkg) ContainsRemoteBranch() (bool, error) {
	out, err := RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "show-ref"))
	if err != nil {
		return false, fmt.Errorf("git %s %s show-ref\n%s\n", this.GitDir, this.WorkTree, err.Error())
	}

	return strings.Contains(out.String(), " "+this.RemoteBranch+"\n"), nil
}

func (this GitPkg) EqualChecksumAndPakbranch() (bool, error) {
    checksumChecksum, err := this.GetChecksumHash(this.Checksum)
    if err != nil {
        return false, err
    }

    pakbranchChecksum, err := this.GetChecksumHash(this.Pakbranch)
    if err != nil {
        return false, err
    }

    return checksumChecksum == pakbranchChecksum, nil
}

func (this GitPkg) GetHeadRefName() (string, error) {
    out, err := RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "symbolic-ref",  "HEAD"))
    if err != nil {
        return "", fmt.Errorf("git %s %s symbolic-ref HEAD", this.GitDir, this.WorkTree)
    }

    refs := out.String()
    return refs[:len(refs)-1], err
}

func (this GitPkg) GetHeadChecksum() (string, error) {
    headBranch, err := this.GetHeadRefName()
    if err != nil {
        return "", err
    }

    out, err := RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "show-ref",  "--hash", headBranch))
    if err != nil {
        return "", fmt.Errorf("git %s %s symbolic-ref HEAD", this.GitDir, this.WorkTree)
    }

    refs := out.String()
    return refs[:len(refs)-1], err
}