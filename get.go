package pak

import (
	"os"
)

type GetOptions struct {
	All    bool
	Repo   string
	Branch string
}

func Get(option GetOptions) {
	_, err := os.Stat(paklock)
	if os.IsNotExist(err) {

	} else {

	}

	if option.All {

	} else {

	}
}
