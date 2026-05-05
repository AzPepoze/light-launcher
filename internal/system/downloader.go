package system

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
	ID          string `json:"ID"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	RepoOwner   string `json:"RepoOwner"`
	RepoName    string `json:"RepoName"`
}

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

func FetchReleases(variantID string) ([]GitHubRelease, error) {
	variants := GetKnownVariants()
	var selected ProtonVariant
	found := false
	for _, variant := range variants {
		if variant.ID == variantID {
			selected = variant
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("unknown variant: %s", variantID)
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?per_page=50", selected.RepoOwner, selected.RepoName)

	client := http.Client{Timeout: 15 * time.Second}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "LightLauncher-App")

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch releases: status %d", response.StatusCode)
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(response.Body).Decode(&releases); err != nil {
		return nil, err
	}
	return releases, nil
}

func InstallProton(url string, versionTag string, onProgress func(percentage int, message string)) error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	targetBase := filepath.Join(currentUser.HomeDir, ".steam/root/compatibilitytools.d")
	if _, err := os.Stat(filepath.Join(currentUser.HomeDir, ".steam/root")); os.IsNotExist(err) {
		targetBase = filepath.Join(currentUser.HomeDir, ".local/share/Steam/compatibilitytools.d")
	}

	if err := os.MkdirAll(targetBase, 0755); err != nil {
		return fmt.Errorf("failed to create compatibilitytools.d: %w", err)
	}

	onProgress(0, "Downloading...")
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", response.Status)
	}

	extension := ".tar.gz"
	if parts := filepath.Base(url); len(parts) > 5 {
		extension = parts
	}

	temporaryFile, err := os.CreateTemp("", "proton-install-*"+extension)
	if err != nil {
		return err
	}
	temporaryName := temporaryFile.Name()
	defer os.Remove(temporaryName)
	defer temporaryFile.Close()

	size := response.ContentLength
	buffer := make([]byte, 32*1024)
	var downloaded int64

	for {
		bytesRead, readErr := response.Body.Read(buffer)
		if bytesRead > 0 {
			_, writeErr := temporaryFile.Write(buffer[:bytesRead])
			if writeErr != nil {
				return writeErr
			}
			downloaded += int64(bytesRead)
			if size > 0 {
				percent := int(float64(downloaded) / float64(size) * 50)
				onProgress(percent, fmt.Sprintf("Downloading... %.0f%%", float64(downloaded)/float64(size)*100))
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	temporaryFile.Close()

	onProgress(50, "Extracting...")

	command := exec.Command("tar", "-xf", temporaryName, "-C", targetBase)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("extraction failed: %v, output: %s", err, string(output))
	}

	onProgress(100, "Installation Complete!")
	return nil
}
