package downloads

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ErrUnsupportedArchive means the file is not an archive this package can open.
var ErrUnsupportedArchive = errors.New("unsupported archive")

// ExtractOptions receives progress and the name of the engine doing the work.
type ExtractOptions struct {
	OnProgress func(done, total int64) // total is 0 when the size is unknown
	OnEngine   func(engine string)     // called once, e.g. "7-Zip" or "built-in zip"
}

func (o ExtractOptions) engine(name string) {
	if o.OnEngine != nil {
		o.OnEngine(name)
	}
}

// Extract unpacks archive into dest. 7-Zip is used for every format when it is installed.
// Without it, zip and rar use the built-in readers, and 7z, tar and gzip fail with a clear message.
func Extract(ctx context.Context, archive, dest string, opts ExtractOptions) error {
	kind := detectArchiveKind(archive)
	if kind == KindUnknown {
		return fmt.Errorf("%w: %s", ErrUnsupportedArchive, filepath.Base(archive))
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	if bin := findSevenZip(); bin != "" {
		return extractWithSevenZip(ctx, bin, archive, dest, opts)
	}
	switch kind {
	case KindZip:
		return extractZip(ctx, archive, dest, opts)
	case KindRar:
		return extractRar(ctx, archive, dest, opts)
	default:
		return fmt.Errorf("install 7-Zip to extract %s files", kind)
	}
}

var driveLetter = regexp.MustCompile(`^[a-zA-Z]:$`)

// entryPath turns an archive entry name into a safe relative path. It fails on anything that
// escapes the destination and reports skip for junk that is never written.
func entryPath(name string) (rel string, skip bool, err error) {
	var parts []string
	for _, p := range strings.Split(strings.ReplaceAll(name, `\`, "/"), "/") {
		switch p {
		case "", ".":
			continue
		case "..":
			return "", false, fmt.Errorf("zip-slip: %s", name)
		}
		parts = append(parts, p)
	}
	if len(parts) == 0 || parts[0] == macOSJunkDir {
		return "", true, nil
	}
	if driveLetter.MatchString(parts[0]) {
		return "", false, fmt.Errorf("zip-slip: %s", name)
	}
	for _, p := range parts {
		if junkFiles[p] {
			return "", true, nil
		}
	}
	return strings.Join(parts, "/"), false, nil
}

// targetPath joins a checked relative path onto dest and refuses anything that lands outside it.
func targetPath(dest, rel string) (string, error) {
	root := filepath.Clean(dest)
	target := filepath.Join(root, filepath.FromSlash(rel))
	if !strings.HasPrefix(target, root+string(filepath.Separator)) {
		return "", fmt.Errorf("zip-slip: %s", rel)
	}
	return target, nil
}

// progressCounter adds up bytes and reports them. A zero total means the size is unknown.
type progressCounter struct {
	done   int64
	total  int64
	report func(done, total int64)
}

func (p *progressCounter) add(n int64) {
	p.setDone(p.done + n)
}

// setDone sets the absolute byte count, for tools that only report percentages.
func (p *progressCounter) setDone(done int64) {
	p.done = done
	if p.report != nil {
		p.report(p.done, p.total)
	}
}

// writeEntry copies src into a new file at target. It stops early when ctx ends.
func writeEntry(ctx context.Context, src io.Reader, target string, progress *progressCounter) (err error) {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := out.Close(); err == nil {
			err = closeErr
		}
	}()
	buf := make([]byte, copyBufferSize)
	_, err = io.CopyBuffer(out, &cancelableReader{ctx: ctx, src: src, progress: progress}, buf)
	return err
}

// cancelableReader checks the context and counts bytes on every read.
type cancelableReader struct {
	ctx      context.Context
	src      io.Reader
	progress *progressCounter
}

func (r *cancelableReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	n, err := r.src.Read(p)
	if n > 0 {
		r.progress.add(int64(n))
	}
	return n, err
}

// ensureDir creates an archive directory entry.
func ensureDir(target string) error {
	return os.MkdirAll(target, 0o755)
}
