package gitpkg

import (
	"fmt"
	. "github.com/theplant/pak/share"
	"os"
	"strings"
)

func RegisterProxy() {
	RegisterPkgProxy(PkgProxyBuilder{IsTracking: IsTracking, NewVCS: NewGitPkg})
}

func IsTracking(pkg string) (bool, error) {
	for true {
		tracking, err := isDirTracked(pkg)
		if err != nil {
			return false, err
		}

		if tracking {
			return true, nil
		} else {
			slashIndex := strings.LastIndex(pkg, "/")
			if slashIndex == -1 {
				return false, nil
			}
			pkg = pkg[0:slashIndex]
		}
	}

	return false, nil
}

// isDirTracked will return true if the the pkg is Git Tracked
func isDirTracked(pkg string) (bool, error) {
	_, err := os.Stat(fmt.Sprintf("%s/src/%s/.git", Gopath, pkg))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}
