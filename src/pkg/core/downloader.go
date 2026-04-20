package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"time"
)

// GitHubRelease represents the structure of a release from GitHub API
type GitHubRelease struct {
	Tag         string        `json:"tag_name"`
	Name        string        `json:"name"`
	PublishedAt string        `json:"published_at"`
	HtmlUrl     string        `json:"html_url"`
	Assets      []GitHubAsset `json:"assets"`
}

type GitHubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadUrl string `json:"browser_download_url"`
	Size               int64  `json:"size"`
	ContentType        string `json:"content_type"`
}

type ProtonVariant struct {
	ID          string
	Name        string
	Description string
	RepoOwner   string
	RepoName    string
}

// GetKnownVariants returns the list of supported Proton providers
func GetKnownVariants() []ProtonVariant {
	return []ProtonVariant{
		{
			ID:          "ge-proton",
			Name:        "GE-Proton (GloriousEggroll)",
			Description: "The most popular custom Proton build. Includes many game fixes and codec patches.",
			RepoOwner:   "GloriousEggroll",
			RepoName:    "proton-ge-custom",
		},
		{
			ID:          "proton-cachyos",
			Name:        "Proton-CachyOS",
			Description: "Optimized for performance with CachyOS patches and schedulers.",
			RepoOwner:   "CachyOS",
			RepoName:    "proton-cachyos",
		},
		{
			ID:          "kron4ek",
			Name:        "Proton-Kron4ek",
			Description: "Vanilla builds and TKG builds. Often smaller and faster updates.",
			RepoOwner:   "Kron4ek",
			RepoName:    "Proton-Builds",
		},
		{
			ID:          "luxtorpeda",
			Name:        "Luxtorpeda (Native Tools)",
			Description: "Runs Windows games using native Linux engines (e.g. GZDoom, ScummVM).",
			RepoOwner:   "luxtorpeda-dev",
			RepoName:    "luxtorpeda",
		},
	}
}

// FetchReleases gets the latest releases from a specific GitHub repository
func FetchReleases(variantID string) ([]GitHubRelease, error) {
	variants := GetKnownVariants()
	var selected ProtonVariant
	found := false
	for _, v := range variants {
		if v.ID == variantID {
			selected = v
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("unknown variant: %s", variantID)
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?per_page=50", selected.RepoOwner, selected.RepoName)

	client := http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "LightLauncher-Launcher")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch releases: status %d", resp.StatusCode)
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}
	return releases, nil
}

// InstallProton downloads and extracts a Proton package using system 'tar'
func InstallProton(url string, versionTag string, onProgress func(percentage int, msg string)) error {
	u, err := user.Current()
	if err != nil {
		return err
	}

	// Target Directory: ~/.steam/root/compatibilitytools.d/
	targetBase := filepath.Join(u.HomeDir, ".steam/root/compatibilitytools.d")
	if _, err := os.Stat(filepath.Join(u.HomeDir, ".steam/root")); os.IsNotExist(err) {
		// Fallback to .local/share/Steam
		targetBase = filepath.Join(u.HomeDir, ".local/share/Steam/compatibilitytools.d")
	}

	if err := os.MkdirAll(targetBase, 0755); err != nil {
		return fmt.Errorf("failed to create compatibilitytools.d: %w", err)
	}

	// 1. Download
	onProgress(0, "Downloading...")
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	// Determine extension from URL to create temp file
	ext := ".tar.gz" // default
	if len(filepath.Ext(url)) > 0 {
		// Handle compound extensions like .tar.xz if possible, but basic ext is fine for temp
		// Actually, creating a temp file with the right suffix helps some tools, but tar -xf autodetects.
		// Let's just use the last part of the URL.
		parts := filepath.Base(url)
		if len(parts) > 5 {
			ext = parts // e.g. "Proton-7.0.tar.xz"
		}
	}

	tmpFile, err := os.CreateTemp("", "proton-install-*"+ext)
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)
	defer tmpFile.Close()

	// Download loop
	size := resp.ContentLength
	buffer := make([]byte, 32*1024)
	var downloaded int64

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			_, wErr := tmpFile.Write(buffer[:n])
			if wErr != nil {
				return wErr
			}
			downloaded += int64(n)
			if size > 0 {
				percent := int(float64(downloaded) / float64(size) * 50)
				onProgress(percent, fmt.Sprintf("Downloading... %.0f%%", float64(downloaded)/float64(size)*100))
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	tmpFile.Close() // Close before passing to tar command

	// 2. Extract using system 'tar'
	// This is much more robust than Go's archive/tar for handling .tar.xz, .tar.zst, permissions, and symlinks.
	onProgress(50, "Extracting (using system tools)...")

	// Command: tar -xf <tmpFile> -C <targetBase>
	cmd := exec.Command("tar", "-xf", tmpName, "-C", targetBase)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("extraction failed: %v, output: %s", err, string(output))
	}

	onProgress(100, "Installation Complete!")
	return nil
}
