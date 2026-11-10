package downloads

import (
	"sort"
	"sync"
)

// runningInstalls keeps the latest snapshot of each install that has not finished yet.
// The zero value is ready to use.
type runningInstalls struct {
	mu   sync.Mutex
	byID map[int64]InstallProgress
}

// update stores an active snapshot, and forgets the install once it is done or failed.
func (r *runningInstalls) update(snapshot InstallProgress) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if snapshot.State != "active" {
		delete(r.byID, snapshot.DownloadID)
		return
	}
	if r.byID == nil {
		r.byID = map[int64]InstallProgress{}
	}
	r.byID[snapshot.DownloadID] = snapshot
}

// list returns the running installs, ordered by download id.
func (r *runningInstalls) list() []InstallProgress {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]InstallProgress, 0, len(r.byID))
	for _, snapshot := range r.byID {
		out = append(out, snapshot)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DownloadID < out[j].DownloadID })
	return out
}
