package pak

import (
	"os"
	// "fmt"
	"errors"
	"io/ioutil"
)

func Init() error {
	_, err := os.Stat(pakfile)
	if err == nil {
		return errors.New("Pakfile already existed.")
	} else if !os.IsNotExist(err) {
		return err
	}

	return ioutil.WriteFile(pakfile, []byte(PakfileTemplate), os.FileMode(0644))
}
