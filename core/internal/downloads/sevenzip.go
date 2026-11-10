package downloads

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// sevenZipNames are the commands tried in order. 7zz is the modern 7-Zip build on Linux.
var sevenZipNames = []string{"7zz", "7z", "7za"}

var (
	sevenZipPath    = regexp.MustCompile(`(?m)^Path = (.*)$`)
	sevenZipSize    = regexp.MustCompile(`(?m)^Size = (\d+)$`)
	sevenZipDivider = regexp.MustCompile(`(?m)^-{10,}$`)
	sevenZipPercent = regexp.MustCompile(`(\d{1,3})%`)
)

// HasSevenZip reports whether a 7-Zip command is on PATH.
func HasSevenZip() bool {
	return findSevenZip() != ""
}

func findSevenZip() string {
	for _, name := range sevenZipNames {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	return ""
}

// extractWithSevenZip lists the archive first and refuses it if any entry escapes dest.
func extractWithSevenZip(ctx context.Context, bin, archive, dest string, opts ExtractOptions) error {
	total, err := listSevenZipEntries(ctx, bin, archive)
	if err != nil {
		return err
	}
	opts.engine("7-Zip")

	cmd := exec.CommandContext(ctx, bin, "x", "-y", "-mmt=on", "-bsp1", "-bso0", "-o"+dest, archive)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	progress := &progressCounter{total: total, report: opts.OnProgress}
	reportSevenZipProgress(stdout, progress)
	waitErr := cmd.Wait()
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	// exit code 1 is a warning that some files were skipped, so it counts as a failure too
	if waitErr != nil {
		return fmt.Errorf("%s failed: %s", filepath.Base(bin), firstLine(stderr.String(), waitErr))
	}
	progress.setDone(total)
	return cleanTree(dest)
}

// listSevenZipEntries reads the entry list, rejects unsafe paths and returns the total uncompressed size.
func listSevenZipEntries(ctx context.Context, bin, archive string) (int64, error) {
	out, err := exec.CommandContext(ctx, bin, "l", "-slt", archive).Output()
	if err != nil {
		return 0, fmt.Errorf("%s could not list the archive", filepath.Base(bin))
	}
	entries := entryBlocks(string(out))
	for _, m := range sevenZipPath.FindAllStringSubmatch(entries, -1) {
		name := m[1]
		if strings.HasPrefix(name, "/") || strings.HasPrefix(name, `\`) {
			return 0, fmt.Errorf("zip-slip: %s", name)
		}
		if _, _, err := entryPath(name); err != nil {
			return 0, err
		}
	}
	var total int64
	for _, m := range sevenZipSize.FindAllStringSubmatch(entries, -1) {
		size, _ := strconv.ParseInt(m[1], 10, 64)
		total += size
	}
	return total, nil
}

// entryBlocks returns the part of `7z l -slt` output that holds the entries, which starts at the first dashed line.
func entryBlocks(listing string) string {
	loc := sevenZipDivider.FindStringIndex(listing)
	if loc == nil {
		return ""
	}
	return listing[loc[0]:]
}

// reportSevenZipProgress reads the percentages 7-Zip prints on stdout and reports the last one per chunk.
func reportSevenZipProgress(stdout io.Reader, progress *progressCounter) {
	buf := make([]byte, 4096)
	for {
		n, err := stdout.Read(buf)
		if n > 0 {
			matches := sevenZipPercent.FindAllSubmatch(buf[:n], -1)
			if len(matches) > 0 {
				last := matches[len(matches)-1]
				pct, _ := strconv.ParseInt(string(last[1]), 10, 64)
				progress.setDone(progress.total * pct / 100)
			}
		}
		if err != nil {
			return
		}
	}
}

// firstLine picks a short error message from 7-Zip's stderr, falling back to the exit error.
func firstLine(stderr string, fallback error) string {
	line := strings.TrimSpace(stderr)
	if before, _, found := strings.Cut(line, "\n"); found {
		line = before
	}
	if line == "" {
		return fallback.Error()
	}
	return line
}
