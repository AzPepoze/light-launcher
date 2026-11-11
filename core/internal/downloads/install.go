package downloads

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/logger"
)

// maxNestedArchives limits how many wrapper archives are opened one inside another.
const maxNestedArchives = 4

// TitleLookup returns the title of a game, which names its folder in the library.
type TitleLookup func(ctx context.Context, gameID int64) (string, error)

// FolderInstaller unpacks a finished archive into its own folder in the library and finds the launcher.
type FolderInstaller struct {
	Root   string      // library folder; every game gets a folder here
	Events api.Events  // optional; receives install progress and warnings
	Titles TitleLookup // optional; without it the folder is named after the archive

	running runningInstalls
}

var _ Installer = (*FolderInstaller)(nil)

// ActiveInstalls returns the installs that are running now, with their latest progress.
func (in *FolderInstaller) ActiveInstalls() []InstallProgress {
	return in.running.list()
}

// installResult is what a finished install leaves behind.
type installResult struct {
	dir        string
	executable string
	size       int64
}

// Install unpacks the archive, finds the launcher and moves the game into the library.
func (in *FolderInstaller) Install(ctx context.Context, job InstallJob) error {
	rep := newInstallReporter(in.Events, &in.running, job.DownloadID, job.GameID)
	result, err := in.install(ctx, job, rep)
	if err != nil {
		logger.Warn(logCategory, "install of download %d failed: %v", job.DownloadID, err)
		rep.fail(err)
		return err
	}
	rep.done(result)
	logger.Info(logCategory, "installed download %d into %s", job.DownloadID, filepath.Base(result.dir))
	return nil
}

func (in *FolderInstaller) install(ctx context.Context, job InstallJob, rep *installReporter) (installResult, error) {
	if !fileExists(job.Archive) {
		return installResult{}, fmt.Errorf("archive is missing: %s", filepath.Base(job.Archive))
	}
	if err := os.MkdirAll(in.Root, 0o755); err != nil {
		return installResult{}, err
	}
	parent, err := resolveFolder(in.Root, job.Subfolder)
	if err != nil {
		return installResult{}, err
	}
	staging := filepath.Join(in.Root, fmt.Sprintf(".staging-%d-%d", job.DownloadID, time.Now().UnixNano()))
	defer os.RemoveAll(staging)
	if err := os.MkdirAll(staging, 0o755); err != nil {
		return installResult{}, err
	}

	title := in.titleOf(ctx, job.GameID)
	rep.titled(title)
	rep.phase("extracting")
	in.warnIfSevenZipMissing(job.Archive)
	err = Extract(ctx, job.Archive, staging, ExtractOptions{
		OnProgress: rep.progress,
		OnEngine:   rep.engine,
	})
	if err != nil {
		return installResult{}, err
	}

	rep.phase("unpacking")
	if err := peelNested(ctx, staging); err != nil {
		return installResult{}, err
	}
	if err := cleanTree(staging); err != nil {
		return installResult{}, err
	}

	rep.phase("finalizing")
	root := singleFolder(staging)
	exe, err := detectExecutable(root)
	if err != nil {
		return installResult{}, err
	}
	target, err := in.targetFolder(parent, job.GameID, job.Archive, title)
	if err != nil {
		return installResult{}, err
	}
	if err := os.Rename(root, target); err != nil {
		return installResult{}, err
	}
	if err := writeManifest(target, job.DownloadID, job.GameID, job.Archive); err != nil {
		return installResult{}, err
	}
	result := installResult{dir: target, size: treeSize(target)}
	if exe != "" {
		rel, err := filepath.Rel(root, exe)
		if err != nil {
			return installResult{}, err
		}
		result.executable = filepath.Join(target, rel)
	}
	return result, nil
}

// warnIfSevenZipMissing tells the user what the slower or failing path means, once per install.
func (in *FolderInstaller) warnIfSevenZipMissing(archivePath string) {
	if HasSevenZip() || in.Events == nil {
		return
	}
	switch detectArchiveKind(archivePath) {
	case KindSevenZip:
		in.Events.Emit(EventToast, Toast{Level: "warn", Message: "Install 7-Zip to unpack .7z files."})
	case KindRar:
		in.Events.Emit(EventToast, Toast{Level: "warn", Message: "7-Zip is not installed. RAR files use the slower built-in reader."})
	}
}

// titleOf returns the game's title from the lookup, or "" when there is none.
func (in *FolderInstaller) titleOf(ctx context.Context, gameID int64) string {
	if in.Titles == nil {
		return ""
	}
	title, err := in.Titles(ctx, gameID)
	if err != nil {
		return ""
	}
	return title
}

// targetFolder picks a free folder in parent for the game. The title names it, or else the archive does.
func (in *FolderInstaller) targetFolder(parent string, gameID int64, archivePath, title string) (string, error) {
	name := strings.TrimSuffix(filepath.Base(archivePath), filepath.Ext(archivePath))
	if title != "" {
		name = title
	}
	base := folderName(name, gameID)
	for n := 1; ; n++ {
		candidate := filepath.Join(parent, base)
		if n > 1 {
			candidate = filepath.Join(parent, fmt.Sprintf("%s (%d)", base, n))
		}
		if !fileExists(candidate) {
			return candidate, nil
		}
	}
}

var (
	bracketTags       = regexp.MustCompile(`\[[^\]]*\]|\([^)]*\)`)
	illegalFolderRune = regexp.MustCompile(`[<>:"/\\|?*]`)
)

// folderName keeps the readable part of a title, without [tags] or (tags).
func folderName(title string, gameID int64) string {
	name := bracketTags.ReplaceAllString(title, " ")
	name = illegalFolderRune.ReplaceAllString(name, " ")
	name = strings.Join(strings.Fields(name), " ")
	name = strings.TrimRight(name, ". ")
	if name == "" {
		return fmt.Sprintf("game-%d", gameID)
	}
	return name
}

// peelNested unwraps archives that only wrap the game, such as a zip inside a zip.
func peelNested(ctx context.Context, dir string) error {
	for level := 0; level < maxNestedArchives; level++ {
		only := singleFile(dir)
		if only == "" || detectArchiveKind(only) == KindUnknown {
			return nil
		}
		if err := Extract(ctx, only, dir, ExtractOptions{}); err != nil {
			return err
		}
		if err := os.Remove(only); err != nil {
			return err
		}
		if err := cleanTree(dir); err != nil {
			return err
		}
	}
	return nil
}

// singleFile returns the path of the only visible entry in dir when it is a regular file, or "".
func singleFile(dir string) string {
	entries := visibleEntries(dir)
	if len(entries) != 1 || !entries[0].Type().IsRegular() {
		return ""
	}
	return filepath.Join(dir, entries[0].Name())
}

// singleFolder returns the one folder that holds the whole game, or dir when there is none.
func singleFolder(dir string) string {
	entries := visibleEntries(dir)
	if len(entries) == 1 && entries[0].IsDir() {
		return filepath.Join(dir, entries[0].Name())
	}
	return dir
}

// visibleEntries lists dir without OS junk.
func visibleEntries(dir string) []os.DirEntry {
	all, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []os.DirEntry
	for _, e := range all {
		if !isJunk(e.Name()) {
			out = append(out, e)
		}
	}
	return out
}

// writeManifest records where the game came from. It stores only the archive's file name, not its folder.
func writeManifest(dir string, downloadID, gameID int64, archivePath string) error {
	metaDir := filepath.Join(dir, ".light-launcher")
	if err := os.MkdirAll(metaDir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(map[string]any{
		"version":     1,
		"downloadId":  downloadID,
		"gameId":      gameID,
		"archive":     filepath.Base(archivePath),
		"installedAt": time.Now().UTC().Format(time.RFC3339),
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(metaDir, "manifest.json"), data, 0o644)
}
