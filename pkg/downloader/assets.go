package downloader

import (
	"encoding/json"
	"fmt"
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

func (d *Downloader) DownloadAssets(assets types.AssetIndex) error {
	assetsDir := filepath.Join(d.cfg.GameDir, "assets")
	indexPath := filepath.Join(assetsDir, "indexes", "5.json") // 1.21.4

	if err := d.downloadWithChecksum(assets.URL, indexPath, assets.SHA1); err != nil {
		return fmt.Errorf("failed to download asset index: %v", err)
	}

	assetIndex, err := d.parseAssetIndex(indexPath)
	if err != nil {
		return fmt.Errorf("failed to parse asset index: %v", err)
	}

	return d.downloadAllAssets(assetIndex)
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

func (d *Downloader) downloadAllAssets(assetIndex *AssetIndex) error {
	jobs := make(chan AssetDownloadJob, len(assetIndex.Objects))
	results := make(chan AssetDownloadResult, len(assetIndex.Objects))

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
			fmt.Printf("Failed to download %s: %v\n", result.Name, result.Error)
			failed++
		} else if result.Skipped {
			skipped++
		} else {
			downloaded++
		}

		total := downloaded + skipped + failed
		if total%100 == 0 {
			fmt.Printf("Progress: %d/%d assets (downloaded: %d, skipped: %d, failed: %d)\n",
				total, len(assetIndex.Objects), downloaded, skipped, failed)
		}
	}

	fmt.Printf("Asset download complete: %d downloaded, %d skipped, %d failed\n", downloaded, skipped, failed)

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

		if err := d.downloadWithChecksum(url, assetPath, job.Hash); err != nil {
			result.Error = err
		}

		results <- result
	}
}
