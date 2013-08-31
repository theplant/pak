package core

import (
	"errors"
	"fmt"
	. "github.com/theplant/pak/share"
	"github.com/wsxiaoys/terminal/color"
	"io/ioutil"
	"os"
	"regexp"
	"time"
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
	var start time.Time
	if option.Verbose {
		color.Printf("Start paking.\n")
		start = time.Now()
	}

	pakInfo, paklockInfo, err := GetPakInfo(GpiParams{
		Type:                 "11",
		Path:                 "",
		DeepParse:            true,
		WithBasicDependences: true,
	})
	if err != nil {
		return err
	}

	if len(paklockInfo) == 0 {
		if option.SkipUncleanPkgs {
			return fmt.Errorf("Can't skip unclean packages because this project has not yet been locked.\nPlease run pak get.")
		}
	}

	var allPakPkgs []PakPkg
	for _, pkgCfg := range pakInfo.Packages {
		pakPkg := NewPakPkg(pkgCfg)

		allPakPkgs = append(allPakPkgs, pakPkg)
	}

	// For: pak [update|get] <package, ...>
	// Pick up PakPkgs to be updated this time
	pakPkgs := []PakPkg{}
	if len(option.PakMeter) != 0 {
		if paklockInfo == nil {
			return fmt.Errorf("Can't pak specific packages because this project has not yet been locked.\nPlease run pak get before getting or updating specific packages.")
		}

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

	if option.Verbose {
		color.Printf("Paking Packages.\n")
	}

	CategorizePakPkgs(&pakPkgs, paklockInfo, option)

	errChan := make(chan error)
	doneChan := make(chan bool)
	checksumChan := make(chan [2]string)

	for i, _ := range pakPkgs {
		pakPkgs[i].Force = option.Force
		pakPkgs[i].Verbose = option.Verbose
		pakPkgs[i].UsingPakMeter = len(option.PakMeter) > 0
		pakPkgs[i].SkipUncleanPkgs = option.SkipUncleanPkgs

		pkg := pakPkgs[i]

		go func() {
			pkgNameAndChecksum, err := pkg.Get()
			if err != nil {
				errChan <- err

				return
			}

			checksumChan <- pkgNameAndChecksum
			doneChan <- true
		}()
	}

	newPaklockInfo := PaklockInfo{}
	if paklockInfo != nil {
		for k, v := range paklockInfo {
			newPaklockInfo[k] = v
		}
	}
	err = waitForPkgsProcessing(&newPaklockInfo, len(pakPkgs), checksumChan, doneChan, errChan)
	if err != nil {
		return err
	}

	if option.Verbose {
		color.Printf("Writing Pakfile.lock.\n")
	}

	lockInfo := PaklockInfo{}
	for _, pkgName := range pakInfo.BasicDependences {
		if newPaklockInfo[pkgName] != "" {
			lockInfo[pkgName] = newPaklockInfo[pkgName]
		}
	}

	err = writePaklockInfo(lockInfo)
	if err != nil {
		return err
	}

	var end time.Time
	if option.Verbose {
		end = time.Now()
		color.Printf("Pak Done.\nTook: %ds.\n", int(end.Sub(start).Seconds()))
	}

	return nil
}

func waitForPkgsProcessing(newPaklockInfo *PaklockInfo, pkgLen int, checksumChan chan [2]string, doneChan chan bool, errChan chan error) (err error) {
	count := 0
	for {
		select {
		case nameAndChecksum := <-checksumChan:
			if nameAndChecksum[1] == "" {
				delete(*newPaklockInfo, nameAndChecksum[0])

				break
			}

			(*newPaklockInfo)[nameAndChecksum[0]] = nameAndChecksum[1]
		case newError := <-errChan:
			// TODO: make a public error chanle to inform other working gorutine to stop running before panic this error
			if err == nil {
				err = fmt.Errorf("Error: %+v", newError)
			} else {
				err = fmt.Errorf("%+v\nError: %+v", err, newError)
			}
			count += 1
			if count == pkgLen {
				return
			}
		case <-doneChan:
			count += 1
			if count == pkgLen {
				return
			}
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
			pakPkgNameReg, err := regexp.Compile(pakPkgName)
			if err != nil {
				return false, PakPkg{}, err
			}

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

func FindPackage(name string) (bool, PakPkg, error) {
	pakInfo, _, err := GetPakInfo(GpiParams{
		Type:                 "10",
		Path:                 "",
		DeepParse:            false,
		WithBasicDependences: false,
	})
	if err != nil {
		return false, PakPkg{}, err
	}

	var allPakPkgs []PakPkg
	for _, pkgCfg := range pakInfo.Packages {
		pakPkg := NewPakPkg(pkgCfg)

		allPakPkgs = append(allPakPkgs, pakPkg)
	}

	return isPkgMatched(allPakPkgs, name)
}
