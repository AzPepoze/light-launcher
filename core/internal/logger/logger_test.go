package logger

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
)

// useWriter sends stderr output to a buffer and restores the logger state when the test ends.
func useWriter(t *testing.T, color bool) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prevWriter, prevColor, prevLevel := outWriter, colorOn, minLevel
	outWriter, colorOn, minLevel = &buf, color, LevelDebug
	t.Cleanup(func() {
		mu.Lock()
		defer mu.Unlock()
		outWriter, colorOn, minLevel = prevWriter, prevColor, prevLevel
		notifier = nil
	})
	return &buf
}

func lookupFrom(values map[string]string) func(string) (string, bool) {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}

func regularFile(t *testing.T) *os.File {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "out")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { file.Close() })
	return file
}

func TestNoEscapeCodesWhenNotTerminal(t *testing.T) {
	buf := useWriter(t, colorEnabled(regularFile(t), lookupFrom(nil)))
	Info("games", "started %d", 3)
	Error("games", "failed")

	if strings.Contains(buf.String(), "\x1b[") {
		t.Fatalf("escape codes written to a non-terminal: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "[games] started 3") {
		t.Fatalf("line missing: %q", buf.String())
	}
}

func TestEscapeCodesOnlyForColorTerminal(t *testing.T) {
	buf := useWriter(t, true)
	Warn("games", "careful")
	if !strings.Contains(buf.String(), "\x1b[") {
		t.Fatalf("color line has no escape codes: %q", buf.String())
	}
}

func TestNoColorWinsOverTerminal(t *testing.T) {
	// A character device stands in for a terminal.
	console, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer console.Close()

	if !colorEnabled(console, lookupFrom(nil)) {
		t.Fatal("terminal without NO_COLOR should get colors")
	}
	if colorEnabled(console, lookupFrom(map[string]string{"NO_COLOR": "1"})) {
		t.Fatal("NO_COLOR should turn colors off")
	}
}

func TestForceColorOnPipe(t *testing.T) {
	file := regularFile(t)
	if colorEnabled(file, lookupFrom(map[string]string{"FORCE_COLOR": "1"})) != true {
		t.Fatal("FORCE_COLOR should turn colors on for a pipe")
	}
	if colorEnabled(file, lookupFrom(map[string]string{"FORCE_COLOR": "0"})) {
		t.Fatal("FORCE_COLOR=0 should not turn colors on")
	}
}

func TestCategoryColorIsStable(t *testing.T) {
	first := categoryColor("runtime")
	if second := categoryColor("runtime"); second != first {
		t.Fatalf("same category, different colors: %q and %q", first, second)
	}
	// Pinned value: a change here would recolor every category between runs.
	const want = "\x1b[38;2;238;196;227m"
	if first != want {
		t.Fatalf("categoryColor(runtime) = %q, want %q", first, want)
	}
}

func TestCategoryColorIsPastel(t *testing.T) {
	for _, category := range []string{"api", "games", "downloads", "plugins", "shell"} {
		r, g, b := rgbOf(t, categoryColor(category))
		if max(r, g, b) < 200 || min(r, g, b) < 100 {
			t.Fatalf("%s is not pastel: %d,%d,%d", category, r, g, b)
		}
	}
}

func rgbOf(t *testing.T, code string) (int, int, int) {
	t.Helper()
	var r, g, b int
	if _, err := fmt.Sscanf(code, "\x1b[38;2;%d;%d;%dm", &r, &g, &b); err != nil {
		t.Fatalf("parse %q: %v", code, err)
	}
	return r, g, b
}

func TestErrorUsesNotifierOnlyWhenSet(t *testing.T) {
	useWriter(t, false)
	var mu sync.Mutex
	var got []string
	Error("api", "no notifier yet")
	SetNotifier(func(msg string) {
		mu.Lock()
		defer mu.Unlock()
		got = append(got, msg)
	})
	Warn("api", "warnings never notify")
	Error("api", "boom")

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 1 || got[0] != "[api] boom" {
		t.Fatalf("notifier calls = %q", got)
	}
}

func TestNotifierIsOffByDefault(t *testing.T) {
	useWriter(t, false)
	mu.Lock()
	defer mu.Unlock()
	if notifier != nil {
		t.Fatal("a notifier is set by default")
	}
}

func TestConcurrentLoggingIsSafe(t *testing.T) {
	useWriter(t, false)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				Info("stress", "line %d", j)
			}
		}()
	}
	wg.Wait()
}
