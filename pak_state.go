package pak

import (
	"strings"
	"github.com/theplant/pak/gitpkg"
)

func ParsePakState(pakfileGitPkgs []gitpkg.GitPkg, paklockInfo PaklockInfo) (newGitPkgs []gitpkg.GitPkg, toUpdateGitPkgs []gitpkg.GitPkg, toRemoveGitPkgs []gitpkg.GitPkg) {
	if paklockInfo != nil {
		for _, gitPkg := range pakfileGitPkgs {
			if paklockInfo[gitPkg.Name] != "" {
				gitPkg.Checksum = paklockInfo[gitPkg.Name]
				toUpdateGitPkgs = append(toUpdateGitPkgs, gitPkg)
				delete(paklockInfo, gitPkg.Name)
			} else {
				newGitPkgs = append(newGitPkgs, gitPkg)
			}
		}
		if len(paklockInfo) != 0 {
			for key, val := range paklockInfo {
				gitPkg := NewGitPkg(key, "", "")
				gitPkg.Checksum = val
				toRemoveGitPkgs = append(toRemoveGitPkgs, gitPkg)
			}
		}
	} else {
		newGitPkgs = pakfileGitPkgs
	}

	return
}

// Supported Pkg description format
// "github.com/theplant/package2"
// "github.com/theplant/package2@dev"
// "github.com/theplant/package2@origin/dev"
func parsePakfile() ([]gitpkg.GitPkg, error) {
	pakInfo, err := GetPakInfo()
	if err != nil {
	    return nil, err
	}

	gitPkgs := []gitpkg.GitPkg{}
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

		gitPkgs = append(gitPkgs, NewGitPkg(name, remote, branch))
	}

	return gitPkgs, nil
}