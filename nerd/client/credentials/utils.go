package credentials

import (
	"errors"
	"os"
	"path/filepath"
)

func TokenFilename() (string, error) {
	homeDir := os.Getenv("HOME") // *nix
	if homeDir == "" {           // Windows
		homeDir = os.Getenv("USERPROFILE")
	}
	if homeDir == "" {
		return "", errors.New("homedir not found")
	}

	return filepath.Join(homeDir, ".nerd", "token"), nil
}
