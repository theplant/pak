package pak

import (
	// "fmt"
	"io/ioutil"
	"launchpad.net/goyaml"
	"os"
	"os/exec"
	"bytes"
	"path/filepath"
	"errors"
)

type PakInfo struct {
	Packages []string
}
type PaklockInfo map[string]string

const (
	pakfile   = "Pakfile"
	paklock   = "Pakfile.lock"
	pakbranch = "pak"
	paktag 	  = "_pak_latest_"
)

var gopath = os.Getenv("GOPATH")

func SamePakInfo(pakInfo1, pakInfo2 PakInfo) (result bool) {
	result = len(pakInfo1.Packages) == len(pakInfo2.Packages)
	if !result {
		return
	}

	for i, pak := range pakInfo1.Packages {
		result = pak == pakInfo2.Packages[i]
		if !result {
			return
		}
	}

	return
}

func RunCmd(cmd *exec.Cmd) (out bytes.Buffer, err error) {
	cmd.Stdout = &out
	err = cmd.Run()

	return
}

func readPakfile() (pakInfo PakInfo, err error) {
	var pakInfoBytes []byte
	pakInfoBytes, err = pakRead(pakfile)
	if err != nil {
		return
	}

	err = goyaml.Unmarshal(pakInfoBytes, &pakInfo)

	return
}

func readPaklockInfo() (paklockInfo PaklockInfo, err error) {
	var content []byte
	content, err = pakRead(paklock)
	if err != nil {
		return
	}

	err = goyaml.Unmarshal(content, &paklockInfo)

	return
}

func pakRead(filePath string) (fileContent []byte, err error) {
	absPakfilePath := ""
	for true {
		absPakfilePath, err = filepath.Abs(filePath)
		if err != nil {
			return
		}
		if absPakfilePath == gopath + "/Pakfile" {
			err = errors.New("Can't find Pakfile.")

			return
		}

		_, err = os.Stat(filePath)
		if os.IsNotExist(err) {
			filePath = "../" + filePath
			continue
		}

		break
	}

	return ioutil.ReadFile(filePath)
}