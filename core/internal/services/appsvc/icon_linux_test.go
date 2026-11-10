package appsvc

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// buildIco makes an ICO with a BMP-style frame (16x16) and a PNG frame (32x32).
func buildIco() []byte {
	bmpData := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	pngData := append(append([]byte{}, pngSignature...), 0, 0, 0, 0)

	header := []byte{0, 0, 1, 0, 2, 0}
	entries := make([]byte, 2*icoEntrySize)
	dataStart := uint32(icoHeaderSize + 2*icoEntrySize)

	entries[0], entries[1] = 16, 16
	binary.LittleEndian.PutUint32(entries[8:], uint32(len(bmpData)))
	binary.LittleEndian.PutUint32(entries[12:], dataStart)

	entries[16], entries[17] = 32, 32
	binary.LittleEndian.PutUint32(entries[24:], uint32(len(pngData)))
	binary.LittleEndian.PutUint32(entries[28:], dataStart+uint32(len(bmpData)))

	var ico bytes.Buffer
	ico.Write(header)
	ico.Write(entries)
	ico.Write(bmpData)
	ico.Write(pngData)
	return ico.Bytes()
}

func TestParseIcoFramesFindsLargestPNGFrame(t *testing.T) {
	frames := parseIcoFrames(buildIco())
	if len(frames) != 2 {
		t.Fatalf("got %d frames, want 2", len(frames))
	}
	largest := largestIcoFrame(frames)
	if largest.index != 1 || largest.width != 32 || !largest.png {
		t.Fatalf("largest frame = %+v", largest)
	}
}

func TestParseIcoFramesRejectsBadHeader(t *testing.T) {
	if frames := parseIcoFrames([]byte("not an icon file")); len(frames) != 0 {
		t.Fatalf("bad header gave %d frames", len(frames))
	}
}

func TestMemoryIconCacheEvictsOldestAndChecksStamp(t *testing.T) {
	cache := newMemoryIconCache(2)
	stamp := fileStamp{mtimeMs: 1, size: 10}
	cache.remember("a.exe", stamp, "icon-a")
	cache.remember("b.exe", stamp, "icon-b")
	cache.remember("c.exe", stamp, "icon-c")

	if _, ok := cache.get("a.exe", stamp); ok {
		t.Error("oldest entry must be evicted")
	}
	if icon, ok := cache.get("c.exe", stamp); !ok || icon != "icon-c" {
		t.Errorf("newest entry = %q, %v", icon, ok)
	}
	if _, ok := cache.get("c.exe", fileStamp{mtimeMs: 2, size: 10}); ok {
		t.Error("a changed file must miss the memory cache")
	}
}

func TestExtractExeIconIgnoresFoldersAndMissingFiles(t *testing.T) {
	ctx := context.Background()
	if icon, err := extractExeIcon(ctx, t.TempDir()); err != nil || icon != "" {
		t.Errorf("folder icon = %q, %v", icon, err)
	}
	if icon, err := extractExeIcon(ctx, filepath.Join(t.TempDir(), "missing.exe")); err != nil || icon != "" {
		t.Errorf("missing file icon = %q, %v", icon, err)
	}
	file := filepath.Join(t.TempDir(), "plain.exe")
	if err := os.WriteFile(file, []byte("not a program"), 0o644); err != nil {
		t.Fatal(err)
	}
	if icon, err := extractExeIcon(ctx, file); err != nil || icon != "" {
		t.Errorf("plain file icon = %q, %v; want no icon and no error", icon, err)
	}
}
