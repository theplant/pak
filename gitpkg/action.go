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

func (this *GitPkg) Pak(option GetOption) (string, error) {
	if this.State.OnPakbranch {
		return this.PakbranchChecksum, nil
	}

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

	// Create Pakbranch
	_, err = RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "checkout", "-b", Pakbranch, ref))
	if err != nil {
		err = fmt.Errorf("git %s %s checkout -b %s %s\n%s\n", this.GitDir, this.WorkTree, Pakbranch, ref, err.Error())
		return "", err
	}

	// Create Paktag
	checksum, err := this.GetChecksum(this.Pakbranch)
	if err != nil {
		return "", err
	}
	_, err = RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "tag", Paktag, checksum))
	if err != nil {
		err = fmt.Errorf("git %s %s tag %s %s\n%s\n", this.GitDir, this.WorkTree, Paktag, checksum, err.Error())
	}

	return checksum, err
}

func (this *GitPkg) Unpak(force bool) (err error) {
	if this.State.ContainsBranchNamedPak && !this.State.OnPakbranch && !force {
		return fmt.Errorf("Package %s Contains Branch Named pak.\n Please use -f flag or manually remove the branch in the package.\n", this.Name)
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
