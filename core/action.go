package pak

import (
	"os"
	"fmt"
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
	allPakPkgs, err := parsePakfile()
	if err != nil {
	    return err
	}
	var paklockInfo PaklockInfo
	if option.UsePakfileLock {
		var err error
		paklockInfo, err = GetPaklockInfo()
		if err != nil && err != PakfileLockNotExist {
		    return err
		}
	}

	pakPkgs := []PakPkg{}
	if len(option.PakMeter) != 0 {
		for _, pakPkgName := range option.PakMeter {
			described := false
			for _, pakPkg := range allPakPkgs {
				if pakPkg.Name == pakPkgName {
					pakPkgs = append(pakPkgs, pakPkg)
					described = true
					break
				}
			}
			if !described {
				return fmt.Errorf("Package %s DO NOT Be Included in Pakfile.", pakPkgName)
			}
		}
	} else {
		pakPkgs = allPakPkgs
	}

	newPakPkgs, toUpdatePakPkgs, toRemovePakPkgs := ParsePakState(pakPkgs, paklockInfo)

	// Assign GetOption && Sync && Report Erorrs
	for i := 0; i < len(pakPkgs); i++ {
		pakPkgs[i].GetOption.Fetch = option.Fetch
		pakPkgs[i].GetOption.Force = option.Force

		err = pakPkgs[i].Sync()
		if err != nil {
		    return err
		}
		err = pakPkgs[i].Report()
		if err != nil {
		    return err
		}
	}

	// Pak
	newPaklockInfo := PaklockInfo{}
	var checksum string
	for i := 0; i < len(newPakPkgs); i++ {
		checksum, err = newPakPkgs[i].Pak(newPakPkgs[i].GetOption)
		if err != nil {
		    return err
		}

		newPaklockInfo[newPakPkgs[i].Name] = checksum
	}
	for i := 0; i < len(toUpdatePakPkgs); i++ {
		checksum, err = toUpdatePakPkgs[i].Pak(toUpdatePakPkgs[i].GetOption)
		if err != nil {
		    return err
		}

		newPaklockInfo[toUpdatePakPkgs[i].Name] = checksum
	}
	for i:=0; i<len(toRemovePakPkgs); i++  {
		err = toRemovePakPkgs[i].Unpak(toRemovePakPkgs[i].Force)
		if err != nil {
		    return err
		}
	}

	return writePaklockInfo(newPaklockInfo)
}