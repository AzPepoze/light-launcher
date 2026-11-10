package appsvc

import (
	"io"
	"os"
	"path/filepath"
)

// moveOrCopy renames src to dest. When a rename is not possible (another disk), it copies and deletes src.
func moveOrCopy(src, dest string) error {
	if err := os.Rename(src, dest); err == nil {
		return nil
	}
	if err := copyTree(src, dest); err != nil {
		return err
	}
	_ = os.RemoveAll(src)
	return nil
}

// copyTree copies a file or folder. Symlinks are copied as links, like fs.cp without dereference.
func copyTree(src, dest string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	switch {
	case info.Mode()&os.ModeSymlink != 0:
		return copySymlink(src, dest)
	case info.IsDir():
		return copyFolder(src, dest, info.Mode().Perm())
	default:
		return copyFile(src, dest, info.Mode().Perm())
	}
}

func copySymlink(src, dest string) error {
	target, err := os.Readlink(src)
	if err != nil {
		return err
	}
	_ = os.Remove(dest)
	return os.Symlink(target, dest)
}

func copyFolder(src, dest string, perm os.FileMode) error {
	if err := os.MkdirAll(dest, perm|0o700); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := copyTree(filepath.Join(src, entry.Name()), filepath.Join(dest, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dest string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
