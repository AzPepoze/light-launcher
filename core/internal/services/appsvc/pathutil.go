package appsvc

import (
	"path/filepath"
	"strings"
)

// normalizePath matches Node's path.normalize: it cleans the path and keeps a trailing separator.
func normalizePath(p string) string {
	if p == "" {
		return "."
	}
	cleaned := filepath.Clean(p)
	separator := string(filepath.Separator)
	if strings.HasSuffix(p, separator) && !strings.HasSuffix(cleaned, separator) {
		return cleaned + separator
	}
	return cleaned
}

// cleanPath trims and normalizes a user-supplied path. Empty input stays empty.
func cleanPath(p string) string {
	if p == "" {
		return ""
	}
	return normalizePath(strings.TrimSpace(p))
}

// isSubPath reports whether target is inside parent.
func isSubPath(parent, target string) bool {
	if parent == "" || target == "" {
		return false
	}
	rel := relativePath(parent, target)
	return !strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel) && rel != ""
}

// relativePath matches Node's path.relative, which resolves both sides to absolute paths first.
func relativePath(from, to string) string {
	absFrom, errFrom := filepath.Abs(from)
	absTo, errTo := filepath.Abs(to)
	if errFrom != nil || errTo != nil {
		return ""
	}
	rel, err := filepath.Rel(absFrom, absTo)
	if err != nil || rel == "." {
		return ""
	}
	return rel
}

// baseName matches Node's path.basename, which returns "" for an empty or root-only path.
func baseName(p string) string {
	if strings.Trim(p, "/"+string(filepath.Separator)) == "" {
		return ""
	}
	return filepath.Base(p)
}

// splitExt matches Node's path.parse for a base name: the dot at index 0 is not an extension.
func splitExt(base string) (name, ext string) {
	if base == ".." {
		return base, ""
	}
	dot := strings.LastIndex(base, ".")
	if dot <= 0 {
		return base, ""
	}
	return base[:dot], base[dot:]
}

// fileNameWithoutExt returns the file name of p without its extension.
func fileNameWithoutExt(p string) string {
	name, _ := splitExt(baseName(p))
	return name
}

// extName returns the extension of p, including the dot, or "".
func extName(p string) string {
	_, ext := splitExt(baseName(p))
	return ext
}

// firstNonEmpty returns the first value that is not empty, or "".
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
