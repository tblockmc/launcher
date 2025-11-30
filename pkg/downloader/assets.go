package downloader

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/havrydotdev/tblock-launcher/pkg/types"
)

const ConcurrentDownloads = 10

type ResourceType int

const (
	ResourcePack ResourceType = iota
	Mod
)

type ResouceData struct {
	Type ResourceType
	URL  string
}

type StaticAsset struct {
	Path string
	Data []byte
}

type AssetIndex struct {
	Objects map[string]AssetObject `json:"objects"`
}

type AssetObject struct {
	Hash string `json:"hash"`
	Size int    `json:"size"`
}

type AssetDownloadJob struct {
	Name string
	Hash string
	Size int
}

type AssetDownloadResult struct {
	Name    string
	Error   error
	Skipped bool
}

func (d *Downloader) DownloadAssets(assets types.AssetIndex, onProgress ProgressCallback) error {
	assetsDir := filepath.Join(d.cfg.GameDir, "assets")
	indexPath := filepath.Join(assetsDir, "indexes", "5.json") // 1.21.4

	if err := d.downloadWithChecksum(assets.URL, indexPath, assets.SHA1, func(downloaded, total int64) {}); err != nil {
		return fmt.Errorf("failed to download asset index: %v", err)
	}

	assetIndex, err := d.parseAssetIndex(indexPath)
	if err != nil {
		return fmt.Errorf("failed to parse asset index: %v", err)
	}

	return d.downloadAllAssets(assetIndex, onProgress)
}

func (d *Downloader) parseAssetIndex(indexPath string) (*AssetIndex, error) {
	file, err := os.Open(indexPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var indexData map[string]interface{}
	if err := json.NewDecoder(file).Decode(&indexData); err != nil {
		return nil, err
	}

	// Extract objects from the correct location in the JSON
	objectsMap, ok := indexData["objects"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid asset index format: missing objects")
	}

	assetIndex := &AssetIndex{
		Objects: make(map[string]AssetObject),
	}

	for name, obj := range objectsMap {
		objMap, ok := obj.(map[string]interface{})
		if !ok {
			continue
		}

		hash := objMap["hash"].(string)
		size := objMap["size"].(float64)

		if hash != "" {
			assetIndex.Objects[name] = AssetObject{
				Hash: hash,
				Size: int(size),
			}
		}
	}

	return assetIndex, nil
}

func (d *Downloader) downloadAllAssets(assetIndex *AssetIndex, onProgress ProgressCallback) error {
	total := len(assetIndex.Objects)
	jobs := make(chan AssetDownloadJob, total)
	results := make(chan AssetDownloadResult, total)

	var wg sync.WaitGroup
	for range ConcurrentDownloads {
		wg.Add(1)
		go func() {
			d.assetDownloadWorker(jobs, results, &wg)
		}()
	}

	go func() {
		for name, obj := range assetIndex.Objects {
			jobs <- AssetDownloadJob{
				Name: name,
				Hash: obj.Hash,
				Size: obj.Size,
			}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var downloaded, skipped, failed int
	for result := range results {
		if result.Error != nil {
			d.log.Error("failed to download asset", slog.String("name", result.Name), slog.String("error", result.Error.Error()))
			failed++
		} else if result.Skipped {
			skipped++
		} else {
			downloaded++
		}

		progress := downloaded + skipped + failed
		if onProgress != nil {
			onProgress(int64(progress), int64(total))
		}

		if progress%100 == 0 {
			d.log.Info("successfully downloaded asset", slog.Int("progress", progress),
				slog.Int("total", total), slog.Int("downloaded", downloaded),
				slog.Int("skipped", skipped), slog.Int("failed", failed))
		}
	}

	if failed > 0 {
		return fmt.Errorf("%d assets failed to download", failed)
	}

	return nil
}

func (d *Downloader) assetDownloadWorker(jobs <-chan AssetDownloadJob, results chan<- AssetDownloadResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		result := AssetDownloadResult{Name: job.Name}

		url := fmt.Sprintf("https://resources.download.minecraft.net/%s/%s", job.Hash[:2], job.Hash)
		assetPath := filepath.Join(d.cfg.GameDir, "assets", "objects", job.Hash[:2], job.Hash)

		if info, err := os.Stat(assetPath); err == nil {
			if info.Size() == int64(job.Size) {
				if err := d.verifyChecksum(assetPath, job.Hash); err == nil {
					result.Skipped = true
					results <- result
					continue
				}
			}
		}

		if err := d.downloadWithChecksum(url, assetPath, job.Hash, func(downloaded, total int64) {}); err != nil {
			result.Error = err
		}

		results <- result
	}
}
