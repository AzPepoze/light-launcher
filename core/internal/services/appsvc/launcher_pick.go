package appsvc

import (
	"cmp"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/types"
)

// junkExecutablePatterns name helper programs that are never a game's launcher.
// They are removed from a game folder's launchers. The user's scan excludes are not changed.
var junkExecutablePatterns = []string{
	"*CrashHandler*.exe",
	"*crashpad*.exe",
	"*CrashReport*.exe",
	"*setup*.exe",
	"*installer*.exe",
	"unins*.exe",
	"*uninstall*.exe",
	"*redist*.exe",
	"dotNetFx*.exe",
	"*flashplayer*.exe",
	"*notification*.exe",
	"*subprocess*.exe",
}

var junkExecutableRegexes = compileWildcardPatterns(junkExecutablePatterns)

// bitness32Hint matches a 32-bit name such as "Game32", "Win32" or "x86" as a whole number.
var bitness32Hint = regexp.MustCompile(`(?i)(^|[^0-9])32($|[^0-9])|x86`)

// rankedLauncher holds the sort keys of one candidate so each is computed once.
type rankedLauncher struct {
	file      scannedFile
	bitness   int
	nameMatch int
}

// pickPrimaryLauncher drops helper programs and picks the main launcher of one game folder.
// remembered is the path the user chose, or "". It returns the launchers with the primary one
// first, or nil when no launcher is left.
func pickPrimaryLauncher(folder string, exes []scannedFile, remembered string) []types.Launcher {
	kept := withoutJunkExecutables(exes)
	if len(kept) == 0 {
		return nil
	}

	shipping := unrealShippingExes(kept)
	primary := choosePrimary(folder, kept, shipping, remembered)
	launchers := make([]types.Launcher, 0, len(kept))
	for _, exe := range kept {
		isPrimary := exe.Path == primary
		launchers = append(launchers, types.Launcher{
			Name:    config.NameWithoutExt(exe.Path),
			Path:    exe.Path,
			Kind:    launcherKind(folder, exe.Path, len(shipping) > 0, isPrimary),
			Primary: isPrimary,
		})
	}
	slices.SortFunc(launchers, comparePrimaryFirst)
	return launchers
}

// isJunkExecutable reports whether a file name is a helper program such as a crash handler or installer.
func isJunkExecutable(name string) bool {
	for _, regex := range junkExecutableRegexes {
		if regex.MatchString(name) {
			return true
		}
	}
	return false
}

func withoutJunkExecutables(exes []scannedFile) []scannedFile {
	kept := make([]scannedFile, 0, len(exes))
	for _, exe := range exes {
		if !isJunkExecutable(filepath.Base(exe.Path)) {
			kept = append(kept, exe)
		}
	}
	return kept
}

// choosePrimary applies the rules in order: the remembered launcher, then Unreal Shipping,
// then the ranking of rankLaunchers.
func choosePrimary(folder string, exes, shipping []scannedFile, remembered string) string {
	if path, ok := findRemembered(exes, remembered); ok {
		return path
	}
	if len(shipping) > 0 {
		return rankLaunchers(folder, shipping).Path
	}
	return rankLaunchers(folder, exes).Path
}

func findRemembered(exes []scannedFile, remembered string) (string, bool) {
	if remembered == "" {
		return "", false
	}
	key := config.PathKey(remembered)
	for _, exe := range exes {
		if config.PathKey(exe.Path) == key {
			return exe.Path, true
		}
	}
	return "", false
}

// rankLaunchers returns the best candidate: 64-bit before 32-bit, then the best name match
// with the folder, then the larger file, then the shorter and alphabetically first path.
func rankLaunchers(folder string, exes []scannedFile) scannedFile {
	ranked := make([]rankedLauncher, 0, len(exes))
	for _, exe := range exes {
		ranked = append(ranked, rankedLauncher{
			file:      exe,
			bitness:   bitnessRank(exe.Path),
			nameMatch: nameMatchScore(folder, exe.Path),
		})
	}
	return slices.MinFunc(ranked, compareRanked).file
}

func compareRanked(a, b rankedLauncher) int {
	return cmp.Or(
		cmp.Compare(a.bitness, b.bitness),
		cmp.Compare(b.nameMatch, a.nameMatch),
		cmp.Compare(b.file.Size, a.file.Size),
		cmp.Compare(len(a.file.Path), len(b.file.Path)),
		strings.Compare(a.file.Path, b.file.Path),
	)
}

// bitnessRank is 1 for a 32-bit name and 0 otherwise, so 32-bit candidates rank last.
func bitnessRank(path string) int {
	if bitness32Hint.MatchString(config.NameWithoutExt(path)) {
		return 1
	}
	return 0
}

// nameMatchScore is 2 when the names match after removing punctuation, 1 when one contains the other.
func nameMatchScore(folder, exePath string) int {
	folderKey := alphanumericKey(config.BaseName(folder))
	exeKey := alphanumericKey(config.NameWithoutExt(exePath))
	switch {
	case folderKey == "" || exeKey == "":
		return 0
	case folderKey == exeKey:
		return 2
	case strings.Contains(exeKey, folderKey) || strings.Contains(folderKey, exeKey):
		return 1
	default:
		return 0
	}
}

func alphanumericKey(name string) string {
	var builder strings.Builder
	for _, char := range strings.ToLower(name) {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

func unrealShippingExes(exes []scannedFile) []scannedFile {
	shipping := []scannedFile{}
	for _, exe := range exes {
		if isUnrealShipping(exe.Path) {
			shipping = append(shipping, exe)
		}
	}
	return shipping
}

// isUnrealShipping matches "<Game>-Win64-Shipping.exe" inside a "Binaries/Win64" folder.
func isUnrealShipping(path string) bool {
	name := strings.ToLower(filepath.Base(path))
	return strings.HasSuffix(name, "-win64-shipping.exe") && inBinariesWin64(path)
}

func inBinariesWin64(path string) bool {
	parts := strings.Split(strings.ToLower(filepath.ToSlash(filepath.Dir(path))), "/")
	for i := 0; i+1 < len(parts); i++ {
		if parts[i] == "binaries" && parts[i+1] == "win64" {
			return true
		}
	}
	return false
}

// launcherKind labels a launcher. The primary one is "main"; the root stub of an Unreal game is
// "unreal-stub"; other 32-bit executables are "x86"; the rest are "other".
func launcherKind(folder, exePath string, shipping, primary bool) types.LauncherKind {
	switch {
	case primary:
		return types.LauncherMain
	case shipping && config.PathKey(filepath.Dir(exePath)) == config.PathKey(folder):
		return types.LauncherUnrealStub
	case bitness32Hint.MatchString(config.NameWithoutExt(exePath)):
		return types.LauncherX86
	default:
		return types.LauncherOther
	}
}

// comparePrimaryFirst puts the primary launcher first, then the rest by path.
func comparePrimaryFirst(a, b types.Launcher) int {
	if a.Primary != b.Primary {
		if a.Primary {
			return -1
		}
		return 1
	}
	return strings.Compare(a.Path, b.Path)
}
