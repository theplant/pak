.DEFAULT_GOAL := test

test:
	# go test -i github.com/theplant/pak/gitpkg
	# go test github.com/theplant/pak/gitpkg
	go test -i github.com/theplant/pak/hgpkg
	go test github.com/theplant/pak/hgpkg
	go test -i github.com/theplant/pak/core
	go test github.com/theplant/pak/core
