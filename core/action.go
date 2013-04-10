package core

import (
	"errors"
	"fmt"
	. "github.com/theplant/pak/share"
	"io/ioutil"
	"os"
	"regexp"
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
		// TODO: remove Fetch option
		allPakPkgs[i].GetOption.Fetch = option.Fetch
		allPakPkgs[i].GetOption.Force = option.Force
		allPakPkgs[i].GetOption.NotGet = option.NotGet // added temporally. TODO: refactor

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

		// TODO: add tests
		// this enable pak to check pak even the local package repo doesn't contain
		// the remote branch.
		// for:
		// 新bug：执行 pak get 时，Pak 没有先执行 fetch 就直接 checkout pak HASH，导致出现无法checkout的错误。重现方法：
		// * 把package的远程分支删除 git -r -d origin/master
		// * 回到项目执行 pak get
		// * 出现错误：`github.com/xxx` does not contain reference `refs/remotes/origin/master`
		//
		err = allPakPkgs[i].GitPkg.Fetch()
		if err != nil {
			return err
		}

		allPakPkgs[i].GitPkg.Sync()

		err = allPakPkgs[i].Report()
		if err != nil {
			return err
		}
	}

	// For: pak update package
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

	// Categorize Pakpkgs
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
