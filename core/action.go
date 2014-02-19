package core

import (
	"bytes"
	"errors"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"text/template"
	"time"
	. "github.com/theplant/pak/share"
	"github.com/wsxiaoys/terminal/color"
)

func Init() (err error) {
	_, err = os.Stat(Pakfile)
	if err == nil {
		err = errors.New("Pakfile already existed.")
		return
	} else if !os.IsNotExist(err) {
		return
	}

	absProjRoot, err := filepath.Abs(".")
	if err != nil {
		return
	}
	pkg, err := build.ImportDir(".", 0)
	if err != nil {
		return
	}
	importsMap := map[string]bool{}
	projRoot := strings.Replace(absProjRoot, Gopath+"/src/", "", 1)
	for _, lib := range pkg.Imports {
		if isStdLib(lib) || strings.Contains(lib, projRoot) {
			continue
		}
		importsMap[lib] = true
	}

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() || strings.HasPrefix(path, ".") {
			return err
		}

		// ignoring errors here for sometimes the sub packages may not be able be compiled.
		pkg, _ := build.ImportDir(path, 0)
		if len(pkg.Imports) == 0 {
			return err
		}

		for _, lib := range pkg.Imports {
			if isStdLib(lib) || strings.Contains(lib, projRoot) {
				continue
			}
			importsMap[lib] = true
		}

		return err
	})
	if err != nil {
		return
	}

	imports := sort.StringSlice{}
	for lib, _ := range importsMap {
		imports = append(imports, lib)
	}
	imports.Sort()

	tmpl := template.Must(template.New("").Parse(PakfileTemplate))
	wr := bytes.NewBuffer([]byte{})
	if err = tmpl.Execute(wr, imports); err != nil {
		return
	}

	return ioutil.WriteFile(Pakfile, wr.Bytes(), os.FileMode(0644))
}

func Get(option PakOption) (err error) {
	var start time.Time
	if option.Verbose {
		start = time.Now()
	}

	pakInfo, paklockInfo, err := GetPakInfo(GpiParams{
		Type:                 "11",
		Path:                 "",
		DeepParse:            true,
		WithBasicDependences: true,
		Verbose:              option.Verbose,
	})
	if err != nil {
		return err
	}

	if len(paklockInfo) == 0 {
		if option.SkipUncleanPkgs {
			return fmt.Errorf("Can't skip unclean packages without Pakfile.lock. Please run pak get first.")
		}
	}

	var allPakPkgs []PakPkg
	for _, pkgCfg := range pakInfo.Packages {
		pakPkg := NewPakPkg(pkgCfg)

		allPakPkgs = append(allPakPkgs, pakPkg)
	}

	pakPkgs, err := getPackages(allPakPkgs, option, paklockInfo)
	if err != nil {
		return
	}

	paklockInfoBackup := PaklockInfo{}
	for k, v := range paklockInfo {
		paklockInfoBackup[k] = v
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

	// For Pak Meter
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

	if reflect.DeepEqual(paklockInfoBackup, newPaklockInfo) {
		if option.Verbose && !option.UsePakfileLock {
			color.Println("Nothing changes in Pakfile.lock.")
		}
	} else {
		if option.Verbose {
			color.Printf("Writing Pakfile.lock.\n")
		}

		lockInfo := PaklockInfo{}
		for _, pkgName := range pakInfo.BasicDependences {
			if newPaklockInfo[pkgName] != "" {
				lockInfo[pkgName] = newPaklockInfo[pkgName]
			}
		}

		err = writePaklockInfo(lockInfo, option)
		if err != nil {
			return err
		}
	}

	var end time.Time
	if option.Verbose {
		end = time.Now()
		color.Printf("Pak Done (Took: %.2fs).\n", end.Sub(start).Seconds())
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
				// err = fmt.Errorf("%+v", newError)
				err = newError
			} else {
				err = fmt.Errorf("%+v\n%+v", err, newError)
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

// For: pak [update|get] <package1 package2 ...>
// Pick up PakPkgs to be updated this time
func getPackages(allPakPkgs []PakPkg, option PakOption, paklockInfo PaklockInfo) (pakPkgs []PakPkg, err error) {
	if len(option.PakMeter) == 0 {
		pakPkgs = allPakPkgs
		return
	}

	if paklockInfo == nil {
		err = fmt.Errorf("Package not yet locked. Please run pak get before getting or updating specific packages.")
		return
	}

	pkgMap := map[string]PakPkg{}
	for _, pakPkgName := range option.PakMeter {
		for _, pkg := range allPakPkgs {
			if strings.Contains(pkg.Name, pakPkgName) {
				pkgMap[pkg.Name] = pkg
			}
		}
	}
	for _, pkg := range pkgMap {
		pakPkgs = append(pakPkgs, pkg)
	}

	if len(pakPkgs) == 0 {
		err = fmt.Errorf("Not packages matched.")
		return
	}
	if option.Verbose {
		info := "Matched Packages:"
		if len(pakPkgs) == 1 {
			info = "Matched Package:"
		}
		spaces := strings.Repeat(" ", len(info)+1)
		for i, pkg := range pakPkgs {
			if i == 0 {
				fmt.Println(info, pkg.Name)
			} else {
				fmt.Printf("%s%s\n", spaces, pkg.Name)
			}
		}
	}

	return
}

func FindPackage(name string) (pakPkg PakPkg, err error) {
	pakInfo, _, err := GetPakInfo(GpiParams{
		Type:                 "10",
		Path:                 "",
		DeepParse:            false,
		WithBasicDependences: false,
	})
	if err != nil {
		return PakPkg{}, err
	}

	var allPakPkgs []PakPkg
	for _, pkgCfg := range pakInfo.Packages {
		pakPkg := NewPakPkg(pkgCfg)

		allPakPkgs = append(allPakPkgs, pakPkg)
	}

	for _, pkg := range allPakPkgs {
		if strings.Contains(pkg.Name, name) {
			pakPkg = pkg
			break
		}
	}

	return
}
