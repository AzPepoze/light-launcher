// Package downloads queues, transfers and extracts game archives.
package downloads

import "time"

// State is the lifecycle stage of one download.
type State string

const (
	StateQueued      State = "queued"
	StateDownloading State = "downloading"
	StatePaused      State = "paused"
	StateDone        State = "done"
	StateError       State = "error"
	StateCanceled    State = "canceled"
)

// Download is one file to fetch. It is also the row shape the Store keeps.
type Download struct {
	ID          int64
	GameID      int64
	URL         string
	Filename    string
	Dest        string
	Size        int64
	BytesDone   int64
	Connections int
	State       State
	Error       string
	ETag        string
	Headers     map[string]string
	Install     bool
	// InstallSubfolder is the folder under the games folder for the install. Empty means the games folder.
	InstallSubfolder string
	// DeleteArchive overrides the KeepArchives setting after an install. Nil uses the setting.
	DeleteArchive *bool
	CreatedAt     time.Time
	StartedAt     time.Time
	FinishedAt    time.Time
}

// Segment is one byte range of a download. End is exclusive; Done bytes are already on disk.
type Segment struct {
	Index int
	Start int64
	End   int64
	Done  int64
}

// Patch changes only the fields that are not nil.
type Patch struct {
	State       *State
	Error       *string
	Size        *int64
	BytesDone   *int64
	Connections *int
	ETag        *string
	StartedAt   *time.Time
	FinishedAt  *time.Time
}
