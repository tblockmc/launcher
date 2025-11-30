package utils

import (
	"encoding/json"
	"os"
	"path"

	"github.com/havrydotdev/tblock-launcher/pkg/config"
)

// TODO: remove hard-coded versions
const (
	McVersion           = "1.21.8"
	FabricLoaderVersion = "0.18.1"
	DefaultMemory       = "4G"
	ConfigPath          = "tblock_settings.json"
	DefaultJavaPath     = ""
)

func GetTblockFolderPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return path.Join(home, ".tblock"), nil
}

func PersistConfig(cfg *config.Config) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	cfgPath := path.Join(cfg.GameDir, ConfigPath)
	file, err := os.Create(cfgPath)
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}

func ReadPersistedConfig(gameDir string) (*config.Config, error) {
	cfgPath := path.Join(gameDir, ConfigPath)
	file, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	var config config.Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
