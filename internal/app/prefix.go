package app

import (
	"os/exec"
	"path/filepath"
	"strings"

	"light-launcher/internal/config"
	"light-launcher/internal/executor/builder"
	"light-launcher/internal/system"
	"light-launcher/internal/types"
)

func (app *App) ListPrefixes() ([]string, error) {
	return config.ListPrefixes()
}

func (app *App) CreatePrefix(name string) error {
	return config.CreatePrefix(name)
}

func (app *App) GetPrefixBaseDir() string {
	return config.GetPrefixBaseDirectory()
}

func (app *App) RemovePrefix(name string) error {
	return config.RemovePrefix(name)
}

func (app *App) SavePrefixConfig(prefixName string, options types.LaunchOptions) error {
	return config.SavePrefixConfig(prefixName, options)
}

func (app *App) LoadPrefixConfig(prefixName string) (*types.LaunchOptions, error) {
	return config.LoadPrefixConfig(prefixName)
}

func (app *App) LoadPrefixConfigWithProton(prefixName string) (*types.PrefixConfigWithProton, error) {
	// Load the prefix config
	cfg, err := config.LoadPrefixConfig(prefixName)
	if err != nil {
		return nil, err
	}

	result := &types.PrefixConfigWithProton{
		Config: *cfg,
	}

	// Get available Proton tools and resolve the saved path
	if cfg.ProtonPath != "" {
		protonTools, _ := system.GetProtonTools()
		match := app.findProtonMatch(cfg.ProtonPath, protonTools)

		if match != nil {
			result.ProtonDisplayName = match.DisplayName
			result.ProtonName = match.Name
			result.ProtonPath = match.Path
			result.ProtonIsSteam = match.IsSteam
		} else {
			// Proton not found, but we have a saved path
			result.ProtonPath = cfg.ProtonPath
			result.ProtonDisplayName = cfg.ProtonPath
		}
	}

	return result, nil
}

func (app *App) findProtonMatch(savedPath string, protonVersions []types.ProtonTool) *types.ProtonTool {
	// Try matching by DisplayName
	for i := range protonVersions {
		if protonVersions[i].DisplayName == savedPath {
			return &protonVersions[i]
		}
	}

	// Try exact path match
	for i := range protonVersions {
		if protonVersions[i].Path == savedPath {
			return &protonVersions[i]
		}
	}

	// Extract directory name
	parts := strings.Split(filepath.Clean(savedPath), string(filepath.Separator))
	if len(parts) == 0 {
		return nil
	}
	dirName := parts[len(parts)-1]
	isSteam := strings.Contains(savedPath, "steamapps/common")

	// Try matching by name + Steam status
	for i := range protonVersions {
		if protonVersions[i].Name == dirName && protonVersions[i].IsSteam == isSteam {
			return &protonVersions[i]
		}
	}

	// Try matching by name only
	for i := range protonVersions {
		if protonVersions[i].Name == dirName {
			return &protonVersions[i]
		}
	}

	return nil
}

func (app *App) RunPrefixTool(prefixPath, toolName, protonPath string) error {
	options := types.LaunchOptions{
		GamePath:   toolName,
		PrefixPath: prefixPath,
		ProtonPath: protonPath,
	}
	commandArguments, environment := builder.BuildCommand(options)
	command := exec.Command(commandArguments[0], commandArguments[1:]...)
	command.Env = environment
	return command.Start()
}
