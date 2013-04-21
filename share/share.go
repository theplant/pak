package share

import (
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

type PakOption struct {
	PakMeter       []string // used for containing spcified packages
	UsePakfileLock bool

	// Force option is designed for asking pak to update package state as long
	// as the package is exist and clean. This option is always true when pak is
	// updating packages because [update] command is designed for checking out
	// the up-to-date commit of a package instead of resetting the package state
	// in accord with Pakfile.lock. In another case, [get] command, Force option
	// will allow pak to forcefully remove branch named [pak] in packages which
	// don't contain a tag named [_pak_latest_].
	Force bool

	// When the option set, pak will stop complaining about unclean
	// packages, either installing or updating packages. That make it less
	// painful for someone who is developing both the main project using pak
	// and its dependent packages tracked by pak, because he no longer has to
	// be clean up dependent packages that he is developing when all he want
	// to is get other dependent packages of the main project updated.
	// But for package that hasn't been locked down by pak, it should be clean.
	SkipUncleanPkgs bool
}

type GetOption struct {
	Force           bool
	Checksum        string
	SkipUncleanPkgs bool
}

type PkgProxy interface {
	Fetch() error
	NewBranch(string) error
	NewTag(string, string) error
	RemoveTag(string) error
	Pak(string) (string, error)
	Unpak() error

	IsClean() (bool, error)
	ContainsRemoteBranch() (bool, error)
	ContainsPakbranch() (bool, error)
	ContainsPaktag() (bool, error)
	GetChecksum(string) (string, error)
	GetHeadChecksum() (string, error)
	GetHeadRefName() (string, error)
	GetPaktagRef() string
	GetPakbranchRef() string
	GetRemoteBranch() string
}

type PkgProxyBuilder struct {
	IsTracking func(name string) (bool, error)
	NewVCS     func(name, remote, branch string) PkgProxy
}

var PkgProxyList = []PkgProxyBuilder{}

func RegisterPkgProxy(newBuilder PkgProxyBuilder) {
	PkgProxyList = append(PkgProxyList, newBuilder)
}
