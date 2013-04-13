package gitpkg

import (
	"fmt"
	"os"
	. "github.com/theplant/pak/share"
)

func init() {
	RegisterVCSPkg(VCSPkgBuilder{IsTracking: IsTracking, NewVCS: NewGitPkg})
}

func IsTracking(pkg string) (bool, error) {
	_, err := os.Stat(fmt.Sprintf("%s/%s/.git", Gopath, pkg))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}