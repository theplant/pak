package gitpkg

import (
	"fmt"
	. "github.com/theplant/pak/share"
	// "os/exec"
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
	// TODO: add tests
	if this.State.OnPakbranch && this.PakbranchChecksum == option.Checksum && !option.Force {
		// Go Get Package
		if !option.NotGet {
			err := this.GoGet()
			if err != nil {
				return "", err
			}
		}

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
	_, err = this.Run("checkout", "-b", Pakbranch, ref)
	if err != nil {
		return "", err
	}

	// Create Paktag
	checksum, err := this.GetChecksum(this.Pakbranch)
	if err != nil {
		return "", err
	}
	_, err = this.Run("tag", Paktag, checksum)
	if err != nil {
		return "", err
	}

	// Go Get Package
	if !option.NotGet {
		err = this.GoGet()
		if err != nil {
			return "", err
		}
	}

	return checksum, err
}

func (this *GitPkg) Unpak(force bool) (err error) {
	if this.State.ContainsBranchNamedPak && !this.State.OwnPakbranch && !force {
		return fmt.Errorf("Package %s Contains Branch Named pak. Please use pak with -f flag or manually remove/rename the branch in the package.", this.Name)
	}

	// Move to Master Branch
	_, err = this.Run("checkout", "master")
	if err != nil {
		return err
	}

	// Delete Pakbranch
	if this.State.ContainsBranchNamedPak {
		_, err = this.Run("branch", "-D", Pakbranch)
		if err != nil {
			return err
		}
	}

	// Delete Paktag
	if this.State.ContainsPaktag {
		_, err = this.Run("tag", "-d", Paktag)
	}

	return
}
