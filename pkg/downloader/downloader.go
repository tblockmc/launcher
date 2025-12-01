package downloader

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/havrydotdev/tblock-launcher/pkg/config"
)

type Downloader struct {
	client *http.Client
	cfg    *config.Config
	log    *slog.Logger
}

type ProgressCallback func(downloaded, total int64)

func New(cfg *config.Config) *Downloader {
	return &Downloader{cfg: cfg, client: http.DefaultClient, log: slog.Default()}
}

func (d *Downloader) WithHTTPClient(client *http.Client) *Downloader {
	d.client = client
	return d
}

func (d *Downloader) WithLogger(log *slog.Logger) *Downloader {
	d.log = log
	return d
}

// TODO
func (d *Downloader) download(url, filepath string, onProgress ProgressCallback) error {
	if err := os.MkdirAll(filepath[:strings.LastIndex(filepath, string(os.PathSeparator))], 0755); err != nil {
		return err
	}

	if info, err := os.Stat(filepath); err == nil && info.Size() > 0 {
		d.log.Warn("file already exists", slog.String("path", filepath))
		return nil
	}

	resp, err := d.client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error status for %s: %s", url, resp.Status)
	}

	tmpPath := filepath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	downloaded := 0
	total := resp.ContentLength
	buffer := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := out.Write(buffer[:n]); writeErr != nil {
				return writeErr
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		downloaded += n

		onProgress(int64(downloaded), total)
	}

	out.Close()

	if err := os.Rename(tmpPath, filepath); err != nil {
		return err
	}

	return nil
}

func (d *Downloader) downloadWithChecksum(url, filepath, expectedSHA1 string, onProgress ProgressCallback) error {
	if err := d.download(url, filepath, onProgress); err != nil {
		return err
	}

	if expectedSHA1 != "" {
		if err := d.verifyChecksum(filepath, expectedSHA1); err != nil {
			os.Remove(filepath)
			return fmt.Errorf("checksum verification failed for %s: %v", filepath, err)
		}
	}

	return nil
}

func (d *Downloader) verifyChecksum(filepath, expectedSHA1 string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	hasher := sha1.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return err
	}

	actualSHA1 := hex.EncodeToString(hasher.Sum(nil))
	if actualSHA1 != expectedSHA1 {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedSHA1, actualSHA1)
	}

	return nil
}
