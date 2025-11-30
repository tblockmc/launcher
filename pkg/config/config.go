package config

type Config struct {
	JavaPath string   `json:"java_path"`
	Memory   string   `json:"memory"`
	Username string   `json:"username"`
	GameDir  string   `json:"game_dir"`
	JvmArgs  string   `json:"jvm_args"`
	Versions Versions `json:"versions"`
}

type Versions struct {
	Minecraft    string `json:"minecraft"`
	Launcher     string `json:"launcher"`
	FabricLoader string `json:"fabric_loader"`
}
