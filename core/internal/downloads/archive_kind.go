package downloads

import (
	"bytes"
	"io"
	"os"
	"strings"
)

// ArchiveKind is the format of a downloaded archive.
type ArchiveKind string

const (
	KindUnknown  ArchiveKind = "unknown"
	KindZip      ArchiveKind = "zip"
	KindRar      ArchiveKind = "rar"
	KindSevenZip ArchiveKind = "7z"
	KindTar      ArchiveKind = "tar"
	KindGzip     ArchiveKind = "gzip"
)

var (
	magicRar4     = []byte{0x52, 0x61, 0x72, 0x21, 0x1a, 0x07, 0x00}
	magicRar5     = []byte{0x52, 0x61, 0x72, 0x21, 0x1a, 0x07, 0x01, 0x00}
	magicSevenZip = []byte{0x37, 0x7a, 0xbc, 0xaf, 0x27, 0x1c}
)

const magicHeaderSize = 512

// detectArchiveKind trusts the bytes over the extension, because some mirrors serve a RAR under a .zip name.
// A tar.gz is a gzip wrapper around a tar, so it stays a tar.
func detectArchiveKind(file string) ArchiveKind {
	name := strings.ToLower(file)
	byName := kindByName(name)
	byMagic := kindByMagic(file)
	if byName == KindTar && byMagic == KindGzip {
		return KindTar
	}
	if byMagic != KindUnknown && byMagic != byName {
		return byMagic
	}
	if byName != KindUnknown {
		return byName
	}
	return byMagic
}

func kindByName(name string) ArchiveKind {
	switch {
	case strings.HasSuffix(name, ".zip"):
		return KindZip
	case strings.HasSuffix(name, ".rar"):
		return KindRar
	case strings.HasSuffix(name, ".7z"):
		return KindSevenZip
	case strings.HasSuffix(name, ".tar"), strings.HasSuffix(name, ".tar.gz"), strings.HasSuffix(name, ".tgz"):
		return KindTar
	case strings.HasSuffix(name, ".gz"), strings.HasSuffix(name, ".gzip"):
		return KindGzip
	}
	return KindUnknown
}

func kindByMagic(file string) ArchiveKind {
	f, err := os.Open(file)
	if err != nil {
		return KindUnknown
	}
	defer f.Close()
	head := make([]byte, magicHeaderSize)
	n, err := io.ReadFull(f, head)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return KindUnknown
	}
	return kindOfHeader(head[:n])
}

func kindOfHeader(b []byte) ArchiveKind {
	switch {
	case isZipHeader(b):
		return KindZip
	case bytes.HasPrefix(b, magicRar5), bytes.HasPrefix(b, magicRar4):
		return KindRar
	case bytes.HasPrefix(b, magicSevenZip):
		return KindSevenZip
	case len(b) >= 2 && b[0] == 0x1f && b[1] == 0x8b:
		return KindGzip
	case len(b) >= 262 && string(b[257:262]) == "ustar":
		return KindTar
	}
	return KindUnknown
}

// isZipHeader matches a local file, end-of-central-directory, or spanned-archive signature.
func isZipHeader(b []byte) bool {
	if len(b) < 4 || b[0] != 'P' || b[1] != 'K' {
		return false
	}
	return (b[2] == 3 && b[3] == 4) || (b[2] == 5 && b[3] == 6) || (b[2] == 7 && b[3] == 8)
}
