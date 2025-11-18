package mc

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const ConcurrentDownloads = 10

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

func (v *VersionManager) DownloadAssets(assetIndexURL, assetIndexSHA1 string) error {
	assetsDir := filepath.Join(v.GameDir, "assets")
	indexPath := filepath.Join(assetsDir, "indexes", "5.json") // 1.21.4

	if err := v.DownloadFileWithChecksum(assetIndexURL, indexPath, assetIndexSHA1); err != nil {
		return fmt.Errorf("failed to download asset index: %v", err)
	}

	assetIndex, err := v.parseAssetIndex(indexPath)
	if err != nil {
		return fmt.Errorf("failed to parse asset index: %v", err)
	}

	return v.downloadAllAssets(assetIndex)
}

func (v *VersionManager) parseAssetIndex(indexPath string) (*AssetIndex, error) {
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

func (v *VersionManager) downloadAllAssets(assetIndex *AssetIndex) error {
	jobs := make(chan AssetDownloadJob, len(assetIndex.Objects))
	results := make(chan AssetDownloadResult, len(assetIndex.Objects))

	var wg sync.WaitGroup
	for i := 0; i < ConcurrentDownloads; i++ {
		wg.Add(1)
		go func() {
			v.assetDownloadWorker(jobs, results, &wg)
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
			fmt.Fprintf(v.Stdout, "Progress: %d/%d assets (downloaded: %d, skipped: %d, failed: %d)\n",
				total, len(assetIndex.Objects), downloaded, skipped, failed)
		}
	}

	fmt.Fprintf(v.Stdout, "Asset download complete: %d downloaded, %d skipped, %d failed\n", downloaded, skipped, failed)

	if failed > 0 {
		return fmt.Errorf("%d assets failed to download", failed)
	}

	return nil
}

func (v *VersionManager) assetDownloadWorker(jobs <-chan AssetDownloadJob, results chan<- AssetDownloadResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		result := AssetDownloadResult{Name: job.Name}

		url := fmt.Sprintf("https://resources.download.minecraft.net/%s/%s", job.Hash[:2], job.Hash)
		assetPath := filepath.Join(v.GameDir, "assets", "objects", job.Hash[:2], job.Hash)

		if info, err := os.Stat(assetPath); err == nil {
			if info.Size() == int64(job.Size) {
				if err := v.verifyChecksum(assetPath, job.Hash); err == nil {
					result.Skipped = true
					results <- result
					continue
				}
			}
		}

		if err := v.DownloadFileWithChecksum(url, assetPath, job.Hash); err != nil {
			result.Error = err
		}

		results <- result
	}
}
