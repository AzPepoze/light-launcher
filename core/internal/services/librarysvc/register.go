// Package librarysvc exposes the game library to the UI as api methods.
package librarysvc

import (
	"light-launcher/core/internal/api"
	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/library"
)

type service struct {
	lib     *library.Library
	sources contract.Sources
	events  api.Events
}

// Register adds the library methods to the router. Each method's payload and result are listed in the report.
func Register(r *api.Router, ev api.Events, lib *library.Library, sources contract.Sources) {
	s := &service{lib: lib, sources: sources, events: ev}

	r.Handle("LibraryList", api.Typed(s.list))
	r.Handle("LibraryGet", api.Typed(s.get))
	r.Handle("LibraryUpdate", api.Typed(s.update))
	r.Handle("LibraryRemove", api.Typed(s.remove))
	r.Handle("LibrarySearch", api.Typed(s.search))
	r.Handle("LibraryPreview", api.Typed(s.preview))
	r.Handle("LibraryAddFromUrl", api.Typed(s.addFromURL))
	r.Handle("LibraryLinkLocal", api.Typed(s.linkLocal))
	r.Handle("LibraryLinkFolder", api.Typed(s.linkFolder))
	r.Handle("LibraryCheckUpdates", api.Typed(s.checkUpdates))
	r.Handle("LibraryUpdatesSummary", api.NoInput(s.updatesSummary))
	r.Handle("LibraryUpdateHistory", api.Typed(s.updateHistory))
	r.Handle("LibrarySetFavorite", api.Typed(s.setFavorite))
	r.Handle("LibraryScanFolder", api.Typed(s.scanFolder))
	r.Handle("LibraryLaunch", api.Typed(s.launch))

	r.Handle("IncomingList", api.NoInput(s.incomingList))
	r.Handle("IncomingLink", api.Typed(s.incomingLink))
	r.Handle("IncomingIgnore", api.Typed(s.incomingIgnore))
	r.Handle("IncomingCreateLocal", api.Typed(s.incomingCreateLocal))
}

// okResult is the empty success reply for actions with nothing to return.
type okResult struct {
	OK bool `json:"ok"`
}

// idPayload addresses one game.
type idPayload struct {
	ID int64 `json:"id"`
}

// changed tells the UI that one game row changed.
func (s *service) changed(id int64) {
	s.events.Emit("library.game", map[string]any{"id": id})
}
