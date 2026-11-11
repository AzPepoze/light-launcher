package plugins

import (
	"encoding/json"
	"sync"
	"testing"
	"time"
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
	lines := env.lines.tail(0)
	if len(lines) != 1 || lines[0].Level != "info" || lines[0].Text != "scanning" {
		t.Fatalf("log lines = %+v", lines)
	}
	if _, err := time.Parse(time.RFC3339Nano, lines[0].Time); err != nil {
		t.Errorf("log time %q is not RFC 3339: %v", lines[0].Time, err)
	}
}

func TestLineRingKeepsLastLines(t *testing.T) {
	ring := newLineRing(3)
	for _, text := range []string{"a", "b", "c", "d", "e"} {
		ring.add(LogLine{Text: text})
	}
	got := ring.tail(2)
	if len(got) != 2 || got[0].Text != "d" || got[1].Text != "e" {
		t.Fatalf("tail(2) = %+v", got)
	}
	if all := ring.tail(0); len(all) != 3 {
		t.Fatalf("tail(0) = %v, want the 3 kept lines", all)
	}
}
