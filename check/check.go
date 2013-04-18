package check

import (
	"bufio"
	"fmt"
	. "github.com/theplant/pak/core"
	"github.com/wsxiaoys/terminal/color"
	"os"
	"path/filepath"
	"regexp"
)

// Check should be invoked in the init function.
// It will check your package dependencies, if the dependencies is inconsistent with Pakfile.lock, it force the program to exit. However, Check allows packages to be dirty or not on pak branch as long as the package is marked managed, that is to say, you has run pak get before calling Check.
func Check() {
	// Parse
	allPakPkgs, err := ParsePakfile()
	if err != nil {
		color.Printf("Force Exited by Pak: @r%s@w. Please run @bpak get@w.\n", err.Error())
		os.Exit(1)
	}

	paklockInfo, err := GetPaklockInfo()
	if err != nil {
		color.Printf("Force Exited by Pak: @r%s@w. Please run @bpak get@w.\n", err.Error())
		os.Exit(1)
	}

	errors := [][2]string{}
	warnings := [][2]string{}
	for _, pakPkg := range allPakPkgs {
		err := pakPkg.Sync()
		if err != nil {
			color.Printf("Force Exited by Pak: @r%s@w. Please run @bpak get@w.\n", err.Error())
			os.Exit(1)
		}

		checksum := paklockInfo[pakPkg.Name]
		if checksum == "" {
			errors = append(errors, [2]string{pakPkg.Name, "New Package, Not Managed by Pak."})
			continue
		}

		if !pakPkg.State.OnPakbranch {
			warnings = append(warnings, [2]string{pakPkg.Name, "Not on Pak Branch."})
			continue
		}

		if pakPkg.PakbranchChecksum != checksum {
			errors = append(errors, [2]string{pakPkg.Name, "Inconsistent with Pakfile.lock."})
			continue
		}
	}

	if len(warnings) > 0 {
		paddingLen := 0
		for _, info := range errors {
			if len(info[0]) > paddingLen {
				paddingLen = len(info[0])
			}
		}

		color.Println("Warning: Packages bellow are out of the control of Pak:")
		paddingStr := fmt.Sprintf(`@g%%-%d`, paddingLen)
		for _, warning := range warnings {
			color.Printf(paddingStr+`s @w-> @r%s`, warning[0], warning[1])
			fmt.Println()
		}
	}

	if len(errors) == 0 {
		return
	}

	paddingLen := 0
	for _, info := range errors {
		if len(info[0]) > paddingLen {
			paddingLen = len(info[0])
		}
	}

	color.Println("Force Exited by Pak. Packages listed bellow is not under the control of Pak:")
	paddingStr := fmt.Sprintf(`@g%%-%d`, paddingLen)
	for _, info := range errors {
		color.Printf(paddingStr+`s @w-> @r%s`, info[0], info[1])
		fmt.Println()
	}

	color.Print("Please run @bpak get@w before restarting your program.\n")
	os.Exit(1)
}

var checkTemplate = `package %s
import "github.com/theplant/pak/check"

func init() {
	check.Check()
}`

func ImportPakCheck() (err error) {
	// pkgName, gofileCount, mainGoFile, err := parsePkg()
	// if err != nil {
	//     return err
	// }
	//
	// if gofileCount > 1 {
	// 	// create a file to auto-check package dependences
	//
	// 	err = ioutil.WriteFile("pak.go", []byte(fmt.Sprintf(checkTemplate), pkgName), os.FileMode(0644))
	// } else if gofileCount == 1 && mainGoFile != "" {
	// 	// modify the only go file to auto-check package dependences
	//
	// 	f, err := os.Open(mainGoFile)
	// 	if err != nil {
	// 	    return err
	// 	}
	//
	// 	reader := bufio.NewReader(f)
	// 	lineBytes, _, err := reader.ReadLine()
	// 	if err != nil {
	// 	    return err
	// 	}
	//
	// 	fileContent := ""
	// 	importContent := "import (\n"
	// 	singleLineImportExp := regexp.MustCompile(`^import ("[^"]+")$`)
	// 	multipleLinesImportStartExp := regexp.MustCompile(`^import ?\( ?$`)
	// 	multipleLinesImportEndExp := regexp.MustCompile(`^ ?) ?$`)
	// 	inImportBlock := false
	//
	// 	for reader.Buffered() != 0 {
	// 		line := string(lineBytes)
	//
	// 		switch{
	// 		case singleLineImportExp.MatchString(line):
	// 			pkg := singleLineImportExp.FindStringSubmatch(line)
	// 			if len(pkg) != 2 {
	// 				return fmt.Errorf("Can't Detect Package Name.")
	// 			}
	//
	// 			line = fmt.Sprintf("import (\n    \"github.com/theplant/pak/check\",\n    %s\n)\n", pkg[1])
	// 			line += "\nfunc init() {\ncheck.Check()\n}\n"
	// 		case multipleLinesImportStartExp.MatchString(line):
	// 			inImportBlock = true
	//
	// 		case inImportBlock && multipleLinesImportEndExp.MatchString(line):
	// 			inImportBlock = false
	// 		case inImportBlock:
	//
	// 		default:
	//
	// 		}
	//
	// 		if inImportBlock {
	// 			importContent += line
	// 		} else {
	// 			fileContent += line
	// 		}
	//
	// 		// pkg := packageExp.FindStringSubmatch(string(line))
	//
	// 		if len(pkg) == 2 {
	// 			pkgName = pkg[1]
	// 			return nil
	// 		}
	//
	// 		lineBytes, _, err = reader.ReadLine()
	// 		if err != nil {
	// 		    return err
	// 		}
	// 	}
	// }
	//
	return err
}

var packageExp *regexp.Regexp = regexp.MustCompile(`^package (\w+)$`)

// Parse out the package name, how many go files it contains, and a go file
// name in the current directory.
func parsePkg() (pkgName string, gofileCount int, mainGoFile string, err error) {
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		gofile, err := filepath.Match("?*.go", path)
		if err != nil {
			return err
		}

		if !gofile {
			return nil
		}

		gofileCount++
		mainGoFile = path

		f, err := os.Open(path)
		if err != nil {
			return err
		}

		reader := bufio.NewReader(f)
		line, _, err := reader.ReadLine()
		if err != nil {
			return err
		}

		for reader.Buffered() != 0 {
			pkg := packageExp.FindStringSubmatch(string(line))

			if len(pkg) == 2 {
				pkgName = pkg[1]
				return nil
			}

			line, _, err = reader.ReadLine()
			if err != nil {
				return err
			}
		}

		return nil
	})

	return
}
