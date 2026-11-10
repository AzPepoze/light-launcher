package appsvc

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// isDirectory reports whether the path exists and is a folder.
func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// imageDataURL reads an image as a data URL, or "" when it is missing or unreadable.
func imageDataURL(imagePath string) string {
	if _, err := os.Stat(imagePath); err != nil {
		return ""
	}
	data, err := os.ReadFile(imagePath)
	if err != nil || len(data) == 0 {
		return ""
	}
	return dataURL(imageMimeType(extName(imagePath)), data)
}

func imageMimeType(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".svg":
		return "image/svg+xml"
	case ".webp":
		return "image/webp"
	case ".ico":
		return "image/x-icon"
	default:
		return "image/png"
	}
}

// openURL opens a link in the default browser. An empty link does nothing.
func openURL(url string) error {
	if url == "" {
		return nil
	}
	return openExternalURL(url)
}

// openFileLocation opens a folder, or shows a file inside its folder.
func openFileLocation(targetPath string) error {
	if targetPath == "" {
		return errors.New("No path provided")
	}
	info, err := os.Stat(targetPath)
	if err != nil {
		return fmt.Errorf("Path does not exist: %s", targetPath)
	}
	if info.IsDir() {
		return openPathInShell(targetPath)
	}
	return revealInFolder(targetPath)
}
