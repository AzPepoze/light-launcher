package appsvc

import (
	"context"
	"strings"
)

// fileFilter is one entry in a file dialog's type list. The "*" extension matches every file.
type fileFilter struct {
	name       string
	extensions []string
}

// dialogRequest describes one file or folder picker. A cancelled dialog returns "".
type dialogRequest struct {
	title     string
	directory bool
	filters   []fileFilter
}

// customFilter is a filter sent by the UI, with patterns such as "*.png;*.jpg".
type customFilter struct {
	DisplayName string `json:"displayName"`
	Pattern     string `json:"pattern"`
}

var exeFilters = []fileFilter{
	{name: "Executables (*.exe)", extensions: []string{"exe"}},
	{name: "All Files", extensions: []string{"*"}},
}

func pickFile(ctx context.Context) (string, error) {
	return runFileDialog(ctx, dialogRequest{title: "Select Game Executable", filters: exeFilters})
}

func pickFolder(ctx context.Context) (string, error) {
	return runFileDialog(ctx, dialogRequest{title: "Select Directory", directory: true})
}

// pickFileCustom uses the UI's filter list. Patterns like "*.png;*.jpg" become extensions "png" and "jpg".
func pickFileCustom(ctx context.Context, title string, filters []customFilter) (string, error) {
	return runFileDialog(ctx, dialogRequest{title: title, filters: fileFiltersFrom(filters)})
}

func fileFiltersFrom(filters []customFilter) []fileFilter {
	converted := []fileFilter{}
	for _, filter := range filters {
		var extensions []string
		for _, pattern := range strings.Split(filter.Pattern, ";") {
			extensions = append(extensions, strings.TrimPrefix(pattern, "*."))
		}
		converted = append(converted, fileFilter{name: filter.DisplayName, extensions: extensions})
	}
	return converted
}

// extensionGlob turns an extension into a file pattern such as "*.exe".
func extensionGlob(extension string) string {
	if extension == "*" {
		return "*"
	}
	return "*." + extension
}
