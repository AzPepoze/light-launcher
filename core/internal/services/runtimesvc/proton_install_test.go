package runtimesvc

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"light-launcher/core/internal/config"
)

// buildProtonArchive returns a .tar.gz holding Proton-Test/proton, like a real Proton build.
func buildProtonArchive(t *testing.T) []byte {
	t.Helper()
	var compressed bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressed)
	tarWriter := tar.NewWriter(gzipWriter)

	content := []byte("#!/bin/sh\n")
	header := &tar.Header{Name: "Proton-Test/proton", Mode: 0755, Size: int64(len(content)), Typeflag: tar.TypeReg}
	if err := tarWriter.WriteHeader(header); err != nil {
		t.Fatalf("write tar header: %v", err)
	}
	if _, err := tarWriter.Write(content); err != nil {
		t.Fatalf("write tar body: %v", err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("close gzip: %v", err)
	}
	return compressed.Bytes()
}

func TestInstallProtonVersionExtractsArchive(t *testing.T) {
	skipUnlessLinux(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	archive := buildProtonArchive(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archive)
	}))
	defer server.Close()

	var messages []string
	var lastPercent int
	err := installProtonVersion(server.URL+"/Proton-Test.tar.gz", "Proton-Test", func(percent int, message string) {
		lastPercent = percent
		messages = append(messages, message)
	})
	if err != nil {
		t.Fatalf("installProtonVersion: %v", err)
	}

	installed := filepath.Join(protonTargetFolder(home), "Proton-Test", "proton")
	if !config.PathExists(installed) {
		t.Errorf("extracted proton not found at %s", installed)
	}
	if lastPercent != 100 || messages[len(messages)-1] != "Installation Complete!" {
		t.Errorf("last progress = %d %q, want 100 Installation Complete!", lastPercent, messages[len(messages)-1])
	}
}

func TestInstallProtonVersionReportsHTTPFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	err := installProtonVersion(server.URL+"/missing.tar.gz", "Proton-Missing", func(int, string) {})
	if err == nil || err.Error() != "Failed to install Proton Proton-Missing: Download failed with status 404" {
		t.Errorf("error = %v, want download status 404 message", err)
	}
}

func TestCheckReleaseStatus(t *testing.T) {
	tests := []struct {
		status  int
		wantErr string
	}{
		{http.StatusOK, ""},
		{http.StatusForbidden, "GitHub API rate limit exceeded. Please wait a few minutes before trying again."},
		{http.StatusTooManyRequests, "GitHub API rate limit exceeded. Please wait a few minutes before trying again."},
		{http.StatusInternalServerError, "GitHub API returned HTTP 500"},
	}
	for _, test := range tests {
		err := checkReleaseStatus(test.status)
		if test.wantErr == "" && err != nil {
			t.Errorf("status %d: unexpected error %v", test.status, err)
		}
		if test.wantErr != "" && (err == nil || err.Error() != test.wantErr) {
			t.Errorf("status %d: error = %v, want %q", test.status, err, test.wantErr)
		}
	}
}

func TestGetProtonReleasesRejectsUnknownVariant(t *testing.T) {
	if _, err := getProtonReleases("no-such-variant"); err == nil || err.Error() != "Unknown variant: no-such-variant" {
		t.Errorf("error = %v, want unknown variant error", err)
	}
}
