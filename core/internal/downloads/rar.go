package downloads

import (
	"context"
	"errors"
	"io"

	"github.com/nwaples/rardecode/v2"
)

// extractRar reads RAR entries with the built-in decoder. Its total size is unknown up front,
// so progress is reported as bytes written.
func extractRar(ctx context.Context, archive, dest string, opts ExtractOptions) error {
	opts.engine("built-in RAR")
	rc, err := rardecode.OpenReader(archive)
	if err != nil {
		return err
	}
	defer rc.Close()

	progress := &progressCounter{report: opts.OnProgress}
	for {
		hdr, err := rc.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		rel, skip, err := entryPath(hdr.Name)
		if err != nil {
			return err
		}
		if skip || hdr.LinkType != 0 {
			continue
		}
		target, err := targetPath(dest, rel)
		if err != nil {
			return err
		}
		if hdr.IsDir {
			if err := ensureDir(target); err != nil {
				return err
			}
			continue
		}
		if err := writeEntry(ctx, rc, target, progress); err != nil {
			return err
		}
	}
}
