package core

import (
	"fmt"
	// "github.com/theplant/pak/gitpkg"
	. "github.com/theplant/pak/share"
	"io/ioutil"
	"launchpad.net/goyaml"
	"os"
	"path/filepath"
)

func GetPakInfo() (pakInfo PakInfo, err error) {
	var pakInfoBytes []byte
	pakInfoBytes, err = pakRead(Pakfile)
	if err != nil {
		return
	}

	err = goyaml.Unmarshal(pakInfoBytes, &pakInfo)

	for i, _ := range pakInfo.Packages {
		if pakInfo.Packages[i].PakName == "" {
			pakInfo.Packages[i].PakName = "pak"
		}
	}

	return
}

func GetPaklockInfo() (paklockInfo PaklockInfo, err error) {
	var content []byte
	content, err = pakRead(Paklock)
	if err != nil {
		return
	}

	err = goyaml.Unmarshal(content, &paklockInfo)

	return
}

var PakfileNotExist = fmt.Errorf("Can't find %s", Pakfile)
var PakfileLockNotExist = fmt.Errorf("Can't find %s", Paklock)

func pakRead(file string) (fileContent []byte, err error) {
	absPakfilePath := ""
	originalFile := file
	for true {
		absPakfilePath, err = filepath.Abs(file)
		if err != nil {
			return nil, err
		}
		if absPakfilePath == Gopath+"/"+originalFile {
			if originalFile == Pakfile {
				return nil, PakfileNotExist
			} else {
				return nil, PakfileLockNotExist
			}
		}

		_, err = os.Stat(file)
		if os.IsNotExist(err) {
			file = "../" + file
			continue
		}

		break
	}

	return ioutil.ReadFile(file)
}

func writePaklockInfo(paklockInfo PaklockInfo) error {
	content, err := goyaml.Marshal(&paklockInfo)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(Paklock, content, os.FileMode(0644))
}
