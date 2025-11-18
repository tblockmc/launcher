package utils

import (
	"os"
	"path"
)

const (
	McVersion       = "1.21.4"
	DefaultMemory   = "4G"
	DefaultJavaPath = "java"
)

func GetTblockFolderPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return path.Join(home, ".tblock"), nil
}
