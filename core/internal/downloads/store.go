package downloads

import (
	"context"
	"errors"
)

// ErrNotFound is returned when a download ID does not exist.
var ErrNotFound = errors.New("download not found")

// Store keeps download rows and their segment state.
type Store interface {
	Insert(ctx context.Context, d Download) (int64, error)
	Get(ctx context.Context, id int64) (Download, error)
	List(ctx context.Context) ([]Download, error)
	Update(ctx context.Context, id int64, p Patch) error
	Delete(ctx context.Context, id int64) error
	SaveSegments(ctx context.Context, id int64, segments []Segment) error
	Segments(ctx context.Context, id int64) ([]Segment, error)
}
