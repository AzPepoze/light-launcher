package lsfg

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

func Install(onProgress func(int, string)) error {
	onProgress(0, "Fetching release info from GitHub...")
	response, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases", Repo))
	if err != nil {
		return fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return fmt.Errorf("failed to fetch releases: HTTP %d", response.StatusCode)
	}

	var releases []struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(response.Body).Decode(&releases); err != nil {
		return fmt.Errorf("failed to parse release data: %w", err)
	}

	if len(releases) == 0 {
		return fmt.Errorf("no releases found for lsfg-vk")
	}

	var downloadURL, assetName string
	found := false

	for _, release := range releases {
		if found {
			break
		}
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
		return fmt.Errorf("no compatible linux package found")
	}

	onProgress(5, fmt.Sprintf("Downloading %s...", assetName))
	extension := ".tar.xz"
	if strings.HasSuffix(assetName, ".tar.zst") {
		extension = ".tar.zst"
	}
	temporaryFile := filepath.Join(os.TempDir(), "lsfg-vk-dl"+extension)

	err = downloadFileWithProgress(downloadURL, temporaryFile, func(current, total int64) {
		if total > 0 {
			percent := float64(current) / float64(total) * 80.0
			onProgress(5+int(percent), fmt.Sprintf("Downloading... %.1f MB / %.1f MB",
				float64(current)/1024/1024, float64(total)/1024/1024))
		}
	})
	if err != nil {
		os.Remove(temporaryFile)
		return fmt.Errorf("download failed: %w", err)
	}
	defer os.Remove(temporaryFile)

	fileInfo, err := os.Stat(temporaryFile)
	if err != nil || fileInfo.Size() == 0 {
		return fmt.Errorf("downloaded file is empty or not found")
	}

	onProgress(85, "Installing to system directories (requires sudo)...")

	extractCommand := []string{"-xf", temporaryFile, "-C", "/usr"}
	if strings.HasSuffix(temporaryFile, ".tar.zst") {
		extractCommand = []string{"--use-compress-program=unzstd", "-xf", temporaryFile, "-C", "/usr"}
	}

	command := exec.Command("pkexec", append([]string{"tar"}, extractCommand...)...)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("installation failed: %w\nDetails: %s", err, string(output))
	}

	systemLibPath := "/usr/lib/liblsfg-vk-layer.so"
	if _, err := os.Stat(ManifestPath); err != nil {
		return fmt.Errorf("verification failed: manifest not found")
	}
	if _, err := os.Stat(systemLibPath); err != nil {
		return fmt.Errorf("verification failed: library not found")
	}

	onProgress(100, "Installation complete!")
	return nil
}

type progressWriter struct {
	total, current int64
	onProgress     func(current, total int64)
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	bytesRead := len(p)
	pw.current += int64(bytesRead)
	pw.onProgress(pw.current, pw.total)
	return bytesRead, nil
}

func downloadFileWithProgress(url string, destination string, onProgress func(current, total int64)) error {
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", "LightLauncher-App")

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %s", response.Status)
	}

	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer file.Close()

	pw := &progressWriter{total: response.ContentLength, onProgress: onProgress}
	_, err = io.Copy(file, io.TeeReader(response.Body, pw))
	return err
}
