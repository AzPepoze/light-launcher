package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"light-launcher/internal/executor"
	"light-launcher/internal/executor/builder"
	lsfgLib "light-launcher/lib/lsfg"

	"github.com/getlantern/systray"
)

//go:embed appicon.png
var applicationIconData []byte

func getIconPath() string {
	if exe, err := os.Executable(); err == nil {
		local := filepath.Join(filepath.Dir(exe), "appicon.png")
		if _, err := os.Stat(local); err == nil {
			return local
		}
	}
	tmpPath := filepath.Join(os.TempDir(), "light-launcher-icon.png")
	if _, err := os.Stat(tmpPath); os.IsNotExist(err) {
		_ = os.WriteFile(tmpPath, applicationIconData, 0644)
	}
	return tmpPath
}

func sendNotification(title, message string) {
	_ = exec.Command("notify-send", "-a", "LightLauncher", "-i", getIconPath(), title, message).Run()
}

func onReady(logPath string) {
	exeName := filepath.Base(gamePath)
	exeNameClean := strings.TrimSuffix(exeName, filepath.Ext(exeName))
	launcherName := filepath.Base(launcherPath)

	systray.SetIcon(applicationIconData)
	systray.SetTitle("LightLauncher: " + exeNameClean)
	systray.SetTooltip("Running: " + exeNameClean)

	sendNotification("LightLauncher", "Launching "+exeNameClean+" ("+launcherName+")...")

	// Setup UI
	mStatus := systray.AddMenuItem("Running: "+exeNameClean, "")
	mStatus.Disable()
	systray.AddSeparator()

	if lsfg {
		setupLsfgMenu()
	}

	mKill := systray.AddMenuItem("End Process", "Stop this game")

	// Start game
	opts := buildLaunchOptions()
	cmdArgs, env := builder.BuildCommand(opts)

	logGameStartup(cmdArgs)

	gameCmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	gameCmd.Env = env
	gameCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	gameCmd.Stdout = logFileHandle
	gameCmd.Stderr = logFileHandle

	if err := gameCmd.Start(); err != nil {
		log.Printf("!!! ERROR: Failed to start game: %v\n", err)
		sendNotification("Launch Error", "Failed to start "+exeNameClean+" ("+launcherName+"): "+err.Error())
		systray.Quit()
		return
	}

	// Internal helper to kill game gracefully
	killGame := func() {
		if gameCmd.Process != nil {
			log.Println("Stopping game process group...")
			executor.StopProcessGroup(gameCmd.Process)
		}
	}

	// Setup signal handling for graceful shutdown (SIGINT/SIGTERM)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Printf("Received system signal: %v, triggering graceful shutdown...\n", sig)
		killGame()
	}()

	// Setup tray kill handler
	go func() {
		<-mKill.ClickedCh
		log.Println("Kill button clicked in tray")
		killGame()
	}()

	// Show logs in terminal if enabled
	if showLogs {
		startLogTerminal(logPath, gameCmd.Process.Pid)
	}

	// Periodically trim log file to keep it manageable (queue: last 500 lines)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			_ = trimLogFile(logPath, 500)
		}
	}()

	// Wait for game to exit
	go func() {
		err := gameCmd.Wait()
		log.Printf("Game process exited with: %v\n", err)

		if err != nil {
			sendNotification("Process Exited", fmt.Sprintf("%s exited with error: %v", exeNameClean, err))
		}

		time.Sleep(1 * time.Second)
		systray.Quit()
	}()
}

// setupLsfgMenu configures the LSFG-VK configuration menu item
func setupLsfgMenu() {
	mLsfgEdit := systray.AddMenuItem("Edit LSFG-VK Config", "Open LSFG-VK configuration")
	log.Printf("LSFG menu item created, waiting for clicks...")

	go func() {
		for {
			log.Printf("LSFG menu handler: waiting for click...")
			<-mLsfgEdit.ClickedCh
			log.Printf("LSFG menu handler: click received!")

			profile, idx, err := lsfgLib.FindProfileForGame(gamePath)
			if err != nil {
				log.Printf("LSFG menu handler: error finding profile: %v", err)
				sendNotification("LSFG-VK Config", fmt.Sprintf("Could not find profile for this game: %v", err))
				continue
			}
			log.Printf("Found LSFG profile for game: %s (index: %d)", profile.Name, idx)

			launchLsfgUI()
			log.Printf("LSFG menu handler: UI launch completed, ready for next click")
		}
	}()
}

// launchLsfgUI launches the light-launcher in LSFG edit mode
func launchLsfgUI() {
	uiBinary := "light-launcher"

	// Try to find local UI binary first
	if exePath, err := os.Executable(); err == nil {
		localBinary := filepath.Join(filepath.Dir(exePath), "light-launcher")
		if _, err := os.Stat(localBinary); err == nil {
			uiBinary = localBinary
			log.Printf("Found local light-launcher binary: %s", localBinary)
		}
	}

	uiCmd := exec.Command(uiBinary)
	env := os.Environ()
	env = append(env, fmt.Sprintf("LIGHT_LAUNCHER_GAME_PATH=%s", gamePath))
	env = append(env, "LIGHT_LAUNCHER_EDIT_LSFG=1")
	uiCmd.Env = env
	uiCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	log.Printf("Launching light-launcher with LIGHT_LAUNCHER_GAME_PATH=%s and LIGHT_LAUNCHER_EDIT_LSFG=1", gamePath)

	if err := uiCmd.Start(); err != nil {
		sendNotification("LSFG-VK Config", fmt.Sprintf("Failed to launch UI: %v", err))
		log.Printf("Error launching UI: %v", err)
	} else {
		log.Printf("UI launched successfully (PID: %d)", uiCmd.Process.Pid)
		go uiCmd.Process.Release()
	}
}

func onExit() {
	if logFileHandle != nil {
		logFileHandle.Close()
	}
}
