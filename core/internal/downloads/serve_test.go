package downloads

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// testFile serves data with optional failure injection. Failures apply to GET requests only.
type testFile struct {
	data      []byte
	etag      string
	noRanges  bool
	slow      bool  // write in small pieces with pauses, so a test can stop mid-transfer
	failGets  int   // answer this many GETs with 500
	limitGets int   // answer this many GETs with 429 and Retry-After: 1
	dropAfter int64 // drop the first ranged GET after this many body bytes (0 = never)

	mu       sync.Mutex
	ranges   []string
	failures int
}

func (f *testFile) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("ETag", f.etag)
	w.Header().Set("Content-Type", "application/octet-stream")
	if f.noRanges {
		w.Header().Set("Content-Length", strconv.Itoa(len(f.data)))
	} else {
		w.Header().Set("Accept-Ranges", "bytes")
	}
	if r.Method == http.MethodHead {
		w.Header().Set("Content-Length", strconv.Itoa(len(f.data)))
		return
	}

	f.mu.Lock()
	slow := f.slow
	if f.failGets > 0 {
		f.failGets--
		f.failures++
		f.mu.Unlock()
		http.Error(w, "temporary failure", http.StatusInternalServerError)
		return
	}
	if f.limitGets > 0 {
		f.limitGets--
		f.mu.Unlock()
		w.Header().Set("Retry-After", "1")
		http.Error(w, "slow down", http.StatusTooManyRequests)
		return
	}
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		f.ranges = append(f.ranges, rangeHeader)
	}
	drop := f.dropAfter
	f.dropAfter = 0
	f.mu.Unlock()

	if rangeHeader == "" || f.noRanges {
		w.Header().Set("Content-Length", strconv.Itoa(len(f.data)))
		writeBody(w, f.data, 0, slow)
		return
	}
	var start, end int64
	if _, err := fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end); err != nil {
		http.Error(w, "bad range", http.StatusRequestedRangeNotSatisfiable)
		return
	}
	body := f.data[start : end+1]
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(f.data)))
	w.WriteHeader(http.StatusPartialContent)
	writeBody(w, body, drop, slow)
}

// setSlow changes the pacing of later responses.
func (f *testFile) setSlow(slow bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.slow = slow
}

func (f *testFile) rangeLog() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.ranges...)
}

// writeBody sends body; when drop > 0 it aborts the connection after drop bytes.
func writeBody(w http.ResponseWriter, body []byte, drop int64, slow bool) {
	const piece = 16 * 1024
	for off := 0; off < len(body); off += piece {
		end := min(off+piece, len(body))
		if drop > 0 && int64(end) > drop {
			w.Write(body[off : off+int(drop)])
			panic(http.ErrAbortHandler)
		}
		w.Write(body[off:end])
		if slow {
			w.(http.Flusher).Flush()
			time.Sleep(5 * time.Millisecond)
		}
	}
}

func patternData(n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = byte(i*31 + i>>9)
	}
	return out
}
