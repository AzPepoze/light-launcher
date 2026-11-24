package drives

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"light-launcher/core/internal/store"
)

const (
	setupVersion = 1
	// appliedFileName lists the letters the launcher last linked into a prefix.
	appliedFileName = "light-launcher-drives.json"
)

// SetupEntry is one drive letter that the launcher sets up for every prefix.
type SetupEntry struct {
	Letter string `json:"letter"`
	Path   string `json:"path"`
}

// setupFile is the saved drive setup. The store only writes JSON objects, so the list sits inside one.
type setupFile struct {
	Entries []SetupEntry `json:"entries"`
}

// appliedFile lists the letters the launcher last linked into one prefix.
type appliedFile struct {
	Letters []string `json:"letters"`
}

// LoadSetup reads the saved drive setup. A missing file means no drives are set up.
func LoadSetup(path string) ([]SetupEntry, error) {
	var file setupFile
	if _, err := store.NewFile(path, setupVersion).Load(&file); err != nil {
		return nil, err
	}
	if file.Entries == nil {
		return []SetupEntry{}, nil
	}
	return file.Entries, nil
}

// SaveSetup checks the entries and writes the drive setup.
func SaveSetup(path string, entries []SetupEntry) error {
	checked, err := checkSetup(entries)
	if err != nil {
		return err
	}
	return store.NewFile(path, setupVersion).Save(setupFile{Entries: checked})
}

// checkSetup returns the entries with letters and folders checked. A letter may appear only once.
func checkSetup(entries []SetupEntry) ([]SetupEntry, error) {
	checked := make([]SetupEntry, 0, len(entries))
	seen := make(map[string]bool)
	for _, entry := range entries {
		letter, err := normalizeLetter(entry.Letter)
		if err != nil {
			return nil, err
		}
		if seen[letter] {
			return nil, ErrLetterTaken
		}
		seen[letter] = true
		target, err := resolveFolder(entry.Path)
		if err != nil {
			return nil, err
		}
		checked = append(checked, SetupEntry{Letter: letter, Path: target})
	}
	return checked, nil
}

// ApplySetup links the saved drives into one prefix. Letters from the last apply that the setup no longer
// lists are released. C: goes back to Wine's own drive when the launcher had set it.
func ApplySetup(prefixPath string, entries []SetupEntry) error {
	dosdevices := filepath.Join(prefixPath, "dosdevices")
	if err := os.MkdirAll(dosdevices, 0755); err != nil {
		return err
	}
	applied := store.NewFile(filepath.Join(prefixPath, appliedFileName), setupVersion)
	var previous appliedFile
	if _, err := applied.Load(&previous); err != nil {
		return err
	}

	wanted := make(map[string]bool, len(entries))
	for _, entry := range entries {
		wanted[entry.Letter] = true
	}

	var errs []error
	for _, letter := range previous.Letters {
		if !wanted[letter] {
			errs = append(errs, releaseLetter(dosdevices, letter))
		}
	}
	letters := make([]string, 0, len(entries))
	for _, entry := range entries {
		errs = append(errs, forceLink(dosdevices, entry.Path, entry.Letter))
		letters = append(letters, entry.Letter)
	}
	errs = append(errs, applied.Save(appliedFile{Letters: letters}))
	return errors.Join(errs...)
}

// releaseLetter undoes a letter the launcher set up earlier. C: returns to Wine's own drive.
func releaseLetter(dosdevices string, letter string) error {
	if letter == systemLetter {
		return forceLink(dosdevices, wineSystemTarget, systemLetter)
	}
	return removeLink(dosdevices, letter)
}

// forceLink points letter at target, replacing a link that is already there.
func forceLink(dosdevices string, target string, letter string) error {
	if err := removeLink(dosdevices, letter); err != nil {
		return err
	}
	return os.Symlink(target, filepath.Join(dosdevices, letter))
}

// removeLink deletes a drive link. A real folder with a drive letter name is never removed.
func removeLink(dosdevices string, letter string) error {
	link := filepath.Join(dosdevices, letter)
	info, err := os.Lstat(link)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if info.Mode()&fs.ModeSymlink == 0 {
		return fmt.Errorf("%s is a folder, not a drive link", letter)
	}
	return os.Remove(link)
}
