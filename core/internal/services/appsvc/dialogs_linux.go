package appsvc

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
)

var errNoDialogTool = errors.New("no file dialog available: install zenity or kdialog")

// runFileDialog shows a zenity dialog, or kdialog when zenity is missing. It returns "" when cancelled.
func runFileDialog(ctx context.Context, request dialogRequest) (string, error) {
	if tool, err := exec.LookPath("zenity"); err == nil {
		return runDialogTool(ctx, tool, zenityArgs(request))
	}
	if tool, err := exec.LookPath("kdialog"); err == nil {
		return runDialogTool(ctx, tool, kdialogArgs(request))
	}
	return "", errNoDialogTool
}

// runDialogTool returns the chosen path. Both tools exit with status 1 when the user cancels.
func runDialogTool(ctx context.Context, tool string, args []string) (string, error) {
	output, err := exec.CommandContext(ctx, tool, args...).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func zenityArgs(request dialogRequest) []string {
	args := []string{"--file-selection", "--title=" + request.title}
	if request.directory {
		return append(args, "--directory")
	}
	for _, filter := range request.filters {
		args = append(args, "--file-filter="+filter.name+" | "+filterGlobs(filter))
	}
	return args
}

func filterGlobs(filter fileFilter) string {
	globs := make([]string, 0, len(filter.extensions))
	for _, extension := range filter.extensions {
		globs = append(globs, extensionGlob(extension))
	}
	return strings.Join(globs, " ")
}

func kdialogArgs(request dialogRequest) []string {
	home, _ := os.UserHomeDir()
	if request.directory {
		return []string{"--getexistingdirectory", home, "--title", request.title}
	}
	args := []string{"--getopenfilename", home}
	if len(request.filters) > 0 {
		args = append(args, kdialogFilters(request.filters))
	}
	return append(args, "--title", request.title)
}

// kdialogFilters writes one "patterns|name" line per filter.
func kdialogFilters(filters []fileFilter) string {
	lines := make([]string, 0, len(filters))
	for _, filter := range filters {
		lines = append(lines, filterGlobs(filter)+"|"+filter.name)
	}
	return strings.Join(lines, "\n")
}
