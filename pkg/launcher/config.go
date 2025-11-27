package launcher

import (
	"encoding/json"
	"os"
	"path"

	"github.com/havrydotdev/tblock-launcher/pkg/utils"
)

const ConfigPath = "tblock_settings.json"

type Config struct {
	JavaPath string `json:"java_path"`
	Memory   string `json:"memory"`
	Username string `json:"username"`
	GameDir  string `json:"game_dir"`
	Version  string `json:"version"`
	JvmArgs  string `json:"jvm_args"`
}

func NewConfig(username, gameDir string) *Config {
	return &Config{
		Username: username, GameDir: gameDir, Version: utils.McVersion,
		JavaPath: utils.DefaultJavaPath, Memory: utils.DefaultMemory, JvmArgs: "",
	}
}

func PersistConfig(gameDir string, cfg *Config) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	cfgPath := path.Join(gameDir, ConfigPath)
	file, err := os.Create(cfgPath)
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}

func ReadPersistedConfig(gameDir string) (*Config, error) {
	cfgPath := path.Join(gameDir, ConfigPath)
	file, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
