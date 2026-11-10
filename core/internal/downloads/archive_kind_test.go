package downloads

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestFile(t *testing.T, name string, data []byte) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(file, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return file
}

func TestRarServedUnderZipNameIsDetectedByMagic(t *testing.T) {
	file := writeTestFile(t, "PIXELDRAIN.zip", append([]byte("Rar!\x1a\x07\x01\x00"), make([]byte, 64)...))
	if got := detectArchiveKind(file); got != KindRar {
		t.Fatalf("kind = %s, want rar", got)
	}
}

func TestZipServedUnderRarNameIsDetectedByMagic(t *testing.T) {
	file := writeTestFile(t, "game.rar", append([]byte("PK\x03\x04"), make([]byte, 64)...))
	if got := detectArchiveKind(file); got != KindZip {
		t.Fatalf("kind = %s, want zip", got)
	}
}

func TestApkWithZipMagicIsZip(t *testing.T) {
	file := writeTestFile(t, "game.apk", append([]byte("PK\x03\x04"), make([]byte, 64)...))
	if got := detectArchiveKind(file); got != KindZip {
		t.Fatalf("kind = %s, want zip", got)
	}
}

func TestSevenZipAndGzipMagicAreRecognised(t *testing.T) {
	seven := writeTestFile(t, "game.7z", append([]byte{0x37, 0x7a, 0xbc, 0xaf, 0x27, 0x1c}, make([]byte, 64)...))
	gz := writeTestFile(t, "game.gz", []byte{0x1f, 0x8b, 0x08, 0x00})
	if got := detectArchiveKind(seven); got != KindSevenZip {
		t.Fatalf("7z kind = %s", got)
	}
	if got := detectArchiveKind(gz); got != KindGzip {
		t.Fatalf("gz kind = %s", got)
	}
}

func TestGzipWrappedTarballStaysTar(t *testing.T) {
	file := writeTestFile(t, "game.tar.gz", []byte{0x1f, 0x8b, 0x08, 0x00})
	if got := detectArchiveKind(file); got != KindTar {
		t.Fatalf("kind = %s, want tar", got)
	}
}

func TestTruncatedZipFallsBackToExtension(t *testing.T) {
	file := writeTestFile(t, "truncated.zip", []byte("not really a zip"))
	if got := detectArchiveKind(file); got != KindZip {
		t.Fatalf("kind = %s, want zip", got)
	}
}

func TestUnknownBytesWithUnknownExtensionStayUnknown(t *testing.T) {
	file := writeTestFile(t, "mystery.bin", []byte("hello"))
	if got := detectArchiveKind(file); got != KindUnknown {
		t.Fatalf("kind = %s, want unknown", got)
	}
}
