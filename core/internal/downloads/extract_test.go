package downloads

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type zipEntry struct {
	name string
	data string
}

// makeZip writes a zip with the given entries, in order, into a new temp folder and returns its path.
func makeZip(t *testing.T, entries []zipEntry) string {
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
	return writeTestFile(t, "game.zip", buf.Bytes())
}

// hideSevenZip makes the 7-Zip lookup fail, so the built-in readers run.
func hideSevenZip(t *testing.T) {
	t.Helper()
	t.Setenv("PATH", t.TempDir())
}

func TestExtractZipSkipsJunkAndReportsProgress(t *testing.T) {
	hideSevenZip(t)
	archive := makeZip(t, []zipEntry{
		{"game/readme.txt", "hello world"},
		{"game/run.sh", "#!/bin/sh\necho hi\n"},
		{"__MACOSX/._readme.txt", "junk"},
		{"game/.DS_Store", "junk"},
	})
	dest := filepath.Join(t.TempDir(), "out")

	var engine string
	var lastDone, lastTotal int64
	err := Extract(context.Background(), archive, dest, ExtractOptions{
		OnEngine:   func(name string) { engine = name },
		OnProgress: func(done, total int64) { lastDone, lastTotal = done, total },
	})
	if err != nil {
		t.Fatal(err)
	}
	if engine != "built-in zip" {
		t.Fatalf("engine = %q", engine)
	}
	if got, _ := os.ReadFile(filepath.Join(dest, "game", "readme.txt")); string(got) != "hello world" {
		t.Fatalf("readme = %q", got)
	}
	if _, err := os.Stat(filepath.Join(dest, "__MACOSX")); !os.IsNotExist(err) {
		t.Fatal("macOS metadata must not be extracted")
	}
	if _, err := os.Stat(filepath.Join(dest, "game", ".DS_Store")); !os.IsNotExist(err) {
		t.Fatal(".DS_Store must not be extracted")
	}
	if lastTotal == 0 || lastDone != lastTotal {
		t.Fatalf("progress ended at %d/%d", lastDone, lastTotal)
	}
}

func TestExtractRejectsZipSlipBeforeWritingAnything(t *testing.T) {
	hideSevenZip(t)
	archive := makeZip(t, []zipEntry{
		{"game/ok.txt", "fine"},
		{"../evil.txt", "pwned"},
	})
	root := t.TempDir()
	dest := filepath.Join(root, "slip")

	err := Extract(context.Background(), archive, dest, ExtractOptions{})
	if err == nil || !strings.Contains(err.Error(), "zip-slip") {
		t.Fatalf("err = %v, want a zip-slip error", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, "evil.txt")); !os.IsNotExist(statErr) {
		t.Fatal("zip-slip entry was written outside the destination")
	}
	if _, statErr := os.Stat(filepath.Join(dest, "game", "ok.txt")); !os.IsNotExist(statErr) {
		t.Fatal("archive must be refused before any entry is written")
	}
}

func TestExtractStopsWhenContextIsCanceled(t *testing.T) {
	hideSevenZip(t)
	archive := makeZip(t, []zipEntry{{"game/a.bin", strings.Repeat("x", 100_000)}})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := Extract(ctx, archive, filepath.Join(t.TempDir(), "out"), ExtractOptions{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestSevenZipWithoutToolAsksForInstall(t *testing.T) {
	hideSevenZip(t)
	archive := writeTestFile(t, "game.7z", append([]byte{0x37, 0x7a, 0xbc, 0xaf, 0x27, 0x1c}, make([]byte, 32)...))
	err := Extract(context.Background(), archive, t.TempDir(), ExtractOptions{})
	if err == nil || !strings.Contains(err.Error(), "install 7-Zip") {
		t.Fatalf("err = %v, want an install 7-Zip message", err)
	}
}

func TestUnknownArchiveIsUnsupported(t *testing.T) {
	archive := writeTestFile(t, "mystery.bin", []byte("hello"))
	err := Extract(context.Background(), archive, t.TempDir(), ExtractOptions{})
	if !errors.Is(err, ErrUnsupportedArchive) {
		t.Fatalf("err = %v, want ErrUnsupportedArchive", err)
	}
}

func TestMalformedRarFailsWithoutPanicking(t *testing.T) {
	hideSevenZip(t)
	archive := writeTestFile(t, "broken.rar", append([]byte("Rar!\x1a\x07\x01\x00"), []byte("garbage data")...))
	if err := Extract(context.Background(), archive, t.TempDir(), ExtractOptions{}); err == nil {
		t.Fatal("a malformed RAR must return an error")
	}
}

func TestExtractWithSevenZipWhenInstalled(t *testing.T) {
	bin := ""
	for _, name := range sevenZipNames {
		if p, err := exec.LookPath(name); err == nil {
			bin = p
			break
		}
	}
	if bin == "" {
		t.Skip("no 7-Zip command on PATH")
	}
	src := t.TempDir()
	if err := os.MkdirAll(filepath.Join(src, "game"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "game", "a.txt"), []byte("seven zip"), 0o644); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(t.TempDir(), "game.7z")
	build := exec.Command(bin, "a", archive, "game")
	build.Dir = src
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building the fixture failed: %v: %s", err, out)
	}

	dest := filepath.Join(t.TempDir(), "out")
	var engine string
	var lastDone, lastTotal int64
	err := Extract(context.Background(), archive, dest, ExtractOptions{
		OnEngine:   func(name string) { engine = name },
		OnProgress: func(done, total int64) { lastDone, lastTotal = done, total },
	})
	if err != nil {
		t.Fatal(err)
	}
	if engine != "7-Zip" {
		t.Fatalf("engine = %q, want 7-Zip", engine)
	}
	if got, _ := os.ReadFile(filepath.Join(dest, "game", "a.txt")); string(got) != "seven zip" {
		t.Fatalf("a.txt = %q", got)
	}
	if lastTotal != 9 || lastDone != 9 {
		t.Fatalf("progress ended at %d/%d, want 9/9", lastDone, lastTotal)
	}
}
