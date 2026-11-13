package library

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"light-launcher/core/internal/contract"
)

// importPageSize is how many games one source.listGames page asks for.
const importPageSize = 100

// Capabilities and method of a source that syncs its own games instead of listing them.
// A source declares one of the two sync capabilities; source.sync answers with games or a status.
const (
	capabilitySyncGames = "sync:games"
	capabilitySyncIndex = "sync:index"
	methodSourceSync    = "source.sync"
)

// listGamesScopeMine is the only scope of source.listGames: the games the user owns, follows or bookmarked.
const listGamesScopeMine = "mine"

var (
	// ErrNoListGames means the source can neither list nor sync the games of the user.
	ErrNoListGames = errors.New("source cannot list its games")
	// ErrImportRunning means a sync of this source is already running.
	ErrImportRunning = errors.New("a sync of this source is already running")
)

// ImportResult counts what an import did. Imported is games written (new or changed), Skipped games
// that were already up to date and Failed the pages that could not be fetched or saved.
type ImportResult struct {
	Imported int `json:"imported"`
	Skipped  int `json:"skipped"`
	Failed   int `json:"failed"`
}

// ImportProgress is the result so far. It is passed to the progress callback after each page.
type ImportProgress = ImportResult

type listGamesParams struct {
	Source string `json:"source"`
	Scope  string `json:"scope"`
	Cursor string `json:"cursor,omitempty"`
	Limit  int    `json:"limit"`
}

type listGamesReply struct {
	Games      []contract.GameDTO `json:"games"`
	NextCursor string             `json:"nextCursor,omitempty"`
}

// importer runs imports in the background and keeps one import per source at a time.
type importer struct {
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	mu      sync.Mutex
	running map[string]bool
}

func newImporter() *importer {
	ctx, cancel := context.WithCancel(context.Background())
	return &importer{ctx: ctx, cancel: cancel, running: map[string]bool{}}
}

// claim marks a source as importing. The returned func releases the source.
func (i *importer) claim(source string) (func(), error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.running[source] {
		return nil, ErrImportRunning
	}
	i.running[source] = true
	return func() {
		i.mu.Lock()
		defer i.mu.Unlock()
		delete(i.running, source)
	}, nil
}

// stop cancels the running imports and waits until they end.
func (i *importer) stop() {
	i.cancel()
	i.wg.Wait()
}

// ImportSource saves every game a source lists for the user, one page at a time, each page in its own
// transaction. It stops at the first page that fails; the pages saved before it stay saved.
func (l *Library) ImportSource(ctx context.Context, plugin, source string, progress func(ImportProgress)) (ImportResult, error) {
	release, err := l.imports.claim(source)
	if err != nil {
		return ImportResult{}, err
	}
	defer release()
	return l.importPages(ctx, plugin, source, progress)
}

// SyncSource fills the library from one source on the caller's goroutine, the way StartImport does in
// the background. A source that lists its games is still imported page by page; a source that only
// syncs runs source.sync and saves the games it returns. It is what scheduled sync jobs call, so a
// scheduled run stores games exactly like a manual sync. It reports ErrNoListGames when the source
// is unknown or cannot fill the library, and ErrImportRunning while another import of it is running.
func (l *Library) SyncSource(ctx context.Context, plugin, source string) (ImportResult, error) {
	info, ok := l.pluginFor(source)
	if !ok || info.Plugin != plugin || !canImport(info) {
		return ImportResult{}, ErrNoListGames
	}
	release, err := l.imports.claim(source)
	if err != nil {
		return ImportResult{}, err
	}
	defer release()
	return l.importSource(ctx, info, nil)
}

// StartImport runs an import of a source in the background and returns at once. A source with
// listGames is read page by page; a source with sync:games or sync:index runs source.sync instead.
// progress is called after each page, or once after a sync, and done once at the end with the final
// result and the error that ended the import.
func (l *Library) StartImport(sourceID string, progress func(ImportProgress), done func(ImportResult, error)) error {
	info, ok := l.pluginFor(sourceID)
	if !ok || !canImport(info) {
		return ErrNoListGames
	}
	release, err := l.imports.claim(sourceID)
	if err != nil {
		return err
	}
	l.imports.wg.Add(1)
	go func() {
		defer l.imports.wg.Done()
		result, err := l.importSource(l.imports.ctx, info, progress)
		release()
		if done != nil {
			done(result, err)
		}
	}()
	return nil
}

// canImport reports whether a source can fill the library: it lists its games page by page, or it
// syncs them itself.
func canImport(info contract.SourceInfo) bool {
	return hasCapability(info, CapabilityListGames) ||
		hasCapability(info, capabilitySyncGames) ||
		hasCapability(info, capabilitySyncIndex)
}

// importSource fills the library from one source. The listGames path is unchanged; a source that
// only syncs runs source.sync and saves the games it returns.
func (l *Library) importSource(ctx context.Context, info contract.SourceInfo, progress func(ImportProgress)) (ImportResult, error) {
	if hasCapability(info, CapabilityListGames) {
		return l.importPages(ctx, info.Plugin, info.SourceID, progress)
	}
	result, err := l.syncDirect(ctx, info)
	reportProgress(progress, result)
	return result, err
}

// syncDirect runs source.sync for a source that does not list its games. A sync:games source
// answers with the games themselves; a sync:index source answers with a status object, which may
// carry no games. A status without games is a success: the source did its work, so no games are
// invented for it.
func (l *Library) syncDirect(ctx context.Context, info contract.SourceInfo) (ImportResult, error) {
	var reply syncReply
	params := map[string]string{"source": info.SourceID}
	if err := l.sources.Call(ctx, info.Plugin, methodSourceSync, params, &reply); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ImportResult{}, ctxErr
		}
		return ImportResult{Failed: 1}, fmt.Errorf("sync %s: %w", info.SourceID, err)
	}
	if err := checkSyncedGames(reply.Games, info.SourceID); err != nil {
		return ImportResult{Failed: 1}, fmt.Errorf("sync %s: %w", info.SourceID, err)
	}
	batch, err := l.saveBatch(ctx, reply.Games)
	if err != nil {
		return ImportResult{Failed: 1}, fmt.Errorf("sync %s: %w", info.SourceID, err)
	}
	l.startArtworkFor(batch.CoverIDs)
	return ImportResult{Imported: batch.Written, Skipped: batch.Skipped}, nil
}

// checkSyncedGames rejects a sync answer that carries a game of another source.
func checkSyncedGames(games []contract.GameDTO, source string) error {
	for _, game := range games {
		if game.Source != source {
			return fmt.Errorf("game %q has source %q, want %q", game.Title, game.Source, source)
		}
	}
	return nil
}

// syncReply is the answer of source.sync. A sync:games source answers with a bare array of games;
// a sync:index source answers with a status object. Both shapes are accepted, and a status object
// may carry no games at all, so an index-only sync leaves the library untouched.
type syncReply struct {
	Games []contract.GameDTO
}

func (r *syncReply) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil
	}
	if trimmed[0] == '[' {
		return json.Unmarshal(trimmed, &r.Games)
	}
	var status struct {
		Games []contract.GameDTO `json:"games"`
	}
	if err := json.Unmarshal(trimmed, &status); err != nil {
		return err
	}
	r.Games = status.Games
	return nil
}

// importPages saves the pages of a source's games until the plugin reports no next page.
func (l *Library) importPages(ctx context.Context, plugin, source string, progress func(ImportProgress)) (ImportResult, error) {
	var result ImportResult
	used := map[string]bool{}
	cursor := ""
	for {
		used[cursor] = true
		batch, next, err := l.importPage(ctx, plugin, source, cursor, used)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return result, ctxErr
			}
			result.Failed++
			reportProgress(progress, result)
			return result, fmt.Errorf("import %s: %w", source, err)
		}
		result.Imported += batch.Written
		result.Skipped += batch.Skipped
		reportProgress(progress, result)
		if next == "" {
			return result, nil
		}
		cursor = next
	}
}

// importPage fetches one page, saves its games and returns the next cursor. used holds the cursors
// already read, so a plugin that loops is stopped before its games are saved again.
func (l *Library) importPage(ctx context.Context, plugin, source, cursor string, used map[string]bool) (batchResult, string, error) {
	page, err := l.fetchGamesPage(ctx, plugin, source, cursor)
	if err != nil {
		return batchResult{}, "", err
	}
	if err := checkPage(page, source, used); err != nil {
		return batchResult{}, "", err
	}
	batch, err := l.saveBatch(ctx, page.Games)
	if err != nil {
		return batchResult{}, "", err
	}
	l.startArtworkFor(batch.CoverIDs)
	return batch, page.NextCursor, nil
}

func (l *Library) fetchGamesPage(ctx context.Context, plugin, source, cursor string) (listGamesReply, error) {
	var reply listGamesReply
	params := listGamesParams{Source: source, Scope: listGamesScopeMine, Cursor: cursor, Limit: importPageSize}
	if err := l.sources.Call(ctx, plugin, methodListGames, params, &reply); err != nil {
		return listGamesReply{}, fmt.Errorf("list games: %w", err)
	}
	return reply, nil
}

// checkPage rejects a page of games from another source, or a cursor the import has already read.
func checkPage(page listGamesReply, source string, used map[string]bool) error {
	for _, game := range page.Games {
		if game.Source != source {
			return fmt.Errorf("game %q has source %q, want %q", game.Title, game.Source, source)
		}
	}
	if page.NextCursor != "" && used[page.NextCursor] {
		return fmt.Errorf("plugin repeated cursor %q", page.NextCursor)
	}
	return nil
}

func reportProgress(progress func(ImportProgress), result ImportResult) {
	if progress != nil {
		progress(result)
	}
}
