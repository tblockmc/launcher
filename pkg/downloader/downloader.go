package downloader

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Downloader struct {
	client  *http.Client
	gameDir string
	stdout  io.Writer
}

func New(gameDir string) *Downloader {
	return &Downloader{
		gameDir: gameDir, client: http.DefaultClient, stdout: os.Stdout,
	}
}

func (d *Downloader) WithHTTPClient(client *http.Client) *Downloader {
	d.client = client
	return d
}

func (d *Downloader) WithStdout(stdout io.Writer) *Downloader {
	d.stdout = stdout
	return d
}

func (d *Downloader) download(url, filepath string) error {
	if err := os.MkdirAll(filepath[:strings.LastIndex(filepath, string(os.PathSeparator))], 0755); err != nil {
		return err
	}

	if info, err := os.Stat(filepath); err == nil && info.Size() > 0 {
		fmt.Printf("File already exists: %s\n", filepath)
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
	}

	out.Close()

	if err := os.Rename(tmpPath, filepath); err != nil {
		return err
	}

	return nil
}

func (d *Downloader) downloadWithChecksum(url, filepath, expectedSHA1 string) error {
	if err := d.download(url, filepath); err != nil {
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
