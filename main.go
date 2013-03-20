package main

import (
	"flag"
	"fmt"
	"github.com/theplant/pak/core"
	. "github.com/theplant/pak/share"
)

var (
	getLatestFlag bool
	forceFlag     bool
)

func init() {
	flag.Usage = func() {
		spaces := "    "
		fmt.Printf("Usage:\n")
		fmt.Printf("%spak init\n", spaces)
		fmt.Printf("%spak [-uf] get\n", spaces)
		fmt.Printf("%spak update [package]\n", spaces)
		fmt.Printf("%spak version\n", spaces)
		// fmt.Printf("%spak open\n", spaces)
		// fmt.Printf("%spak list\n", spaces)
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
		pak.Init()
	case "get":
		err := pak.Get(PakOption{UsePakfileLock: true, Fetch: getLatestFlag, Force: forceFlag})
		if err != nil {
			fmt.Printf("%s\n", err)
		}
	case "update":
		option := PakOption{UsePakfileLock: false, Fetch: true, Force: true}
		option.PakMeter = flag.Args()[1:]
		err := pak.Get(option)
		if err != nil {
			fmt.Printf("%s\n", err)
		}
	case "version":
		fmt.Println("1.0.0")
	default:
		flag.Usage()
	}
}
