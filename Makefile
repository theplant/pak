all: test

test:
	go test github.com/theplant/pak/gitpkg
	go test github.com/theplant/pak/core