package library

import (
	"context"
	"strings"
)

// Patch changes only the fields that are set. A nil field is left alone.
// For text fields an empty string clears the value to NULL.
type Patch struct {
	Title            *string
	URL              *string
	SourceID         *string
	Version          *string
	Notes            *string
	Status           *string
	Tags             *[]string
	Favorite         *bool
	Hidden           *bool
	ExePath          *string
	ExeArgs          *string
	InstallDir       *string
	InstallSize      *int64
	InstalledVersion *string
	PlaytimeS        *int64
	LastPlayedAt     *string
	SourceUpdatedAt  *string
	LastCheckedAt    *string
	CoverPath        *string
	CoverURL         *string
}

// setClause builds the SET part of an UPDATE from the fields that are present.
type setClause struct {
	parts []string
	args  []any
}

func (c *setClause) add(column string, value any) {
	c.parts = append(c.parts, column+" = ?")
	c.args = append(c.args, value)
}

func (c *setClause) addText(column string, value *string) {
	if value != nil {
		c.add(column, nullIfEmpty(*value))
	}
}

func (c *setClause) addInt(column string, value *int64) {
	if value != nil {
		c.add(column, *value)
	}
}

func (c *setClause) addBool(column string, value *bool) {
	if value == nil {
		return
	}
	if *value {
		c.add(column, 1)
	} else {
		c.add(column, 0)
	}
}

func (p Patch) clause() (setClause, error) {
	var c setClause
	c.addText("title", p.Title)
	c.addText("url", p.URL)
	c.addText("source_id", p.SourceID)
	c.addText("version", p.Version)
	c.addText("notes", p.Notes)
	c.addText("status", p.Status)
	c.addText("exe_path", p.ExePath)
	c.addText("exe_args", p.ExeArgs)
	c.addText("install_dir", p.InstallDir)
	c.addText("installed_version", p.InstalledVersion)
	c.addText("last_played_at", p.LastPlayedAt)
	c.addText("source_updated_at", p.SourceUpdatedAt)
	c.addText("last_checked_at", p.LastCheckedAt)
	c.addText("cover_path", p.CoverPath)
	c.addText("cover_url", p.CoverURL)
	c.addInt("install_size", p.InstallSize)
	c.addInt("playtime_s", p.PlaytimeS)
	c.addBool("favorite", p.Favorite)
	c.addBool("hidden", p.Hidden)
	if p.Tags != nil {
		tags, err := encodeJSON(*p.Tags)
		if err != nil {
			return setClause{}, err
		}
		c.add("tags", tags)
	}
	return c, nil
}

// Patch updates the given fields of one game and returns the stored row.
func (s *Store) Patch(ctx context.Context, id int64, p Patch) (Game, error) {
	c, err := p.clause()
	if err != nil {
		return Game{}, err
	}
	if len(c.parts) == 0 {
		return s.Get(ctx, id)
	}
	c.add("updated_at", nowStamp())
	c.args = append(c.args, id)
	query := "update games set " + strings.Join(c.parts, ", ") + " where id = ?"
	if _, err := s.db.ExecContext(ctx, query, c.args...); err != nil {
		return Game{}, err
	}
	return s.Get(ctx, id)
}
