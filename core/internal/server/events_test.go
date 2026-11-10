package server

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"light-launcher/core/internal/api"
)

// openStream connects to the event stream and cancels it when the test ends.
func openStream(t *testing.T, base string) *bufio.Scanner {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/api/events", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set(tokenHeader, testToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { res.Body.Close() })
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", res.StatusCode)
	}
	return bufio.NewScanner(res.Body)
}

func TestEventsStreamHubEvents(t *testing.T) {
	fx, base := newLiveFixture(t)
	scanner := openStream(t, base)

	// The hub only reaches subscribers, so keep emitting until the stream is open.
	done := make(chan struct{})
	t.Cleanup(func() { close(done) })
	go func() {
		for {
			select {
			case <-done:
				return
			case <-time.After(10 * time.Millisecond):
				fx.hub.Emit("game:update", map[string]int{"progress": 50})
			}
		}
	}()

	for scanner.Scan() {
		payload, found := strings.CutPrefix(scanner.Text(), "data: ")
		if !found {
			continue
		}
		var event api.Event
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			t.Fatalf("decode event: %v", err)
		}
		if event.Channel != "game:update" {
			t.Fatalf("channel = %q", event.Channel)
		}
		return
	}
	t.Fatalf("stream ended without an event: %v", scanner.Err())
}

func TestEventsSendHeartbeat(t *testing.T) {
	previous := heartbeatEvery
	t.Cleanup(func() { heartbeatEvery = previous })
	heartbeatEvery = 10 * time.Millisecond

	_, base := newLiveFixture(t)
	scanner := openStream(t, base)
	for scanner.Scan() {
		if scanner.Text() == ": heartbeat" {
			return
		}
	}
	t.Fatalf("stream ended without a heartbeat: %v", scanner.Err())
}
