package downloads

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestInstallCarriesTitleAndClearsRunningInstall(t *testing.T) {
	hideSevenZip(t)
	archive := makeZip(t, []zipEntry{{"Game.x86_64", "binary"}})
	events := &eventLog{}
	in := &FolderInstaller{
		Root:   t.TempDir(),
		Events: events,
		Titles: func(context.Context, int64) (string, error) { return "Harbor Tale", nil },
	}

	if err := in.Install(context.Background(), InstallJob{DownloadID: 1, GameID: 1, Archive: archive}); err != nil {
		t.Fatal(err)
	}
	installs := events.installs()
	for _, snapshot := range installs {
		if snapshot.Title != "Harbor Tale" {
			t.Fatalf("snapshot title = %q, want Harbor Tale in %+v", snapshot.Title, snapshot)
		}
	}
	last := installs[len(installs)-1]
	if filepath.Base(last.InstallDir) != "Harbor Tale" {
		t.Errorf("install folder = %s, want it named after the title", last.InstallDir)
	}
	if last.ETA != nil {
		t.Errorf("a finished install has no eta, got %v", *last.ETA)
	}
	if active := in.ActiveInstalls(); len(active) != 0 {
		t.Errorf("active installs after the install = %+v, want none", active)
	}
}

func TestRunningInstallsKeepOnlyActiveSnapshots(t *testing.T) {
	var running runningInstalls
	running.update(InstallProgress{DownloadID: 2, State: "active"})
	running.update(InstallProgress{DownloadID: 1, State: "active"})
	if got := running.list(); len(got) != 2 || got[0].DownloadID != 1 || got[1].DownloadID != 2 {
		t.Fatalf("list = %+v, want downloads 1 and 2 in order", got)
	}
	running.update(InstallProgress{DownloadID: 1, State: "done"})
	if got := running.list(); len(got) != 1 || got[0].DownloadID != 2 {
		t.Errorf("after done, list = %+v, want only download 2", got)
	}
}

func TestEtaSecondsNeedsSpeedAndTotal(t *testing.T) {
	if eta := etaSeconds(50, 100, 10); eta == nil || *eta != 5 {
		t.Errorf("eta = %v, want 5 seconds", eta)
	}
	cases := map[string]struct {
		done, total int64
		speed       float64
	}{
		"unknown total": {done: 0, total: 0, speed: 10},
		"no speed yet":  {done: 50, total: 100, speed: 0},
		"finished":      {done: 100, total: 100, speed: 10},
	}
	for name, c := range cases {
		if eta := etaSeconds(c.done, c.total, c.speed); eta != nil {
			t.Errorf("%s: eta = %v, want nil", name, *eta)
		}
	}
}

func TestViewIsRateLimitedOnlyWhileDownloadingUnderLimit(t *testing.T) {
	m := newTestManager(t, nil, t.TempDir(), Settings{SpeedLimitKBps: 100})
	now := time.Now()
	if v := m.viewLocked(Download{State: StateDownloading}, now); !v.RateLimited {
		t.Error("a running download under a speed limit must be rate limited")
	}
	if v := m.viewLocked(Download{State: StateQueued}, now); v.RateLimited {
		t.Error("a queued download is not rate limited")
	}

	unlimited := newTestManager(t, nil, t.TempDir(), Settings{})
	if v := unlimited.viewLocked(Download{State: StateDownloading}, now); v.RateLimited {
		t.Error("without a speed limit nothing is rate limited")
	}
}

func TestStatusReportsLimitAndSevenZip(t *testing.T) {
	hideSevenZip(t)
	m := newTestManager(t, nil, t.TempDir(), Settings{SpeedLimitKBps: 512})
	status := m.Status()
	if status.SpeedLimitKBps != 512 || status.SevenZip {
		t.Errorf("status = %+v, want limit 512 and no 7-Zip", status)
	}
	if status.Installs == nil {
		t.Error("installs must be an empty list, not null")
	}
}
