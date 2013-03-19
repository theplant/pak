package gitpkg

import(
	"fmt"
	"os/exec"
)

// func (this *GitPkg) Report() error {
//
// }
//
func (this *GitPkg) Get(fetchLatest bool, force bool, useChecksum bool) (string, error) {
    if !this.State.IsClean {
		return "", fmt.Errorf("Package %s is not clean. Please clean it and re-start pak.", this.Name)
	}

    if !this.State.IsRemoteBranchExist {
		return "", fmt.Errorf("`%s` does not contain reference `%s`", this.Name, this.RemoteBranch)
	}

	// fetch pkg before check out
	if fetchLatest {
		err := this.Fetch()
		if err != nil {
			return "", err
		}
	}

    if this.State.UnderPak {
        err := this.RemovePak()
        if err != nil {
            return "", err
        }
    }

	// create pakbranch
	var checksum string
    var err error
    checksum, err = this.GetPak()
	if err != nil {
		return "", err
	}

    return checksum, err
}

func (this *GitPkg) Unpak() (err error) {
    return
    // containPakbranch, err := this.ContainsPakbranch()
    // if err != nil {
    //     return "", err
    // }
    //
    // // Remove pakbranch
    // onPakbranch := false
    // if containPakbranch {
    //     removable, err := this.PakbranchRemovable();
    //     if err != nil {
    //         return "", err
    //     }
    //
    //     if !removable && !force {
    //         return "", fmt.Errorf(unrecognizedPakbranch)
    //     }
    //
    //     // using checksum means pak is getting pkgs by Pakfile.lock
    //     if useChecksum {
    //         var err error
    //         onPakbranch, err = this.EqualChecksumAndPakbranch()
    //         if err != nil {
    //             return "", err
    //         }
    //     }
    //
    //     if !onPakbranch {
    //         err := this.RemovePak()
    //         if err != nil {
    //             return "", err
    //         }
    //     }
    // }
}

// NOTE: will also try to delete pak tag
func (this *GitPkg) RemovePak() (err error) {
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

func (this *GitPkg) removePaktag() (err error) {
	_, err = RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "tag", "-d", paktag))
	if err != nil {
		err = fmt.Errorf("git %s %s tag -d %s\n%s\n", this.GitDir, this.WorkTree, paktag, err.Error())
	}

	return
}

func (this *GitPkg) GetPak() (string, error) {
	return this.metaGetPak(this.Remote+"/"+this.Branch)
}

func (this *GitPkg) GetPakByChecksum() (string, error) {
	return this.metaGetPak(this.Checksum)
}

// try to figure out the plumbing command for git-checkout
func (this *GitPkg) metaGetPak(checksumOrRemoteBranchRef string) (checksum string, err error) {
	_, err = RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "checkout", "-b", pakbranch, checksumOrRemoteBranchRef))
	if err != nil {
		err = fmt.Errorf("git %s %s checkout -b %s %s\n%s\n", this.GitDir, this.WorkTree, pakbranch, checksumOrRemoteBranchRef, err.Error())
		return
	}

	checksum, err = this.GetChecksum(this.Pakbranch)
	if err != nil {
		return
	}
	err = this.tagPakbranch(checksum)

	return
}

func (this *GitPkg) tagPakbranch(checksumHash string) (err error) {
	_, err = RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "tag", paktag, checksumHash))
	if err != nil {
		err = fmt.Errorf("git %s %s tag %s %s\n%s\n", this.GitDir, this.WorkTree, paktag, checksumHash, err.Error())
	}

	return
}
