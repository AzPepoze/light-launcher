package library

import (
	"context"
	"database/sql"
	"errors"
	"sync/atomic"

	"light-launcher/core/internal/contract"
)

// Store is typed CRUD over the games table.
type Store struct {
	db *sql.DB
	// writes counts the writes to the games table, so cached summaries know when they are stale.
	writes atomic.Uint64
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// Get returns one game by id, or ErrNotFound.
func (s *Store) Get(ctx context.Context, id int64) (Game, error) {
	return s.queryOne(ctx, `select `+gameColumns+` from games where id = ?`, id)
}

// BySource finds the row for a plugin source and its id, or ErrNotFound.
func (s *Store) BySource(ctx context.Context, source, sourceID string) (Game, error) {
	return s.queryOne(ctx, `select `+gameColumns+` from games where source = ? and source_id = ?`, source, sourceID)
}

// ByInstallDir finds the row that owns an install folder, or ErrNotFound.
func (s *Store) ByInstallDir(ctx context.Context, dir string) (Game, error) {
	return s.queryOne(ctx, `select `+gameColumns+` from games where install_dir = ? limit 1`, dir)
}

// All returns every game, newest update first.
func (s *Store) All(ctx context.Context) ([]Game, error) {
	return s.queryMany(ctx, `select `+gameColumns+` from games order by updated_at desc`)
}

// Upsert saves a plugin game, creating or refreshing the row for its source id.
func (s *Store) Upsert(ctx context.Context, dto contract.GameDTO) (int64, error) {
	batch, err := s.saveBatch(ctx, []contract.GameDTO{dto})
	if err != nil {
		return 0, err
	}
	return batch.IDs[0], nil
}

// Relink moves a row onto a real source identity and keeps its install data.
func (s *Store) Relink(ctx context.Context, id int64, dto contract.GameDTO) error {
	tags, screenshots, extra, err := encodeGameLists(dto)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `update games set source=?, source_id=?, url=?, title=?, version=?, developer=?,
		engine=?, tags=?, cover_url=?, cover_path=null, screenshots=?, description=?, changelog=?, rating=?, status=?,
		source_updated_at=?, extra=?, updated_at=? where id=?`,
		dto.Source, dto.SourceID, dto.URL, dto.Title, nullIfEmpty(dto.Version), nullIfEmpty(dto.Developer),
		nullIfEmpty(dto.Engine), tags, nullIfEmpty(dto.CoverURL), screenshots, nullIfEmpty(dto.DescriptionHTML),
		nullIfEmpty(dto.Changelog), dto.Rating, statusOr(dto.Status, "completed"), nullIfEmpty(dto.UpdatedAt), extra,
		nowStamp(), id)
	if err != nil {
		return err
	}
	s.markWritten()
	return nil
}

// Remove deletes a row. Files and checks go with it through foreign keys.
func (s *Store) Remove(ctx context.Context, id int64) error {
	if _, err := s.db.ExecContext(ctx, `delete from games where id = ?`, id); err != nil {
		return err
	}
	s.markWritten()
	return nil
}

// markWritten records a change to the games table.
func (s *Store) markWritten() {
	s.writes.Add(1)
}

// writeCount returns how many writes to the games table this store has made.
func (s *Store) writeCount() uint64 {
	return s.writes.Load()
}

func (s *Store) queryOne(ctx context.Context, query string, args ...any) (Game, error) {
	game, err := scanGame(s.db.QueryRowContext(ctx, query, args...))
	if errors.Is(err, sql.ErrNoRows) {
		return Game{}, ErrNotFound
	}
	return game, err
}

func (s *Store) queryMany(ctx context.Context, query string, args ...any) ([]Game, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	games := []Game{}
	for rows.Next() {
		game, err := scanGame(rows)
		if err != nil {
			return nil, err
		}
		games = append(games, game)
	}
	return games, rows.Err()
}

// encodeGameLists turns a plugin game's list fields into the JSON text the columns hold.
func encodeGameLists(dto contract.GameDTO) (tags, screenshots, extra any, err error) {
	return encodeLists(dto.Tags, dto.Screenshots, dto.Extra)
}

func mergeGameLists(dto contract.GameDTO, existing Game) (tags, screenshots, extra any, err error) {
	tagList := dto.Tags
	if len(tagList) == 0 {
		tagList = existing.Tags
	}
	shotList := dto.Screenshots
	if len(shotList) == 0 {
		shotList = existing.Screenshots
	}
	extraMap := dto.Extra
	if extraMap == nil {
		extraMap = existing.Extra
	}
	return encodeLists(tagList, shotList, extraMap)
}

func encodeLists(tagList, shotList []string, extraMap map[string]any) (tags, screenshots, extra any, err error) {
	if tagList == nil {
		tagList = []string{}
	}
	if shotList == nil {
		shotList = []string{}
	}
	tagJSON, err := encodeJSON(tagList)
	if err != nil {
		return nil, nil, nil, err
	}
	shotJSON, err := encodeJSON(shotList)
	if err != nil {
		return nil, nil, nil, err
	}
	var extraJSON any
	if extraMap != nil {
		raw, err := encodeJSON(extraMap)
		if err != nil {
			return nil, nil, nil, err
		}
		extraJSON = raw
	}
	return tagJSON, shotJSON, extraJSON, nil
}

func nullablePtr(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func statusOr(status, fallback string) string {
	if status == "" {
		return fallback
	}
	return status
}
