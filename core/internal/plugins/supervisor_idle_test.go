package plugins

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func idleTimeout(value time.Duration) *time.Duration {
	return &value
}

// waitForTimer waits until the watcher is waiting for the next timer at d from now.
func waitForTimer(t *testing.T, clock *fakeClock, d time.Duration) {
	t.Helper()
	waitUntil(t, func() bool { return clock.HasTimer(d) })
}

func TestIdleStopsPluginAfterTimeout(t *testing.T) {
	clock := newFakeClock()
	sup, _, recorder := newFixtureSupervisorWith(t, Permissions{}, idleTimeout(time.Minute), clock)
	if _, err := callFixture(sup, "source.parseRef", map[string]string{}); err != nil {
		t.Fatal(err)
	}
	waitForTimer(t, clock, time.Minute)

	clock.Advance(59 * time.Second)
	if !sup.Running() {
		t.Fatal("the plugin stopped before the idle timeout")
	}

	clock.Advance(time.Second)
	waitUntil(t, func() bool { return !sup.Running() })
	if !sup.Idle() {
		t.Fatal("the supervisor does not report the idle stop")
	}
	if sup.crashCount() != 0 || sup.Disabled() || recorder.count() != 0 {
		t.Fatal("an idle stop must not count as a crash")
	}
}

func TestActivityResetsIdleTimer(t *testing.T) {
	clock := newFakeClock()
	sup, _, _ := newFixtureSupervisorWith(t, Permissions{}, idleTimeout(time.Minute), clock)
	if _, err := callFixture(sup, "source.parseRef", map[string]string{}); err != nil {
		t.Fatal(err)
	}

	clock.Advance(50 * time.Second)
	if _, err := callFixture(sup, "source.echo", map[string]string{"x": "1"}); err != nil {
		t.Fatal(err)
	}
	waitForTimer(t, clock, time.Minute)
	clock.Advance(50 * time.Second)
	if !sup.Running() {
		t.Fatal("a call did not reset the idle timer")
	}

	clock.Advance(10 * time.Second)
	waitUntil(t, func() bool { return !sup.Running() })
}

func TestInFlightCallBlocksIdle(t *testing.T) {
	clock := newFakeClock()
	sup, _, _ := newFixtureSupervisorWith(t, Permissions{}, idleTimeout(time.Minute), clock)

	done := make(chan error, 1)
	go func() {
		_, err := callFixture(sup, "source.wait", map[string]int{"ms": 200})
		done <- err
	}()
	waitUntil(t, func() bool {
		sup.mu.Lock()
		defer sup.mu.Unlock()
		return sup.inFlight > 0
	})

	clock.Advance(10 * time.Minute)
	if !sup.Running() {
		t.Fatal("an in-flight call did not keep the plugin awake")
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	waitUntil(t, func() bool {
		sup.mu.Lock()
		defer sup.mu.Unlock()
		return sup.inFlight == 0
	})

	deadline := time.Now().Add(2 * time.Second)
	for sup.Running() && time.Now().Before(deadline) {
		clock.Advance(10 * time.Second)
		time.Sleep(time.Millisecond)
	}
	if sup.Running() {
		t.Fatal("the plugin did not stop after the call ended")
	}
}

func TestUnfinishedProgressJobBlocksIdle(t *testing.T) {
	clock := newFakeClock()
	sup, _, _ := newFixtureSupervisorWith(t, Permissions{}, idleTimeout(time.Minute), clock)
	if _, err := callFixture(sup, "source.progress", map[string]any{"id": "job", "done": 1, "total": 2}); err != nil {
		t.Fatal(err)
	}
	waitForTimer(t, clock, time.Minute)

	clock.Advance(5 * time.Minute)
	time.Sleep(20 * time.Millisecond)
	if !sup.Running() {
		t.Fatal("an unfinished progress job did not keep the plugin awake")
	}

	if _, err := callFixture(sup, "source.progress", map[string]any{"id": "job", "done": 2, "total": 2}); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for sup.Running() && time.Now().Before(deadline) {
		clock.Advance(10 * time.Second)
		time.Sleep(time.Millisecond)
	}
	if sup.Running() {
		t.Fatal("the plugin did not stop after the progress job finished")
	}
}

func TestCallDuringIdleShutdownWaitsThenStartsFresh(t *testing.T) {
	clock := newFakeClock()
	sup, _, _ := newFixtureSupervisorWith(t, Permissions{}, idleTimeout(time.Minute), clock)

	slowDone := make(chan error, 1)
	go func() {
		_, err := callFixture(sup, "source.wait", map[string]int{"ms": 300})
		slowDone <- err
	}()
	waitUntil(t, func() bool {
		sup.mu.Lock()
		defer sup.mu.Unlock()
		return sup.inFlight > 0
	})

	shutdownDone := make(chan struct{})
	go func() {
		sup.Shutdown()
		close(shutdownDone)
	}()
	waitUntil(t, func() bool {
		sup.mu.Lock()
		defer sup.mu.Unlock()
		return sup.conn != nil && sup.conn.stopping
	})

	if _, err := callFixture(sup, "source.parseRef", map[string]string{}); err != nil {
		t.Fatalf("call during shutdown: %v", err)
	}
	<-shutdownDone
	<-slowDone
	if !sup.Running() {
		t.Fatal("the plugin did not start again after shutdown")
	}
	if sup.crashCount() != 0 || sup.Disabled() {
		t.Fatal("the restart after shutdown counted as a crash")
	}
}

func TestIdleRestartReadsSettingsAgain(t *testing.T) {
	clock := newFakeClock()
	sup, env, _ := newFixtureSupervisorWith(t, Permissions{}, idleTimeout(time.Minute), clock)
	settingsPath := filepath.Join(env.dataDir, settingsFileName)
	writeSettings := func(region string) {
		t.Helper()
		data, err := json.Marshal(map[string]string{"region": region})
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(settingsPath, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	writeSettings("us")
	if region := probeRegion(t, sup); region != "us" {
		t.Fatalf("region before the idle stop = %q", region)
	}
	waitForTimer(t, clock, time.Minute)
	clock.Advance(time.Minute)
	waitUntil(t, func() bool { return !sup.Running() })

	writeSettings("eu")
	if region := probeRegion(t, sup); region != "eu" {
		t.Fatalf("region after the restart = %q", region)
	}
	if sup.Idle() {
		t.Fatal("a call must clear the idle state")
	}
}

func TestZeroIdleTimeoutNeverStops(t *testing.T) {
	clock := newFakeClock()
	sup, _, _ := newFixtureSupervisorWith(t, Permissions{}, idleTimeout(0), clock)
	if _, err := callFixture(sup, "source.parseRef", map[string]string{}); err != nil {
		t.Fatal(err)
	}
	clock.Advance(24 * time.Hour)
	time.Sleep(20 * time.Millisecond)
	if !sup.Running() {
		t.Fatal("idle timeout 0 must keep the plugin running")
	}
}

func probeRegion(t *testing.T, sup *Supervisor) string {
	t.Helper()
	raw, err := callFixture(sup, "source.probe", map[string]any{"service": "settings.get", "args": nil})
	if err != nil {
		t.Fatalf("settings.get: %v", err)
	}
	var answer probeAnswer
	if err := json.Unmarshal(raw, &answer); err != nil {
		t.Fatalf("probe answer %s: %v", raw, err)
	}
	var result struct {
		Settings map[string]any `json:"settings"`
	}
	if err := json.Unmarshal(answer.Result, &result); err != nil {
		t.Fatalf("settings result %s: %v", answer.Result, err)
	}
	region, _ := result.Settings["region"].(string)
	return region
}

func TestIdleWatcherStopsOnShutdown(t *testing.T) {
	clock := newFakeClock()
	sup, _, _ := newFixtureSupervisorWith(t, Permissions{}, idleTimeout(time.Hour), clock)
	if _, err := callFixture(sup, "source.parseRef", map[string]string{}); err != nil {
		t.Fatal(err)
	}
	sup.Shutdown()
	waitUntil(t, func() bool { return !sup.Running() })
	if sup.Idle() {
		t.Fatal("a requested shutdown is not an idle stop")
	}
}
