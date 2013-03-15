package pak

import (
	"fmt"
	"os/exec"
	"bytes"
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

func (this GitPkg) DeletePakbranch(branch string) {
	out := MustRun(exec.Command("git", this.GitDir, this.WorkTree, "branch", "-d", pakbranch))
}

func (this GitPkg) NewBranch(targetBranch string) (result bool) {
	out := MustRun(exec.Command("git", this.GitDir, this.WorkTree, "checkout", "-b", pakbranch, targetBranch))

	result = out.String() == fmt.Sprintf("Switched to a new branch '%s'", pakbranch)

	if result {
		this.TagPakbranch()
	}

	return
}

func (this GitPkg) TagPakbranch(checksumHash string) {
	MustRun(exec.Command("git", this.GitDir, this.WorkTree, "tag", paktag, checksumHash))
}

func (this GitPkg) DeletePaktag() {
	MustRun(exec.Command("git", this.GitDir, this.WorkTree, "tag", "-d", paktag))
}

func (this GitPkg) EqualPakBranchAndTag() bool {
	pakbranchHash := this.GetChecksumHash(this.Pakbranch)
	paktagHash := this.GetChecksumHash(this.Paktag)

	if paktagHash == pakbranchHash {
		return true
	}

	return false
}

func (this GitPkg) ContainsPakbranch() bool {
	return this.GetChecksumHash(this.Pakbranch) != ""
}

func (this GitPkg) ContainPaktag() bool {
	return this.GetChecksumHash(this.Paktag) != ""
}

func (this GitPkg) IsClean() bool {
	out := MustRun(exec.Command("git", this.GitDir, this.WorkTree, "status", "--porcelain", "--untracked-files=no"))

	return out.String() == ""
}

func (this GitPkg) GetChecksumHash(target string) string {
	out := MustRun(exec.Command("git", this.GitDir, this.WorkTree, "show-ref", target, "--hash"))

	return out.String()
}