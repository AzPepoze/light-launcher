package plugins

import (
	"encoding/json"
	"sync"
	"testing"
)

// recordingEvents remembers which channels were emitted.
type recordingEvents struct {
	mu       sync.Mutex
	channels []string
}

func (r *recordingEvents) Emit(channel string, _ any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.channels = append(r.channels, channel)
}

func (r *recordingEvents) has(channel string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, name := range r.channels {
		if name == channel {
			return true
		}
	}
	return false
}

func TestNotifyEmitsProgressAndKeepsLogLines(t *testing.T) {
	events := &recordingEvents{}
	env := testEnv(t, Permissions{}, nil)
	env.deps.events = events

	env.notify("progress", json.RawMessage(`{"id":"scan","done":2,"total":5}`))
	env.notify("log", json.RawMessage(`{"level":"info","message":"scanning"}`))

	if !events.has("plugin-progress") {
		t.Fatalf("progress did not reach the event sink")
	}
	if lines := env.lines.tail(0); len(lines) != 1 || lines[0] != "[info] scanning" {
		t.Fatalf("log lines = %v", lines)
	}
}

func TestLineRingKeepsLastLines(t *testing.T) {
	ring := newLineRing(3)
	for _, line := range []string{"a", "b", "c", "d", "e"} {
		ring.add(line)
	}
	got := ring.tail(2)
	if len(got) != 2 || got[0] != "d" || got[1] != "e" {
		t.Fatalf("tail(2) = %v", got)
	}
	if all := ring.tail(0); len(all) != 3 {
		t.Fatalf("tail(0) = %v, want the 3 kept lines", all)
	}
}
