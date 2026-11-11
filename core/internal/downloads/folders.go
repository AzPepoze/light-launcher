package downloads

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"light-launcher/core/internal/contract"
)

const (
	// maxNameBytes is the Windows file name limit, so a folder works on every drive.
	maxNameBytes = 255
	// maxSubfolderBytes leaves room for the root and the file name inside the folder.
	maxSubfolderBytes = 240
)

const illegalNameChars = `<>:"|?*\`

// reservedWindowsNames are device names that cannot be a folder or file name on Windows.
var reservedWindowsNames = map[string]bool{
	"CON": true, "PRN": true, "AUX": true, "NUL": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true,
	"COM6": true, "COM7": true, "COM8": true, "COM9": true,
	"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true,
	"LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
	"COM¹": true, "COM²": true, "COM³": true,
	"LPT¹": true, "LPT²": true, "LPT³": true,
}

// ValidateSubfolder checks a relative folder path. It allows only plain names joined by "/".
// An empty path is valid and means the root itself.
func ValidateSubfolder(sub string) error {
	if sub == "" {
		return nil
	}
	if len(sub) > maxSubfolderBytes {
		return badFolder(fmt.Sprintf("the folder path cannot be longer than %d bytes", maxSubfolderBytes))
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

// ValidateName checks one folder or file name against the rules of every supported OS.
func ValidateName(name string) error {
	return checkFolderName(name)
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
	if i := strings.IndexFunc(name, func(r rune) bool { return strings.ContainsRune(illegalNameChars, r) }); i >= 0 {
		return badFolder(fmt.Sprintf("folder names cannot contain %q", name[i]))
	}
	if strings.HasSuffix(name, ".") || strings.HasSuffix(name, " ") {
		return badFolder("folder names cannot end with a dot or a space")
	}
	if len(name) > maxNameBytes {
		return badFolder(fmt.Sprintf("folder names cannot be longer than %d bytes", maxNameBytes))
	}
	if stem := reservedStem(name); reservedWindowsNames[stem] {
		return badFolder(fmt.Sprintf("folder names cannot be the reserved Windows name %q", stem))
	}
	return nil
}

// reservedStem returns the part Windows checks for a device name, without an extension.
func reservedStem(name string) string {
	if i := strings.IndexByte(name, '.'); i >= 0 {
		name = name[:i]
	}
	return strings.ToUpper(strings.TrimRight(name, " ."))
}

// truncateName shortens a name to maxNameBytes without splitting a UTF-8 character.
func truncateName(name string) string {
	return truncateBytes(name, maxNameBytes)
}

func truncateBytes(name string, max int) string {
	if len(name) <= max {
		return name
	}
	cut := max
	for cut > 0 && !utf8.RuneStart(name[cut]) {
		cut--
	}
	return name[:cut]
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
