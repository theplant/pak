package hgpkg

import (
	. "github.com/theplant/pak/share"
)

func RegisterProxy() {
	RegisterPkgProxy(PkgProxyBuilder{IsTracking: IsTracking, NewVCS: NewHgPkg})
}

func IsTracking(pkg string) (bool, error) {
	return IsTrackingImpl(pkg, "hg")
}
