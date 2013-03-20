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
	Pull     bool
	Force    bool
	Checksum string
}

type PakOption struct {
	PakMeter       []string
	UsePakfileLock bool
	Pull           bool
	Force          bool
}
