package core

import (
	"fmt"
	. "github.com/theplant/pak/share"
	"io/ioutil"
	"launchpad.net/goyaml"
	"os"
	"path/filepath"
	"strings"
)

type PakInfo struct {
	Packages []PkgCfg
}

func GetPakInfo(path string) (pakInfo PakInfo, err error) {
	var pakInfoBytes []byte
	if path == "" {
		path = Pakfile
	} else {
		path += "/" + Pakfile
	}
	pakInfoBytes, err = pakRead(path)
	if err != nil {
		return
	}

	if pakInfoBytes == nil {
		return PakInfo{}, nil
	}

	err = goyaml.Unmarshal(pakInfoBytes, &pakInfo)

	subPakInfos := []PakInfo{}
	for i, _ := range pakInfo.Packages {
		if pakInfo.Packages[i].PakName == "" {
			pakInfo.Packages[i].PakName = Pakbranch
		}

		subPakInfo, err := GetPakInfo(Gopath + "/src/" + pakInfo.Packages[i].Name)
		if err != nil {
			return pakInfo, err
		}
		if len(subPakInfo.Packages) > 0 {
			subPakInfos = append(subPakInfos, subPakInfo)
		}
	}

	subCount := len(subPakInfos)
	for i, _ := range subPakInfos {
		pakInfo.Packages = append(subPakInfos[subCount-i-1].Packages, pakInfo.Packages...)
	}

	return
}

func GetPaklockInfo(path string) (paklockInfo PaklockInfo, err error) {
	var content []byte
	if path == "" {
		path = Paklock
	} else {
		path += "/" + Paklock
	}
	content, err = pakRead(path)
	if err != nil {
		return
	}

	err = goyaml.Unmarshal(content, &paklockInfo)

	return
}

var PakfileNotExist = fmt.Errorf("Can't find %s", Pakfile)
var PakfileLockNotExist = fmt.Errorf("Can't find %s", Paklock)

func pakRead(path string) (fileContent []byte, err error) {
	absPakfilePath := ""
	for true {
		absPakfilePath, err = filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		if absPakfilePath == Gopath+"/"+Pakfile || absPakfilePath == Gopath+"/"+Paklock {
			if strings.Contains(path, Pakfile) {
				return nil, nil
			} else {
				return nil, nil
			}
		}

		_, err = os.Stat(path)
		if os.IsNotExist(err) {
			index := strings.LastIndex(path, Pakfile)
			if index == -1 {
				index = strings.LastIndex(path, Paklock)
				if index == -1 {
					return nil, fmt.Errorf("Illegal path for pakRead")
				}
			}
			path = path[:index] + "../" + path[index:]
			continue
		}

		break
	}

	return ioutil.ReadFile(path)
}

func writePaklockInfo(paklockInfo PaklockInfo) error {
	content, err := goyaml.Marshal(&paklockInfo)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(Paklock, content, os.FileMode(0644))
}
