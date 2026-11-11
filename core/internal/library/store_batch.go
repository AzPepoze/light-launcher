package library

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"light-launcher/core/internal/contract"
)

// batchResult says what a batch of plugin games did.
type batchResult struct {
	// IDs holds the library id of each input game, in the same order.
	IDs []int64
	// Written counts rows inserted or changed.
	Written int
	// Skipped counts rows already stored with the same updatedAt and version.
	Skipped int
	// CoverIDs lists written rows whose cover must be cached.
	CoverIDs []int64
}

// gameKey identifies a plugin game: its source and its id at that source.
type gameKey struct {
	source   string
	sourceID string
}

// lookupChunk limits how many games one lookup names, to stay under the SQLite parameter limit.
const lookupChunk = 500

// saveGameSQL writes one game as a new row, or refreshes the row that has the same source id.
const saveGameSQL = `insert into games (source, source_id, url, title, version, developer, engine, tags, cover_url,
	cover_path, screenshots, description, changelog, rating, status, source_updated_at, extra, added_at, updated_at)
	values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
	on conflict(source, source_id) do update set url=excluded.url, title=excluded.title, version=excluded.version,
	developer=excluded.developer, engine=excluded.engine, tags=excluded.tags, cover_url=excluded.cover_url,
	cover_path=excluded.cover_path, screenshots=excluded.screenshots, description=excluded.description,
	changelog=excluded.changelog, rating=excluded.rating, status=excluded.status,
	source_updated_at=excluded.source_updated_at, extra=excluded.extra, updated_at=excluded.updated_at
	returning id`

// saveBatch saves plugin games in one transaction, with one prepared statement. Games stored with the
// same updatedAt and version are skipped. Hidden and favorite are never changed here.
// The database has one connection, so nothing inside the transaction may call other Store methods.
func (s *Store) saveBatch(ctx context.Context, games []contract.GameDTO) (batchResult, error) {
	result := batchResult{IDs: make([]int64, len(games))}
	if len(games) == 0 {
		return result, nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return batchResult{}, fmt.Errorf("begin game batch: %w", err)
	}
	defer tx.Rollback()

	stored, err := lookupGames(ctx, tx, games)
	if err != nil {
		return batchResult{}, err
	}
	save, err := tx.PrepareContext(ctx, saveGameSQL)
	if err != nil {
		return batchResult{}, fmt.Errorf("prepare game save: %w", err)
	}
	defer save.Close()

	writer := batchWriter{save: save, stored: stored, now: nowStamp(), result: &result}
	last := lastIndexes(games)
	for i, dto := range games {
		if last[gameKeyOf(dto)] != i {
			continue
		}
		id, err := writer.write(ctx, dto)
		if err != nil {
			return batchResult{}, err
		}
		result.IDs[i] = id
	}
	for i, dto := range games {
		result.IDs[i] = result.IDs[last[gameKeyOf(dto)]]
	}
	if err := tx.Commit(); err != nil {
		return batchResult{}, fmt.Errorf("commit game batch: %w", err)
	}
	if result.Written > 0 {
		s.markWritten()
	}
	return result, nil
}

// batchWriter writes the games of one batch inside its transaction.
type batchWriter struct {
	save   *sql.Stmt
	stored map[gameKey]Game
	now    string
	result *batchResult
}

// write saves one game and returns its library id.
func (w *batchWriter) write(ctx context.Context, dto contract.GameDTO) (int64, error) {
	existing, found := w.stored[gameKeyOf(dto)]
	if found && unchanged(dto, existing) {
		w.result.Skipped++
		return existing.ID, nil
	}
	args, err := saveArgs(dto, existing, w.now)
	if err != nil {
		return 0, err
	}
	var id int64
	if err := w.save.QueryRowContext(ctx, args...).Scan(&id); err != nil {
		return 0, fmt.Errorf("save game %q: %w", dto.Title, err)
	}
	w.result.Written++
	if coverChanged(dto, existing) {
		w.result.CoverIDs = append(w.result.CoverIDs, id)
	}
	return id, nil
}

// lookupGames loads the stored rows for the games of a batch, keyed by source and source id.
func lookupGames(ctx context.Context, tx *sql.Tx, games []contract.GameDTO) (map[gameKey]Game, error) {
	stored := map[gameKey]Game{}
	for start := 0; start < len(games); start += lookupChunk {
		end := min(start+lookupChunk, len(games))
		if err := lookupChunkInto(ctx, tx, games[start:end], stored); err != nil {
			return nil, err
		}
	}
	return stored, nil
}

func lookupChunkInto(ctx context.Context, tx *sql.Tx, games []contract.GameDTO, stored map[gameKey]Game) error {
	args := make([]any, 0, 2*len(games))
	for _, dto := range games {
		args = append(args, dto.Source, dto.SourceID)
	}
	query := `select ` + gameColumns + ` from games where (source, source_id) in (values ` + pairPlaceholders(len(games)) + `)`
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("look up games: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		game, err := scanGame(rows)
		if err != nil {
			return err
		}
		stored[gameKey{game.Source, game.SourceID}] = game
	}
	return rows.Err()
}

// pairPlaceholders returns n "(?,?)" groups separated by commas.
func pairPlaceholders(n int) string {
	return strings.TrimSuffix(strings.Repeat("(?,?),", n), ",")
}

// lastIndexes maps each game key to the position of its last occurrence in the batch.
func lastIndexes(games []contract.GameDTO) map[gameKey]int {
	last := make(map[gameKey]int, len(games))
	for i, dto := range games {
		last[gameKeyOf(dto)] = i
	}
	return last
}

func gameKeyOf(dto contract.GameDTO) gameKey {
	return gameKey{source: dto.Source, sourceID: dto.SourceID}
}

// unchanged reports whether the stored row already has the source's updatedAt and version.
func unchanged(dto contract.GameDTO, stored Game) bool {
	return dto.UpdatedAt != "" && dto.UpdatedAt == derefString(stored.SourceUpdatedAt) &&
		dto.Version == derefString(stored.Version)
}

// coverChanged reports whether the plugin sent a cover that the row does not have yet.
func coverChanged(dto contract.GameDTO, stored Game) bool {
	return dto.CoverURL != "" && (stored.CoverURL == nil || *stored.CoverURL != dto.CoverURL)
}

// saveArgs returns the values of saveGameSQL. Empty fields keep the stored value; a new row starts from empty.
func saveArgs(dto contract.GameDTO, stored Game, now string) ([]any, error) {
	tags, screenshots, extra, err := mergeGameLists(dto, stored)
	if err != nil {
		return nil, err
	}
	coverPath := nullablePtr(stored.CoverPath)
	if coverChanged(dto, stored) {
		coverPath = nil
	}
	return []any{
		dto.Source, dto.SourceID, dto.URL,
		firstNonEmpty(dto.Title, stored.Title),
		pickString(dto.Version, stored.Version),
		pickString(dto.Developer, stored.Developer),
		pickString(dto.Engine, stored.Engine),
		tags,
		pickString(dto.CoverURL, stored.CoverURL),
		coverPath,
		screenshots,
		pickString(dto.DescriptionHTML, stored.Description),
		pickString(dto.Changelog, stored.Changelog),
		pickRating(dto.Rating, stored.Rating),
		statusOr(firstNonEmpty(dto.Status, stored.Status), "unknown"),
		pickString(dto.UpdatedAt, stored.SourceUpdatedAt),
		extra,
		firstNonEmpty(stored.AddedAt, now),
		now,
	}, nil
}
