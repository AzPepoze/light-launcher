package appsvc

import (
	"regexp"
	"strings"

	"light-launcher/core/internal/config"
)

// gameNameSources holds the candidate names of one game. Empty fields are skipped.
type gameNameSources struct {
	userName      string // the name the user saved in the game config
	exeName       string // the launcher file name without its extension
	folderName    string // the game folder name, raw; empty when the game is a loose file
	product       string // ProductName from the version info
	description   string // FileDescription from the version info
	multiLauncher bool   // the folder holds several launchers
}

// gameName returns the card name of a game root. savedName is the name in the primary launcher's config.
func gameName(root gameRoot, savedName string) string {
	primary := primaryLauncher(root)
	names := cachedVersionStrings(root.exe)
	return resolveGameName(gameNameSources{
		userName:      userChosenName(savedName, primary.Path),
		exeName:       config.NameWithoutExt(primary.Path),
		folderName:    gameFolderName(root, primary.Path),
		product:       names.productName,
		description:   names.fileDescription,
		multiLauncher: len(root.launchers) > 1,
	})
}

// resolveGameName returns the first usable name: the user's name, the product name, the cleaned folder
// name when the launcher name says little, and the launcher name.
func resolveGameName(src gameNameSources) string {
	if src.userName != "" {
		return src.userName
	}
	if name := embeddedName(src); name != "" {
		return name
	}
	if src.multiLauncher || isGenericExeName(src.exeName, src.folderName) {
		if name := cleanFolderName(src.folderName); name != "" {
			return name
		}
	}
	return src.exeName
}

// embeddedName returns the product name, or the file description, that is not generic.
func embeddedName(src gameNameSources) string {
	for _, name := range []string{src.product, src.description} {
		if name != "" && !isGenericName(name) {
			return name
		}
	}
	return ""
}

// userChosenName returns the config name when the user set it. The scan stores the exe name by default, which is not a choice.
func userChosenName(savedName, exePath string) string {
	saved := strings.TrimSpace(savedName)
	if saved == config.NameWithoutExt(exePath) {
		return ""
	}
	return saved
}

// gameFolderName returns the game folder, or "" when the game is a loose executable in the watched folder.
func gameFolderName(root gameRoot, exePath string) string {
	if config.PathKey(root.folder) == config.PathKey(exePath) {
		return ""
	}
	return config.BaseName(root.folder)
}

// genericNames are names that say nothing about a game. Compared in lower case.
var genericNames = map[string]bool{
	"game":          true,
	"launcher":      true,
	"start":         true,
	"play":          true,
	"run":           true,
	"main":          true,
	"app":           true,
	"client":        true,
	"win":           true,
	"pc":            true,
	"nw":            true,
	"unity":         true,
	"unreal engine": true,
	"ren'py":        true,
	"godot engine":  true,
	"nw.js":         true,
	"electron":      true,
}

// isGenericName reports whether a product or launcher name says nothing about the game.
func isGenericName(name string) bool {
	lower := strings.ToLower(strings.TrimSpace(name))
	if lower == "" || genericNames[lower] {
		return true
	}
	if strings.HasPrefix(lower, "unityplayer") {
		return true
	}
	return strings.HasPrefix(lower, "microsoft") && (strings.Contains(lower, "windows") || strings.Contains(lower, "corporation"))
}

// isGenericExeName reports whether a launcher file name should give way to the folder name.
func isGenericExeName(exeName, folderName string) bool {
	if isGenericName(exeName) {
		return true
	}
	lower := strings.ToLower(exeName)
	if folder := strings.ToLower(folderName); folder != "" && (lower == folder+"-32" || lower == folder+"-64") {
		return true
	}
	return strings.HasSuffix(lower, "-win64-shipping") || strings.HasSuffix(lower, "-win32-shipping")
}

// Version and platform suffixes that a folder name carries but a game title does not.
var (
	versionSuffixPattern  = regexp.MustCompile(`(?i)\s*\[v?\d[^\]]*\]$|[-_ ]v?\d+(\.\d+)+$|[-_ ]v\d+$`)
	platformSuffixPattern = regexp.MustCompile(`(?i)[-_ ](win(32|64)?|pc)$`)
	folderSeparators      = strings.NewReplacer("_", " ", ".", " ")
)

// cleanFolderName turns a folder name into a game title: "Mad_Island_0.34" becomes "Mad Island".
// It returns the folder name unchanged when cleaning would leave nothing.
func cleanFolderName(folder string) string {
	cleaned := strings.TrimSpace(folder)
	for {
		stripped := stripFolderSuffixes(cleaned)
		if stripped == cleaned {
			break
		}
		cleaned = stripped
	}
	if !strings.Contains(cleaned, " ") {
		cleaned = strings.Join(strings.Fields(folderSeparators.Replace(cleaned)), " ")
	}
	if cleaned == "" {
		return folder
	}
	return cleaned
}

// stripFolderSuffixes removes one round of version and platform suffixes.
func stripFolderSuffixes(name string) string {
	withoutVersion := versionSuffixPattern.ReplaceAllString(name, "")
	return strings.TrimSpace(platformSuffixPattern.ReplaceAllString(withoutVersion, ""))
}
