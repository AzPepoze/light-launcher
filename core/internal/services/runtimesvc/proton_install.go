package runtimesvc

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"light-launcher/core/internal/config"
)

// installProtonVersion downloads a Proton archive and extracts it into the Steam compatibility tools folder.
// Progress goes to onProgress: download 0-50%, extract 50%, done 100%.
func installProtonVersion(url string, versionTag string, onProgress func(percent int, message string)) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	targetBase := protonTargetFolder(home)
	if err := os.MkdirAll(targetBase, 0755); err != nil {
		return err
	}
	onProgress(0, "Downloading...")

	tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("proton-install-%d%s", time.Now().UnixMilli(), archiveExtension(url)))
	defer removeIfExists(tempFile)

	if err := downloadFile(url, tempFile, protonDownloadProgress(onProgress)); err != nil {
		return installFailed(versionTag, err)
	}

	onProgress(50, "Extracting Proton package...")
	if err := extractTarArchive(tempFile, targetBase); err != nil {
		return installFailed(versionTag, err)
	}
	onProgress(100, "Installation Complete!")
	return nil
}

// protonTargetFolder prefers ~/.steam when it exists, else the Steam share folder.
func protonTargetFolder(home string) string {
	if config.PathExists(filepath.Join(home, ".steam/root")) {
		return filepath.Join(home, ".steam/root/compatibilitytools.d")
	}
	return filepath.Join(home, ".local/share/Steam/compatibilitytools.d")
}

func archiveExtension(url string) string {
	if strings.HasSuffix(url, ".tar.zst") {
		return ".tar.zst"
	}
	return ".tar.gz"
}

// protonDownloadProgress maps download bytes to the first half of the progress bar.
func protonDownloadProgress(onProgress func(percent int, message string)) func(current, total int64) {
	return func(current, total int64) {
		if total <= 0 {
			return
		}
		percent := int(math.Round(float64(current) / float64(total) * 50))
		message := fmt.Sprintf("Downloading... %.1f MB / %.1f MB", megabytes(current), megabytes(total))
		onProgress(percent, message)
	}
}

func megabytes(bytes int64) float64 {
	return float64(bytes) / 1024 / 1024
}

// extractTarArchive unpacks a .tar.gz or .tar.zst archive into destination.
func extractTarArchive(archive string, destination string) error {
	arguments := []string{"-xf", archive, "-C", destination}
	if strings.HasSuffix(archive, ".tar.zst") {
		arguments = append([]string{"--use-compress-program=unzstd"}, arguments...)
	}
	output, err := exec.Command("tar", arguments...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func installFailed(versionTag string, err error) error {
	return fmt.Errorf("Failed to install Proton %s: %v", versionTag, err)
}
