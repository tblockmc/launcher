package downloader

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// TODO why is this value here and not in config??
const (
	JavaVersion = "21"
	JavaRelease = "21.0.9+10"
)

// TODO add graalvm and see if its decreasing start up times
// so players dont rethink their life choices while waiting
// for this fuckass game to boot up
var (
	javaDownloadUrl = fmt.Sprintf("https://github.com/adoptium/temurin21-binaries/releases/download/jdk-%s", url.PathEscape(JavaRelease))
)

func (d *Downloader) GetJavaPath() string {
	javaBaseFolder := path.Join(d.cfg.GameDir, "java", fmt.Sprintf("jdk-%s", JavaRelease))
	if runtime.GOOS == "windows" {
		return filepath.Join(javaBaseFolder, "bin")
	}

	if runtime.GOOS == "darwin" {
		return filepath.Join(javaBaseFolder, "Contents", "Home", "bin")
	}

	return filepath.Join(javaBaseFolder, "bin", "java")
}

func (d *Downloader) DownloadJava() error {
	javaURL := d.getJavaDownloadURL()
	d.log.Info("downloading java", slog.String("url", javaURL))
	if javaURL == "" {
		return fmt.Errorf("unsupported platform: %s %s", runtime.GOOS, runtime.GOARCH)
	}

	zipPath := filepath.Join(d.cfg.GameDir, "java.zip")
	if err := d.downloadJava(javaURL, zipPath); err != nil {
		return err
	}

	if err := d.extractJava(zipPath); err != nil {
		return err
	}

	return os.Remove(zipPath)
}

// e.g. OpenJDK21U-jdk_x64_mac_hotspot_21.0.9_10.tar.gz
func (d *Downloader) getJavaDownloadURL() string {
	arch := formatJavaReleaseArch()

	os := runtime.GOOS
	if runtime.GOOS == "darwin" {
		os = "mac"
	}

	release := strings.ReplaceAll(JavaRelease, "+", "_")
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}

	return fmt.Sprintf("%s/OpenJDK%sU-jdk_%s_%s_hotspot_%s.%s", javaDownloadUrl, JavaVersion, arch, os, release, ext)
}

func formatJavaReleaseArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x64"
	case "arm64":
		return "aarch64"
	case "386":
		return "x86-32"
	}

	return ""
}

func (d *Downloader) downloadJava(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	buffer := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			out.Write(buffer[:n])
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	return nil
}

func (d *Downloader) extractJava(zipPath string) error {
	javaDir := filepath.Join(d.cfg.GameDir, "java")
	os.RemoveAll(javaDir)
	err := os.MkdirAll(javaDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to make java installation directory: %s", err.Error())
	}

	if runtime.GOOS == "windows" {
		return extractZip(zipPath, javaDir)
	}

	return extractTarGz(zipPath, javaDir)
}

// windows
func extractZip(zipPath, dest string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		path := filepath.Join(dest, f.Name)
		os.MkdirAll(filepath.Dir(path), 0755)

		out, err := os.Create(path)
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(out, rc)
		out.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// macos/unix
func extractTarGz(tarPath, dest string) error {
	file, err := os.Open(tarPath)
	if err != nil {
		return err
	}

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}

		target := filepath.Join(dest, header.Name)
		err = os.MkdirAll(filepath.Dir(target), 0755)
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			f.Close()
		}
	}
}
