package pak

import (
	"fmt"
	"io/ioutil"
	"launchpad.net/goyaml"
	"os"
	// "os/exec"
	// "bytes"
	"path/filepath"
)

type PakInfo struct {
	Packages []string
}
type PaklockInfo map[string]string

const (
	pakfile   = "Pakfile"
	paklock   = "Pakfile.lock"
	pakbranch = "pak"
	paktag 	  = "_pak_lastest_"
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

func MustRun(cmd *exec.Cmd) (out bytes.Buffer) {
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	return
}

func readPakfile() (exist bool, pakInfo PakInfo) {
	pakfilePath := "./Pakfile"
	for true {
		absPakfilePath, err := filepath.Abs(pakfilePath)
		if err != nil {
			panic(err)
		}
		if absPakfilePath == gopath + "/Pakfile" {
			fmt.Println("Can't find Pakfile.")

			exist = false
			return
		}

		_, err = os.Stat(pakfilePath)
		if os.IsNotExist(err) {
			pakfilePath = "../" + pakfilePath
			continue
		}

		break
	}

	exist = true
	pakInfoBytes, err := ioutil.ReadFile(pakfilePath)
	if err != nil {
		panic(err)
	}
	goyaml.Unmarshal(pakInfoBytes, &pakInfo)

	return
}