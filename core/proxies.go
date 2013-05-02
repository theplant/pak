package core

import (
	"github.com/theplant/pak/gitpkg"
	"github.com/theplant/pak/hgpkg"
)

func init() {
	gitpkg.RegisterProxy()
	hgpkg.RegisterProxy()
}
