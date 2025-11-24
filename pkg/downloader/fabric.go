package downloader

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	FabricLoaderVersion = "0.18.1"
)

type FabricInstaller struct {
	GameDir string
}

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

func NewFabricInstaller(gameDir string) *FabricInstaller {
	return &FabricInstaller{
		GameDir: gameDir,
	}
}

func (f *FabricInstaller) InstallFabric(minecraftVersion string) error {
	fmt.Printf("Installing Fabric %s for Minecraft %s\n", FabricLoaderVersion, minecraftVersion)

	versionName := fmt.Sprintf("fabric-loader-%s-%s", FabricLoaderVersion, minecraftVersion)
	versionDir := filepath.Join(f.GameDir, "versions", versionName)
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return err
	}

	profileURL := fmt.Sprintf("https://meta.fabricmc.net/v2/versions/loader/%s/%s/profile/json",
		minecraftVersion, FabricLoaderVersion)

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

	if err := f.downloadFabricLibraries(profile.Libraries); err != nil {
		return err
	}

	fmt.Printf("Fabric installed: %s\n", versionName)
	return nil
}

func (f *FabricInstaller) downloadFabricLibraries(libraries []FabricLibrary) error {
	for _, library := range libraries {
		if err := f.downloadLibrary(library); err != nil {
			return fmt.Errorf("failed to download library %s: %v", library.Name, err)
		}
	}
	return nil
}

func (f *FabricInstaller) downloadLibrary(library FabricLibrary) error {
	path := f.mavenToPath(library.Name)
	url := f.mavenToURL(library.Name)

	log.Printf("Downloading fabric library: %s", library.Name)

	libraryPath := filepath.Join(f.GameDir, "libraries", path)

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

func (f *FabricInstaller) mavenToPath(name string) string {
	parts := splitMavenName(name)
	return filepath.Join(parts[0], parts[1], parts[2], parts[1]+"-"+parts[2]+".jar")
}

func (f *FabricInstaller) mavenToURL(name string) string {
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
