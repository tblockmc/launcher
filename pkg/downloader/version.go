package downloader

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/havrydotdev/tblock-launcher/pkg/types"
)

func (d *Downloader) GetVersionURL() (string, error) {
	url := "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json"

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch version manifest: %v", err)
	}
	defer resp.Body.Close()

	var manifest types.VersionManifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return "", fmt.Errorf("failed to parse version manifest: %v", err)
	}

	var versionURL string
	for _, version := range manifest.Versions {
		if version.ID == d.cfg.Versions.Minecraft {
			versionURL = version.URL
			break
		}
	}

	if versionURL == "" {
		return "", fmt.Errorf("version %s not found", d.cfg.Versions.Minecraft)
	}

	return versionURL, nil
}

func (d *Downloader) GetVersionDetails(versionURL string) (*types.VersionDetails, error) {
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
