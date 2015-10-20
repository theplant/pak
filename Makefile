.DEFAULT_GOAL := test

test:
	go test -i github.com/theplant/pak/gitpkg
	go test github.com/theplant/pak/gitpkg
	go test -i github.com/theplant/pak/hgpkg
	go test github.com/theplant/pak/hgpkg
	go test -i github.com/theplant/pak/core
	go test github.com/theplant/pak/core

install:
	go get github.com/theplant/pak
	go get github.com/wsxiaoys/terminal/color
	go get launchpad.net/gocheck
	go get gopkg.in/go-yaml/v2/yaml
