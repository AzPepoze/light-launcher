package server

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"light-launcher/core/internal/logger"
)

// Upload limits. They are variables so tests can lower them.
var (
	maxUploadBytes int64 = 512 << 20
	maxUploadFiles       = 5000
)

// uploadError is a refused upload, with the HTTP status for the reply.
type uploadError struct {
	status int
	msg    string
}

func (e *uploadError) Error() string { return e.msg }

func refuse(msg string) error {
	return &uploadError{status: http.StatusBadRequest, msg: msg}
}

// handleUpload stores the files of a plugin drop and replies with the path to install.
func (s *server) handleUpload(w http.ResponseWriter, r *http.Request) {
	folder, err := s.cfg.Uploads.NewFolder()
	if err != nil {
		logger.Error("upload", "%v", err)
		writeJSON(w, http.StatusInternalServerError, envelope{Error: "could not store the upload"})
		return
	}
	path, err := storeUpload(w, r, folder)
	if err != nil {
		if rmErr := s.cfg.Uploads.Remove(folder); rmErr != nil {
			logger.Warn("upload", "remove refused upload: %v", rmErr)
		}
		writeUploadError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, envelope{Success: true, Data: map[string]string{"path": path}})
}

// storeUpload writes every file part of the request into folder and returns the path to install.
func storeUpload(w http.ResponseWriter, r *http.Request, folder string) (string, error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	reader, err := r.MultipartReader()
	if err != nil {
		return "", refuse("upload must be multipart/form-data")
	}
	var names []string
	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", readError(err)
		}
		if len(names) == maxUploadFiles {
			return "", &uploadError{status: http.StatusRequestEntityTooLarge, msg: fmt.Sprintf("upload has more than %d files", maxUploadFiles)}
		}
		name, err := writeUploadPart(folder, part)
		if err != nil {
			return "", err
		}
		names = append(names, name)
	}
	if len(names) == 0 {
		return "", refuse("upload has no files")
	}
	return installPath(folder, names)
}

// readError keeps a too-large body as is, and treats any other read failure as a bad body.
func readError(err error) error {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		return err
	}
	return refuse("bad multipart body")
}

// writeUploadError sends the reply for a failed upload.
func writeUploadError(w http.ResponseWriter, err error) {
	var tooLarge *http.MaxBytesError
	var refused *uploadError
	switch {
	case errors.As(err, &tooLarge):
		writeJSON(w, http.StatusRequestEntityTooLarge, envelope{Error: "upload is larger than 512 MiB"})
	case errors.As(err, &refused):
		writeJSON(w, refused.status, envelope{Error: refused.msg})
	default:
		logger.Error("upload", "%v", err)
		writeJSON(w, http.StatusInternalServerError, envelope{Error: "could not store the upload"})
	}
}

// writeUploadPart saves one file part under folder and returns its relative path.
func writeUploadPart(folder string, part *multipart.Part) (string, error) {
	name, err := partPath(part)
	if err != nil {
		return "", err
	}
	if err := createUploadFile(folder, name, part); err != nil {
		return "", err
	}
	return name, nil
}

// partPath returns the relative path the UI sent as the file name. The multipart package keeps
// only the last element of a file name, so the raw Content-Disposition value is read instead.
func partPath(part *multipart.Part) (string, error) {
	_, params, err := mime.ParseMediaType(part.Header.Get("Content-Disposition"))
	name, ok := params["filename"]
	if err != nil || !ok {
		return "", refuse("every part must be a file")
	}
	if err := checkRelativePath(name); err != nil {
		return "", err
	}
	return name, nil
}

// checkRelativePath refuses names that are empty, absolute, have a drive or backslash, or contain
// "." and ".." parts.
func checkRelativePath(name string) error {
	bad := name == "" || strings.ContainsAny(name, "\\:\x00") || strings.HasPrefix(name, "/") || strings.HasSuffix(name, "/")
	for _, segment := range strings.Split(name, "/") {
		if segment == "" || segment == "." || segment == ".." {
			bad = true
		}
	}
	if bad {
		return refuse(fmt.Sprintf("bad file path %q", name))
	}
	return nil
}

// createUploadFile writes body to root/name. Each folder on the way must be a real folder, and the file
// must not exist yet, so a link or an earlier file can never be written through.
func createUploadFile(root, name string, body io.Reader) error {
	segments := strings.Split(name, "/")
	dir := root
	for _, segment := range segments[:len(segments)-1] {
		dir = filepath.Join(dir, segment)
		if err := ensureRealDir(dir); err != nil {
			return err
		}
	}
	file, err := os.OpenFile(filepath.Join(dir, segments[len(segments)-1]), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, fs.ErrExist) {
		return refuse(fmt.Sprintf("file %q is sent twice", name))
	}
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, body)
	return err
}

// ensureRealDir creates dir, or accepts it when it is already a real folder. A file or link is refused.
func ensureRealDir(dir string) error {
	info, err := os.Lstat(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return os.Mkdir(dir, 0o755)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return refuse(fmt.Sprintf("path %q is not a folder", filepath.Base(dir)))
	}
	return nil
}

// installPath returns what to install from folder: the one package file sent, or the folder that holds plugin.json.
func installPath(folder string, names []string) (string, error) {
	if len(names) == 1 && isPackageFile(names[0]) {
		return filepath.Join(folder, names[0]), nil
	}
	if fileExists(filepath.Join(folder, "plugin.json")) {
		return folder, nil
	}
	if top, ok := singleTopFolder(names); ok && fileExists(filepath.Join(folder, top, "plugin.json")) {
		return filepath.Join(folder, top), nil
	}
	return "", refuse("upload has no plugin.json and no .llplugin or .zip file")
}

func isPackageFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".llplugin" || ext == ".zip"
}

// singleTopFolder returns the one folder that every name is inside, if there is one.
func singleTopFolder(names []string) (string, bool) {
	top := ""
	for _, name := range names {
		first, _, nested := strings.Cut(name, "/")
		if !nested || (top != "" && first != top) {
			return "", false
		}
		top = first
	}
	return top, top != ""
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
