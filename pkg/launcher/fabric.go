package launcher

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/havrydotdev/tblock-launcher/internal/utils"
	"github.com/havrydotdev/tblock-launcher/pkg/auth"
	"github.com/havrydotdev/tblock-launcher/pkg/config"
	"github.com/havrydotdev/tblock-launcher/pkg/downloader"
)

type FabricLauncher struct {
	cfg               *config.Config
	fabricVersionName string
}

func NewFabricLauncher(cfg *config.Config) *FabricLauncher {
	return &FabricLauncher{
		cfg:               cfg,
		fabricVersionName: fmt.Sprintf("fabric-loader-%s-%s", cfg.Versions.FabricLoader, cfg.Versions.Minecraft),
	}
}

func (f *FabricLauncher) Launch() error {
	if !f.IsFabricInstalled() {
		return fmt.Errorf("fabric is not installed. please install it first")
	}

	cmd, err := f.buildFabricCommand(f.fabricVersionName)
	if err != nil {
		return err
	}

	cmd.Dir = f.cfg.GameDir
	cmd.Stdout = os.Stdout

	fmt.Printf("Launching Minecraft with Fabric %s...\n", f.cfg.Versions.FabricLoader)
	return cmd.Run()
}

func (f *FabricLauncher) IsFabricInstalled() bool {
	profilePath := filepath.Join(f.cfg.GameDir, "versions", f.fabricVersionName, f.fabricVersionName+".json")
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func (f *FabricLauncher) buildFabricCommand(versionName string) (*exec.Cmd, error) {
	profile, err := f.loadFabricProfile(versionName)
	if err != nil {
		return nil, err
	}

	classpath, err := f.buildFabricClasspath()
	if err != nil {
		return nil, err
	}

	jvmArgs := f.buildFabricJVMArgs(profile, classpath)

	gameArgs := f.buildFabricGameArgs(versionName)

	allArgs := append(jvmArgs, gameArgs...)

	javaBinary := "java"
	if runtime.GOOS == "windows" {
		javaBinary = "java.exe"
	}

	javaExec := filepath.Join(f.cfg.JavaPath, javaBinary)

	log.Println(javaExec, allArgs)

	return exec.Command(javaExec, allArgs...), nil
}

func (f *FabricLauncher) loadFabricProfile(versionName string) (*downloader.FabricProfile, error) {
	profilePath := filepath.Join(f.cfg.GameDir, "versions", versionName, versionName+".json")
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return nil, err
	}

	var profile downloader.FabricProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

func (f *FabricLauncher) buildFabricClasspath() (string, error) {
	var classpathElements []string

	mcJar := filepath.Join(f.cfg.GameDir, "versions", utils.McVersion, fmt.Sprintf("%s.jar", utils.McVersion))
	classpathElements = append(classpathElements, mcJar)

	librariesDir := filepath.Join(f.cfg.GameDir, "libraries")
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

	modsDir := filepath.Join(f.cfg.GameDir, "mods")
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

	fabricJar := filepath.Join(f.cfg.GameDir, "versions",
		f.fabricVersionName,
		f.fabricVersionName+".jar")
	if _, err := os.Stat(fabricJar); err == nil {
		classpathElements = append(classpathElements, fabricJar)
	}

	return strings.Join(classpathElements, string(os.PathListSeparator)), nil
}

func (f *FabricLauncher) buildFabricJVMArgs(profile *downloader.FabricProfile, classpath string) []string {
	natives := filepath.Join(f.cfg.GameDir, "natives")
	args := []string{
		"-Xmx" + f.cfg.Memory,
		"-Djava.library.path=" + natives,
		"-Djna.tmpdir=" + natives,
		"-Dorg.lwjgl.system.SharedLibraryExtractPath=" + natives,
		"-Dio.netty.native.workdir=" + natives,
		"-Dminecraft.launcher.brand=tblock",
		// TODO
		"-Dminecraft.launcher.version=0.2.1",
		"-cp", classpath,
	}

	if f.cfg.JvmArgs != "" {
		args = append(args, f.cfg.JvmArgs)
	}

	if runtime.GOOS == "darwin" {
		args = append(args, "-XstartOnFirstThread")
	}

	for _, arg := range profile.Arguments.JVM {
		if strArg, ok := arg.(string); ok {
			resolvedArg := f.resolvePlaceholders(strArg)
			args = append(args, resolvedArg)
		}
	}

	args = append(args, "net.fabricmc.loader.impl.launch.knot.KnotClient")

	return args
}

func (f *FabricLauncher) buildFabricGameArgs(versionName string) []string {
	auth := auth.NewOfflineAuth(f.cfg.Username)
	username, uuid := auth.GetAuthData()

	args := []string{
		"--username", username,
		"--version", versionName,
		"--gameDir", f.cfg.GameDir,
		"--assetsDir", filepath.Join(f.cfg.GameDir, "assets"),
		"--assetIndex", "5", // For 1.21.4
		"--accessToken", "0",
		"--userType", "legacy",
		"--uuid", uuid,
	}

	return args
}

func (f *FabricLauncher) resolvePlaceholders(arg string) string {
	replacements := map[string]string{
		"${natives_directory}": filepath.Join(f.cfg.GameDir, "natives"),
		"${launcher_name}":     "TBlock Launcher",
		"${launcher_version}":  "1.0.0",
		"${classpath}":         "dummy", // This gets handled separately
	}

	for placeholder, value := range replacements {
		arg = strings.ReplaceAll(arg, placeholder, value)
	}

	return strings.ReplaceAll(arg, " ", "")
}
