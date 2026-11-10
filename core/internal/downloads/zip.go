package downloads

import (
	"archive/zip"
	"context"
	"io/fs"
)

// extractZip uses the standard library. Every entry name is checked before the first byte is written.
func extractZip(ctx context.Context, archive, dest string, opts ExtractOptions) error {
	opts.engine("built-in zip")
	r, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer r.Close()

	total, err := zipPayloadSize(r)
	if err != nil {
		return err
	}
	progress := &progressCounter{total: total, report: opts.OnProgress}
	for _, f := range r.File {
		rel, skip, err := entryPath(f.Name)
		if err != nil {
			return err
		}
		if skip || isSkippedZipEntry(f) {
			continue
		}
		target, err := targetPath(dest, rel)
		if err != nil {
			return err
		}
		if f.FileInfo().IsDir() {
			if err := ensureDir(target); err != nil {
				return err
			}
			continue
		}
		if err := writeZipEntry(ctx, f, target, progress); err != nil {
			return err
		}
	}
	return nil
}

// zipPayloadSize sums the uncompressed size of the entries that will be written.
func zipPayloadSize(r *zip.ReadCloser) (int64, error) {
	var total int64
	for _, f := range r.File {
		_, skip, err := entryPath(f.Name)
		if err != nil {
			return 0, err
		}
		if !skip && !isSkippedZipEntry(f) && !f.FileInfo().IsDir() {
			total += int64(f.UncompressedSize64)
		}
	}
	return total, nil
}

// isSkippedZipEntry is true for symbolic links, which are never extracted.
func isSkippedZipEntry(f *zip.File) bool {
	return f.Mode()&fs.ModeSymlink != 0
}

func writeZipEntry(ctx context.Context, f *zip.File, target string, progress *progressCounter) error {
	src, err := f.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	return writeEntry(ctx, src, target, progress)
}
