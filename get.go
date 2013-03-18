package pak

import (
	// "os"
	// "fmt"
	// "errors"
	// "launchpad.net/goyaml"
)

type GetOptions struct {
	GitPkgs     []GitPkg
	FetchLatest bool
	Force       bool
	UseChecksum bool
}

// TODO: GitPkgs generation need be verified by Pakfile at first
func Get(options GetOptions) (PaklockInfo, error) {
	paklockInfo := PaklockInfo{}

	for _, gitPkg := range options.GitPkgs {
		checksum, err := gitPkg.Get(options.FetchLatest, options.Force, options.UseChecksum)
		if err != nil {
		    return nil, err
		}
		paklockInfo[gitPkg.Name] = checksum
	}

	return paklockInfo, nil
}
