package library

import "context"

// detailChangeLimit is how many recent version checks a game page lists.
const detailChangeLimit = 100

// GameDetail is one game with the download candidates and version history shown on its page.
// The game fields are flattened into the JSON object.
type GameDetail struct {
	Game
	Files   []GameFile `json:"files"`
	Changes []Check    `json:"changes"`
}

// Detail returns a game with its stored files and its recent version checks.
func (l *Library) Detail(ctx context.Context, id int64) (GameDetail, error) {
	game, err := l.Get(ctx, id)
	if err != nil {
		return GameDetail{}, err
	}
	files, err := l.ListFiles(ctx, id)
	if err != nil {
		return GameDetail{}, err
	}
	changes, err := l.ListChecks(ctx, id, detailChangeLimit)
	if err != nil {
		return GameDetail{}, err
	}
	return GameDetail{Game: game, Files: files, Changes: changes}, nil
}
