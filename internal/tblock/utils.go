package tblock

import (
	"fmt"
	"runtime"
)

func getReleaseArchive() string {
	os := runtime.GOOS
	if os == "darwin" {
		os = "mac"
	}

	arch := runtime.GOARCH
	return fmt.Sprintf("TBlockMC-%s-%s.zip", os, arch)
}

func getBinaryName() string {
	switch runtime.GOOS {
	case "darwin":
		return "tblock-launcher"
	case "windows":
		return "tblockmc.exe"
	default:
		return "tblockmc"
	}
}
