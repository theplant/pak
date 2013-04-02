package share

import (
	"os"
)

const (
	Pakfile   = "Pakfile"
	Paklock   = "Pakfile.lock"
	Pakbranch = "pak"
	Paktag    = "_pak_latest_"
)

var Gopath = os.Getenv("GOPATH")

type PakInfo struct {
	Packages []string
}

type PaklockInfo map[string]string

type GetOption struct {
	Fetch    bool
	Force    bool
	Checksum string
	NotGet   bool // add temporally for pass tests. TODO: remove it.
}

type PakOption struct {
	PakMeter       []string // used for containing spcified packages
	UsePakfileLock bool
	Fetch          bool
	Force          bool
	NotGet         bool // add temporally for pass tests. TODO: remove it.
}
