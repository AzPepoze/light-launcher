package plugins

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// packFolder writes a folder with plugin.json as a package zip at dst. The folder's own SHA256SUMS
// and SIGNATURE are left out, and a fresh SHA256SUMS is added, so the package is unsigned and its
// checksums are still checked like any other package.
func packFolder(dir, dst string) error {
	if _, err := os.Stat(filepath.Join(dir, "plugin.json")); err != nil {
		return fmt.Errorf("folder has no plugin.json")
	}
	names, err := folderFiles(dir)
	if err != nil {
		return err
	}
	sums, err := folderChecksums(dir, names)
	if err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	writer := zip.NewWriter(out)
	for _, name := range names {
		if err := addFileToZip(writer, dir, name); err != nil {
			return err
		}
	}
	if err := addBytesToZip(writer, checksumsFile, sums); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return out.Close()
}

// folderFiles lists the slash paths of the regular files in dir, sorted. Links and other special
// files are refused, and the root SHA256SUMS and SIGNATURE are skipped.
func folderFiles(dir string) ([]string, error) {
	var names []string
	err := filepath.WalkDir(dir, func(full string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("folder entry %q is not a regular file", entry.Name())
		}
		rel, err := filepath.Rel(dir, full)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)
		if name == checksumsFile || name == signatureFile {
			return nil
		}
		names = append(names, name)
		return nil
	})
	sort.Strings(names)
	return names, err
}

// folderChecksums returns a SHA256SUMS file for names, in the format parseChecksums reads.
func folderChecksums(dir string, names []string) ([]byte, error) {
	var lines strings.Builder
	for _, name := range names {
		digest, err := fileSHA256(filepath.Join(dir, filepath.FromSlash(name)))
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(&lines, "%s  %s\n", digest, name)
	}
	return []byte(lines.String()), nil
}

func addFileToZip(writer *zip.Writer, dir, name string) error {
	info, err := os.Stat(filepath.Join(dir, filepath.FromSlash(name)))
	if err != nil {
		return err
	}
	header := &zip.FileHeader{Name: name, Method: zip.Deflate}
	header.SetMode(info.Mode().Perm())
	out, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}
	in, err := os.Open(filepath.Join(dir, filepath.FromSlash(name)))
	if err != nil {
		return err
	}
	defer in.Close()
	_, err = io.Copy(out, in)
	return err
}

func addBytesToZip(writer *zip.Writer, name string, data []byte) error {
	out, err := writer.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Deflate})
	if err != nil {
		return err
	}
	_, err = out.Write(data)
	return err
}
