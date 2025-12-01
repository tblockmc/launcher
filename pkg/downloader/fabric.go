package downloader

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type FabricVersion struct {
	Loader struct {
		Version string `json:"version"`
	} `json:"loader"`
}

type FabricProfile struct {
	ID           string `json:"id"`
	InheritsFrom string `json:"inheritsFrom"`
	MainClass    string `json:"mainClass"`
	Arguments    struct {
		Game []interface{} `json:"game"`
		JVM  []interface{} `json:"jvm"`
	} `json:"arguments"`
	Libraries []FabricLibrary `json:"libraries"`
}

type FabricLibrary struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

func (d *Downloader) InstallFabric() error {
	mcVersion := d.cfg.Versions.Minecraft
	loaderVersion := d.cfg.Versions.FabricLoader

	d.log.Info("installing fabric", slog.String("loader_version", loaderVersion), slog.String("mc_version", mcVersion))

	versionName := fmt.Sprintf("fabric-loader-%s-%s", loaderVersion, mcVersion)
	versionDir := filepath.Join(d.cfg.GameDir, "versions", versionName)
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return err
	}

	profileURL := fmt.Sprintf("https://meta.fabricmc.net/v2/versions/loader/%s/%s/profile/json", mcVersion, loaderVersion)
	resp, err := http.Get(profileURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var profile FabricProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return err
	}

	profilePath := filepath.Join(versionDir, versionName+".json")
	profileFile, err := os.Create(profilePath)
	if err != nil {
		return err
	}
	defer profileFile.Close()

	profileData, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return err
	}

	if _, err := profileFile.Write(profileData); err != nil {
		return err
	}

	if err := d.downloadFabricLibraries(profile.Libraries); err != nil {
		return err
	}

	d.log.Info("installed fabric", slog.String("version", versionName))
	return nil
}

func (d *Downloader) downloadFabricLibraries(libraries []FabricLibrary) error {
	for _, library := range libraries {
		if err := d.downloadFabricLibrary(library); err != nil {
			return fmt.Errorf("failed to download library %s: %v", library.Name, err)
		}
	}
	return nil
}

// TODO why am i not using download function here?
func (d *Downloader) downloadFabricLibrary(library FabricLibrary) error {
	path := d.mavenToPath(library.Name)
	url := d.mavenToURL(library.Name)

	d.log.Info("downloading fabric library", slog.String("name", library.Name))

	libraryPath := filepath.Join(d.cfg.GameDir, "libraries", path)

	if err := os.MkdirAll(filepath.Dir(libraryPath), 0755); err != nil {
		return err
	}

	if _, err := os.Stat(libraryPath); err == nil {
		return nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(libraryPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (d *Downloader) mavenToPath(name string) string {
	parts := splitMavenName(name)
	return filepath.Join(parts[0], parts[1], parts[2], parts[1]+"-"+parts[2]+".jar")
}

func (d *Downloader) mavenToURL(name string) string {
	parts := splitMavenName(name)
	baseURL := "https://maven.fabricmc.net/"
	return baseURL + parts[0] + "/" + parts[1] + "/" + parts[2] + "/" + parts[1] + "-" + parts[2] + ".jar"
}

func splitMavenName(name string) []string {
	parts := make([]string, 3)
	split1 := strings.Split(name, ":")
	if len(split1) >= 3 {
		parts[0] = strings.ReplaceAll(split1[0], ".", "/")
		parts[1] = split1[1]
		parts[2] = split1[2]
	}

	return parts
}
