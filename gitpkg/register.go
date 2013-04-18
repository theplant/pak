package gitpkg

import (
	"fmt"
	"os"
	. "github.com/theplant/pak/share"
)

func RegisterProxy() {
	RegisterPkgProxy(PkgProxyBuilder{IsTracking: IsTracking, NewVCS: NewGitPkg})
}

func IsTracking(pkg string) (bool, error) {
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