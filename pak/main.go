package main

import (
	"github.com/theplant/pak"
	"flag"
	"fmt"
)

var (
	getLatestFlag bool
)

func init() {
	flag.Usage = func() {
		spaces := "    "
		fmt.Printf("Usage:\n")
		fmt.Printf("%spak init\n", spaces)
		fmt.Printf("%spak [-u] get [package]\n", spaces)
		fmt.Printf("%spak update\n", spaces)
		// fmt.Printf("%spak open\n", spaces)
		// fmt.Printf("%spak list\n", spaces)
		// fmt.Printf("%spak scan\n", spaces)
		// flag.PrintDefaults()
	}
	flag.BoolVar(&getLatestFlag, "u", false, "Download the lastest revisions from remote repo before checkout")
}

func main() {
	flag.Parse()
	switch flag.Arg(0) {
	case "init":
		pak.Init()
	case "get":
		// pak.Get()
	case "update":
		pak.Update()
	}
}