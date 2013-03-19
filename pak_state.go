package pak

import (
	"github.com/theplant/pak/gitpkg"
	. "github.com/theplant/pak/share"
	"strings"
)

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
				pakPkg := PakPkg{GitPkg: gitpkg.NewGitPkg(key, "", "")}
                _ = val
                pakPkg.Checksum = val
				toRemovePkgs = append(toRemovePkgs, pakPkg)
			}
		}
	} else {
		newPkgs = pakfilePakPkgs
	}

	return
}

// Supported Pkg description format
// "github.com/theplant/package2"
// "github.com/theplant/package2@dev"
// "github.com/theplant/package2@origin/dev"
func parsePakfile() ([]PakPkg, error) {
	pakInfo, err := GetPakInfo()
	if err != nil {
		return nil, err
	}

    // gitPkgs := []PakPkg{}
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

        pakPkgs = append(pakPkgs, PakPkg{GitPkg: gitpkg.NewGitPkg(name, remote, branch)})
	}

	return pakPkgs, nil
}
