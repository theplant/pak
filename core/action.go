package pak

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
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

	// Assign GetOption && Sync && Report Erorrs
	for i := 0; i < len(allPakPkgs); i++ {
		allPakPkgs[i].GetOption.Fetch = option.Fetch
		allPakPkgs[i].GetOption.Force = option.Force

		err = allPakPkgs[i].Sync()
		if err != nil {
			return err
		}
		err = allPakPkgs[i].Report()
		if err != nil {
			return err
		}
	}

	pakPkgs := []PakPkg{}
	newPaklockInfo := PaklockInfo{}
	if len(option.PakMeter) != 0 {
		for _, pakPkgName := range option.PakMeter {
			described := false
			for _, pakPkg := range allPakPkgs {
				if pakPkg.Name == pakPkgName {
					pakPkgs = append(pakPkgs, pakPkg)
					described = true
					break
				// } else {
					// newPaklockInfo :=
				}
			}
			if !described {
				return fmt.Errorf("Package %s Is Not Included in Pakfile.", pakPkgName)
			}
		}
	} else {
		pakPkgs = allPakPkgs
	}

	// categorize pakpkgs
	newPakPkgs, toUpdatePakPkgs, toRemovePakPkgs := ParsePakState(pakPkgs, paklockInfo)

	// Pak dependencies
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
	for i := 0; i < len(toRemovePakPkgs); i++ {
		err = toRemovePakPkgs[i].Unpak(toRemovePakPkgs[i].Force)
		if err != nil {
			return err
		}
	}

	return writePaklockInfo(newPaklockInfo)
}
