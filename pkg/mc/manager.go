package mc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/havrydotdev/tblock-launcher/types"
)

type VersionManager struct {
	GameDir    string
	Client     *http.Client
	OptionsTxt []byte
	ServersDat []byte

	Stdout io.Writer
}

func NewVersionManager(gameDir string, stdout io.Writer, options, servers []byte) *VersionManager {
	return &VersionManager{
		GameDir:    gameDir,
		OptionsTxt: options,
		ServersDat: servers,
		Client: &http.Client{
			Timeout: 30 * time.Minute,
		},
		Stdout: stdout,
	}
}

func (v *VersionManager) GetVersionManifest() (*types.VersionManifest, error) {
	url := "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json"

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch version manifest: %v", err)
	}
	defer resp.Body.Close()

	var manifest types.VersionManifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("failed to parse version manifest: %v", err)
	}

	return &manifest, nil
}

func (v *VersionManager) GetVersionDetails(versionURL string) (*types.VersionDetails, error) {
	resp, err := http.Get(versionURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch version details: %v", err)
	}
	defer resp.Body.Close()

	var details types.VersionDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, fmt.Errorf("failed to parse version details: %v", err)
	}

	return &details, nil
}

func (v *VersionManager) InstallVersion(versionID string) error {
	fmt.Printf("Installing Minecraft %s...\n", versionID)

	manifest, err := v.GetVersionManifest()
	if err != nil {
		return err
	}

	var versionURL string
	for _, version := range manifest.Versions {
		if version.ID == versionID {
			versionURL = version.URL
			break
		}
	}

	if versionURL == "" {
		return fmt.Errorf("version %s not found", versionID)
	}

	details, err := v.GetVersionDetails(versionURL)
	if err != nil {
		return err
	}

	if err := v.downloadClient(details); err != nil {
		return err
	}

	if err := v.downloadLibraries(details.Libraries); err != nil {
		return err
	}

	if err := v.downloadAssets(details.AssetIndex); err != nil {
		return err
	}

	if err := v.writeDefaultFiles(); err != nil {
		return err
	}

	fmt.Printf("Successfully installed Minecraft %s\n", versionID)
	return nil
}

func (v *VersionManager) downloadClient(details *types.VersionDetails) error {
	client := details.Downloads.Client
	clientPath := filepath.Join(v.GameDir, "versions", details.ID, details.ID+".jar")

	return v.DownloadFileWithChecksum(client.URL, clientPath, client.SHA1)
}

func (v *VersionManager) downloadLibraries(libraries []types.Library) error {
	for i, library := range libraries {
		if !v.shouldDownloadLibrary(library) {
			continue
		}

		artifact := library.Downloads.Artifact
		if artifact.URL == "" {
			continue
		}

		libraryPath := filepath.Join(v.GameDir, "libraries", artifact.Path)

		fmt.Fprintf(v.Stdout, "[%d/%d] Downloading library: %s\n", i+1, len(libraries), filepath.Base(libraryPath))

		if err := v.DownloadFileWithChecksum(artifact.URL, libraryPath, artifact.SHA1); err != nil {
			return fmt.Errorf("failed to download library %s: %v", library.Name, err)
		}
	}

	return nil
}

func (v *VersionManager) shouldDownloadLibrary(library types.Library) bool {
	if len(library.Rules) == 0 {
		return true
	}

	// TODO
	for _, rule := range library.Rules {
		if rule.Action == "allow" {
			return true
		}
		if rule.Action == "disallow" {
			return false
		}
	}

	return true
}

func (v *VersionManager) downloadAssets(assetIndex types.AssetIndex) error {
	return v.DownloadAssets(assetIndex.URL, assetIndex.SHA1)
}

func (v *VersionManager) writeDefaultFiles() error {
	err := v.writeDefaultFile("options.txt", v.OptionsTxt)
	if err != nil {
		return err
	}

	return v.writeDefaultFile("servers.dat", v.ServersDat)
}

func (v *VersionManager) writeDefaultFile(relPath string, content []byte) error {
	filePath := path.Join(v.GameDir, relPath)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	_, err = file.Write(content)
	return err
}
