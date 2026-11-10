package plugins

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	idPattern         = regexp.MustCompile(`^[a-z][a-z0-9-]{1,39}$`)
	itemIDPattern     = regexp.MustCompile(`^[a-z][a-z0-9-]{0,39}$`)
	versionPattern    = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)
	apiVersionPattern = regexp.MustCompile(`^[0-9]+$`)
	entryKeyPattern   = regexp.MustCompile(`^(linux|windows|mac)-(x64|arm64)$`)
	hostPattern       = regexp.MustCompile(`^(\*\.)?[a-z0-9-]+(\.[a-z0-9-]+)+$`)
	fsNamePattern     = regexp.MustCompile(`^\$(GAMES|DOWNLOADS|DATA|APPDATA_[A-Z][A-Z0-9_]*)$`)
	methodPattern     = regexp.MustCompile(`^[a-z][a-z0-9-]*(\.[A-Za-z][A-Za-z0-9]*)+$`)
)

var validCapabilities = map[string]bool{
	"browse": true, "auth": true, "updateCheck": true, "resolveFiles": true,
	"sync:games": true, "sync:index": true,
}

var validHooks = map[string]bool{
	"onGameLaunch": true, "onGameExit": true, "onGameInstalled": true,
}

// Validate checks every field against the rules in schema/plugin.schema.json.
func (m Manifest) Validate() error {
	checks := []func() error{
		m.validateIdentity,
		m.validateInfo,
		m.validateEntry,
		m.validateContributes,
		m.validatePermissions,
	}
	for _, check := range checks {
		if err := check(); err != nil {
			return err
		}
	}
	return nil
}

func (m Manifest) validateIdentity() error {
	if !idPattern.MatchString(m.ID) {
		return fieldError("id", "must be 2 to 40 lowercase letters, digits or dashes, starting with a letter")
	}
	if n := len(m.Name); n < 1 || n > 80 {
		return fieldError("name", "must be 1 to 80 characters")
	}
	if !versionPattern.MatchString(m.Version) {
		return fieldError("version", "must be a semver version such as 1.0.0")
	}
	if !apiVersionPattern.MatchString(m.APIVersion) {
		return fieldError("apiVersion", "must be a major version number such as \"1\"")
	}
	if _, err := parseSemver(m.MinHost); err != nil {
		return fieldError("minHost", "must be a semver version such as 2.0.0")
	}
	return nil
}

// validateInfo checks the optional author, description and homepage. The SDK schema may not list them yet.
func (m Manifest) validateInfo() error {
	if len(m.Author) > 80 {
		return fieldError("author", "must be at most 80 characters")
	}
	if len(m.Description) > 1000 {
		return fieldError("description", "must be at most 1000 characters")
	}
	if m.Homepage != "" && !isHTTPURL(m.Homepage) {
		return fieldError("homepage", "must be an http or https URL")
	}
	return nil
}

func isHTTPURL(raw string) bool {
	parsed, err := url.Parse(raw)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}

func (m Manifest) validateEntry() error {
	if len(m.Entry) == 0 {
		return fieldError("entry", "must name at least one platform")
	}
	for key, path := range m.Entry {
		if !entryKeyPattern.MatchString(key) {
			return fieldError("entry."+key, "unknown platform key, use linux-x64, windows-x64, mac-arm64 and so on")
		}
		if !isRelativePath(path) {
			return fieldError("entry."+key, "must be a relative path inside the package without \"..\"")
		}
	}
	return nil
}

func (m Manifest) validateContributes() error {
	c := m.Contributes
	if err := validateSources(c.Sources); err != nil {
		return err
	}
	if c.Settings != nil && !isRelativePath(c.Settings.Ref) {
		return fieldError("contributes.settings.$ref", "must be a relative path inside the package")
	}
	for i, page := range c.Pages {
		if err := validateItem(fmt.Sprintf("contributes.pages[%d]", i), page.ID, page.Title); err != nil {
			return err
		}
		if !isRelativePath(page.Path) {
			return fieldError(fmt.Sprintf("contributes.pages[%d].path", i), "must be a relative path inside the package")
		}
	}
	if err := validateCommands(c.Commands); err != nil {
		return err
	}
	for _, group := range [][]NamedItem{c.LaunchAdapters, c.Importers} {
		if err := validateNamedItems(group); err != nil {
			return err
		}
	}
	return validateHooksAndThemes(c)
}

func validateSources(sources []SourceContribution) error {
	seen := map[string]bool{}
	for i, source := range sources {
		field := fmt.Sprintf("contributes.sources[%d]", i)
		if err := validateItem(field, source.ID, source.Label); err != nil {
			return err
		}
		if seen[source.ID] {
			return fieldError(field+".id", "duplicate source id %q", source.ID)
		}
		seen[source.ID] = true
		if err := validateCapabilities(field, source.Capabilities); err != nil {
			return err
		}
	}
	return nil
}

// validateCapabilities checks names, duplicates, and that sync:games and sync:index are not both set.
func validateCapabilities(field string, capabilities []string) error {
	seen := map[string]bool{}
	for _, capability := range capabilities {
		if !validCapabilities[capability] {
			return fieldError(field+".capabilities", "unknown capability %q", capability)
		}
		if seen[capability] {
			return fieldError(field+".capabilities", "duplicate capability %q", capability)
		}
		seen[capability] = true
	}
	if seen["sync:games"] && seen["sync:index"] {
		return fieldError(field+".capabilities", "use sync:games or sync:index, not both")
	}
	return nil
}

func validateCommands(commands []CommandContribution) error {
	for i, command := range commands {
		field := fmt.Sprintf("contributes.commands[%d]", i)
		if err := validateItem(field, command.ID, command.Label); err != nil {
			return err
		}
		if command.Method != "" && !methodPattern.MatchString(command.Method) {
			return fieldError(field+".method", "invalid method %q, use a name such as demo-site.importKey", command.Method)
		}
	}
	return nil
}

func validateNamedItems(items []NamedItem) error {
	for i, item := range items {
		if err := validateItem(fmt.Sprintf("contributes[%d]", i), item.ID, item.Label); err != nil {
			return err
		}
	}
	return nil
}

func validateHooksAndThemes(c Contributes) error {
	seenHooks := map[string]bool{}
	for _, hook := range c.Hooks {
		if !validHooks[hook] {
			return fieldError("contributes.hooks", "unknown hook %q", hook)
		}
		if seenHooks[hook] {
			return fieldError("contributes.hooks", "duplicate hook %q", hook)
		}
		seenHooks[hook] = true
	}
	for i, theme := range c.Themes {
		field := fmt.Sprintf("contributes.themes[%d]", i)
		if err := validateItem(field, theme.ID, theme.Label); err != nil {
			return err
		}
		if !isRelativePath(theme.Path) {
			return fieldError(field+".path", "must be a relative path inside the package")
		}
	}
	return nil
}

func validateItem(field, id, label string) error {
	if !itemIDPattern.MatchString(id) {
		return fieldError(field+".id", "must be lowercase letters, digits or dashes, starting with a letter")
	}
	if label == "" {
		return fieldError(field+".label", "must not be empty")
	}
	return nil
}

func (m Manifest) validatePermissions() error {
	seen := map[string]bool{}
	for _, host := range m.Permissions.Network {
		if !hostPattern.MatchString(host) {
			return fieldError("permissions.network", "invalid host %q, use example.com or *.example.com", host)
		}
		if seen[host] {
			return fieldError("permissions.network", "duplicate host %q", host)
		}
		seen[host] = true
	}
	for _, folder := range m.Permissions.FS {
		if !fsNamePattern.MatchString(folder) {
			return fieldError("permissions.fs", "invalid folder %q, use $GAMES, $DOWNLOADS, $DATA or $APPDATA_<NAME>", folder)
		}
	}
	return nil
}

// isRelativePath rejects absolute paths, drive letters, backslashes and ".." segments.
func isRelativePath(path string) bool {
	if path == "" || strings.HasPrefix(path, "/") || strings.ContainsAny(path, "\\") {
		return false
	}
	if len(path) > 1 && path[1] == ':' {
		return false
	}
	for _, segment := range strings.Split(path, "/") {
		if segment == ".." {
			return false
		}
	}
	return true
}

func fieldError(field, format string, args ...any) error {
	return fmt.Errorf("%s: %s", field, fmt.Sprintf(format, args...))
}
