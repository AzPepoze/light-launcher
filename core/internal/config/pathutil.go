package config

import (
	"os"
	"path/filepath"
	"strings"
)

// NormalizePath matches Node's path.normalize: it cleans the path and keeps a trailing separator.
func NormalizePath(p string) string {
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

// CleanPath trims and normalizes a user-supplied path. Empty input stays empty.
func CleanPath(p string) string {
	if p == "" {
		return ""
	}
	return NormalizePath(strings.TrimSpace(p))
}

// ExpandPath resolves a leading "~" to the home folder and normalizes the path.
func ExpandPath(p string) string {
	trimmed := strings.TrimSpace(p)
	if trimmed == "" {
		return ""
	}
	home, err := os.UserHomeDir()
	if err == nil {
		switch {
		case trimmed == "~":
			return home
		case strings.HasPrefix(trimmed, "~/"):
			return filepath.Join(home, trimmed[2:])
		}
	}
	return CleanPath(trimmed)
}

// PathKey normalizes a path for case-insensitive comparison. Both slash styles map to the same key.
func PathKey(p string) string {
	if p == "" {
		return ""
	}
	slashed := strings.ReplaceAll(p, `\`, "/")
	return strings.ToLower(filepath.Clean(filepath.FromSlash(slashed)))
}

// PathExists reports whether target exists. It never fails.
func PathExists(target string) bool {
	if target == "" {
		return false
	}
	_, err := os.Stat(ExpandPath(target))
	return err == nil
}

// BaseName matches Node's path.basename, which returns "" for an empty or root-only path.
func BaseName(p string) string {
	if strings.Trim(p, "/"+string(filepath.Separator)) == "" {
		return ""
	}
	return filepath.Base(p)
}

// NameWithoutExt returns the base name of p without its extension. A leading dot is part of the name.
func NameWithoutExt(p string) string {
	name, _ := splitExt(BaseName(p))
	return name
}

// ExtName returns the extension of p, including the dot, or "".
func ExtName(p string) string {
	_, ext := splitExt(BaseName(p))
	return ext
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

// FirstNonEmpty returns the first value that is not empty, or "".
func FirstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

// PrefixBaseDir returns the custom prefix folder when one is set, else the default prefixes folder.
func PrefixBaseDir(customDir string) string {
	if trimmed := strings.TrimSpace(customDir); trimmed != "" {
		return ExpandPath(trimmed)
	}
	return GetPrefixBaseDirectory()
}

// IsSubPath reports whether target is inside parent.
func IsSubPath(parent, target string) bool {
	if parent == "" || target == "" {
		return false
	}
	rel := RelativePath(parent, target)
	return !strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel) && rel != ""
}

// RelativePath matches Node's path.relative, which resolves both sides to absolute paths first.
func RelativePath(from, to string) string {
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
