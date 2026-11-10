package runtimesvc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	lsfgReleasesURL   = "https://api.github.com/repos/Pumboo/lsfg-vk/releases"
	lsfgManifestPath  = "/usr/share/vulkan/implicit_layer.d/lsfg-vk.json"
	lsfgLibraryPath   = "/usr/lib/liblsfg-vk-layer.so"
	lsfgAuthFailedMsg = "Authentication failed or was cancelled during installation."
)

type lsfgRelease struct {
	Assets []lsfgAsset `json:"assets"`
}

type lsfgAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// installLsfg downloads the newest lsfg-vk Linux package and unpacks it into /usr with pkexec.
func installLsfg(onProgress func(percent int, message string)) error {
	onProgress(0, "Fetching release info from GitHub...")
	releases, err := fetchLsfgReleases()
	if err != nil {
		return err
	}
	if len(releases) == 0 {
		return errors.New("No releases found for lsfg-vk")
	}

	downloadURL, assetName := findLsfgAsset(releases)
	if downloadURL == "" {
		return errors.New("No compatible Linux package found for lsfg-vk")
	}

	onProgress(10, fmt.Sprintf("Downloading %s...", assetName))
	tempFile := filepath.Join(os.TempDir(), "lsfg-vk-dl"+lsfgArchiveExtension(assetName))
	defer removeIfExists(tempFile)
	if err := downloadFile(downloadURL, tempFile, nil); err != nil {
		return err
	}

	onProgress(85, "Installing to system directories (requires sudo)...")
	if err := extractLsfgArchive(tempFile); err != nil {
		return errors.New(lsfgAuthFailedMsg)
	}
	onProgress(100, "Installation complete!")
	return nil
}

func fetchLsfgReleases() ([]lsfgRelease, error) {
	request, err := http.NewRequest(http.MethodGet, lsfgReleasesURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", userAgent)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %d", response.StatusCode)
	}
	var releases []lsfgRelease
	if err := json.NewDecoder(response.Body).Decode(&releases); err != nil {
		return nil, err
	}
	return releases, nil
}

// findLsfgAsset returns the first x86_64 .tar.zst or linux .tar.xz asset across all releases.
func findLsfgAsset(releases []lsfgRelease) (string, string) {
	for _, release := range releases {
		for _, asset := range release.Assets {
			name := strings.ToLower(asset.Name)
			isZst := strings.Contains(name, "x86_64") && strings.HasSuffix(name, ".tar.zst")
			isXz := strings.Contains(name, "linux") && strings.HasSuffix(name, ".tar.xz")
			if isZst || isXz {
				return asset.BrowserDownloadURL, asset.Name
			}
		}
	}
	return "", ""
}

func lsfgArchiveExtension(assetName string) string {
	if strings.HasSuffix(assetName, ".tar.zst") {
		return ".tar.zst"
	}
	return ".tar.xz"
}

// extractLsfgArchive unpacks the archive into /usr as root, asking for authentication through pkexec.
func extractLsfgArchive(archive string) error {
	extractCommand := fmt.Sprintf(`tar -xf "%s" -C /usr`, archive)
	if strings.HasSuffix(archive, ".tar.zst") {
		extractCommand = fmt.Sprintf(`tar --use-compress-program=unzstd -xf "%s" -C /usr`, archive)
	}
	return exec.Command("pkexec", "sh", "-c", extractCommand).Run()
}

// uninstallLsfg removes the lsfg-vk layer files from the system with pkexec.
func uninstallLsfg() error {
	err := exec.Command("pkexec", "rm", "-f", lsfgManifestPath, lsfgLibraryPath).Run()
	if err != nil {
		return errors.New("Authentication failed or was cancelled during uninstallation.")
	}
	return nil
}
