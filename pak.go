package pak

import (
	"fmt"
	"io/ioutil"
	"launchpad.net/goyaml"
	"os"
	"path/filepath"
)

type PakInfo struct {
	Packages []string
}

const (
	pakfile   = "Pakfile"
	pakBranch = "pak"
)

func SamePakInfo(pakInfo1, pakInfo2 PakInfo) (result bool) {
	result = len(pakInfo1.Packages) == len(pakInfo2.Packages)
	if !result {
		return
	}

	for i, pak := range pakInfo1.Packages {
		result = pak == pakInfo2.Packages[i]
		if !result {
			break
		}
	}

	return
}

func Init() {
	_, err := os.Stat(pakfile)
	if err == nil {
		fmt.Println("Pakfile already existed.")
		return
	} else if !os.IsNotExist(err) {
		panic(err)
	}

	err = ioutil.WriteFile(pakfile, []byte(PakfileTemplate), os.FileMode(0644))
	if err != nil {
		panic(err)
	}

	return
}

type GetOptions struct {
	All    bool
	Repo   string
	Branch string
}

func Get(option GetOptions) {
	if option.All {

	} else {

	}
}

func readPakfile() (pakInfo PakInfo) {
	pakfilePath := "./Pakfile"
	for true {
		absPakfilePath, err := filepath.Abs(pakfilePath)
		if err != nil {
			panic(err)
		}
		if absPakfilePath == os.Getenv("GOPATH")+"./Pakfile" {
			fmt.Println("Can't find Pakfile.")
			return
		}

		_, err = os.Stat(pakfilePath)
		if os.IsNotExist(err) {
			pakfilePath = "../" + pakfilePath
			continue
		}

		break
	}

	pakInfoBytes, err := ioutil.ReadFile(pakfilePath)
	if err != nil {
		panic(err)
	}
	goyaml.Unmarshal(pakInfoBytes, &pakInfo)

	return
}

func Update() {
	pakInfo := readPakfile()
	// gopath := os.Getenv("GOPATH")

	if len(pakInfo.Packages) == 0 {
		fmt.Println("No packages need to be updated.")
		return
	}

}
