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

func Get(option PakOption) error {
	// Parse
	pakPkgs, err := parsePakfile()
	if err != nil {
	    return err
	}
	paklockInfo, err := GetPaklockInfo()
	if err != nil {
	    return err
	}

	newPakPkgs, toUpdatePakPkgs, toRemovePakPkgs := ParsePakState(pakPkgs, paklockInfo)

	// Assign GetOption && Sync && Report Erorrs
	for _, pakPkg := range pakPkgs {
		pakPkg.GetOption.Fetch = option.Fetch
		pakPkg.GetOption.Force = option.Force

		err = pakPkg.Sync()
		if err != nil {
		    return err
		}
		err = pakPkg.Report()
		if err != nil {
		    return err
		}
	}

	// Pak
	newPaklockInfo := PaklockInfo{}
	var checksum string
	for _, pakPkg := range newPakPkgs {
		checksum, err = pakPkg.Pak(pakPkg.GetOption)
		if err != nil {
		    return err
		}

		newPaklockInfo[pakPkg.Name] = checksum
	}
	for _, pakPkg := range toUpdatePakPkgs {
		checksum, err = pakPkg.Pak(pakPkg.GetOption)
		if err != nil {
		    return err
		}

		newPaklockInfo[pakPkg.Name] = checksum
	}
	for _, pakPkg := range toRemovePakPkgs {
		err = pakPkg.Unpak(pakPkg.Force)
		if err != nil {
		    return err
		}
	}

	return writePaklockInfo(newPaklockInfo)
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
