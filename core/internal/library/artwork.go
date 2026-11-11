package library

import (
	"context"
	"sync"

	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/logger"
)

// artworkSlots is how many games download artwork at the same time.
const artworkSlots = 4

// artworkQueue runs artwork jobs in the background, a few at a time, until it is stopped.
type artworkQueue struct {
	ctx    context.Context
	cancel context.CancelFunc
	slots  chan struct{}
	wg     sync.WaitGroup
}

func newArtworkQueue() *artworkQueue {
	ctx, cancel := context.WithCancel(context.Background())
	return &artworkQueue{ctx: ctx, cancel: cancel, slots: make(chan struct{}, artworkSlots)}
}

// start runs job in its own goroutine once a slot is free. Jobs still waiting when the queue stops are dropped.
func (q *artworkQueue) start(job func(ctx context.Context)) {
	q.wg.Add(1)
	go func() {
		defer q.wg.Done()
		select {
		case q.slots <- struct{}{}:
		case <-q.ctx.Done():
			return
		}
		defer func() { <-q.slots }()
		job(q.ctx)
	}()
}

// stop cancels running downloads and waits until they end.
func (q *artworkQueue) stop() {
	q.cancel()
	q.wg.Wait()
}

// Upsert saves a plugin game and caches its artwork in the background. Sync and plugin uploads go through here.
func (l *Library) Upsert(ctx context.Context, dto contract.GameDTO) (int64, error) {
	id, err := l.Store.Upsert(ctx, dto)
	if err != nil {
		return 0, err
	}
	l.startArtwork(id)
	return id, nil
}

// startArtwork caches the cover and screenshots of a saved game in the background.
func (l *Library) startArtwork(id int64) {
	l.artwork.start(func(ctx context.Context) {
		l.cacheArtwork(ctx, id)
	})
}

// cacheArtwork downloads the artwork of one game. It tells the UI when a new cover file arrives.
func (l *Library) cacheArtwork(ctx context.Context, id int64) {
	game, err := l.Get(ctx, id)
	if err != nil {
		warnUnlessStopped(ctx, "artwork for game %d: %v", id, err)
		return
	}
	if l.cacheCover(ctx, game) {
		l.notify("library.game", map[string]any{"id": id})
	}
	l.CacheScreenshots(ctx, game)
}

// cacheCover stores the cover of a game that has a remote cover but no local file. It reports whether it did.
func (l *Library) cacheCover(ctx context.Context, game Game) bool {
	if game.CoverURL == nil || game.CoverPath != nil {
		return false
	}
	if _, err := l.AttachCover(ctx, game); err != nil {
		warnUnlessStopped(ctx, "cover for %s: %v", game.Title, err)
		return false
	}
	return true
}

// warnUnlessStopped logs a failed artwork step. Steps cut short by a shutdown are expected and stay quiet.
func warnUnlessStopped(ctx context.Context, format string, args ...any) {
	if ctx.Err() == nil {
		logger.Warn(logCategory, format, args...)
	}
}
