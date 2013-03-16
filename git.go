package pak

import (
	"fmt"
	"os/exec"
)

type GitPkg struct {
	Name string
	Path string
	WorkTree string
	GitDir string
	Pakbranch string
	Paktag string
}

func NewGitPkg(name string) (gitPkg GitPkg) {
	gitPkg.Pakbranch = "refs/heads/" + pakbranch
	gitPkg.Paktag = "refs/tags/" + paktag
	gitPkg.Name = name
	gitPkg.Path = getGitPkgPath(name)
	gitPkg.WorkTree = getGitWorkTreeOpt(gitPkg.Path)
	gitPkg.GitDir = getGitDirOpt(gitPkg.Path)

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

func (this GitPkg) DeletePakbranch(branch string) (err error) {
	_, err = RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "branch", "-d", pakbranch))

	return
}

func (this GitPkg) NewBranch(targetBranch string) (result bool, err error) {
	out, err := RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "checkout", "-b", pakbranch, targetBranch))
	if err != nil {
	    return
	}

	result = out.String() == fmt.Sprintf("Switched to a new branch '%s'", pakbranch)
	if result {
		checksum, err := this.GetChecksumHash(this.Pakbranch)
		if err != nil {
		    return false, err
		}
		err = this.TagPakbranch(checksum)
	}

	return
}

func (this GitPkg) TagPakbranch(checksumHash string) (err error) {
	_, err = RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "tag", paktag, checksumHash))

	return
}

func (this GitPkg) DeletePaktag() (err error) {
	_, err = RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "tag", "-d", paktag))

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

	if paktagHash == pakbranchHash {
		return true, err
	}

	return false, err
}

func (this GitPkg) ContainsPakbranch() (bool, error) {
	contain, err := this.GetChecksumHash(this.Pakbranch)

	return contain != "", err
}

func (this GitPkg) ContainPaktag() (bool, error) {
	contain, err := this.GetChecksumHash(this.Paktag)
	return contain != "", err
}

func (this GitPkg) IsClean() (bool, error) {
	out, err := RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "status", "--porcelain", "--untracked-files=no"))

	return out.String() == "", err
}

func (this GitPkg) GetChecksumHash(target string) (string, error) {
	out, err := RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "show-ref", target, "--hash"))

	return out.String(), err
}

func (this GitPkg) Fetch() (error) {
	_, err := RunCmd(exec.Command("git", this.GitDir, this.WorkTree, "fetch"))

	return err
}