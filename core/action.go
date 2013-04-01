package core

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
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
	allPakPkgs, err := ParsePakfile()
	if err != nil {
		return err
	}
	var paklockInfo PaklockInfo
	newPaklockInfo := PaklockInfo{}
	paklockInfo, err = GetPaklockInfo()
	if err != nil {
		if err == PakfileLockNotExist {
			paklockInfo = nil
		} else {
			return err
		}
	} else {
		newPaklockInfo = paklockInfo
	}
	if !option.UsePakfileLock {
		paklockInfo = nil
	}

	// Assign GetOption && Sync && Report Erorrs
	for i := 0; i < len(allPakPkgs); i++ {
		allPakPkgs[i].GetOption.Fetch = option.Fetch
		allPakPkgs[i].GetOption.Force = option.Force

		err = allPakPkgs[i].Sync()
		if err != nil {
			// go get package when the package is not downloaded before
			if !allPakPkgs[i].State.IsPkgExist {
				err = allPakPkgs[i].GoGet()
				if err != nil {
					return err
				}

				err = allPakPkgs[i].Sync()
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		err = allPakPkgs[i].Report()
		if err != nil {
			return err
		}
	}

	// Pick up PakPkg to be updated this time
	pakPkgs := []PakPkg{}
	if len(option.PakMeter) != 0 {
		for _, pakPkgName := range option.PakMeter {
			described := false
			// full-name matching
			for _, pakPkg := range allPakPkgs {
				if pakPkg.Name == pakPkgName {
					pakPkgs = append(pakPkgs, pakPkg)
					described = true
					break
				}
			}
			// partial matching
			if !described {
				matchedResult := []string{}
				for _, pakPkg := range allPakPkgs {
					pakPkgNameReg := regexp.MustCompile(pakPkgName)
					if pakPkgNameReg.MatchString(pakPkg.Name) {
						pakPkgs = append(pakPkgs, pakPkg)
						described = true

						matchedResult = append(matchedResult, pakPkg.Name)
					}
				}
				if len(matchedResult) > 1 {
					nameString := ""
					for _, pkg := range matchedResult {
						nameString = fmt.Sprintf("%s    %s\n", nameString, pkg)
					}
					return fmt.Errorf("More than 1 matched packages:\n%sCan't update with ambiguous package matching.", nameString)
				}
			}

			// non-matching package
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

		// TODO: Add tests for removing Pakfile.lock record
		delete(newPaklockInfo, toRemovePakPkgs[i].Name)
	}

	// TODO: Add Tests for Writing Pakfile.Lock
	return writePaklockInfo(newPaklockInfo)
}
