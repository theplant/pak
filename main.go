package main

import (
	"bytes"
	"flag"
	"github.com/theplant/pak/core"
	. "github.com/theplant/pak/share"
	"github.com/wsxiaoys/terminal/color"
	"os/exec"
)

var (
	getLatestFlag bool
	forceFlag     bool
)

func init() {
	flag.Usage = func() {
		spaces := "    "
		color.Printf("@gPackages Management Tool Pak.\n@wUsage:\n")
		color.Printf("%spak init\n", spaces)
		color.Printf("%spak [-uf] get\n", spaces)
		color.Printf("%spak update [package]\n", spaces)
		color.Printf("%spak open [package]\n", spaces)
		color.Printf("%spak list\n", spaces)
		color.Printf("%spak version\n", spaces)
		// fmt.Printf("%spak scan\n", spaces)
		// flag.PrintDefaults()
	}
	flag.BoolVar(&getLatestFlag, "u", false, "Download the lastest revisions from remote repo before checkout.")
	flag.BoolVar(&forceFlag, "f", false, "Force pak to remove pak branch.")
}

func main() {
	flag.Parse()
	switch flag.Arg(0) {
	case "init":
		core.Init()
	case "get":
		err := core.Get(PakOption{UsePakfileLock: true, Force: forceFlag})
		if err != nil {
			color.Printf("@r%s\n", err)
		}
	case "update":
		option := PakOption{UsePakfileLock: false, Force: true}
		option.PakMeter = flag.Args()[1:]
		err := core.Get(option)
		if err != nil {
			color.Printf("@r%s\n", err)
			return
		}
	case "open":
		name := flag.Args()[1]
		matched, pkg, err := core.FindPackage(name)
		if err != nil {
			color.Printf("@r%s\n", err)
			return
		}
		if !matched {
			color.Printf("@rPackage %s Not Exist.\n", name)
			return
		}

		cmd := exec.Command("open", "-t", Gopath+"/src/"+pkg.Name)
		cmd.Stderr = &bytes.Buffer{}
		err = cmd.Run()
		if err != nil {
			color.Printf("@r%s", cmd.Stderr.(*bytes.Buffer).String())
		}
	case "list":
		allPakPkgs, err := core.ParsePakfile()
		if err != nil {
			color.Printf("@r%s\n", err)
			return
		}

		color.Println("All Packages Depended In this Package:")
		for _, pkg := range allPakPkgs {
			color.Printf("@g    %s\n", pkg.Name)
		}
	case "version":
		color.Println("@g1.1.1")
	default:
		flag.Usage()
	}
}
