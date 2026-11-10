package appsvc

import (
	"context"
	"encoding/base64"
	"os/exec"
	"strings"
	"syscall"
	"unicode/utf16"
)

// runFileDialog shows a WinForms dialog through PowerShell. It returns "" when cancelled.
func runFileDialog(ctx context.Context, request dialogRequest) (string, error) {
	cmd := exec.CommandContext(ctx, "powershell.exe",
		"-NoProfile", "-NonInteractive", "-STA", "-EncodedCommand", encodePowerShell(dialogScript(request)))
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// encodePowerShell returns the UTF-16LE base64 form that -EncodedCommand expects. It avoids quoting problems.
func encodePowerShell(script string) string {
	units := utf16.Encode([]rune(script))
	raw := make([]byte, 0, len(units)*2)
	for _, unit := range units {
		raw = append(raw, byte(unit), byte(unit>>8))
	}
	return base64.StdEncoding.EncodeToString(raw)
}

func dialogScript(request dialogRequest) string {
	statements := []string{"Add-Type -AssemblyName System.Windows.Forms"}
	if request.directory {
		statements = append(statements,
			"$dialog = New-Object System.Windows.Forms.FolderBrowserDialog",
			"$dialog.Description = "+psQuote(request.title),
			"if ($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) { Write-Output $dialog.SelectedPath }")
		return strings.Join(statements, "; ")
	}
	statements = append(statements,
		"$dialog = New-Object System.Windows.Forms.OpenFileDialog",
		"$dialog.Title = "+psQuote(request.title))
	if len(request.filters) > 0 {
		statements = append(statements, "$dialog.Filter = "+psQuote(windowsFilter(request.filters)))
	}
	statements = append(statements,
		"if ($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) { Write-Output $dialog.FileName }")
	return strings.Join(statements, "; ")
}

// windowsFilter builds the "Name|*.ext;*.ext|Name|*.*" form that OpenFileDialog expects.
func windowsFilter(filters []fileFilter) string {
	parts := make([]string, 0, len(filters)*2)
	for _, filter := range filters {
		globs := make([]string, 0, len(filter.extensions))
		for _, extension := range filter.extensions {
			globs = append(globs, windowsGlob(extension))
		}
		parts = append(parts, filter.name, strings.Join(globs, ";"))
	}
	return strings.Join(parts, "|")
}

// windowsGlob maps "*" to "*.*", which is what the Windows dialog expects for every file.
func windowsGlob(extension string) string {
	if extension == "*" {
		return "*.*"
	}
	return extensionGlob(extension)
}

// psQuote wraps text in a PowerShell single-quoted string.
func psQuote(text string) string {
	return "'" + strings.ReplaceAll(text, "'", "''") + "'"
}
