package main

import (

	// "bytes"
	"flag"
	"os"
	"os/exec"
	"runtime"

	"github.com/theplant/pak/check"
	"github.com/theplant/pak/core"
	. "github.com/theplant/pak/share"
	"github.com/wsxiaoys/terminal/color"
)

var (
	// getLatest       bool
	force           bool
	skipUncleanPkgs bool
	gogeterror      bool
)

func init() {
	flag.Usage = func() {
		spaces := "    "
		color.Printf("@gPackages Management Tool Pak.\n@wUsage:\n")
		color.Printf("%spak init\n", spaces)
		color.Printf("%spak [-sfe] get [package]\n", spaces)
		color.Printf("%spak [-se] update [package]\n", spaces)
		color.Printf("%spak open [package]\n", spaces)
		color.Printf("%spak list\n", spaces)
		color.Printf("%spak version\n", spaces)
		// fmt.Printf("%spak scan\n", spaces)
		flag.PrintDefaults()
	}
	// flag.BoolVar(&getLatest, "u", false, "Download the lastest revisions from remote repo before checkout.")
	flag.BoolVar(&force, "f", false, "Force pak to remove pak branch.")
	flag.BoolVar(&skipUncleanPkgs, "s", false, "Left out unclean packages.")
	flag.BoolVar(&gogeterror, "e", false, "Print out go get error")
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()
	core.PrintGoGetError = gogeterror
	switch flag.Arg(0) {
	case "init":
		initPak()
	case "get":
		getPakPkgs()
	case "update":
		updatePakPkgs()
	case "open":
		openPkgWithPakEditor()
	case "list":
		listPakfilePkgs()
	case "check":
		docheck()
	case "version":
		color.Println("@g1.5.0")
	default:
		flag.Usage()
	}
}

func initPak() {
	err := core.Init()
	if err != nil {
		color.Printf("@r%+v@w\n", err)
		os.Exit(1)
	}
}

func getPakPkgs() {
	err := core.Get(PakOption{
		PakMeter:        flag.Args()[1:],
		UsePakfileLock:  true,
		Force:           force,
		SkipUncleanPkgs: skipUncleanPkgs,
		Verbose:         true,
	})

	if err != nil {
		color.Printf("@r%s\n", err)
		color.Println("Pak Failed.")
		os.Exit(1)
	}
}

func updatePakPkgs() {
	err := core.Get(PakOption{
		UsePakfileLock:  false,
		Force:           true,
		SkipUncleanPkgs: skipUncleanPkgs,
		PakMeter:        flag.Args()[1:],
		Verbose:         true,
	})

	if err != nil {
		color.Printf("@r%s\n", err)
		os.Exit(1)
	}
}

func openPkgWithPakEditor() {
	pakEditor := os.Getenv("PAK_OPEN_EDITOR")
	if pakEditor == "" {
		color.Printf("@rPAK_OPEN_EDITOR is Not Configured.@w")
		color.Println(" Please configure it like below in your ~/.bash_profile or anywhere that is accessible to pak.")
		color.Println("    @gexport PAK_OPEN_EDITOR={mate or whatever editor that you like}@w")
		return
	}

	name := flag.Args()[1]
	pkg, err := core.FindPackage(name)
	if err != nil {
		color.Printf("@r%s\n", err)
		return
	}
	if pkg.Name == "" {
		color.Printf("@rPackage %s Not Exist.\n", name)
		return
	}

	cmd := exec.Command(pakEditor, Gopath+"/src/"+pkg.Name)
	err = cmd.Run()
	if err != nil {
		color.Printf("@r%s@w\n", err)
		os.Exit(1)
	}
}

func listPakfilePkgs() {
	pakInfo, _, err := core.GetPakInfo(core.GpiParams{
		Type:                 "10",
		Path:                 "",
		DeepParse:            false,
		WithBasicDependences: true,
	})
	if err != nil {
		color.Printf("@r%s\n", err)
		return
	}

	var allPakPkgs []core.PakPkg
	for _, pkgCfg := range pakInfo.Packages {
		pakPkg := core.NewPakPkg(pkgCfg)

		allPakPkgs = append(allPakPkgs, pakPkg)
	}

	color.Println("All Packages Depended In this Package:")
	for _, pkg := range allPakPkgs {
		color.Printf("@g    %s\n", pkg.Name)
	}
}

func docheck() {
	check.Check()
}
