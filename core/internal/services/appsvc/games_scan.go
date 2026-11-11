package appsvc

import (
	"sync"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/types"
)

// scanCache keeps the last auto-scan result until a folder or game changes.
var scanCache struct {
	mu     sync.Mutex
	valid  bool
	groups []types.ScannedFolderGroup
}

func resetScanCache() {
	scanCache.mu.Lock()
	defer scanCache.mu.Unlock()
	scanCache.valid = false
	scanCache.groups = nil
}

// getAutoScannedGames scans every watched folder. Executables without a config get one created.
func (s *service) getAutoScannedGames(force bool) ([]types.ScannedFolderGroup, error) {
	if !force {
		if groups, ok := cachedScanGroups(); ok {
			return groups, nil
		}
	}

	settings := s.settings.Load()
	index := newScanIndex(settings.ScanFolders, s.games.List(), s.preferences.Preferred())
	blacklist := blacklistSet(settings.Blacklist)

	groups := []types.ScannedFolderGroup{}
	for _, folderCfg := range settings.ScanFolderConfigs {
		group, err := s.scanFolderGroup(folderCfg, index, blacklist)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}

	storeScanGroups(groups)
	return groups, nil
}

func cachedScanGroups() ([]types.ScannedFolderGroup, bool) {
	scanCache.mu.Lock()
	defer scanCache.mu.Unlock()
	return scanCache.groups, scanCache.valid
}

func storeScanGroups(groups []types.ScannedFolderGroup) {
	scanCache.mu.Lock()
	defer scanCache.mu.Unlock()
	scanCache.valid = true
	scanCache.groups = groups
}

// scanIndex answers "which config belongs to this executable" during a scan.
type scanIndex struct {
	manualPaths   map[string]bool
	configsByPath map[string]types.LaunchOptions
	remembered    map[string]string
}

// newScanIndex records configs added by hand, which are never auto-scanned twice.
// remembered maps each game folder to the launcher the user chose.
func newScanIndex(scanFolders []string, configs []types.LaunchOptions, remembered map[string]string) scanIndex {
	index := scanIndex{
		manualPaths:   map[string]bool{},
		configsByPath: map[string]types.LaunchOptions{},
		remembered:    remembered,
	}
	for _, cfg := range configs {
		p := config.CleanPath(config.FirstNonEmpty(cfg.LauncherPath, cfg.GamePath))
		if p != "" && !isInAnyFolder(scanFolders, p) {
			index.manualPaths[config.PathKey(p)] = true
		}
		if cfg.GamePath != "" {
			index.configsByPath[config.PathKey(cfg.GamePath)] = cfg
		}
		if cfg.LauncherPath != "" {
			index.configsByPath[config.PathKey(cfg.LauncherPath)] = cfg
		}
	}
	return index
}

func blacklistSet(blacklist []string) map[string]bool {
	set := map[string]bool{}
	for _, p := range blacklist {
		set[config.PathKey(p)] = true
	}
	return set
}

func (s *service) scanFolderGroup(folderCfg types.ScanFolderConfig, index scanIndex, blacklist map[string]bool) (types.ScannedFolderGroup, error) {
	folder := config.CleanPath(folderCfg.Path)
	excludeNames := folderCfg.ExcludeNames
	if excludeNames == nil {
		excludeNames = config.DefaultScanExcludeNames()
	}

	exes := keepScannable(scanExecutables(folder, folderCfg.Depth, excludeNames), blacklist, index)
	games := []types.GameInfo{}
	for _, root := range groupGameRoots(folder, exes, index.remembered) {
		game, err := s.scannedGame(root, index)
		if err != nil {
			return types.ScannedFolderGroup{}, err
		}
		games = append(games, game)
	}

	return types.ScannedFolderGroup{
		FolderPath: folder,
		FolderName: config.BaseName(folder),
		Games:      games,
	}, nil
}

// keepScannable drops executables the user blacklisted or added by hand.
func keepScannable(exes []scannedFile, blacklist map[string]bool, index scanIndex) []scannedFile {
	kept := make([]scannedFile, 0, len(exes))
	for _, exe := range exes {
		normalized := config.PathKey(exe.Path)
		if blacklist[normalized] || index.manualPaths[normalized] {
			continue
		}
		kept = append(kept, exe)
	}
	return kept
}

// scannedGame returns the card for one game root. Every launcher keeps its own config.
func (s *service) scannedGame(root gameRoot, index scanIndex) (types.GameInfo, error) {
	var primaryConfig types.LaunchOptions
	for _, launcher := range root.launchers {
		cfg, err := s.launcherConfig(launcher.Path, index)
		if err != nil {
			return types.GameInfo{}, err
		}
		if launcher.Primary {
			primaryConfig = cfg
		}
	}
	primary := primaryLauncher(root)
	return types.GameInfo{
		Name:          gameName(root),
		Path:          primary.Path,
		Icon:          "",
		Config:        primaryConfig,
		IsRecent:      false,
		IsAutoScanned: true,
		Folder:        root.folder,
		Launchers:     root.launchers,
		LauncherCount: len(root.launchers),
	}, nil
}

// launcherConfig returns the config of one executable, saving a default config the first time it is seen.
func (s *service) launcherConfig(exePath string, index scanIndex) (types.LaunchOptions, error) {
	normalized := config.PathKey(exePath)
	if cfg, known := index.configsByPath[normalized]; known {
		return cfg, nil
	}
	cfg := newDefaultGameConfig(config.NameWithoutExt(exePath), exePath)
	if err := s.games.Save(&cfg); err != nil {
		return types.LaunchOptions{}, err
	}
	index.configsByPath[normalized] = cfg
	return cfg, nil
}
