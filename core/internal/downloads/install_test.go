package downloads

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"light-launcher/core/internal/api"
)

// eventLog records every event an install or manager sends.
type eventLog struct {
	mu     sync.Mutex
	events []api.Event
}

func (l *eventLog) Emit(channel string, data any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = append(l.events, api.Event{Channel: channel, Data: data})
}

func (l *eventLog) installs() []InstallProgress {
	l.mu.Lock()
	defer l.mu.Unlock()
	var out []InstallProgress
	for _, e := range l.events {
		if p, ok := e.Data.(InstallProgress); ok && e.Channel == EventInstall {
			out = append(out, p)
		}
	}
	return out
}

func (l *eventLog) toasts() []Toast {
	l.mu.Lock()
	defer l.mu.Unlock()
	var out []Toast
	for _, e := range l.events {
		if t, ok := e.Data.(Toast); ok && e.Channel == EventToast {
			out = append(out, t)
		}
	}
	return out
}

// zipBytes builds a zip in memory, for archives that wrap other archives.
func zipBytes(t *testing.T, entries []zipEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, e := range entries {
		f, err := w.Create(e.name)
		if err != nil {
			t.Fatal(err)
		}
		f.Write([]byte(e.data))
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func launcherName() string {
	if runtime.GOOS == "windows" {
		return "Game.exe"
	}
	return "Game.x86_64"
}

func TestInstallUnpacksAndFindsLauncher(t *testing.T) {
	hideSevenZip(t)
	library := t.TempDir()
	archive := makeZip(t, []zipEntry{
		{"Game-1.0/" + launcherName(), "binary"},
		{"Game-1.0/data/level.dat", "level"},
		{"__MACOSX/._" + launcherName(), "junk"},
	})
	events := &eventLog{}
	in := &FolderInstaller{Root: library, Events: events}

	if err := in.Install(context.Background(), 7, 3, archive); err != nil {
		t.Fatal(err)
	}
	installs := events.installs()
	last := installs[len(installs)-1]
	if last.State != "done" || last.Progress == nil || *last.Progress != 1 {
		t.Fatalf("last progress = %+v", last)
	}
	if want := filepath.Join(library, "game"); last.InstallDir != want {
		t.Fatalf("install dir = %s", last.InstallDir)
	}
	if _, err := os.Stat(last.ExePath); err != nil {
		t.Fatalf("launcher not in place: %v", err)
	}
	if _, err := os.Stat(filepath.Join(last.InstallDir, ".light-launcher", "manifest.json")); err != nil {
		t.Fatal("manifest missing")
	}
	entries, _ := os.ReadDir(library)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".staging-") {
			t.Fatalf("staging folder left behind: %s", e.Name())
		}
	}
}

func TestInstallPeelsWrapperArchive(t *testing.T) {
	hideSevenZip(t)
	inner := zipBytes(t, []zipEntry{{"game.bin", "payload"}})
	archive := writeTestFile(t, "wrapper.zip", zipBytes(t, []zipEntry{{"inner.zip", string(inner)}}))
	events := &eventLog{}
	in := &FolderInstaller{Root: t.TempDir(), Events: events}

	if err := in.Install(context.Background(), 1, 1, archive); err != nil {
		t.Fatal(err)
	}
	last := events.installs()[len(events.installs())-1]
	data, err := os.ReadFile(filepath.Join(last.InstallDir, "game.bin"))
	if err != nil || string(data) != "payload" {
		t.Fatalf("inner archive not unpacked: %v", err)
	}
}

func TestInstallNamesFolderFromTitle(t *testing.T) {
	hideSevenZip(t)
	archive := makeZip(t, []zipEntry{{"Game.x86_64", "binary"}})
	library := t.TempDir()
	in := &FolderInstaller{
		Root: library,
		Titles: func(context.Context, int64) (string, error) {
			return "Dungeon: Crawl [v2] (Final).", nil
		},
	}
	if err := in.Install(context.Background(), 2, 9, archive); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(library, "Dungeon Crawl")); err != nil {
		t.Fatalf("folder not named from title: %v", err)
	}
}

func TestInstallFailureCleansUpAndReports(t *testing.T) {
	hideSevenZip(t)
	archive := writeTestFile(t, "broken.zip", []byte("PK\x03\x04 not a real zip"))
	library := t.TempDir()
	events := &eventLog{}
	in := &FolderInstaller{Root: library, Events: events}

	if err := in.Install(context.Background(), 4, 2, archive); err == nil {
		t.Fatal("a broken archive must fail the install")
	}
	entries, _ := os.ReadDir(library)
	if len(entries) != 0 {
		t.Fatalf("library should be empty after a failed install, found %d entries", len(entries))
	}
	installs := events.installs()
	if last := installs[len(installs)-1]; last.State != "error" || last.Error == "" {
		t.Fatalf("failure not reported: %+v", last)
	}
}

func TestInstallWarnsWhenSevenZipIsMissing(t *testing.T) {
	hideSevenZip(t)
	archive := writeTestFile(t, "game.7z", append([]byte{0x37, 0x7a, 0xbc, 0xaf, 0x27, 0x1c}, make([]byte, 32)...))
	events := &eventLog{}
	in := &FolderInstaller{Root: t.TempDir(), Events: events}

	if err := in.Install(context.Background(), 5, 5, archive); err == nil {
		t.Fatal("a .7z without 7-Zip cannot be installed")
	}
	toasts := events.toasts()
	if len(toasts) != 1 || !strings.Contains(toasts[0].Message, "7-Zip") {
		t.Fatalf("toasts = %+v, want one 7-Zip warning", toasts)
	}
}

func TestFolderNameDropsTagsAndIllegalCharacters(t *testing.T) {
	cases := map[string]string{
		"Game [Final] (v1.2)": "Game",
		"A: B/C?":             "A B C",
		"   ":                 "game-7",
		"Name.":               "Name",
	}
	for in, want := range cases {
		if got := folderName(in, 7); got != want {
			t.Errorf("folderName(%q) = %q, want %q", in, got, want)
		}
	}
}
