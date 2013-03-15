package pak

import (
	"fmt"
	"io/ioutil"
	"launchpad.net/goyaml"
	"os"
	"os/exec"
	"bytes"
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
)

var (
	gopath = os.Getenv("GOPATH")
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
	_, err := os.Stat(paklock)
	if os.IsNotExist(err) {

	} else {

	}

	if option.All {

	} else {

	}
}

func isPackageClean(pkg string) (result bool) {
	pkgPath := getPackagePath(pkg)
	cmd := exec.Command("git", gitDir(pkgPath), gitWorkTree(pkgPath), "status", "-s")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	return out.String() == ""
}

func checkoutPakbranch(pkg, sha string) (result bool) {
	pkgPath := getPackagePath(pkg)
	cmd := exec.Command("git", gitDir(pkgPath), gitWorkTree(pkgPath), "checkout", sha, "-b", pakbranch)
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	return out.String() == fmt.Sprintf("Switched to a new branch '%s'", pakbranch)
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

	if len(pakInfo.Packages) == 0 {
		fmt.Println("No packages need to be updated.")
		return
	}

	paklockInfo := PaklockInfo{}
	for _, pkg := range pakInfo.Packages {
		shaVal := getPackageSHA(pkg)
		paklockInfo[pkg] = shaVal[0:len(shaVal)-1]
	}

	paklockInfoBts, err := goyaml.Marshal(paklockInfo)
	if err != nil {
	    panic(err)
	}

	ioutil.WriteFile(paklock, paklockInfoBts, os.FileMode(0644))
}

func getPackageSHA(pkg string) (sha string) {
	pkgPath := getPackagePath(pkg)

	cmd := exec.Command("git", gitDir(pkgPath), gitWorkTree(pkgPath), "rev-parse", "HEAD")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	return out.String()
}

func gitDir(pkgPath string) string {
	return fmt.Sprintf("--git-dir=%s/.git", pkgPath)
}

func gitWorkTree(pkgPath string) string {
	return fmt.Sprintf("--work-tree=%s", pkgPath)
}

func getPackagePath(pkg string) string {
	return fmt.Sprintf("%s/src/%s", gopath, pkg)
}
