package library

import (
	"database/sql"
	"encoding/json"
	"time"
)

// rowScanner is satisfied by both *sql.Row and *sql.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

const gameColumns = `id, source, source_id, url, title, version, developer, engine, tags, cover_path, cover_url,
	screenshots, description, changelog, rating, status, favorite, hidden, notes, installed_version, install_dir,
	install_size, exe_path, exe_args, last_played_at, playtime_s, source_updated_at, last_checked_at, added_at,
	updated_at, extra`

func scanGame(row rowScanner) (Game, error) {
	var (
		g                                                 Game
		version, developer, engine, coverPath, coverURL   sql.NullString
		tags, screenshots, description, changelog         sql.NullString
		rating                                            sql.NullFloat64
		notes, installedVersion, installDir, exePath      sql.NullString
		exeArgs, lastPlayedAt, sourceUpdatedAt, lastCheck sql.NullString
		installSize                                       sql.NullInt64
		extra                                             sql.NullString
	)
	err := row.Scan(
		&g.ID, &g.Source, &g.SourceID, &g.URL, &g.Title, &version, &developer, &engine, &tags,
		&coverPath, &coverURL, &screenshots, &description, &changelog, &rating, &g.Status,
		&g.Favorite, &g.Hidden, &notes, &installedVersion, &installDir, &installSize, &exePath,
		&exeArgs, &lastPlayedAt, &g.PlaytimeS, &sourceUpdatedAt, &lastCheck, &g.AddedAt, &g.UpdatedAt, &extra,
	)
	if err != nil {
		return Game{}, err
	}
	g.Version = stringPtr(version)
	g.Developer = stringPtr(developer)
	g.Engine = stringPtr(engine)
	g.CoverPath = stringPtr(coverPath)
	g.CoverURL = stringPtr(coverURL)
	g.Description = stringPtr(description)
	g.Changelog = stringPtr(changelog)
	g.Notes = stringPtr(notes)
	g.InstalledVersion = stringPtr(installedVersion)
	g.InstallDir = stringPtr(installDir)
	g.ExePath = stringPtr(exePath)
	g.ExeArgs = stringPtr(exeArgs)
	g.LastPlayedAt = stringPtr(lastPlayedAt)
	g.SourceUpdatedAt = stringPtr(sourceUpdatedAt)
	g.LastCheckedAt = stringPtr(lastCheck)
	if rating.Valid {
		g.Rating = &rating.Float64
	}
	if installSize.Valid {
		g.InstallSize = &installSize.Int64
	}
	g.Tags = decodeStrings(tags.String)
	g.Screenshots = decodeStrings(screenshots.String)
	g.Extra = decodeExtra(extra)
	return g, nil
}

func scanFile(row rowScanner) (GameFile, error) {
	var (
		f               GameFile
		filename, host  sql.NullString
		size, rangeHint sql.NullInt64
	)
	err := row.Scan(&f.ID, &f.GameID, &f.Label, &f.URL, &filename, &size, &host, &f.Kind, &f.Platform,
		&rangeHint, &f.RequiresAuth)
	if err != nil {
		return GameFile{}, err
	}
	f.Filename = stringPtr(filename)
	f.Host = stringPtr(host)
	if size.Valid {
		f.Size = &size.Int64
	}
	if rangeHint.Valid {
		hint := int(rangeHint.Int64)
		f.RangeHint = &hint
	}
	return f, nil
}

func stringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

// nullIfEmpty stores an empty string as NULL, the way the old API treated missing values.
func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

// pickString keeps the previous value when the source sent nothing new.
func pickString(next string, previous *string) any {
	if next != "" {
		return next
	}
	if previous == nil {
		return nil
	}
	return *previous
}

// pickRating keeps the previous rating when the source sent none.
func pickRating(next *float64, previous *float64) any {
	if next != nil {
		return *next
	}
	if previous == nil {
		return nil
	}
	return *previous
}

func decodeStrings(raw string) []string {
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil || out == nil {
		return []string{}
	}
	return out
}

func decodeExtra(raw sql.NullString) map[string]any {
	if !raw.Valid || raw.String == "" {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw.String), &out); err != nil {
		return nil
	}
	return out
}

func encodeJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	return string(data), err
}

// stamp formats time like JavaScript's toISOString, which the old database used.
func stamp(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

func nowStamp() string {
	return stamp(time.Now())
}
