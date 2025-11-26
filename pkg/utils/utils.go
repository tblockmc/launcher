package utils

import (
	"os"
	"path"
)

// TODO: remove hard-coded versions
const (
	McVersion       = "1.21.8"
	DefaultMemory   = "4G"
	DefaultJavaPath = ""
)

func GetTblockFolderPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return path.Join(home, ".tblock"), nil
}
