package server

import (
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// serveUI serves built files and falls back to index.html so client-side routes still load.
func serveUI(ui fs.FS) http.Handler {
	files := http.FileServerFS(ui)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isUIFile(ui, r.URL.Path) {
			files.ServeHTTP(w, r)
			return
		}
		http.ServeFileFS(w, r, ui, "index.html")
	})
}

// isUIFile reports whether urlPath names a file. The root and folders fall back to index.html.
func isUIFile(ui fs.FS, urlPath string) bool {
	name := strings.TrimPrefix(path.Clean("/"+urlPath), "/")
	if name == "" {
		return false
	}
	info, err := fs.Stat(ui, name)
	return err == nil && !info.IsDir()
}

// contentSecurityPolicy allows the app's own files, data and blob images, and the local socket origin only.
func contentSecurityPolicy(port int) string {
	return fmt.Sprintf("default-src 'self'; img-src 'self' data: blob:; connect-src 'self' ws://127.0.0.1:%d", port)
}
