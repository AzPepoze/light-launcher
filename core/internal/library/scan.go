package library

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/logger"
)

// ScanReport counts what a folder scan did.
type ScanReport struct {
	Folders int `json:"folders"`
	Added   int `json:"added"`
	Skipped int `json:"skipped"`
}

// manifestDir and manifestFile mark folders that describe their own game.
const (
	manifestDir  = ".interesting-manager"
	manifestFile = "manifest.json"
)

// notAGame are folder names that hold downloads or caches, never a game.
var notAGame = map[string]bool{"downloads": true, "download": true, "cache": true, "trash": true, "__macosx": true}

var (
	underscores    = regexp.MustCompile(`_+`)
	windowsWord    = regexp.MustCompile(`(?i)[-_ ]*windows?\b`)
	spaces         = regexp.MustCompile(`\s+`)
	trialSuffix    = regexp.MustCompile(`(?i)\s*\((?:patreon|trial|demo)\)\s*$`)
	platformSuffix = regexp.MustCompile(`(?i)[-_ ]*(?:win|x64|64bit|64-bit)$`)
)

// ScanFolder imports the game folders under root that are not yet in the library.
func (l *Library) ScanFolder(ctx context.Context, root string) (ScanReport, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return ScanReport{}, fmt.Errorf("read folder: %w", err)
	}
	report := ScanReport{}
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		report.Folders++
		dir := filepath.Join(root, entry.Name())
		added, err := l.importFolder(ctx, dir)
		if err != nil {
			logger.Debug(logCategory, "scan failed for %s: %v", dir, err)
			report.Skipped++
			continue
		}
		if added {
			report.Added++
		} else {
			report.Skipped++
		}
	}
	logger.Info(logCategory, "scan: %d folders, %d added, %d skipped", report.Folders, report.Added, report.Skipped)
	return report, nil
}

// importFolder imports one folder. It reports false when the folder was skipped.
func (l *Library) importFolder(ctx context.Context, dir string) (bool, error) {
	manifest := filepath.Join(dir, manifestDir, manifestFile)
	if _, err := os.Stat(manifest); err == nil {
		return l.importManifest(ctx, dir, manifest)
	}
	return l.importUnknown(ctx, dir)
}

// manifestData is the content of an install manifest written by the old launcher.
type manifestData struct {
	Source      string `json:"source"`
	SourceID    string `json:"sourceId"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	GameVersion string `json:"gameVersion"`
	Developer   string `json:"developer"`
}

func (l *Library) importManifest(ctx context.Context, dir, file string) (bool, error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return false, err
	}
	var data manifestData
	if err := json.Unmarshal(raw, &data); err != nil {
		return false, err
	}
	if data.Source == "" {
		data.Source = "manual"
	}
	if data.SourceID == "" {
		data.SourceID = filepath.Base(dir)
	}
	if data.URL == "" {
		data.URL = "file://" + dir
	}
	if data.Title == "" {
		data.Title = filepath.Base(dir)
	}
	dto := contract.GameDTO{
		Source:    data.Source,
		SourceID:  data.SourceID,
		URL:       data.URL,
		Title:     data.Title,
		Version:   data.GameVersion,
		Developer: data.Developer,
		Status:    "completed",
		Extra:     map[string]any{"imported": true},
	}
	return l.addInstalled(ctx, dir, dto, data.GameVersion)
}

func (l *Library) importUnknown(ctx context.Context, dir string) (bool, error) {
	base := filepath.Base(dir)
	if notAGame[strings.ToLower(base)] || strings.HasPrefix(base, ".") {
		return false, nil
	}
	sum := sha1.Sum([]byte(dir))
	sourceID := hex.EncodeToString(sum[:])[:16]
	if _, err := l.BySource(ctx, "manual", sourceID); err == nil {
		return false, nil
	}
	if DetectExecutable(dir) == "" {
		return false, nil
	}
	version := versionFromText(base)
	dto := contract.GameDTO{
		Source:   "manual",
		SourceID: sourceID,
		URL:      "file://" + dir,
		Title:    prettyTitle(base),
		Version:  version,
		Status:   "completed",
		Extra:    map[string]any{"localPath": dir},
	}
	return l.addInstalled(ctx, dir, dto, version)
}

// addInstalled saves a game and points it at an install folder. Folders already owned by a row are skipped.
func (l *Library) addInstalled(ctx context.Context, dir string, dto contract.GameDTO, installedVersion string) (bool, error) {
	if _, err := l.ByInstallDir(ctx, dir); err == nil {
		return false, nil
	}
	existing, err := l.BySource(ctx, dto.Source, dto.SourceID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return false, err
	}
	if err == nil && existing.InstallDir != nil {
		return false, nil
	}
	id, err := l.Upsert(ctx, dto)
	if err != nil {
		return false, err
	}
	size := dirSize(dir)
	exe := DetectExecutable(dir)
	patch := Patch{InstallDir: &dir, InstallSize: &size, InstalledVersion: stringOrNil(installedVersion)}
	if exe != "" {
		patch.ExePath = &exe
	}
	if _, err := l.Patch(ctx, id, patch); err != nil {
		return false, err
	}
	l.notify("library.game", map[string]any{"id": id})
	return true, nil
}

// prettyTitle turns a folder name into a title: "Mad_Island_0.34" becomes "Mad Island 0.34".
func prettyTitle(folder string) string {
	title := underscores.ReplaceAllString(folder, " ")
	title = windowsWord.ReplaceAllString(title, " ")
	title = strings.TrimSpace(spaces.ReplaceAllString(title, " "))
	title = strings.TrimSpace(trialSuffix.ReplaceAllString(title, ""))
	title = strings.TrimSpace(platformSuffix.ReplaceAllString(title, ""))
	if title == "" {
		return folder
	}
	return title
}
