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
	// Retrieve PakPkgs from Pakfile
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

	// For: pak update package
	// Pick up PakPkg to be updated this time
	pakPkgs := []PakPkg{}
	if len(option.PakMeter) != 0 {
		var matched bool
		var pakPkg PakPkg
		for _, pakPkgName := range option.PakMeter {
			matched, pakPkg, err = isPkgMatched(allPakPkgs, pakPkgName)
			if err != nil {
			    return err
			}

			if matched {
				pakPkgs = append(pakPkgs, pakPkg)
			} else {
				return fmt.Errorf("Package %s Is Not Included in Pakfile.", pakPkgName)
			}
		}
	} else {
		pakPkgs = allPakPkgs
	}

	// Assign GetOption && Sync && Report Erorrs
	err = loadPkgs(&pakPkgs, option)
	if err != nil {
	    return err
	}

	err = pakDependencies(pakPkgs, paklockInfo, &newPaklockInfo)
	if err != nil {
	    return err
	}

	return writePaklockInfo(newPaklockInfo)
}

func loadPkgs(allPakPkgs *[]PakPkg, option PakOption) (err error) {
	for i := 0; i < len((*allPakPkgs)); i++ {
		// TODO: remove Fetch option
		(*allPakPkgs)[i].GetOption.Fetch = option.Fetch
		(*allPakPkgs)[i].GetOption.Force = option.Force

		// Go Get Package when the Package is not Downloaded Before
		isPkgExist, err := (*allPakPkgs)[i].IsPkgExist()
		if err != nil {
		    return err
		}
		if !isPkgExist {
			err = (*allPakPkgs)[i].GoGet()
			if err != nil {
				return err
			}
		}

		// Fetch Before Hand can Make Sure That the Package Contains Up-To-Date Remote Branch
		err = (*allPakPkgs)[i].GitPkg.Fetch()
		if err != nil {
			return err
		}

		(*allPakPkgs)[i].GitPkg.Sync()

		err = (*allPakPkgs)[i].Report()
		if err != nil {
			return err
		}
	}

	return
}

func isPkgMatched(allPakPkgs []PakPkg, pakPkgName string) (bool, PakPkg, error) {
	matched := false
	matchedPakPkg := PakPkg{}
	// full-name matching
	for _, pakPkg := range allPakPkgs {
		if pakPkg.Name == pakPkgName {
			// pakPkgs = append(pakPkgs, pakPkg)
			matchedPakPkg = pakPkg
			matched = true
			break
		}
	}

	// partial matching
	if !matched {
		matchedResult := []string{}
		for _, pakPkg := range allPakPkgs {
			pakPkgNameReg := regexp.MustCompile(pakPkgName) // TODO: should not use MustCompile here.
			if pakPkgNameReg.MatchString(pakPkg.Name) {
				// pakPkgs = append(pakPkgs, pakPkg)
				matchedPakPkg = pakPkg
				matched = true

				matchedResult = append(matchedResult, pakPkg.Name)
			}
		}

		if len(matchedResult) > 1 {
			nameString := ""
			for _, pkg := range matchedResult {
				nameString = fmt.Sprintf("%s    %s\n", nameString, pkg)
			}

			err := fmt.Errorf("More than 1 matched packages:\n%sStop updating for ambiguous package matching.", nameString)
			return false, matchedPakPkg, err
		}
	}

	return matched, matchedPakPkg, nil
}

func pakDependencies(pakPkgs []PakPkg, paklockInfo PaklockInfo, newPaklockInfo *PaklockInfo) error {
	newPakPkgs, toUpdatePakPkgs, toRemovePakPkgs := ParsePakState(pakPkgs, paklockInfo)

	var (
		 checksum string
		 err error
	)

	for i := 0; i < len(newPakPkgs); i++ {
		checksum, err = newPakPkgs[i].Pak(newPakPkgs[i].GetOption)
		if err != nil {
			return err
		}

		(*newPaklockInfo)[newPakPkgs[i].Name] = checksum
	}
	for i := 0; i < len(toUpdatePakPkgs); i++ {
		checksum, err = toUpdatePakPkgs[i].Pak(toUpdatePakPkgs[i].GetOption)
		if err != nil {
			return err
		}

		(*newPaklockInfo)[toUpdatePakPkgs[i].Name] = checksum
	}
	for i := 0; i < len(toRemovePakPkgs); i++ {
		err = toRemovePakPkgs[i].Unpak(toRemovePakPkgs[i].Force)
		if err != nil {
			return err
		}

		// TODO: Add tests for removing Pakfile.lock record
		delete((*newPaklockInfo), toRemovePakPkgs[i].Name)
	}

	return err
}