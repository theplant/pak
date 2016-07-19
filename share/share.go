package share

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	Pakfile   = "Pakfile"
	Paklock   = "Pakfile.lock"
	Pakbranch = "pak"
	Paktag    = "_pak_latest_"
)

var Gopath = os.Getenv("GOPATH")

func init() {
	if Gopath[len(Gopath)-1] == '/' {
		Gopath = Gopath[0 : len(Gopath)-1]
	}
}

type PaklockInfo map[string]string

// For core function #Get
type PakOption struct {
	PakMeter       []string // used for containing spcified packages
	UsePakfileLock bool
	Verbose        bool

	// Force option is designed for asking pak to update package state as long
	// as the package is exist and clean. This option is always true when pak is
	// updating packages because [update] command is designed for checking out
	// the up-to-date commit of a package instead of resetting the package state
	// in accord with Pakfile.lock. In another case, [get] command, Force option
	// will allow pak to forcefully remove branch named [pak] in packages which
	// don't contain a tag named [_pak_latest_].
	Force bool

	// When set, pak will stop complaining about unclean
	// packages, either installing or updating packages. That make it less
	// painful for someone who is developing both the main project using pak
	// and its dependent packages tracked by pak, because the developer no longer
	// has to be clean up dependent packages that he is developing when all he
	// want to is get other dependent packages of the main project updated.
	// But for package that hasn't been locked down by pak, it should be clean.
	SkipUncleanPkgs bool

	Concurrent bool
}

type PkgProxy interface {
	Fetch() error
	NewBranch(string) error
	// NewTag(string, string) error
	// RemoveTag(string) error
	Pak(string) (string, error)
	Unpak() error

	IsClean() (bool, error)
	IsChecksumExist(string) (bool, error)
	ContainsRemoteBranch() (bool, error)
	ContainsPakbranch() (bool, error)
	GetChecksum(string) (string, error)
	GetHeadChecksum() (string, error)
	GetHeadRefName() (string, error)
	GetPakbranchRef() string
	GetRemoteBranch() string
}

type PkgProxyBuilder struct {
	IsTracking func(name string) (bool, error)
	NewVCS     func(name, remote, branch, pakBranch string) PkgProxy
}

var PkgProxyList = []PkgProxyBuilder{}

func RegisterPkgProxy(newBuilder PkgProxyBuilder) {
	PkgProxyList = append(PkgProxyList, newBuilder)
}

func MustRun(params ...string) (cmd *exec.Cmd) {
	cmd = exec.Command(params[0], params[1:]...)

	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(stderr.String())

		panic(err)
	}

	return
}

/**
 * Package Root Detector
 */
func IsTrackingImpl(pkg, detector string) (bool, error) {
	for true {
		tracking, err := isDirTracked(pkg, detector)
		if err != nil {
			return false, err
		}

		if tracking {
			return true, nil
		} else {
			slashIndex := strings.LastIndex(pkg, "/")
			if slashIndex == -1 {
				return false, nil
			}
			pkg = pkg[0:slashIndex]
		}
	}

	return false, nil
}

// isDirTracked will return true if the the pkg is Git Tracked
func isDirTracked(pkg, detector string) (bool, error) {
	_, err := os.Stat(fmt.Sprintf("%s/src/%s/.%s", Gopath, pkg, detector))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

func GetPkgRootImpl(pkg, detector string) (string, error) {
	originalPkg := pkg
	for true {
		tracking, err := isDirTracked(pkg, detector)
		if err != nil {
			return "", err
		}

		if tracking {
			return pkg, nil
		} else {
			slashIndex := strings.LastIndex(pkg, "/")
			if slashIndex == -1 {
				return "", fmt.Errorf("Pakcage %s: Can't Resolve Package Root.", originalPkg)
			}
			pkg = pkg[0:slashIndex]
		}
	}

	return "", nil
}
