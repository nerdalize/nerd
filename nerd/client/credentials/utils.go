package credentials

import (
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

func TokenFilename() (string, error) {
	f, err := homedir.Expand("~/.nerd/token")
	if err != nil {
		return "", errors.Wrap(err, "failed to retreive homedir path")
	}
	return f, nil
}
