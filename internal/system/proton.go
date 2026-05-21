package system

import (
	"fmt"
	"light-launcher/internal/config"
	"light-launcher/internal/types"
	"os"
	"path/filepath"
	"sort"
)

func GetProtonTools() ([]types.ProtonTool, error) {
	searchPaths := []struct {
		path    string
		isSteam bool
	}{
		{"~/.steam/root/compatibilitytools.d", false},
		{"~/.local/share/Steam/compatibilitytools.d", false},
		{"/usr/share/steam/compatibilitytools.d", false},
		{"~/.steam/root/steamapps/common", true},
		{"~/.local/share/Steam/steamapps/common", true},
		{filepath.Join(config.GetBaseDirectory(), config.ProtonsDirName), false},
	}

	var tools []types.ProtonTool
	seenPaths := make(map[string]bool)

	for _, searchPath := range searchPaths {
		absolutePath := config.ExpandPath(searchPath.path)
		realPath, err := filepath.EvalSymlinks(absolutePath)
		if err != nil {
			realPath = absolutePath
		}

		if seenPaths[realPath] {
			continue
		}
		seenPaths[realPath] = true

		discoveredTools, _ := scanProtonDirectory(realPath, searchPath.isSteam)
		tools = append(tools, discoveredTools...)
	}

	sort.Slice(tools, func(i, j int) bool {
		if tools[i].IsSteam && !tools[j].IsSteam {
			return true
		}
		if !tools[i].IsSteam && tools[j].IsSteam {
			return false
		}
		return tools[i].Name < tools[j].Name
	})

	return tools, nil
}

func scanProtonDirectory(path string, isSteam bool) ([]types.ProtonTool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var tools []types.ProtonTool
	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			protonExecutablePath := filepath.Join(path, name, "proton")
			if _, err := os.Stat(protonExecutablePath); os.IsNotExist(err) {
				continue
			}

			displayName := name
			if isSteam {
				displayName = fmt.Sprintf("(Steam) %s", name)
			}
			tools = append(tools, types.ProtonTool{
				Name:        name,
				Path:        filepath.Join(path, name),
				IsSteam:     isSteam,
				DisplayName: displayName,
			})
		}
	}
	return tools, nil
}
