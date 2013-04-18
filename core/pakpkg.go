package core

import (
	"bytes"
	"fmt"
	. "github.com/theplant/pak/share"
	"os"
	"os/exec"
	"strings"
)

type PakPkg struct {
	PkgProxy
	GetOption

	Name              string
	Remote            string
	Branch            string
	RemoteBranch      string
	Path              string
	HeadRefName       string
	HeadChecksum      string
	PakbranchChecksum string
	PaktagChecksum    string
	PakbranchRef      string
	PaktagRef         string

	// Note:
	// Containing branch named pak does not mean that pkg is managed by pak.
	// Containing tag named _pak_latest_ means this pkg is managed by pak, but
	// still can't make sure the pkg is on the pak branch or it's status is wanted
	// by Pakfile or Pakfile.lock.
	PkgExist               bool
	IsRemoteBranchExist    bool
	ContainsBranchNamedPak bool
	ContainsPaktag         bool
	OnPakbranch            bool
	OwnPakbranch           bool
	IsClean                bool
}

func NewPakPkg(name, remote, branch string) PakPkg {
	pkg := PakPkg{}
	pkg.Name = name
	pkg.Remote = remote
	pkg.Branch = branch
	pkg.Path = fmt.Sprintf("%s/src/%s", Gopath, name)

	return pkg
}

func (this *PakPkg) IsPkgExist() (bool, error) {
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

// Dial will try to figure out what kind version control system that the
// package is using and initialize the PkgProxy accordingly.
// TODO: invoke this function inside of Sync
func (this *PakPkg) Dial() error {
	for _, engine := range PkgProxyList {
		tracking, err := engine.IsTracking(this.Name)
		if err != nil {
			return err
		}
		if tracking {
			this.PkgProxy = engine.NewVCS(this.Name, this.Remote, this.Branch)
			return nil
		}
	}

	return fmt.Errorf("%s: Can't Detect What Kind of VCS it's Using.", this.Name)
}

// GoGet invokes command `go get {package}`.
func (this *PakPkg) GoGet() error {
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

func (this *PakPkg) Sync() (err error) {
	this.PakbranchRef = this.GetPakbranchRef()
	this.PaktagRef = this.GetPaktagRef()
	this.RemoteBranch = this.GetRemoteBranch()

	// Should be Under the Control of Git
	this.HeadRefName, err = this.GetHeadRefName()
	if err != nil {
		return
	}

	this.HeadChecksum, err = this.GetHeadChecksum()
	if err != nil {
		return err
	}

	// Retrieve info

	// Branch Named Pak
	this.ContainsBranchNamedPak, err = this.ContainsPakbranch()
	if err != nil {
		return
	}

	if this.ContainsBranchNamedPak {
		this.PakbranchChecksum, err = this.GetChecksum(this.PakbranchRef)
		if err != nil {
			return
		}
	}

	// Paktag _pak_latest_
	this.ContainsPaktag, err = this.PkgProxy.ContainsPaktag()
	if err != nil {
		return
	}
	if this.ContainsPaktag {
		this.PaktagChecksum, err = this.GetChecksum(this.PaktagRef)
		if err != nil {
			return
		}
	}

	this.OwnPakbranch = this.ContainsBranchNamedPak && this.ContainsPaktag &&
		this.PaktagChecksum == this.PakbranchChecksum

	// on Pakbranch
	// TODO: add OnPakbranch Test(same checksum but different refs)
	this.OnPakbranch = this.ContainsBranchNamedPak && this.ContainsPaktag &&
		this.PaktagChecksum == this.PakbranchChecksum &&
		this.PakbranchChecksum == this.HeadChecksum &&
		this.PakbranchRef == this.HeadRefName

	this.IsClean, err = this.PkgProxy.IsClean()
	if err != nil {
		return
	}

	this.IsRemoteBranchExist, err = this.ContainsRemoteBranch()
	if err != nil {
		return
	}

	return
}

func (this *PakPkg) Report() error {
	if !this.IsClean {
		return fmt.Errorf("Package %s is not clean. Please clean it up before running pak.", this.Name)
	}

	if !this.IsRemoteBranchExist {
		return fmt.Errorf("`%s` does not contain reference `%s`", this.Name, this.RemoteBranch)
	}

	return nil
}

func (this *PakPkg) Pak(option GetOption) (string, error) {
	// TODO: add tests
	if this.OnPakbranch && this.PakbranchChecksum == option.Checksum && !option.Force {
		err := this.GoGet()
		if err != nil {
			return "", err
		}

		return this.PakbranchChecksum, nil
	}

	err := this.Unpak(option.Force)
	if err != nil {
		return "", err
	}

	var ref = this.RemoteBranch
	if option.Checksum != "" {
		ref = option.Checksum
	}

	checksum, err := this.PkgProxy.Pak(ref)
	if err != nil {
		return "", err
	}

	err = this.GoGet()
	if err != nil {
		return "", err
	}

	return checksum, nil
}

func (this *PakPkg) Unpak(force bool) (err error) {
	if this.ContainsBranchNamedPak && !this.OwnPakbranch && !force {
		return fmt.Errorf("Package %s Contains Branch Named pak. Please use pak with -f flag or manually remove/rename the branch in the package.", this.Name)
	}

	return this.PkgProxy.Unpak()
}

func ParsePakState(pakfilePakPkgs []PakPkg, paklockInfo PaklockInfo) (newPkgs []PakPkg, toUpdatePkgs []PakPkg, toRemovePkgs []PakPkg) {
	if paklockInfo != nil {
		for _, pakPkg := range pakfilePakPkgs {
			if paklockInfo[pakPkg.Name] != "" {
				pakPkg.Checksum = paklockInfo[pakPkg.Name]
				toUpdatePkgs = append(toUpdatePkgs, pakPkg)
				delete(paklockInfo, pakPkg.Name)
			} else {
				newPkgs = append(newPkgs, pakPkg)
			}
		}
		if len(paklockInfo) != 0 {
			for key, val := range paklockInfo {
				pakPkg := NewPakPkg(key, "", "")
				pakPkg.Checksum = val
				toRemovePkgs = append(toRemovePkgs, pakPkg)
			}
		}
	} else {
		newPkgs = pakfilePakPkgs
	}

	return
}

/**
 * Supported Pkg Description Formats:
 * 		"github.com/theplant/package2"
 * 		"github.com/theplant/package2@dev"
 * 		"github.com/theplant/package2@origin/dev"
 */
func ParsePakfile() ([]PakPkg, error) {
	pakInfo, err := GetPakInfo()

	if err != nil {
		return nil, err
	}

	pakPkgs := []PakPkg{}
	for _, pkg := range pakInfo.Packages {
		atIndex := strings.LastIndex(pkg, "@")
		var name, remote, branch string
		if atIndex != -1 {
			name = pkg[:atIndex]
			branchInfo := pkg[atIndex+1:]
			if strings.Contains(branchInfo, "/") {
				slashIndex := strings.Index(branchInfo, "/")
				remote = branchInfo[:slashIndex]
				branch = branchInfo[slashIndex+1:]
			} else {
				remote = "origin"
				branch = branchInfo
			}
		} else {
			name = pkg
			remote = "origin"
			branch = "master"
		}

		pakPkg := NewPakPkg(name, remote, branch)
		pakPkgs = append(pakPkgs, pakPkg)
	}

	return pakPkgs, nil
}
