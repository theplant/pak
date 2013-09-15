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
	Packages         []PkgCfg
	BasicDependences []string
}

type GpiParams struct {
	Type                 string // 11 => both Pakfile and Pakfile.lock, 10 => Pkafile only
	Path                 string
	DeepParse            bool
	WithBasicDependences bool
	Verbose              bool
}

func GetPakInfo(params GpiParams) (pakInfo PakInfo, paklockInfo PaklockInfo, err error) {
	var pakInfoBytes, paklockInfoBytes []byte
	var pakfilePath, pakfileLockPath string
	if params.Path == "" {
		pakfilePath = Pakfile
		pakfileLockPath = Paklock
	} else {
		pakfilePath = params.Path + "/" + Pakfile
		pakfileLockPath = params.Path + "/" + Paklock
	}

	pakInfoBytes, err = pakRead(pakfilePath)
	if err != nil {
		return
	}

	if params.Type[1] == '1' {
		paklockInfoBytes, err = pakRead(pakfileLockPath)
		if err != nil {
			return
		}
		// TODO: Report UnPak dependences which contains a Pakfile but without Pakfile.lock
		// if len(paklockInfoBytes) == 0 && params.Path != "" && params.Verbose {
		// 	name := params.Path[len(Gopath+"/src/"):]
		// 	color.Printf("@g%s@ydoes not contains Pakfile.lock. It's recommended to also pak this package.", name)
		// }
	}

	// Pakfile must exist
	if pakInfoBytes == nil {
		return PakInfo{}, PaklockInfo{}, nil
	}

	err = goyaml.Unmarshal(pakInfoBytes, &pakInfo)
	if err != nil {
		return PakInfo{}, PaklockInfo{}, nil
	}

	paklockInfo = PaklockInfo{}
	err = goyaml.Unmarshal(paklockInfoBytes, &paklockInfo)
	if err != nil {
		return PakInfo{}, PaklockInfo{}, nil
	}

	if params.WithBasicDependences {
		for _, pkg := range pakInfo.Packages {
			pakInfo.BasicDependences = append(pakInfo.BasicDependences, pkg.Name)
		}
	}

	if !params.DeepParse {
		for i, _ := range pakInfo.Packages {
			if pakInfo.Packages[i].PakName == "" {
				pakInfo.Packages[i].PakName = Pakbranch
			}
		}
		return
	}

	subPakInfos := []PakInfo{}
	subPaklockInfos := []PaklockInfo{}
	for i, _ := range pakInfo.Packages {
		if pakInfo.Packages[i].PakName == "" {
			pakInfo.Packages[i].PakName = Pakbranch
		}

		subPakInfo, subPaklockInfo, err := GetPakInfo(GpiParams{
			Path:      Gopath + "/src/" + pakInfo.Packages[i].Name,
			Type:      params.Type,
			DeepParse: params.DeepParse,
		})
		if err != nil {
			return pakInfo, paklockInfo, err
		}
		if len(subPakInfo.Packages) > 0 {
			subPakInfos = append(subPakInfos, subPakInfo)
		}
		if len(subPaklockInfo) > 0 {
			subPaklockInfos = append(subPaklockInfos, subPaklockInfo)
		}
	}

	// to detect package dependence conflicts.
	subCount := len(subPakInfos)
	for i, _ := range subPakInfos {
		pakInfo.Packages = append(subPakInfos[subCount-i-1].Packages, pakInfo.Packages...)
	}
	if params.Type[1] == '1' {
		for _, subPaklockInfo := range subPaklockInfos {
			for k, v := range subPaklockInfo {
				paklockInfo[k] = v
			}
		}
	}

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
