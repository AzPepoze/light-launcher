package downloads

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"light-launcher/core/internal/contract"
)

// ValidateSubfolder checks a relative folder path. It allows only plain names joined by "/".
// An empty path is valid and means the root itself.
func ValidateSubfolder(sub string) error {
	if sub == "" {
		return nil
	}
	if strings.ContainsAny(sub, `\:`) {
		return badFolder("use / between names, and no drive letters")
	}
	if strings.HasPrefix(sub, "/") {
		return badFolder("the folder must be relative")
	}
	for _, name := range strings.Split(sub, "/") {
		if err := checkFolderName(name); err != nil {
			return err
		}
	}
	return nil
}

// ValidateRequest checks the folder, root and install fields of a download request.
func ValidateRequest(req contract.DownloadRequest) error {
	if req.Root != "" && req.Root != contract.RootDownloads && req.Root != contract.RootGames {
		return badFolder(`root must be "downloads" or "games"`)
	}
	if err := ValidateSubfolder(req.Subfolder); err != nil {
		return err
	}
	if err := ValidateSubfolder(req.InstallSubfolder); err != nil {
		return err
	}
	return checkInstallOptions(req)
}

// checkInstallOptions rejects install settings that have no install to apply to.
func checkInstallOptions(req contract.DownloadRequest) error {
	if req.Install {
		return ValidateSubfolder(req.InstallSubfolder)
	}
	if req.InstallSubfolder != "" || req.DeleteArchive != nil {
		return fmt.Errorf("%w: installSubfolder and deleteArchive need install set to true", contract.ErrBadRequest)
	}
	return nil
}

func checkFolderName(name string) error {
	switch {
	case name == "" || name == "." || name == "..":
		return badFolder("folder names cannot be empty, . or ..")
	case strings.IndexFunc(name, unicode.IsControl) >= 0:
		return badFolder("folder names cannot contain control characters")
	}
	return nil
}

// resolveFolder returns root joined with sub and creates the folders that are missing.
// It refuses any segment that is a link or a file, so the result always stays inside root.
func resolveFolder(root, sub string) (string, error) {
	if err := ValidateSubfolder(sub); err != nil {
		return "", err
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}
	dir := root
	if sub == "" {
		return dir, nil
	}
	for _, name := range strings.Split(sub, "/") {
		dir = filepath.Join(dir, name)
		if err := ensureRealDir(dir); err != nil {
			return "", err
		}
	}
	return dir, nil
}

// ensureRealDir creates dir when it is missing. An existing link or file at that path is an error.
func ensureRealDir(dir string) error {
	info, err := os.Lstat(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return os.Mkdir(dir, 0o755)
	}
	if err != nil {
		return err
	}
	if info.Mode()&fs.ModeSymlink != 0 || !info.IsDir() {
		return badFolder(fmt.Sprintf("%q is a link or a file, not a folder", filepath.Base(dir)))
	}
	return nil
}

func badFolder(detail string) error {
	return fmt.Errorf("%w: %s", contract.ErrBadFolder, detail)
}
