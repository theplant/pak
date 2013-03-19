package pak

import (
	"fmt"
	"io/ioutil"
	"launchpad.net/goyaml"
	"os"
	// "os/exec"
	// "bytes"
	"path/filepath"
	// "errors"
)

type PakInfo struct {
	Packages []string
}
type PaklockInfo map[string]string

const (
	pakfile   = "Pakfile"
	paklock   = "Pakfile.lock"
	pakbranch = "pak"
	paktag    = "_pak_latest_"
)

var gopath = os.Getenv("GOPATH")

func GetPakInfo() (pakInfo PakInfo, err error) {
	var pakInfoBytes []byte
	pakInfoBytes, err = pakRead(pakfile)
	if err != nil {
		return
	}

	err = goyaml.Unmarshal(pakInfoBytes, &pakInfo)

	return
}

func GetPaklockInfo() (paklockInfo PaklockInfo, err error) {
	var content []byte
	content, err = pakRead(paklock)
	if err != nil {
		return
	}

	err = goyaml.Unmarshal(content, &paklockInfo)

	return
}

// var ErrorNotFound = errors.New("Can't find.")
func pakRead(file string) (fileContent []byte, err error) {
	absPakfilePath := ""
	originalFile := file
	for true {
		absPakfilePath, err = filepath.Abs(file)
		if err != nil {
			return
		}
		if absPakfilePath == gopath+"/Pakfile" {
			return nil, fmt.Errorf("Can't find %s" + originalFile)
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

	return ioutil.WriteFile(paklock, content, os.FileMode(0644))
}
