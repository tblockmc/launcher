package tblock

import (
	"fmt"
	"log"
	"runtime"

	"fyne.io/fyne/v2"
	"github.com/havrydotdev/tblock-launcher/internal/utils"
	"github.com/havrydotdev/tblock-launcher/pkg/config"
)

func getReleaseArchive() string {
	os := runtime.GOOS
	if os == "darwin" {
		os = "mac"
	}

	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x64"
	}

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

func ReadPersistedConfigOrDefault(app fyne.App) (*config.Config, error) {
	gameDir, err := utils.GetTblockFolderPath()
	if err != nil {
		return nil, fmt.Errorf("failed to determine game folder: %s", err.Error())
	}

	cfg, err := utils.ReadPersistedConfig(gameDir)
	if err != nil {
		log.Println("Failed to read config file: ", err)
		// return default config
		return &config.Config{
			Username: "", GameDir: gameDir, JavaPath: utils.DefaultJavaPath,
			Memory: utils.DefaultMemory, JvmArgs: "", Versions: config.Versions{
				Minecraft: utils.McVersion, Launcher: app.Metadata().Version,
				FabricLoader: utils.FabricLoaderVersion,
			},
		}, nil
	}

	return cfg, nil
}
