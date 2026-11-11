package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestGeneratedFileIsCurrent fails when wire.generated.ts was not regenerated
// after a Go wire type changed. It runs with go test ./core/... and needs no bun.
func TestGeneratedFileIsCurrent(t *testing.T) {
	root := repositoryRoot(t)
	want, err := generate(root)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	path := filepath.Join(root, defaultOut)
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", defaultOut, err)
	}
	if bytes.Equal(got, want) {
		return
	}
	t.Fatalf("%s is stale; run go run ./core/cmd/gentypes\n%s", defaultOut, firstDifference(got, want))
}

func TestGenerateIsDeterministic(t *testing.T) {
	root := repositoryRoot(t)
	first, err := generate(root)
	if err != nil {
		t.Fatalf("first generate: %v", err)
	}
	second, err := generate(root)
	if err != nil {
		t.Fatalf("second generate: %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("generate returned different output for the same sources")
	}
}

func TestStructListHasNoDuplicateTypeNames(t *testing.T) {
	names := map[string]string{}
	for _, ref := range structTypes {
		if previous, exists := names[ref.tsName]; exists {
			t.Fatalf("%s is used for both %s and %s", ref.tsName, previous, ref.key())
		}
		names[ref.tsName] = ref.key()
	}
	for _, ref := range aliasTypes {
		if previous, exists := names[ref.tsName]; exists {
			t.Fatalf("%s is used for both %s and %s", ref.tsName, previous, ref.key())
		}
		names[ref.tsName] = ref.key()
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate gentypes source")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

// firstDifference names the first line that differs, which makes a stale file
// quick to understand.
func firstDifference(got, want []byte) string {
	gotLines := strings.Split(string(got), "\n")
	wantLines := strings.Split(string(want), "\n")
	limit := min(len(gotLines), len(wantLines))
	for i := 0; i < limit; i++ {
		if gotLines[i] != wantLines[i] {
			return fmt.Sprintf("first difference at line %d:\n  file: %q\n  want: %q", i+1, gotLines[i], wantLines[i])
		}
	}
	return fmt.Sprintf("file has %d lines, generated output has %d", len(gotLines), len(wantLines))
}
