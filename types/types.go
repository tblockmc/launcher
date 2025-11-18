package types

type Rule struct {
	Action   string          `json:"action"`
	OS       *OSRule         `json:"os,omitempty"`
	Features map[string]bool `json:"features,omitempty"`
}

type OSRule struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Arch    string `json:"arch,omitempty"`
}

// Version manifest from Mojang
type VersionManifest struct {
	Latest   LatestVersions `json:"latest"`
	Versions []Version      `json:"versions"`
}

type LatestVersions struct {
	Release  string `json:"release"`
	Snapshot string `json:"snapshot"`
}

type Version struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	Time        string `json:"time"`
	ReleaseTime string `json:"releaseTime"`
}

type VersionDetails struct {
	ID           string     `json:"id"`
	InheritsFrom string     `json:"inheritsFrom"`
	MainClass    string     `json:"mainClass"`
	Libraries    []Library  `json:"libraries"`
	Downloads    Downloads  `json:"downloads"`
	AssetIndex   AssetIndex `json:"assetIndex"`
}

type Library struct {
	Name      string           `json:"name"`
	Downloads LibraryDownloads `json:"downloads"`
	Rules     []Rule           `json:"rules,omitempty"`
}

type LibraryDownloads struct {
	Artifact Artifact `json:"artifact"`
}

type Artifact struct {
	Path string `json:"path"`
	URL  string `json:"url"`
	SHA1 string `json:"sha1"`
	Size int    `json:"size"`
}

type Downloads struct {
	Client Artifact `json:"client"`
	Server Artifact `json:"server"`
}

type AssetIndex struct {
	ID        string `json:"id"`
	SHA1      string `json:"sha1"`
	Size      int    `json:"size"`
	TotalSize int    `json:"totalSize"`
	URL       string `json:"url"`
}
