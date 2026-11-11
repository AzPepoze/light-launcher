package library

import "errors"

// ErrNotFound is returned when a row does not exist.
var ErrNotFound = errors.New("not found")

// Game is one library row. Nullable columns are pointers so they reach the UI as null.
// Favorite, Hidden and RequiresAuth stay 0 or 1 to match the old API.
type Game struct {
	ID               int64          `json:"id"`
	Source           string         `json:"source"`
	SourceID         string         `json:"sourceId"`
	URL              string         `json:"url"`
	Title            string         `json:"title"`
	Version          *string        `json:"version"`
	Developer        *string        `json:"developer"`
	Engine           *string        `json:"engine"`
	Tags             []string       `json:"tags"`
	CoverPath        *string        `json:"coverPath"`
	CoverURL         *string        `json:"coverUrl"`
	CoverLocal       *string        `json:"coverLocal"`
	Screenshots      []string       `json:"screenshots"`
	ScreenshotsLocal []*string      `json:"screenshotsLocal"`
	Description      *string        `json:"description"`
	DescriptionText  *string        `json:"descriptionText"`
	Changelog        *string        `json:"changelog"`
	Rating           *float64       `json:"rating"`
	Status           string         `json:"status"`
	Running          bool           `json:"running"`
	Favorite         int            `json:"favorite"`
	Hidden           int            `json:"hidden"`
	Notes            *string        `json:"notes"`
	InstalledVersion *string        `json:"installedVersion"`
	InstallDir       *string        `json:"installDir"`
	InstallSize      *int64         `json:"installSize"`
	ExePath          *string        `json:"exePath"`
	ExeArgs          *string        `json:"exeArgs"`
	LastPlayedAt     *string        `json:"lastPlayedAt"`
	PlaytimeS        int64          `json:"playtimeS"`
	SourceUpdatedAt  *string        `json:"sourceUpdatedAt"`
	LastCheckedAt    *string        `json:"lastCheckedAt"`
	AddedAt          string         `json:"addedAt"`
	UpdatedAt        string         `json:"updatedAt"`
	Extra            map[string]any `json:"extra"`
}

// GameFile is one download candidate stored for a game.
type GameFile struct {
	ID           int64   `json:"id"`
	GameID       int64   `json:"gameId"`
	Label        string  `json:"label"`
	URL          string  `json:"url"`
	Filename     *string `json:"filename"`
	Size         *int64  `json:"size"`
	Host         *string `json:"host"`
	Kind         string  `json:"kind"`
	Platform     string  `json:"platform"`
	RangeHint    *int    `json:"rangeHint"`
	RequiresAuth int     `json:"requiresAuth"`
}

// DeletedGame is a history entry kept after a game is removed from the library.
type DeletedGame struct {
	ID           int64    `json:"id"`
	GameID       int64    `json:"gameId"`
	Source       string   `json:"source"`
	SourceID     *string  `json:"sourceId"`
	URL          string   `json:"url"`
	Title        string   `json:"title"`
	Version      *string  `json:"version"`
	Developer    *string  `json:"developer"`
	Tags         []string `json:"tags"`
	InstallDir   *string  `json:"installDir"`
	DeletedFiles bool     `json:"deletedFiles"`
	DeletedAt    string   `json:"deletedAt"`
}

// Check is one recorded update: the version seen before and after.
type Check struct {
	ID         int64   `json:"id"`
	GameID     int64   `json:"gameId"`
	CheckedAt  string  `json:"checkedAt"`
	OldVersion *string `json:"oldVersion"`
	NewVersion *string `json:"newVersion"`
	Changelog  *string `json:"changelog"`
	Seen       int     `json:"seen"`
	Applied    int     `json:"applied"`
}
