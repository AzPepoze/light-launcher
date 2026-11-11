package server

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type uploadPart struct {
	name string
	body string
}

// uploadRequest builds a multipart upload with one file part per entry.
func uploadRequest(t *testing.T, fx *fixture, parts ...uploadPart) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, part := range parts {
		out, err := writer.CreateFormFile("files", part.name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := out.Write([]byte(part.body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	req := fx.request(http.MethodPost, "/api/plugin-upload", body.String())
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

// lowerLimits sets the upload limits for one test and restores them after.
func lowerLimits(t *testing.T, bytesLimit int64, filesLimit int) {
	t.Helper()
	oldBytes, oldFiles := maxUploadBytes, maxUploadFiles
	maxUploadBytes, maxUploadFiles = bytesLimit, filesLimit
	t.Cleanup(func() {
		maxUploadBytes, maxUploadFiles = oldBytes, oldFiles
	})
}

func TestUploadZipPackageReturnsFile(t *testing.T) {
	fx := newFixture(t, testPort)
	rec := fx.serve(uploadRequest(t, fx, uploadPart{name: "demo.llplugin", body: "package-bytes"}))
	env := decodeEnvelope(t, rec)
	if rec.Code != http.StatusOK || !env.Success {
		t.Fatalf("status %d, envelope %+v", rec.Code, env)
	}
	path := env.Data.(map[string]any)["path"].(string)
	if !filepath.IsAbs(path) || filepath.Base(path) != "demo.llplugin" || !strings.HasPrefix(path, fx.uploadDir) {
		t.Fatalf("path = %q, want demo.llplugin inside %q", path, fx.uploadDir)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "package-bytes" {
		t.Fatalf("stored file = %q, %v", data, err)
	}
}

func TestUploadFolderReturnsFolderWithPluginJSON(t *testing.T) {
	fx := newFixture(t, testPort)
	rec := fx.serve(uploadRequest(t, fx,
		uploadPart{name: "myplugin/plugin.json", body: "{}"},
		uploadPart{name: "myplugin/bin/linux-amd64/demo", body: "binary"},
	))
	env := decodeEnvelope(t, rec)
	if rec.Code != http.StatusOK || !env.Success {
		t.Fatalf("status %d, envelope %+v", rec.Code, env)
	}
	path := env.Data.(map[string]any)["path"].(string)
	if filepath.Base(path) != "myplugin" || !strings.HasPrefix(path, fx.uploadDir) {
		t.Fatalf("path = %q, want the myplugin folder", path)
	}
	if !fileExists(filepath.Join(path, "bin", "linux-amd64", "demo")) {
		t.Fatalf("folder %q lost its files", path)
	}
}

func TestUploadRootPluginJSONReturnsUploadFolder(t *testing.T) {
	fx := newFixture(t, testPort)
	rec := fx.serve(uploadRequest(t, fx, uploadPart{name: "plugin.json", body: "{}"}))
	env := decodeEnvelope(t, rec)
	if rec.Code != http.StatusOK || !env.Success {
		t.Fatalf("status %d, envelope %+v", rec.Code, env)
	}
	path := env.Data.(map[string]any)["path"].(string)
	if filepath.Dir(path) != fx.uploadDir || !fileExists(filepath.Join(path, "plugin.json")) {
		t.Fatalf("path = %q, want the upload folder that holds plugin.json", path)
	}
}

func TestUploadRefusesBadRequests(t *testing.T) {
	names := []string{"../evil.llplugin", "/abs/evil.llplugin", `a\b.llplugin`, "C:evil.llplugin", "pkg/../../evil.llplugin", "pkg//x.zip", "./x.zip", "pkg/./x.zip"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			fx := newFixture(t, testPort)
			rec := fx.serve(uploadRequest(t, fx, uploadPart{name: name, body: "x"}))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", rec.Code)
			}
			if fileExists(filepath.Join(filepath.Dir(fx.uploadDir), "evil.llplugin")) {
				t.Fatalf("a refused upload wrote outside the store")
			}
			if entries, _ := os.ReadDir(fx.uploadDir); len(entries) != 0 {
				t.Fatalf("a refused upload left %d folders behind", len(entries))
			}
		})
	}
}

func TestUploadRefusesFolderWithoutPluginJSON(t *testing.T) {
	fx := newFixture(t, testPort)
	rec := fx.serve(uploadRequest(t, fx, uploadPart{name: "myplugin/bin/demo", body: "binary"}))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestUploadRefusesNonMultipartBody(t *testing.T) {
	fx := newFixture(t, testPort)
	rec := fx.serve(fx.request(http.MethodPost, "/api/plugin-upload", `{"a":1}`))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestUploadNeedsToken(t *testing.T) {
	fx := newFixture(t, testPort)
	req := uploadRequest(t, fx, uploadPart{name: "demo.llplugin", body: "x"})
	req.Header.Del(tokenHeader)
	if rec := fx.serve(req); rec.Code != http.StatusUnauthorized {
		t.Fatalf("status without token = %d, want 401", rec.Code)
	}
}

func TestUploadRefusesOversizeBody(t *testing.T) {
	lowerLimits(t, 1024, maxUploadFiles)
	fx := newFixture(t, testPort)
	rec := fx.serve(uploadRequest(t, fx, uploadPart{name: "demo.llplugin", body: strings.Repeat("x", 4096)}))
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", rec.Code)
	}
}

func TestUploadRefusesTooManyFiles(t *testing.T) {
	lowerLimits(t, maxUploadBytes, 2)
	fx := newFixture(t, testPort)
	rec := fx.serve(uploadRequest(t, fx,
		uploadPart{name: "p/a", body: "a"},
		uploadPart{name: "p/b", body: "b"},
		uploadPart{name: "p/c", body: "c"},
	))
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", rec.Code)
	}
}
