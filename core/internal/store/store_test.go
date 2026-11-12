package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

type payload struct {
	Hello string `json:"hello"`
	Count int    `json:"count"`
}

func TestSaveAndLoadRoundTripAddsVersion(t *testing.T) {
	target := filepath.Join(t.TempDir(), "nested", "value.json")
	file := NewFile(target, 3)
	if err := file.Save(payload{Hello: "world", Count: 7}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	var raw map[string]any
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("file is not JSON: %v", err)
	}
	if raw["version"] != float64(3) {
		t.Errorf("version = %v, want 3", raw["version"])
	}

	var loaded payload
	found, err := file.Load(&loaded)
	if err != nil || !found {
		t.Fatalf("Load = %v, %v; want found", found, err)
	}
	if loaded.Hello != "world" || loaded.Count != 7 {
		t.Fatalf("loaded = %+v", loaded)
	}
}

func TestLoadAcceptsLegacyVersionZero(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    payload
	}{
		{"fields", `{"hello": "world", "count": 7}`, payload{Hello: "world", Count: 7}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			target := filepath.Join(t.TempDir(), "value.json")
			if err := os.WriteFile(target, []byte(test.content), 0o644); err != nil {
				t.Fatal(err)
			}
			var loaded payload
			found, err := NewFile(target, 1).Load(&loaded)
			if err != nil || !found {
				t.Fatalf("Load = %v, %v; want found", found, err)
			}
			if loaded != test.want {
				t.Fatalf("loaded = %+v, want %+v", loaded, test.want)
			}
		})
	}
}

func TestSaveAfterLegacyLoadWritesVersionOne(t *testing.T) {
	target := filepath.Join(t.TempDir(), "value.json")
	if err := os.WriteFile(target, []byte(`{"hello": "world", "count": 7}`), 0o644); err != nil {
		t.Fatal(err)
	}
	file := NewFile(target, 1)
	var loaded payload
	if _, err := file.Load(&loaded); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := file.Save(loaded); err != nil {
		t.Fatalf("Save: %v", err)
	}

	var raw map[string]any
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("file is not JSON: %v", err)
	}
	if raw["version"] != float64(1) {
		t.Errorf("version = %v, want 1", raw["version"])
	}
	if raw["hello"] != "world" || raw["count"] != float64(7) {
		t.Errorf("fields = %v, want hello=world count=7", raw)
	}
}

func TestLoadMissingFileReportsNotFound(t *testing.T) {
	file := NewFile(filepath.Join(t.TempDir(), "missing.json"), 1)
	var loaded payload
	found, err := file.Load(&loaded)
	if err != nil || found {
		t.Fatalf("Load = %v, %v; want not found without error", found, err)
	}
}

func TestLoadRejectsEmptyCorruptAndUnknownVersions(t *testing.T) {
	dir := t.TempDir()
	tests := []struct {
		name    string
		content string
		want    error
	}{
		{"empty", "", nil},
		{"corrupt", "{", nil},
		{"version 2", `{"version": 2, "hello": "world"}`, ErrVersion},
		{"unknown version", `{"version": 99, "hello": "world"}`, ErrVersion},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			target := filepath.Join(dir, test.name+".json")
			if err := os.WriteFile(target, []byte(test.content), 0o644); err != nil {
				t.Fatal(err)
			}
			var loaded payload
			found, err := NewFile(target, 1).Load(&loaded)
			if err == nil {
				t.Fatal("Load succeeded, want an error")
			}
			if found {
				t.Error("found must be false on error")
			}
			if test.want != nil && !errors.Is(err, test.want) {
				t.Errorf("error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestSaveCreatesMissingDirectory(t *testing.T) {
	target := filepath.Join(t.TempDir(), "a", "b", "value.json")
	if err := NewFile(target, 1).Save(payload{Hello: "x"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("file missing after Save: %v", err)
	}
}

func TestConcurrentSavesLeaveOneCompleteFile(t *testing.T) {
	target := filepath.Join(t.TempDir(), "value.json")
	file := NewFile(target, 1)

	var wait sync.WaitGroup
	for i := 0; i < 20; i++ {
		wait.Add(1)
		go func(count int) {
			defer wait.Done()
			if err := file.Save(payload{Hello: "concurrent", Count: count}); err != nil {
				t.Errorf("Save: %v", err)
			}
		}(i)
	}
	wait.Wait()

	var loaded payload
	found, err := file.Load(&loaded)
	if err != nil || !found {
		t.Fatalf("Load = %v, %v", found, err)
	}
	if loaded.Hello != "concurrent" {
		t.Fatalf("loaded = %+v", loaded)
	}
	entries, err := os.ReadDir(filepath.Dir(target))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("directory holds %d files, want only value.json", len(entries))
	}
}

func TestMarshalKeepsHTMLCharacters(t *testing.T) {
	data, err := Marshal(map[string]string{"path": "a&b<c>"})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(data) != "{\n  \"path\": \"a&b<c>\"\n}" {
		t.Fatalf("Marshal = %s", data)
	}
}
