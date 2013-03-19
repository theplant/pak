package gitpkg

import (
	"fmt"
	. "github.com/theplant/pak/share"
	"os/exec"
)

// Usage of Gitpkg
// gitPkg.New
// gitPkg.Sync
// gitPkg.Report
// gitPkg.Get
// gitPkg.Sync

func (this *GitPkg) Report() error {
	if !this.State.IsClean {
		return fmt.Errorf("Package %s is not clean. Please clean it and re-start pak.", this.Name)
	}

	if !this.State.IsRemoteBranchExist {
		return fmt.Errorf("`%s` does not contain reference `%s`", this.Name, this.RemoteBranch)
	}

	return nil
}

func (this *GitPkg) Get(option GetOption) (string, error) {
	// Fetch pkg Before Check Out
	if option.Fetch {
		err := this.Fetch()
		if err != nil {
			return "", err
		}
	}

	err := this.Unpak(option.Force)
	if err != nil {
		return "", err
	}

	var ref = this.RemoteBranch
	if option.Checksum != "" {
		ref = option.Checksum
	}

	return this.Pak(ref)
}

func (this *GitPkg) Unpak(force bool) (err error) {
	if this.State.OnPakbranch && !force {
		return nil
	}

	if this.State.ContainsBranchNamedPak && !this.State.UnderPak && !force {
		return fmt.Errorf("Package %s is not Managed by pak.\n Please use -f flag or remove branch named pak in the package.\n", this.Name)
	}

	// Move to Master Branch
	_, err = RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "checkout", "master"))
	if err != nil {
		return fmt.Errorf("git %s %s checkout master\n%s", this.GitDir, this.WorkTree, err.Error())
	}

	// Delete Pakbranch
	if this.State.ContainsBranchNamedPak {
		_, err = RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "branch", "-d", Pakbranch))
		if err != nil {
			return fmt.Errorf("git %s %s branch -d %s\n%s\n", this.GitDir, this.WorkTree, Pakbranch, err.Error())
		}
	}

	// Delete Paktag
	if this.State.ContainsPaktag {
		_, err = RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "tag", "-d", Paktag))
		if err != nil {
			err = fmt.Errorf("git %s %s tag -d %s\n%s\n", this.GitDir, this.WorkTree, Paktag, err.Error())
		}
	}

	return
}

func (this *GitPkg) Pak(checksumOrRemoteBranchRef string) (checksum string, err error) {
	if this.State.OnPakbranch {
		return "", nil
	}

	// Create Pakbranch
	_, err = RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "checkout", "-b", Pakbranch, checksumOrRemoteBranchRef))
	if err != nil {
		err = fmt.Errorf("git %s %s checkout -b %s %s\n%s\n", this.GitDir, this.WorkTree, Pakbranch, checksumOrRemoteBranchRef, err.Error())
		return
	}

	// Create Paktag
	checksum, err = this.GetChecksum(this.Pakbranch)
	if err != nil {
		return
	}
	_, err = RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "tag", Paktag, checksum))
	if err != nil {
		err = fmt.Errorf("git %s %s tag %s %s\n%s\n", this.GitDir, this.WorkTree, Paktag, checksum, err.Error())
	}

	return
}
