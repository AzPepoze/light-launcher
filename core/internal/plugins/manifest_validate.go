package plugins

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/semver"
)

// maxSourceActions is the most buttons one source can add to the Library.
const maxSourceActions = 8

var (
	idPattern         = regexp.MustCompile(`^[a-z][a-z0-9-]{1,39}$`)
	iconPattern       = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	itemIDPattern     = regexp.MustCompile(`^[a-z][a-z0-9-]{0,39}$`)
	versionPattern    = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)
	apiVersionPattern = regexp.MustCompile(`^[0-9]+$`)
	entryKeyPattern   = regexp.MustCompile(`^(linux|windows|mac)-(x64|arm64)$`)
	hostPattern       = regexp.MustCompile(`^(\*\.)?[a-z0-9-]+(\.[a-z0-9-]+)+$`)
	fsNamePattern     = regexp.MustCompile(`^\$(GAMES|DOWNLOADS|DATA|APPDATA_[A-Z][A-Z0-9_]*)$`)
	methodPattern     = regexp.MustCompile(`^[a-z][a-z0-9-]*(\.[A-Za-z][A-Za-z0-9]*)+$`)
	repositoryPattern = regexp.MustCompile(`^github:[A-Za-z0-9._-]+/[A-Za-z0-9._-]+$`)
)

var validCapabilities = map[string]bool{
	"browse": true, "auth": true, "updateCheck": true, "resolveFiles": true,
	"sync:games": true, "sync:index": true, "listGames": true, capabilityBrowseFilters: true,
}

var validTiles = map[string]bool{
	TilePoster: true, TileWide: true, TileSquare: true,
}

var validActionScopes = map[string]bool{
	contract.ActionScopeLibrary: true, contract.ActionScopeGame: true,
}

// cardTextFields are the game fields a card can show as its title or subtitle.
var cardTextFields = map[string]bool{
	"sourceId": true, "title": true, "version": true, "developer": true, "engine": true, "status": true, "updatedAt": true,
}

// cardBadges are the game flags a card can show as badges.
var cardBadges = map[string]bool{
	"favorite": true, "updateAvailable": true, "installed": true,
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
	if _, err := semver.Parse(m.MinHost); err != nil {
		return fieldError("minHost", "must be a semver version such as 2.0.0")
	}
	return nil
}

// validateInfo checks the optional author, description, homepage and repository. The SDK schema may not list them yet.
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
	if m.Repository != "" && !repositoryPattern.MatchString(m.Repository) {
		return fieldError("repository", "must look like github:owner/repo")
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
	if err := validateCommands(c.Commands); err != nil {
		return err
	}
	return validateUnsupportedContributions(c)
}

func validateSources(sources []SourceContribution) error {
	seen := map[string]bool{}
	for i, source := range sources {
		field := fmt.Sprintf("contributes.sources[%d]", i)
		if err := validateSource(field, source); err != nil {
			return err
		}
		if seen[source.ID] {
			return fieldError(field+".id", "duplicate source id %q", source.ID)
		}
		seen[source.ID] = true
	}
	return nil
}

// validateSource checks one source's id, capabilities, card shape, cover, actions, card and sidebar.
func validateSource(field string, source SourceContribution) error {
	if err := validateItem(field, source.ID, source.Label); err != nil {
		return err
	}
	if err := validateCapabilities(field, source.Capabilities); err != nil {
		return err
	}
	if err := validateTile(field, source.Tile); err != nil {
		return err
	}
	if err := validateCover(field, source.Cover); err != nil {
		return err
	}
	if err := validateActions(field, source.Actions); err != nil {
		return err
	}
	if err := validateCard(field, source.Card); err != nil {
		return err
	}
	return validateSidebar(field, source.Sidebar, source.Capabilities)
}

// validateTile accepts an empty tile (the default poster) or one of the known card shapes.
func validateTile(field, tile string) error {
	if tile == "" || validTiles[tile] {
		return nil
	}
	return fieldError(field+".tile", "must be poster, wide or square, not %q", tile)
}

// validateCover accepts a missing cover. A present one must be a ratio the host can draw.
func validateCover(field string, cover *CoverContribution) error {
	if cover == nil || contract.ValidCover(cover.Ratio) {
		return nil
	}
	return fieldError(field+".cover.ratio", "must be W:H with sides 1 to 32 and a width to height ratio from 0.25 to 4, not %q", cover.Ratio)
}

// validateActions checks the source's buttons: at most maxSourceActions, each with a unique id.
func validateActions(field string, actions []contract.SourceAction) error {
	if len(actions) > maxSourceActions {
		return fieldError(field+".actions", "at most %d actions, got %d", maxSourceActions, len(actions))
	}
	seen := map[string]bool{}
	for i, action := range actions {
		actionField := fmt.Sprintf("%s.actions[%d]", field, i)
		if err := validateActionItem(actionField, action); err != nil {
			return err
		}
		if seen[action.ID] {
			return fieldError(actionField+".id", "duplicate action id %q", action.ID)
		}
		seen[action.ID] = true
	}
	return nil
}

// validateActionItem checks one action's id, label, method, scope and icon.
func validateActionItem(field string, action contract.SourceAction) error {
	if err := validateItem(field, action.ID, action.Label); err != nil {
		return err
	}
	if !methodPattern.MatchString(action.Method) {
		return fieldError(field+".method", "invalid method %q, use a name such as demo-site.importKey", action.Method)
	}
	if action.Scope != "" && !validActionScopes[action.Scope] {
		return fieldError(field+".scope", "must be library or game, not %q", action.Scope)
	}
	if action.Icon != "" && !iconPattern.MatchString(action.Icon) {
		return fieldError(field+".icon", "must be a Material icon name such as sync, not %q", action.Icon)
	}
	return nil
}

// validateCard checks the optional card. Title and subtitle must be game fields, badges must be known flags.
func validateCard(field string, card *contract.SourceCard) error {
	if card == nil {
		return nil
	}
	if err := validateCardField(field+".card.title", card.Title); err != nil {
		return err
	}
	if err := validateCardField(field+".card.subtitle", card.Subtitle); err != nil {
		return err
	}
	return validateBadges(field+".card.badges", card.Badges)
}

// validateCardField accepts an empty field (the default) or a text field of the game.
func validateCardField(field, name string) error {
	if name == "" || cardTextFields[name] {
		return nil
	}
	return fieldError(field, "unknown game field %q", name)
}

func validateBadges(field string, badges []string) error {
	seen := map[string]bool{}
	for _, badge := range badges {
		if !cardBadges[badge] {
			return fieldError(field, "unknown badge %q, use favorite, updateAvailable or installed", badge)
		}
		if seen[badge] {
			return fieldError(field, "duplicate badge %q", badge)
		}
		seen[badge] = true
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

// validateUnsupportedContributions rejects contributions the API version 1 host does not run yet.
// Silently accepting them makes a plugin look broken to its author and to the user.
func validateUnsupportedContributions(c Contributes) error {
	switch {
	case len(c.Pages) > 0:
		return fieldError("contributes.pages", "pages are not supported in apiVersion 1 yet")
	case len(c.Hooks) > 0:
		return fieldError("contributes.hooks", "hooks are not supported in apiVersion 1 yet")
	case len(c.LaunchAdapters) > 0:
		return fieldError("contributes.launchAdapters", "launch adapters are not supported in apiVersion 1 yet")
	case len(c.Importers) > 0:
		return fieldError("contributes.importers", "importers are not supported in apiVersion 1 yet")
	case len(c.Themes) > 0:
		return fieldError("contributes.themes", "themes are not supported in apiVersion 1 yet")
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
