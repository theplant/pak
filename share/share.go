package share

import (
	"fmt"
	"os"
)

const (
	Pakfile   = "Pakfile"
	Paklock   = "Pakfile.lock"
	Pakbranch = "pak"
	Paktag    = "_pak_latest_"
)

var Gopath = os.Getenv("GOPATH")

type PakInfo struct {
	Packages []string
}

type PaklockInfo map[string]string

type GetOption struct {
	Force    bool
	Checksum string
}

type PakOption struct {
	PakMeter       []string // used for containing spcified packages
	UsePakfileLock bool
	Force          bool
}

type VCSPkg interface {
	Sync() error
	Fetch() error
	Pak(GetOption) (string, error)
	Unpak(bool) error
	GoGet() error
	IsPkgExist() (bool, error)
	Report() error
}

type VCSPkgBuilder struct {
	IsTracking func(name string) (bool, error)
	NewVCS func(name, remote, branch string) VCSPkg
}

var VCSPkgList = []VCSPkgBuilder{}
func RegisterVCSPkg(newBuilder VCSPkgBuilder) {
	VCSPkgList = append(VCSPkgList, newBuilder)
}

func NewVCSPkg(pkg, remote, branch string) (VCSPkg, error) {
	for _, engine := range VCSPkgList {
		tracking, err := engine.IsTracking(pkg)
		if err != nil {
		    return nil, err
		}
		if tracking {
			return engine.NewVCS(pkg, remote, branch), nil
		}
	}

	return nil, fmt.Errorf("%s: Can't Detect What Kind of VCS it's Using.", pkg)
}