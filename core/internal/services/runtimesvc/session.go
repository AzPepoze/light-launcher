package runtimesvc

import (
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/types"
)

// sessionTracker remembers when each running instance was first seen, so a session keeps its start time.
type sessionTracker struct {
	mu        sync.Mutex
	firstSeen map[int]int64
}

var runningSessions = &sessionTracker{firstSeen: map[int]int64{}}

// getRunningSessions lists running game instances and syncs them into the activity store.
func getRunningSessions() ([]types.RunningSession, error) {
	return runningSessions.refresh(runningInstancePids(), readProcessCmdline, time.Now().UnixMilli())
}

// refresh builds the session list from the running instance PIDs. readCmdline returns a process's NUL-separated arguments.
func (t *sessionTracker) refresh(pids []int, readCmdline func(pid int) (string, error), now int64) ([]types.RunningSession, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(pids) == 0 {
		clear(t.firstSeen)
		return []types.RunningSession{}, syncActivitySessions(nil)
	}
	t.forgetExcept(pids)

	sessions := []types.RunningSession{}
	for _, pid := range pids {
		cmdline, err := readCmdline(pid)
		if err != nil {
			continue
		}
		gamePath, gameName := parseGameArguments(cmdline)
		if gamePath == "" {
			continue
		}
		cleanPath := filepath.Clean(gamePath)
		if gameName == "" {
			gameName = nameWithoutExt(filepath.Base(cleanPath))
		}
		if _, seen := t.firstSeen[pid]; !seen {
			t.firstSeen[pid] = now
		}
		sessions = append(sessions, types.RunningSession{
			Pid:       pid,
			GamePath:  cleanPath,
			GameName:  gameName,
			StartedAt: t.firstSeen[pid],
		})
	}
	return sessions, syncActivitySessions(sessions)
}

// forgetExcept drops start times for processes that are no longer running.
func (t *sessionTracker) forgetExcept(pids []int) {
	active := map[int]bool{}
	for _, pid := range pids {
		active[pid] = true
	}
	for pid := range t.firstSeen {
		if !active[pid] {
			delete(t.firstSeen, pid)
		}
	}
}

// parseGameArguments returns the last --game and --game-name values in a NUL-separated argument list.
func parseGameArguments(cmdline string) (string, string) {
	args := strings.Split(cmdline, "\x00")
	var gamePath, gameName string
	for i := range args {
		if i+1 >= len(args) {
			break
		}
		switch args[i] {
		case "--game":
			gamePath = args[i+1]
		case "--game-name":
			gameName = args[i+1]
		}
	}
	return gamePath, gameName
}

// killSession sends SIGINT so the instance can shut down cleanly. A missing process is not an error.
func killSession(pid int) error {
	if pid <= 0 {
		return nil
	}
	if err := interruptProcess(pid); err != nil && !errors.Is(err, errNoSuchProcess) {
		logger.Error("Session", "Failed to send SIGINT to pid %d: %v", pid, err)
	}
	return nil
}
