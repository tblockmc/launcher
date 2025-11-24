package downloader

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/havrydotdev/tblock-launcher/pkg/types"
)

func (d *Downloader) DownloadClient(details *types.VersionDetails) error {
	client := details.Downloads.Client
	clientPath := filepath.Join(d.gameDir, "versions", details.ID, details.ID+".jar")

	return d.downloadWithChecksum(client.URL, clientPath, client.SHA1)
}

func (d *Downloader) DownloadLibraries(libraries []types.Library) error {
	for i, library := range libraries {
		if !d.shouldDownloadLibrary(library) {
			continue
		}

		artifact := library.Downloads.Artifact
		if artifact.URL == "" {
			continue
		}

		// hack, fabric required another version of asm
		if filepath.Base(artifact.Path) == "asm-9.6.jar" {
			continue
		}

		libraryPath := filepath.Join(d.gameDir, "libraries", artifact.Path)

		fmt.Fprintf(d.stdout, "[%d/%d] Downloading library: %s\n", i+1, len(libraries), filepath.Base(libraryPath))

		if err := d.downloadWithChecksum(artifact.URL, libraryPath, artifact.SHA1); err != nil {
			return fmt.Errorf("failed to download library %s: %v", library.Name, err)
		}
	}

	return nil
}

func (d *Downloader) DownloadMods(mods []ModData) error {
	modsDir := path.Join(d.gameDir, "mods")
	err := os.Mkdir(modsDir, 0755)
	if err != nil {
		return err
	}

	for _, mod := range mods {
		err := d.download(mod.URL, path.Join(modsDir, mod.Name+".jar"))
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

func (d *Downloader) WriteStaticFiles(static []StaticAsset) error {
	for _, s := range static {
		filePath := path.Join(d.gameDir, s.Path)
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
