package kubeconfig

import (
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

// GetPath returns the expanded and normalized kube config path
func GetPath(kubeConfig string) (string, error) {
	hdir, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	if kubeConfig == "" {
		kubeConfig = filepath.Join(hdir, ".kube", "config")
	}
	kubeConfig, err = homedir.Expand(kubeConfig)
	if err != nil {
		return "", errors.Wrap(err, "failed to expand home directory in kube config file path")
	}
	//Normalize all slashes to native platform slashes (e.g. / to \ on Windows)
	kubeConfig = filepath.FromSlash(kubeConfig)
	return kubeConfig, nil
}
