package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func (app *App) runSystemPicker(title string, folder bool, filters []application.FileFilter) (string, bool) {
	if _, err := exec.LookPath("zenity"); err == nil {
		arguments := []string{"--file-selection", "--title=" + title}
		if folder {
			arguments = append(arguments, "--directory")
		}
		if len(filters) > 0 {
			for _, filter := range filters {
				pattern := strings.ReplaceAll(filter.Pattern, ";", " ")
				arguments = append(arguments, "--file-filter="+filter.DisplayName+"|"+pattern)
			}
		}
		command := exec.Command("zenity", arguments...)
		output, err := command.Output()
		if err == nil {
			return strings.TrimSpace(string(output)), true
		}
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return "", true
		}
	}

	if _, err := exec.LookPath("kdialog"); err == nil {
		var arguments []string
		if folder {
			arguments = []string{"--getexistingdirectory", ".", "--title", title}
		} else {
			filterString := ""
			if len(filters) > 0 {
				var parts []string
				for _, filter := range filters {
					pattern := strings.ReplaceAll(filter.Pattern, ";", " ")
					parts = append(parts, filter.DisplayName+" ("+pattern+")")
				}
				filterString = strings.Join(parts, ";;")
			}
			arguments = []string{"--getopenfilename", ".", filterString, "--title", title}
		}
		command := exec.Command("kdialog", arguments...)
		output, err := command.Output()
		if err == nil {
			return strings.TrimSpace(string(output)), true
		}
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return "", true
		}
	}

	return "", false
}

func (app *App) SearchExecutables(folderPath string, maxDepth int, excludeNames []string) ([]string, error) {
	var executables []string
	cleanPath := filepath.Clean(folderPath)

	var excludeRegex []*regexp.Regexp
	for _, pattern := range excludeNames {
		if pattern == "" {
			continue
		}
		regexString := pattern
		if strings.Contains(regexString, "*") {
			regexString = strings.ReplaceAll(regexp.QuoteMeta(regexString), "\\*", ".*")
		}
		if regex, err := regexp.Compile("(?i)" + regexString); err == nil {
			excludeRegex = append(excludeRegex, regex)
		}
	}

	var scanDir func(dir string, currentDepth int)
	scanDir = func(dir string, currentDepth int) {
		if maxDepth != -1 && currentDepth > maxDepth {
			return
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}

		for _, entry := range entries {
			name := entry.Name()

			// Check exclusions
			excluded := false
			for _, regex := range excludeRegex {
				if regex.MatchString(name) {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}

			fullPath := filepath.Join(dir, name)
			if entry.IsDir() {
				scanDir(fullPath, currentDepth+1)
			} else if strings.EqualFold(filepath.Ext(name), ".exe") {
				executables = append(executables, fullPath)
			}
		}
	}

	scanDir(cleanPath, 0)
	return executables, nil
}

func (app *App) PickFile() (string, error) {
	if path, ok := app.runSystemPicker("Select Game Executable", false, []application.FileFilter{
		{DisplayName: "Executables (*.exe)", Pattern: "*.exe"},
		{DisplayName: "All Files", Pattern: "*.*"},
	}); ok {
		return path, nil
	}

	return application.Get().Dialog.OpenFileWithOptions(&application.OpenFileDialogOptions{
		Title: "Select Game Executable",
		Filters: []application.FileFilter{
			{DisplayName: "Executables (*.exe)", Pattern: "*.exe"},
			{DisplayName: "All Files", Pattern: "*.*"},
		},
	}).PromptForSingleSelection()
}

func (app *App) PickFolder() (string, error) {
	if path, ok := app.runSystemPicker("Select Directory", true, nil); ok {
		return path, nil
	}

	return application.Get().Dialog.OpenFileWithOptions(&application.OpenFileDialogOptions{
		Title:                "Select Directory",
		CanChooseDirectories: true,
		CanChooseFiles:       false,
	}).PromptForSingleSelection()
}

func (app *App) PickFileCustom(title string, filters []application.FileFilter) (string, error) {
	if path, ok := app.runSystemPicker(title, false, filters); ok {
		return path, nil
	}

	return application.Get().Dialog.OpenFileWithOptions(&application.OpenFileDialogOptions{
		Title:   title,
		Filters: filters,
	}).PromptForSingleSelection()
}
