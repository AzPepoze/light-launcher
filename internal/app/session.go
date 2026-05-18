package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"light-launcher/internal/types"
)

func (app *App) GetRunningSessions() ([]types.RunningSession, error) {
	output, _ := exec.Command("pgrep", "light-launcher-instance").Output()
	if len(output) == 0 {
		output, _ = exec.Command("pgrep", "light-launcher-instan").Output()
	}

	pids := strings.Split(strings.TrimSpace(string(output)), "\n")
	sessions := make([]types.RunningSession, 0)

	for _, pidString := range pids {
		if pidString == "" {
			continue
		}
		pid, err := strconv.Atoi(pidString)
		if err != nil {
			continue
		}

		cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", pid)
		content, err := os.ReadFile(cmdlinePath)
		if err != nil {
			continue
		}

		arguments := strings.Split(string(content), "\x00")
		gamePath := ""
		for index, argument := range arguments {
			if argument == "--game" && index+1 < len(arguments) {
				gamePath = arguments[index+1]
				break
			}
		}

		if gamePath != "" {
			name := filepath.Base(gamePath)
			name = strings.TrimSuffix(name, filepath.Ext(name))

			cleanedPath := filepath.Clean(gamePath)
			if absolutePath, err := filepath.Abs(cleanedPath); err == nil {
				cleanedPath = absolutePath
			}

			sessions = append(sessions, types.RunningSession{
				Pid:      pid,
				GamePath: cleanedPath,
				GameName: name,
			})
		}
	}
	return sessions, nil
}

func (app *App) KillSession(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Signal(os.Interrupt)
}
