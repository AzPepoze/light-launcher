package downloads

import (
	"mime"
	"net/url"
	"path"
	"regexp"
	"strings"
)

const (
	maxFilenameRunes = 180
	fallbackFilename = "download.bin"
)

var (
	illegalFilenameChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	extendedDisposition  = regexp.MustCompile(`(?i)filename\*\s*=\s*(?:([\w-]+)'[^']*')?([^;]+)`)
	quotedDisposition    = regexp.MustCompile(`(?i)filename\s*=\s*"([^"]+)"`)
	plainDisposition     = regexp.MustCompile(`(?i)filename\s*=\s*([^;]+)`)
	genericStem          = regexp.MustCompile(`\.[a-z0-9]{1,8}$`)
	hasExtension         = regexp.MustCompile(`(?i)\.[a-z0-9]{2,5}$`)
)

// genericStems are names hosts use for every file, so they say nothing about the game.
var genericStems = map[string]bool{
	"download": true, "downloads": true, "file": true, "files": true, "get": true,
	"attachment": true, "attachments": true, "index": true, "api": true, "data": true,
	"blob": true, "unknown": true, "archive": true, "setup": true, "install": true,
	"installer": true, "redirect": true, "link": true, "media": true, "content": true,
	"proxy": true, "stream": true,
}

// fileHostStems are file host names that a mirror label often uses instead of a file name.
var fileHostStems = map[string]bool{
	"pixeldrain": true, "mediafire": true, "mega": true, "mega.nz": true, "gofile": true,
	"datanodes": true, "krakenfiles": true, "workupload": true, "mixdrop": true,
	"vikingfile": true, "buzzheavier": true, "bzzhr": true, "yourfilestore": true,
	"dropmefiles": true, "wdho": true, "bunkr": true,
}

// extensionByType maps the content types seen on download hosts to file extensions.
var extensionByType = map[string]string{
	"application/zip":                               ".zip",
	"application/x-zip-compressed":                  ".zip",
	"application/vnd.rar":                           ".rar",
	"application/x-rar-compressed":                  ".rar",
	"application/x-7z-compressed":                   ".7z",
	"application/x-tar":                             ".tar",
	"application/gzip":                              ".gz",
	"application/x-gzip":                            ".gz",
	"application/x-msdownload":                      ".exe",
	"application/vnd.microsoft.portable-executable": ".exe",
	"application/x-msi":                             ".msi",
	"application/vnd.android.package-archive":       ".apk",
}

// sanitizeFilename keeps one safe path component, so a name can never leave the download folder.
func sanitizeFilename(name string) string {
	base := path.Base(strings.ReplaceAll(name, `\`, "/"))
	if base == "." || base == "/" {
		base = ""
	}
	base = strings.TrimSpace(strings.TrimLeft(illegalFilenameChars.ReplaceAllString(base, "_"), "."))
	cleaned := strings.TrimRight(base, ". ")
	if cleaned == "" {
		return fallbackFilename
	}
	return truncateKeepingExtension(cleaned)
}

// truncateKeepingExtension shortens long names but keeps the extension so archive detection still works.
func truncateKeepingExtension(name string) string {
	runes := []rune(name)
	if len(runes) <= maxFilenameRunes {
		return name
	}
	ext := path.Ext(name)
	if extRunes := []rune(ext); ext != "" && len(extRunes) <= 10 {
		return string(runes[:maxFilenameRunes-len(extRunes)]) + ext
	}
	return string(runes[:maxFilenameRunes])
}

// filenameFromDisposition reads the name from a Content-Disposition header, including the UTF-8 filename* form.
func filenameFromDisposition(value string) string {
	if value == "" {
		return ""
	}
	if m := extendedDisposition.FindStringSubmatch(value); m != nil {
		raw := strings.Trim(strings.TrimSpace(m[2]), `"`)
		if decoded, err := url.PathUnescape(raw); err == nil {
			return sanitizeFilename(decoded)
		}
		return sanitizeFilename(raw)
	}
	if m := quotedDisposition.FindStringSubmatch(value); m != nil {
		return sanitizeFilename(m[1])
	}
	if m := plainDisposition.FindStringSubmatch(value); m != nil {
		return sanitizeFilename(strings.TrimSpace(m[1]))
	}
	return ""
}

// usefulFilename returns the sanitized name, or "" when the name is generic.
func usefulFilename(name string) string {
	if name == "" {
		return ""
	}
	clean := sanitizeFilename(name)
	if isGenericFilename(clean) {
		return ""
	}
	return clean
}

func isGenericFilename(name string) bool {
	lower := strings.ToLower(sanitizeFilename(name))
	if lower == fallbackFilename {
		return true
	}
	stem := genericStem.ReplaceAllString(lower, "")
	return stem == "" || genericStems[stem] || fileHostStems[stem]
}

// ensureFilenameExtension adds an extension from the content type when the name has none.
func ensureFilenameExtension(name, contentType string) string {
	if hasExtension.MatchString(name) {
		return name
	}
	if ext := extensionForContentType(contentType); ext != "" {
		return name + ext
	}
	return name + ".bin"
}

func extensionForContentType(value string) string {
	if value == "" {
		return ""
	}
	mediaType, _, err := mime.ParseMediaType(value)
	if err != nil {
		mediaType = strings.ToLower(strings.TrimSpace(strings.Split(value, ";")[0]))
	}
	return extensionByType[mediaType]
}
