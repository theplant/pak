package gitpkg

import (
	. "github.com/theplant/pak/share"
)

func RegisterProxy() {
	RegisterPkgProxy(PkgProxyBuilder{IsTracking: IsTracking, NewVCS: NewGitPkg})
}

func IsTracking(pkg string) (bool, error) {
	return IsTrackingImpl(pkg, "git")
}
