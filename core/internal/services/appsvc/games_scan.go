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
	index := newScanIndex(settings.ScanFolders, s.games.List())
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
}

// newScanIndex records configs added by hand, which are never auto-scanned twice.
func newScanIndex(scanFolders []string, configs []types.LaunchOptions) scanIndex {
	index := scanIndex{
		manualPaths:   map[string]bool{},
		configsByPath: map[string]types.LaunchOptions{},
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

	games := []types.GameInfo{}
	for _, exePath := range searchExecutables(folder, folderCfg.Depth, excludeNames) {
		normalized := config.PathKey(exePath)
		if blacklist[normalized] || index.manualPaths[normalized] {
			continue
		}
		game, err := s.scannedGame(exePath, normalized, index)
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

// scannedGame returns the game for one executable, saving a default config the first time it is seen.
func (s *service) scannedGame(exePath, normalized string, index scanIndex) (types.GameInfo, error) {
	name := config.NameWithoutExt(exePath)
	cfg, known := index.configsByPath[normalized]
	if !known {
		cfg = newDefaultGameConfig(name, exePath)
		if err := s.games.Save(&cfg); err != nil {
			return types.GameInfo{}, err
		}
		index.configsByPath[normalized] = cfg
	}
	return types.GameInfo{
		Name:          name,
		Path:          exePath,
		Icon:          "",
		Config:        cfg,
		IsRecent:      false,
		IsAutoScanned: true,
	}, nil
}
