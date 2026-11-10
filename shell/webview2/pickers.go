//go:build windows

package main

import (
	"bytes"
	"os/exec"
	"strings"
	"syscall"
)

// createNoWindow stops PowerShell from opening a console window.
const createNoWindow = 0x08000000

// The dialogs run in PowerShell, which is simpler than binding the COM file dialog directly.
const (
	folderDialogScript = `Add-Type -AssemblyName System.Windows.Forms; ` +
		`$d = New-Object System.Windows.Forms.FolderBrowserDialog; ` +
		`if ($d.ShowDialog() -eq 'OK') { Write-Output $d.SelectedPath }`
	fileDialogScript = `Add-Type -AssemblyName System.Windows.Forms; ` +
		`$d = New-Object System.Windows.Forms.OpenFileDialog; ` +
		`if ($d.ShowDialog() -eq 'OK') { Write-Output $d.FileName }`
)

var pickerScripts = map[string]string{
	"pickFolder": folderDialogScript,
	"pickFile":   fileDialogScript,
}

// showPicker runs one dialog and returns the chosen path. An empty path means the user cancelled.
func showPicker(script string) (string, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-STA", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}
