package core

import (
	"bytes"
	"fmt"
	. "github.com/theplant/pak/share"
	"github.com/wsxiaoys/terminal/color"
	"os"
	"os/exec"
	"strings"
)

// For PakPkg#Get
type GetOption struct {
	Checksum        string
	Force           bool
	SkipUncleanPkgs bool
	UsingPakMeter   bool
	Verbose         bool
	ActionType      string // values ranging from => New, Update, Remove
}

type PkgCfg struct {
	Name                   string
	PakName                string // default: pak
	TargetBranch           string
	AutoMatchingHostBranch bool
}

type PakPkg struct {
	PkgProxy
	GetOption
	PkgCfg

	// Name              string // moved to PkgCfg
	Remote            string
	Branch            string
	RemoteBranch      string
	Path              string
	HeadRefName       string
	HeadChecksum      string
	PakbranchChecksum string
	// PaktagChecksum    string
	PakbranchRef string
	// PaktagRef    string

	// Note:
	// Containing abranch named pak does not mean that pkg is managed by pak.
	// However, containing a tag named _pak_latest_ means this pkg is managed by
	// pak, but still can't make sure the pkg is on the pak branch or it's status
	// is consistent with Pakfile.lock.
	IsRemoteBranchExist bool
	// ContainsBranchNamedPak bool
	// ContainsPaktag         bool
	HasPakBranch bool
	OnPakbranch  bool
	// OwnPakbranch           bool
	IsClean bool
}

// func NewPakPkg(name, remote, branch string) PakPkg {
func NewPakPkg(cfg PkgCfg) PakPkg {
	pkg := PakPkg{PkgCfg: cfg}
	// pkg.Name = cfg.Name

	var remote, branch string
	if strings.Contains(cfg.TargetBranch, "/") {
		tb := strings.Split(cfg.TargetBranch, "/")
		remote = tb[0]
		branch = tb[1]
	} else {
		remote = ""
		branch = cfg.TargetBranch
	}
	pkg.Remote = remote
	pkg.Branch = branch

	pkg.Path = fmt.Sprintf("%s/src/%s", Gopath, pkg.Name)

	return pkg
}

// TODO: refactor to multiple small methods
func (this *PakPkg) Get() (nameAndChecksum [2]string, err error) {
	if this.Verbose {
		color.Printf("Checking @g%s@w.\n", this.Name)
	}

	// Go Get Package when the Package is not existing
	isPkgExist, err := this.IsPkgExist()
	if err != nil {
		return nameAndChecksum, err
	}
	if !isPkgExist && this.ActionType != "Remove" {
		err = this.GoGet()
		if err != nil {
			return nameAndChecksum, err
		}
	}

	err = this.Dial()
	if err != nil {
		return nameAndChecksum, err
	}

	// Fetch Before Hand can Make Sure That the Package Contains Up-To-Date Remote Branch
	if this.ActionType != "Remove" {
		err = this.Fetch()
		if err != nil {
			return nameAndChecksum, err
		}
	}

	err = this.Sync()
	if err != nil {
		return nameAndChecksum, err
	}

	if this.ActionType != "Remove" {
		err = this.Report()
		if err != nil {
			return nameAndChecksum, err
		}
	}

	switch this.ActionType {
	case "New":
		if this.Verbose {
			color.Printf("Getting @g%s@w.\n", this.Name)
		}

		if !this.IsClean {
			return nameAndChecksum, fmt.Errorf("Package %s is a New Package and is Not Clean.", this.Name)
		}

		checksum, err := this.Pak()
		if err != nil {
			return nameAndChecksum, err
		}

		// newPaklockInfo[newPakPkgs[i].Name] = checksum
		nameAndChecksum[0] = this.Name
		nameAndChecksum[1] = checksum
	case "Update":
		if this.Verbose {
			color.Printf("Updating @g%s@w.\n", this.Name)
		}

		// Can't Update/Get(by PakMeter) specific packages which is not clean
		if !this.IsClean {
			if this.UsingPakMeter {
				return nameAndChecksum, fmt.Errorf("Package %s is Not Clean.", this.Name)
			}

			nameAndChecksum[0] = this.Name
			nameAndChecksum[1] = this.PakbranchChecksum

			return nameAndChecksum, nil
		}

		checksum, err := this.Pak()
		if err != nil {
			return nameAndChecksum, err
		}

		// newPaklockInfo[toUpdatePakPkgs[i].Name] = checksum
		nameAndChecksum[0] = this.Name
		nameAndChecksum[1] = checksum
	case "Remove":
		if this.Verbose {
			color.Printf("@rUnpaking @g%s@w.\n", this.Name)
		}

		if this.OnPakbranch && this.IsClean {
			err = this.Unpak(this.Force)
			if err != nil {
				return nameAndChecksum, err
			}
		}

		// TODO: Add tests for removing Pakfile.lock record
		nameAndChecksum[0] = this.Name
		nameAndChecksum[1] = ""
	}

	return
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
			this.PkgProxy = engine.NewVCS(this.Name, this.Remote, this.Branch, this.PakName)
			return nil
		}
	}

	return fmt.Errorf("Package %s: Can't Find Out Type of Versino Control System.", this.Name)
}

// GoGet invokes command `go get {package}`.
// TODO: duplicate with fetch, pick one
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
	// this.PaktagRef = this.GetPaktagRef()
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

	// ============ Retrieve info ============

	// Branch Named Pak
	this.HasPakBranch, err = this.ContainsPakbranch()
	if err != nil {
		return
	}

	if this.HasPakBranch {
		this.PakbranchChecksum, err = this.GetChecksum(this.PakbranchRef)
		if err != nil {
			return
		}
	}

	// // Paktag _pak_latest_
	// this.ContainsPaktag, err = this.PkgProxy.ContainsPaktag()
	// if err != nil {
	// 	return
	// }
	// if this.ContainsPaktag {
	// 	this.PaktagChecksum, err = this.GetChecksum(this.PaktagRef)
	// 	if err != nil {
	// 		return
	// 	}
	// }

	// this.OwnPakbranch = this.ContainsBranchNamedPak && this.ContainsPaktag &&
	// 	this.PaktagChecksum == this.PakbranchChecksum

	// // on Pakbranch
	// // TODO: add OnPakbranch Test(same checksum but different refs)
	// this.OnPakbranch = this.ContainsBranchNamedPak && this.ContainsPaktag &&
	// 	this.PaktagChecksum == this.PakbranchChecksum &&
	// 	this.PakbranchChecksum == this.HeadChecksum &&
	// 	this.PakbranchRef == this.HeadRefName
	this.OnPakbranch = this.HasPakBranch && this.PakbranchChecksum == this.HeadChecksum && this.PakbranchRef == this.HeadRefName

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
	if !this.IsClean && !this.SkipUncleanPkgs {
		return fmt.Errorf("Package %s is not clean. Please clean it up before running pak.", this.Name)
	}

	if !this.IsRemoteBranchExist {
		return fmt.Errorf("`%s` does not contain reference `%s`", this.Name, this.RemoteBranch)
	}

	return nil
}

func (this *PakPkg) Pak() (string, error) {
	// TODO: add tests
	option := this.GetOption
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
	// if this.ContainsBranchNamedPak && !this.OwnPakbranch && !force {
	if !this.HasPakBranch {
		return
	}
	// if !force {
	// 	return fmt.Errorf("Package %s Contains Branch Named pak that can't be recognized by pak.\nPlease use pak with -f flag to allow pak to remove the branch or you can manually remove/rename the branch in the package.", this.Name)
	// }

	return this.PkgProxy.Unpak()
}

func CategorizePakPkgs(pakfilePakPkgs *[]PakPkg, paklockInfo PaklockInfo, option PakOption) {
	// if paklockInfo == nil {
	if !option.UsePakfileLock {
		for i, _ := range *pakfilePakPkgs {
			(*pakfilePakPkgs)[i].ActionType = "New"
		}
		return
	}

	for i, pakPkg := range *pakfilePakPkgs {
		if paklockInfo[pakPkg.Name] == "" {
			(*pakfilePakPkgs)[i].ActionType = "New"
		} else {
			(*pakfilePakPkgs)[i].Checksum = paklockInfo[pakPkg.Name]
			(*pakfilePakPkgs)[i].ActionType = "Update"

			delete(paklockInfo, pakPkg.Name)
		}
	}

	if len(option.PakMeter) > 0 {
		return
	}

	if len(paklockInfo) != 0 {
		for key, val := range paklockInfo {
			// pakPkg := NewPakPkg(key, "", "")
			pakPkg := NewPakPkg(PkgCfg{Name: key, PakName: Pakbranch})
			pakPkg.Checksum = val
			pakPkg.ActionType = "Remove"

			pkgs := append(*pakfilePakPkgs, pakPkg)
			*pakfilePakPkgs = pkgs
		}
	}

	return
}

/**
 * This format is ONLY supported in Pakfile-1.0
 * Supported Pkg Description Formats:
 * 		"github.com/theplant/package2"
 * 		"github.com/theplant/package2@dev"
 * 		"github.com/theplant/package2@origin/dev"
 */
// TODO: make it compartible with Pakfile-1.0 by adding a PakfileVersion configration
func ParsePakfile() ([]PakPkg, error) {
	pakInfo, err := GetPakInfo("")

	if err != nil {
		return nil, err
	}

	pakPkgs := []PakPkg{}
	for _, pkgCfg := range pakInfo.Packages {
		// atIndex := strings.LastIndex(pkg, "@")
		// var name, remote, branch string
		// if atIndex != -1 {
		// 	name = pkg[:atIndex]
		// 	branchInfo := pkg[atIndex+1:]
		// 	if strings.Contains(branchInfo, "/") {
		// 		slashIndex := strings.Index(branchInfo, "/")
		// 		remote = branchInfo[:slashIndex]
		// 		branch = branchInfo[slashIndex+1:]
		// 	} else {
		// 		remote = "origin"
		// 		branch = branchInfo
		// 	}
		// } else {
		// 	name = pkg
		// 	remote = "origin"
		// 	branch = "master"
		// }

		// pakPkg := NewPakPkg(name, remote, branch)

		pakPkg := NewPakPkg(pkgCfg)

		pakPkgs = append(pakPkgs, pakPkg)
	}

	return pakPkgs, nil
}
