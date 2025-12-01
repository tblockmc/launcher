package downloader

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/havrydotdev/tblock-launcher/pkg/types"
)

func (d *Downloader) DeleteModsAndResourcepacks() error {
	err := os.RemoveAll(filepath.Join(d.cfg.GameDir, "mods"))
	if err != nil {
		return err
	}

	return os.RemoveAll(filepath.Join(d.cfg.GameDir, "resourcepacks"))
}

func (d *Downloader) DeleteVersion() error {
	err := os.Remove(d.getClientPath())
	if err != nil {
		return err
	}

	err = os.RemoveAll(d.getLibrariesPath())
	if err != nil {
		return err
	}

	err = os.RemoveAll(d.getAssetsPath())
	if err != nil {
		return err
	}

	return os.RemoveAll(d.getNativesPath())
}

func (d *Downloader) getClientPath() string {
	return filepath.Join(d.cfg.GameDir, "versions", d.cfg.Versions.Minecraft, "minecraft.jar")
}

func (d *Downloader) getAssetsPath() string {
	return filepath.Join(d.cfg.GameDir, "assets")
}

func (d *Downloader) getLibrariesPath() string {
	return filepath.Join(d.cfg.GameDir, "libraries")
}

func (d *Downloader) getNativesPath() string {
	return filepath.Join(d.cfg.GameDir, "natives")
}

func (d *Downloader) DownloadClient(details *types.VersionDetails, onProgress ProgressCallback) error {
	client := details.Downloads.Client
	clientPath := d.getClientPath()

	return d.downloadWithChecksum(client.URL, clientPath, client.SHA1, onProgress)
}

func (d *Downloader) DownloadLibraries(libraries []types.Library, onProgress ProgressCallback) error {
	librariesPath := d.getLibrariesPath()
	for i, library := range libraries {
		if !d.shouldDownloadLibrary(library) {
			continue
		}

		artifact := library.Downloads.Artifact
		if artifact.URL == "" {
			continue
		}

		// TODO why is this happening?
		// hack, fabric requires another version of asm
		if filepath.Base(artifact.Path) == "asm-9.6.jar" {
			continue
		}

		libraryPath := filepath.Join(librariesPath, artifact.Path)

		d.log.Info("downloading library", slog.Int("progress", i+1), slog.Int("total", len(libraries)), slog.String("name", artifact.Path))
		onProgress(int64(i+1), int64(len(libraries)))

		if err := d.downloadWithChecksum(artifact.URL, libraryPath, artifact.SHA1, func(downloaded, total int64) {}); err != nil {
			return fmt.Errorf("failed to download library %s: %v", library.Name, err)
		}
	}

	return nil
}

// dowloads mods & resoucepacks
func (d *Downloader) DownloadResouces(resources []ResouceData) error {
	modsDir := path.Join(d.cfg.GameDir, "mods")
	err := os.Mkdir(modsDir, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	packsDir := path.Join(d.cfg.GameDir, "resourcepacks")
	err = os.Mkdir(packsDir, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	paths := map[ResourceType]string{
		Mod:          modsDir,
		ResourcePack: packsDir,
	}

	for _, r := range resources {
		dir := paths[r.Type]
		name := path.Base(r.URL)
		err := d.download(r.URL, path.Join(dir, name), func(downloaded, total int64) {})
		if err != nil {
			return err
		}
	}

	return nil
}

func mcRuleToOs(mcOs string) string {
	if mcOs == "osx" {
		return "darwin"
	}

	return mcOs
}

// TODO rules actually have a lot more conditions
// also, when i figure this out launcher itself wont be so 1.21.8 dependent
func (d *Downloader) shouldDownloadLibrary(library types.Library) bool {
	if len(library.Rules) == 0 {
		return true
	}

	for _, rule := range library.Rules {
		if rule.Action == "allow" && runtime.GOOS != mcRuleToOs(rule.OS.Name) {
			return false
		}
	}

	return true
}

// TODO host them on cdn?
func (d *Downloader) WriteOverrides(overrides []StaticAsset) error {
	for _, s := range overrides {
		filePath := path.Join(d.cfg.GameDir, s.Path)
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}

		_, err = file.Write(s.Data)
		if err != nil {
			return err
		}
	}

	return nil
}
