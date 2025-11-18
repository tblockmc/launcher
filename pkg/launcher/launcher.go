package launcher

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/havrydotdev/tblock-launcher/pkg/auth"
	"github.com/havrydotdev/tblock-launcher/pkg/mc"
)

type GameLauncher struct {
	config         *Config
	versionManager *mc.VersionManager
}

func NewGameLauncher(cfg *Config, vm *mc.VersionManager) *GameLauncher {
	return &GameLauncher{
		config:         cfg,
		versionManager: vm,
	}
}

func (g *GameLauncher) IsInstalled() bool {
	versionDir := filepath.Join(g.config.GameDir, "versions", g.config.Version)
	clientJar := filepath.Join(versionDir, g.config.Version+".jar")

	_, err := os.Stat(clientJar)

	return !os.IsNotExist(err)
}

func (g *GameLauncher) Install() error {
	return g.versionManager.InstallVersion(g.config.Version)
}

func (g *GameLauncher) Launch(username string) error {
	username, uuid := auth.NewOfflineAuth(username).GetAuthData()

	classpath, err := g.buildClasspath()
	if err != nil {
		return err
	}

	args := g.buildArgs(classpath, username, uuid)
	cmd := exec.Command(g.config.JavaPath, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	fmt.Printf("Launching Minecraft %s...\n", g.config.Version)
	return cmd.Run()
}

func (g *GameLauncher) buildArgs(classpath, username, uuid string) []string {
	jvmArgs := g.buildJvmArgs(classpath)
	gameArgs := g.buildGameArgs(username, uuid)

	return append(jvmArgs, gameArgs...)
}

func (g *GameLauncher) buildJvmArgs(classpath string) []string {
	args := []string{
		"-Xmx" + g.config.Memory,
		"-Djava.library.path=" + filepath.Join(g.config.GameDir, "natives"),
		"-cp", classpath,
	}

	if runtime.GOOS == "darwin" {
		args = append(args, "-XstartOnFirstThread")
	}

	args = append(args, "net.minecraft.client.main.Main")

	return args
}

func (g *GameLauncher) buildGameArgs(username, uuid string) []string {
	return []string{
		"--username", username,
		"--version", g.config.Version,
		"--gameDir", g.config.GameDir,
		"--assetsDir", filepath.Join(g.config.GameDir, "assets"),
		"--assetIndex", "5", // 1.21.4
		"--accessToken", "0",
		"--userType", "legacy",
		"--uuid", uuid,
	}
}

func (g *GameLauncher) buildClasspath() (string, error) {
	var classpathElements []string

	clientJar := filepath.Join(g.config.GameDir, "versions", g.config.Version, g.config.Version+".jar")
	classpathElements = append(classpathElements, clientJar)

	librariesDir := filepath.Join(g.config.GameDir, "libraries")
	err := filepath.Walk(librariesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".jar") {
			classpathElements = append(classpathElements, path)
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	modsDir := filepath.Join(g.config.GameDir, "mods")
	if err := filepath.Walk(modsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".jar") {
			classpathElements = append(classpathElements, path)
		}
		return nil
	}); err != nil && !os.IsNotExist(err) {
		return "", err
	}

	return strings.Join(classpathElements, string(os.PathListSeparator)), nil
}
