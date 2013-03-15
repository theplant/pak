package pak

import (
	"os"
	"fmt"
	"io/ioutil"
)

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
