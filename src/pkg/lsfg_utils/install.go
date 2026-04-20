package lsfg_utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Install downloads and installs LSFG-VK
func Install(onProgress func(int, string)) error {
	onProgress(0, "Fetching release info from GitHub...")
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases", Repo))
	if err != nil {
		return fmt.Errorf("failed to fetch releases: network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to fetch releases: HTTP %d", resp.StatusCode)
	}

	var releases []struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return fmt.Errorf("failed to parse release data: %w", err)
	}

	if len(releases) == 0 {
		return fmt.Errorf("no releases found for lsfg-vk")
	}

	var downloadURL, assetName string
	found := false

	// Search through releases (newest first)
	for _, release := range releases {
		if found {
			break
		}
		// Look for compatible assets - prefer x86_64 tar.zst, then linux tar.xz
		for _, asset := range release.Assets {
			name := strings.ToLower(asset.Name)
			if (strings.Contains(name, "x86_64") && strings.HasSuffix(name, ".tar.zst")) ||
				(strings.Contains(name, "linux") && strings.HasSuffix(name, ".tar.xz")) {
				downloadURL = asset.BrowserDownloadURL
				assetName = asset.Name
				found = true
				break
			}
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no compatible linux package found in releases")
	}

	onProgress(5, fmt.Sprintf("Downloading %s...", assetName))
	ext := ".tar.xz"
	if strings.HasSuffix(assetName, ".tar.zst") {
		ext = ".tar.zst"
	}
	tmpFile := filepath.Join(os.TempDir(), "lsfg-vk-dl"+ext)

	err = downloadFileWithProgress(downloadURL, tmpFile, func(current, total int64) {
		if total > 0 {
			percent := float64(current) / float64(total) * 80.0
			onProgress(5+int(percent), fmt.Sprintf("Downloading... %.1f MB / %.1f MB",
				float64(current)/1024/1024, float64(total)/1024/1024))
		}
	})
	if err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("download failed: %w", err)
	}
	defer os.Remove(tmpFile)

	// Verify downloaded file exists and has content
	fileInfo, err := os.Stat(tmpFile)
	if err != nil || fileInfo.Size() == 0 {
		return fmt.Errorf("downloaded file is empty or not found")
	}

	onProgress(85, "Installing to system directories (requires sudo)...")

	// Combine extraction and installation into one command using pkexec tar directly to /usr
	// This satisfies the "combine command" and "just move to /usr" request efficiently
	extractCmd := []string{"-xf", tmpFile, "-C", "/usr"}
	if strings.HasSuffix(tmpFile, ".tar.zst") {
		extractCmd = []string{"--use-compress-program=unzstd", "-xf", tmpFile, "-C", "/usr"}
	}

	cmd := exec.Command("pkexec", append([]string{"tar"}, extractCmd...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("installation failed: %w\nDetails: %s", err, string(output))
	}

	// Verify installation - check if files were actually installed
	systemManifestPath := ManifestPath
	systemLibPath := "/usr/lib/liblsfg-vk-layer.so"
	if _, err := os.Stat(systemManifestPath); err != nil {
		return fmt.Errorf("installation verification failed: manifest not found in /usr/share")
	}
	if _, err := os.Stat(systemLibPath); err != nil {
		return fmt.Errorf("installation verification failed: library not found in /usr/lib")
	}

	onProgress(100, "Installation complete!")
	return nil
}

type progressWriter struct {
	total, current int64
	onProgress     func(current, total int64)
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.current += int64(n)
	pw.onProgress(pw.current, pw.total)
	return n, nil
}

func downloadFileWithProgress(url string, dest string, onProgress func(current, total int64)) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "LightLauncher-App")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	pw := &progressWriter{total: resp.ContentLength, onProgress: onProgress}
	_, err = io.Copy(out, io.TeeReader(resp.Body, pw))
	return err
}
