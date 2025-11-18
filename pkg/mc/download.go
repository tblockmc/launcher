package mc

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func (v *VersionManager) DownloadFile(url, filepath string) error {
	if err := os.MkdirAll(filepath[:strings.LastIndex(filepath, "/")], 0755); err != nil {
		return err
	}

	if info, err := os.Stat(filepath); err == nil && info.Size() > 0 {
		fmt.Printf("File already exists: %s\n", filepath)
		return nil
	}

	fmt.Fprintf(v.Stdout, "Downloading: %s\n", url)

	resp, err := v.Client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status for %s: %s", url, resp.Status)
	}

	tmpPath := filepath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer out.Close()

	var downloaded int64
	buffer := make([]byte, 32*1024)

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := out.Write(buffer[:n]); writeErr != nil {
				return writeErr
			}
			downloaded += int64(n)
			fmt.Fprintf(v.Stdout, "\rDownloaded: %.2f MB", float64(downloaded)/(1024*1024))
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	if err := os.Rename(tmpPath, filepath); err != nil {
		return err
	}

	fmt.Fprintf(v.Stdout, "Downloaded: %s\n", filepath)

	return nil
}

func (v *VersionManager) DownloadFileWithChecksum(url, filepath, expectedSHA1 string) error {
	if err := v.DownloadFile(url, filepath); err != nil {
		return err
	}

	if expectedSHA1 != "" {
		if err := v.verifyChecksum(filepath, expectedSHA1); err != nil {
			os.Remove(filepath)
			return fmt.Errorf("checksum verification failed for %s: %v", filepath, err)
		}
	}

	return nil
}

func (v *VersionManager) verifyChecksum(filepath, expectedSHA1 string) error {
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
