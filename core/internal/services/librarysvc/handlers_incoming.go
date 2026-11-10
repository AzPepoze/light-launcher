package librarysvc

import (
	"context"

	"light-launcher/core/internal/library"
)

type incomingLinkPayload struct {
	Name   string `json:"name"`
	GameID int64  `json:"gameId"`
}

type namePayload struct {
	Name string `json:"name"`
}

func (s *service) incomingList(ctx context.Context) (library.IncomingReport, error) {
	return s.lib.Incoming(ctx)
}

func (s *service) incomingLink(ctx context.Context, in incomingLinkPayload) (okResult, error) {
	if err := s.lib.IncomingLink(ctx, in.Name, in.GameID); err != nil {
		return okResult{}, err
	}
	s.changed(in.GameID)
	return okResult{OK: true}, nil
}

func (s *service) incomingIgnore(ctx context.Context, in namePayload) (okResult, error) {
	if err := s.lib.IncomingIgnore(ctx, in.Name); err != nil {
		return okResult{}, err
	}
	return okResult{OK: true}, nil
}

func (s *service) incomingCreateLocal(ctx context.Context, in namePayload) (library.Game, error) {
	game, err := s.lib.IncomingCreateLocal(ctx, in.Name)
	if err != nil {
		return library.Game{}, err
	}
	s.changed(game.ID)
	return game, nil
}
