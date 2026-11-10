package library

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"light-launcher/core/internal/contract"
)

// Incoming is the list of watched files that need a person to decide what they are.

const (
	ignoredKey       = "watch:ignored"
	maxIgnored       = 500
	fileSettledAge   = 3 * time.Second
	suggestionLimit  = 3
	downloadsWarning = "The watch folder is the downloads folder. Pick a different folder in settings."
)

// IncomingSuggestion is a game a file might belong to.
type IncomingSuggestion struct {
	GameID     int64   `json:"gameId"`
	Title      string  `json:"title"`
	Cover      *string `json:"cover"`
	Version    *string `json:"version"`
	Installed  bool    `json:"installed"`
	Confidence float64 `json:"confidence"`
}

// IncomingItem is one watched file.
type IncomingItem struct {
	Name        string               `json:"name"`
	Size        int64                `json:"size"`
	MTime       string               `json:"mtime"`
	Ready       bool                 `json:"ready"`
	Ignored     bool                 `json:"ignored"`
	Auto        bool                 `json:"auto"`
	Suggestions []IncomingSuggestion `json:"suggestions"`
}

// IncomingReport is the Incoming page. Count is the number of files that need a decision.
type IncomingReport struct {
	Configured bool           `json:"configured"`
	Folder     string         `json:"folder"`
	Items      []IncomingItem `json:"items"`
	Count      int            `json:"count"`
	Warning    string         `json:"warning,omitempty"`
}

// ignoredFile is an entry in the ignore list. Keys match the old meta value.
type ignoredFile struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	MTimeMs int64  `json:"mtimeMs"`
}

// Incoming lists the watched files with their suggestions.
func (l *Library) Incoming(ctx context.Context) (IncomingReport, error) {
	folder := l.watchFolder()
	if folder == "" {
		return IncomingReport{}, nil
	}
	if sameDir(folder, l.opts.DownloadsDir) {
		return IncomingReport{Configured: true, Folder: folder, Warning: downloadsWarning}, nil
	}
	files, err := listWatchable(folder)
	if err != nil {
		return IncomingReport{}, err
	}
	games, err := l.All(ctx)
	if err != nil {
		return IncomingReport{}, err
	}
	ignored, err := l.ignoredNames(ctx)
	if err != nil {
		return IncomingReport{}, err
	}
	titles := gameTitles(games)

	report := IncomingReport{Configured: true, Folder: folder, Items: []IncomingItem{}}
	for _, file := range files {
		_, match := bestConfidentMatch(file.Name, titles)
		item := IncomingItem{
			Name:        file.Name,
			Size:        file.Size,
			MTime:       file.ModTime.UTC().Format(time.RFC3339),
			Ready:       time.Since(file.ModTime) > fileSettledAge,
			Ignored:     ignored[file.Name],
			Auto:        match,
			Suggestions: suggestionsFor(file.Name, games, titles),
		}
		if !item.Ignored && !item.Auto {
			report.Count++
		}
		report.Items = append(report.Items, item)
	}
	sort.SliceStable(report.Items, func(i, j int) bool { return report.Items[i].MTime > report.Items[j].MTime })
	return report, nil
}

func suggestionsFor(name string, games []Game, titles []string) []IncomingSuggestion {
	ranked := rankTitles(name, titles, suggestionLimit)
	suggestions := make([]IncomingSuggestion, 0, len(ranked))
	for _, hit := range ranked {
		game := games[hit.Index]
		suggestions = append(suggestions, IncomingSuggestion{
			GameID:     game.ID,
			Title:      game.Title,
			Cover:      coverOf(game),
			Version:    game.Version,
			Installed:  game.InstallDir != nil,
			Confidence: math.Round(hit.Score*100) / 100,
		})
	}
	return suggestions
}

func coverOf(game Game) *string {
	if game.CoverPath != nil {
		return game.CoverPath
	}
	return game.CoverURL
}

// IncomingLink imports a watched file for a game.
func (l *Library) IncomingLink(ctx context.Context, name string, gameID int64) error {
	game, err := l.Get(ctx, gameID)
	if err != nil {
		return err
	}
	folder, err := l.requireWatchFolder()
	if err != nil {
		return err
	}
	info, err := watchedFileInfo(folder, name)
	if err != nil {
		return err
	}
	if time.Since(info.ModTime()) < fileSettledAge {
		return errors.New("this file is still being written; try again in a moment")
	}
	return l.importWatched(ctx, folder, name, game.ID)
}

// IncomingIgnore hides a watched file from the Incoming list and from automatic import.
func (l *Library) IncomingIgnore(ctx context.Context, name string) error {
	folder, err := l.requireWatchFolder()
	if err != nil {
		return err
	}
	info, err := watchedFileInfo(folder, name)
	if err != nil {
		return err
	}
	rows, err := l.ignoredList(ctx)
	if err != nil {
		return err
	}
	rows = withoutName(rows, name)
	rows = append(rows, ignoredFile{Name: name, Size: info.Size(), MTimeMs: info.ModTime().UnixMilli()})
	if len(rows) > maxIgnored {
		rows = rows[len(rows)-maxIgnored:]
	}
	if err := l.setMetaJSON(ctx, ignoredKey, rows); err != nil {
		return err
	}
	l.notify("library.incoming", map[string]any{})
	return nil
}

// IncomingCreateLocal makes a local game from a watched file name, without importing the file.
func (l *Library) IncomingCreateLocal(ctx context.Context, name string) (Game, error) {
	if err := validFileName(name); err != nil {
		return Game{}, err
	}
	title := prettyTitle(strings.TrimSuffix(name, filepath.Ext(name)))
	sum := sha1.Sum([]byte("watch:" + name))
	id, err := l.Upsert(ctx, contract.GameDTO{
		Source:   "manual",
		SourceID: hex.EncodeToString(sum[:])[:16],
		URL:      "file://" + name,
		Title:    title,
		Status:   "completed",
		Extra:    map[string]any{"fromWatch": true},
	})
	if err != nil {
		return Game{}, err
	}
	return l.Get(ctx, id)
}

func (l *Library) requireWatchFolder() (string, error) {
	folder := l.watchFolder()
	if folder == "" {
		return "", errors.New("no watch folder is configured")
	}
	return folder, nil
}

// watchedFileInfo checks that name is a plain file inside the watch folder and returns its info.
func watchedFileInfo(folder, name string) (os.FileInfo, error) {
	if err := validFileName(name); err != nil {
		return nil, err
	}
	info, err := os.Lstat(filepath.Join(folder, name))
	if err != nil {
		return nil, errors.New("file no longer exists in the watch folder")
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("symbolic links are not accepted")
	}
	return info, nil
}

// validFileName rejects anything that is not a bare file name, so a request cannot leave the folder.
func validFileName(name string) error {
	if name == "" || name == "." || name == ".." || name != filepath.Base(name) || strings.ContainsAny(name, `/\`) {
		return errors.New("invalid file name")
	}
	return nil
}

// ignoredList reads the ignore list from the meta table.
func (l *Library) ignoredList(ctx context.Context) ([]ignoredFile, error) {
	var rows []ignoredFile
	if err := l.getMetaJSON(ctx, ignoredKey, &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

// ignoredNames returns the set of ignored file names.
func (l *Library) ignoredNames(ctx context.Context) (map[string]bool, error) {
	rows, err := l.ignoredList(ctx)
	if err != nil {
		return nil, err
	}
	names := make(map[string]bool, len(rows))
	for _, row := range rows {
		names[row.Name] = true
	}
	return names, nil
}

// restoreIncoming removes a file from the ignore list once it has been imported.
func (l *Library) restoreIncoming(ctx context.Context, name string) error {
	rows, err := l.ignoredList(ctx)
	if err != nil {
		return err
	}
	kept := withoutName(rows, name)
	if len(kept) == len(rows) {
		return nil
	}
	return l.setMetaJSON(ctx, ignoredKey, kept)
}

func withoutName(rows []ignoredFile, name string) []ignoredFile {
	kept := make([]ignoredFile, 0, len(rows))
	for _, row := range rows {
		if row.Name != name {
			kept = append(kept, row)
		}
	}
	return kept
}
