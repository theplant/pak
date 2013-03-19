package pak

import (
	"os"
	// "fmt"
	"errors"
	"io/ioutil"
	// "github.com/theplant/pak/gitpkg"
	. "github.com/theplant/pak/share"
)

func Init() error {
	_, err := os.Stat(Pakfile)
	if err == nil {
		return errors.New("Pakfile already existed.")
	} else if !os.IsNotExist(err) {
		return err
	}

	return ioutil.WriteFile(Pakfile, []byte(PakfileTemplate), os.FileMode(0644))
}

// type GetOptions struct {
// 	GitPkgs     []gitpkg.GitPkg
// 	FetchLatest bool
// 	Force       bool
// 	UseChecksum bool
// }

// TODO: GitPkgs generation need be verified by Pakfile at first
func Get(options GetOption) (PaklockInfo, error) {
	// paklockInfo := PaklockInfo{}
	//
	// for _, gitPkg := range options.GitPkgs {
	// 	checksum, err := gitPkg.Get(option)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	paklockInfo[gitPkg.Name] = checksum
	// }
	//
	// return paklockInfo, nil
	return nil, nil
}

func Update() {
	// pakfileExist, pakInfo := readPakfile()
	//
	// if len(pakInfo.Packages) == 0 {
	// 	fmt.Println("No packages need to be updated.")
	// 	return
	// }
	//
	// paklockInfo := PaklockInfo{}
	// for _, pkg := range pakInfo.Packages {
	// 	shaVal := getPackageSHA(pkg)
	// 	paklockInfo[pkg] = shaVal[0:len(shaVal)-1]
	// }
	//
	// paklockInfoBts, err := goyaml.Marshal(paklockInfo)
	// if err != nil {
	//     panic(err)
	// }
	//
	// ioutil.WriteFile(paklock, paklockInfoBts, os.FileMode(0644))
}
