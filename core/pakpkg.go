package core

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	. "github.com/theplant/pak/share"
	"github.com/wsxiaoys/terminal/color"
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

	Remote            string
	Branch            string
	RemoteBranch      string
	Path              string
	HeadRefName       string
	HeadChecksum      string
	PakbranchChecksum string
	PakbranchRef      string

	IsRemoteBranchExist bool
	IsChecksumExist     bool // Checksum in GetOption
	HasPakBranch        bool
	OnPakbranch         bool
	IsClean             bool

	ShouldFetchBeforePak bool
}

func NewPakPkg(cfg PkgCfg) PakPkg {
	pkg := PakPkg{PkgCfg: cfg}

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
	// if this.Verbose {
	// 	color.Printf("Checking @g%s@w.\n", this.Name)
	// }

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

	// if this.ActionType != "Remove" {
	// 	err = this.Fetch()
	// 	if err != nil {
	// 		return nameAndChecksum, err
	// 	}
	// }

	err = this.Sync()
	if err != nil {
		return nameAndChecksum, err
	}

	isFetched := false
	if this.ShouldFetchBeforePak {
		err := this.Fetch()
		if err != nil {
			return nameAndChecksum, err
		}
		isFetched = true

		this.IsRemoteBranchExist, err = this.ContainsRemoteBranch()
		if err != nil {
			return nameAndChecksum, err
		}

		if this.Checksum != "" {
			this.IsChecksumExist, err = this.PkgProxy.IsChecksumExist(this.Checksum)
			if err != nil {
				return nameAndChecksum, err
			}
		}
	}

	if this.ActionType != "Remove" {
		err = this.Report()
		if err != nil {
			return nameAndChecksum, err
		}
	}

	switch this.ActionType {
	case "New":
		if !this.IsClean {
			return nameAndChecksum, fmt.Errorf("Package %s is a New Package and is Not Clean.\n", this.Name)
		}

		if !isFetched {
			err = this.Fetch()
			if err != nil {
				return nameAndChecksum, err
			}
		}

		originalChecksum := this.Checksum
		originalHeadRef := this.HeadRefName
		this.Checksum = ""
		checksum, err := this.Pak()
		if err != nil {
			return nameAndChecksum, err
		}

		if this.Verbose && (originalChecksum != checksum || originalHeadRef != this.HeadRefName) {
			color.Printf("Paked @g%s@w.\n", this.Name)
		}

		nameAndChecksum[0] = this.Name
		nameAndChecksum[1] = checksum
	case "Update":
		// Can't Update/Get(by PakMeter) specific packages which is not clean
		if !this.IsClean {
			if this.UsingPakMeter {
				return nameAndChecksum, fmt.Errorf("Package %s is Not Clean.\n", this.Name)
			}

			nameAndChecksum[0] = this.Name
			nameAndChecksum[1] = this.PakbranchChecksum

			return nameAndChecksum, nil
		}

		originalChecksum := this.HeadChecksum
		originalHeadRef := this.HeadRefName
		checksum, err := this.Pak()
		if err != nil {
			return nameAndChecksum, err
		}

		if this.Verbose && (originalChecksum != checksum || originalHeadRef != this.HeadRefName) {
			color.Printf("Synced @g%s@w.\n", this.Name)
		}

		nameAndChecksum[0] = this.Name
		nameAndChecksum[1] = checksum
	case "Remove":
		if this.OnPakbranch && this.IsClean {
			err = this.Unpak(this.Force)
			if err != nil {
				return nameAndChecksum, err
			}
		}

		if this.Verbose {
			color.Printf("@yUnPaked @g%s@w.\n", this.Name)
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

	return fmt.Errorf("Package %s: Can't Find Out Type of Versino Control System.\n", this.Name)
}

// GoGet invokes command `go get {package}`.
// TODO: duplicate with fetch, pick one
func (this *PakPkg) GoGet() error {
	return GoGetImpl(this.Name)
}

var GoGetImpl = func(name string) error {
	cmd := exec.Command("go", "get", name)
	cmd.Stderr = &bytes.Buffer{}
	cmd.Stdout = &bytes.Buffer{}
	err := cmd.Run()
	if err != nil {
		getErr := cmd.Stderr.(*bytes.Buffer).String()
		if getErr == "" {
			getErr = cmd.Stdout.(*bytes.Buffer).String()
		}
		err = fmt.Errorf("go get %s:\n%s", name, getErr)
		color.Printf("@r%s\n", err)
	}

	// GO GET error should not get in the way of paking
	return nil
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

	this.OnPakbranch = this.HasPakBranch && this.PakbranchChecksum == this.HeadChecksum && this.PakbranchRef == this.HeadRefName

	this.IsClean, err = this.PkgProxy.IsClean()
	if err != nil {
		return
	}

	this.IsRemoteBranchExist, err = this.ContainsRemoteBranch()
	if err != nil {
		return
	}

	if this.Checksum != "" {
		this.IsChecksumExist, err = this.PkgProxy.IsChecksumExist(this.Checksum)
		if err != nil {
			return
		}
	}

	this.ShouldFetchBeforePak = !this.IsRemoteBranchExist || (this.Checksum != "" && !this.IsChecksumExist)

	return
}

func (this *PakPkg) Report() error {
	if !this.IsClean && !this.SkipUncleanPkgs {
		return fmt.Errorf("Package %s is Not Clean. Please clean it up before running pak.\n", this.Name)
	}

	if !this.IsRemoteBranchExist {
		return fmt.Errorf("`%s` does not contain reference `%s`\n", this.Name, this.RemoteBranch)
	}

	if this.Checksum != "" {
		if !this.IsChecksumExist {
			return fmt.Errorf("`%s` does not contain commit `%s`\n", this.Name, this.Checksum)
		}
	}

	return nil
}

func (this *PakPkg) Pak() (string, error) {
	// TODO: add tests
	if this.ShouldNotUpdate() {
		err := this.GoGet()
		if err != nil {
			return "", err
		}

		return this.PakbranchChecksum, nil
	}

	option := this.GetOption
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

	// Ignore error handling for go get action for concurrent paking sometimes
	// will affect compilation if some packages are dependening on other packages
	// which are still under updating. That would leads to compile error.
	this.GoGet()
	// err = this.GoGet()
	// if err != nil {
	// 	return "", err
	// }

	this.HeadRefName = this.PakbranchRef

	return checksum, nil
}

func (this *PakPkg) ShouldNotUpdate() bool {
	return this.OnPakbranch && this.PakbranchChecksum == this.GetOption.Checksum && !this.GetOption.Force
}

func (this *PakPkg) Unpak(force bool) (err error) {
	if !this.HasPakBranch {
		return
	}

	return this.PkgProxy.Unpak()
}

func CategorizePakPkgs(pakfilePakPkgs *[]PakPkg, paklockInfo PaklockInfo, option PakOption) {
	if !option.UsePakfileLock {
		for i, pakPkg := range *pakfilePakPkgs {
			(*pakfilePakPkgs)[i].Checksum = paklockInfo[pakPkg.Name]
			(*pakfilePakPkgs)[i].ActionType = "New"
		}

		return
	}

	for i, pakPkg := range *pakfilePakPkgs {
		(*pakfilePakPkgs)[i].Checksum = paklockInfo[pakPkg.Name]
		if paklockInfo[pakPkg.Name] == "" {
			(*pakfilePakPkgs)[i].ActionType = "New"
		} else {
			(*pakfilePakPkgs)[i].ActionType = "Update"

			delete(paklockInfo, pakPkg.Name)
		}
	}

	if len(option.PakMeter) > 0 {
		return
	}

	if len(paklockInfo) != 0 {
		for key, val := range paklockInfo {
			pakPkg := NewPakPkg(PkgCfg{Name: key, PakName: Pakbranch})
			pakPkg.Checksum = val
			pakPkg.ActionType = "Remove"

			pkgs := append(*pakfilePakPkgs, pakPkg)
			*pakfilePakPkgs = pkgs
		}
	}

	return
}

// /**
//  * This format is ONLY supported in Pakfile-1.0
//  * Supported Pkg Description Formats:
//  * 		"github.com/theplant/package2"
//  * 		"github.com/theplant/package2@dev"
//  * 		"github.com/theplant/package2@origin/dev"
//  */
// // TODO: make it compartible with Pakfile-1.0 by adding a PakfileVersion configration
// func ParsePakfile() ([]PakPkg, PaklockInfo, error) {
// 	// pakInfo, err := GetPakInfo("")
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	pakPkgs := []PakPkg{}
// 	for _, pkgCfg := range pakInfo.Packages {
// 		// atIndex := strings.LastIndex(pkg, "@")
// 		// var name, remote, branch string
// 		// if atIndex != -1 {
// 		// 	name = pkg[:atIndex]
// 		// 	branchInfo := pkg[atIndex+1:]
// 		// 	if strings.Contains(branchInfo, "/") {
// 		// 		slashIndex := strings.Index(branchInfo, "/")
// 		// 		remote = branchInfo[:slashIndex]
// 		// 		branch = branchInfo[slashIndex+1:]
// 		// 	} else {
// 		// 		remote = "origin"
// 		// 		branch = branchInfo
// 		// 	}
// 		// } else {
// 		// 	name = pkg
// 		// 	remote = "origin"
// 		// 	branch = "master"
// 		// }

// 		// pakPkg := NewPakPkg(name, remote, branch)

// 		pakPkg := NewPakPkg(pkgCfg)

// 		pakPkgs = append(pakPkgs, pakPkg)
// 	}

// 	return pakPkgs, nil
// }
