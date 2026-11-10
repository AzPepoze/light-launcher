package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/logger"
)

const (
	lockFileName    = "launcher.lock"
	instanceTimeout = 2 * time.Second
)

// instanceInfo is what a running launcher writes to the lock file so a second launch can reach it.
type instanceInfo struct {
	Port  int    `json:"port"`
	Token string `json:"token"`
}

// lockFilePath is in the config folder, which only the user can read. The lock holds the API token.
func lockFilePath() string {
	return filepath.Join(config.GetBaseDirectory(), lockFileName)
}

func writeInstance(path string, info instanceInfo) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func readInstance(path string) (instanceInfo, error) {
	var info instanceInfo
	data, err := os.ReadFile(path)
	if err != nil {
		return info, err
	}
	err = json.Unmarshal(data, &info)
	return info, err
}

// removeInstance deletes the lock only if it still names this run's port, so a newer launcher keeps its lock.
func removeInstance(path string, port int) {
	info, err := readInstance(path)
	if err == nil && info.Port == port {
		_ = os.Remove(path)
	}
}

// forwardToRunning returns true when a running launcher took this launch over. With an empty
// gamePath it only checks that the running launcher answers.
func forwardToRunning(lockPath, gamePath string) bool {
	info, err := readInstance(lockPath)
	if err != nil {
		return false
	}
	if err := sendToRunning(info, gamePath); err != nil {
		logger.Warn("main", "no running launcher answered, starting a new one: %v", err)
		return false
	}
	logger.Info("main", "launcher is already running, request sent to it")
	return true
}

func sendToRunning(info instanceInfo, gamePath string) error {
	if gamePath == "" {
		return callInstance(info, http.MethodGet, "/api/platform", nil)
	}
	return callInstance(info, http.MethodPost, "/api/open", openRequest{Path: gamePath})
}

// openRequest mirrors the body of POST /api/open in the server package.
type openRequest struct {
	Path string `json:"path"`
}

// callInstance sends one authenticated request to the running launcher and fails on a non-2xx reply.
func callInstance(info instanceInfo, method, route string, body any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}
	url := fmt.Sprintf("http://127.0.0.1:%d%s", info.Port, route)
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-LL-Token", info.Token)

	client := &http.Client{Timeout: instanceTimeout}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return fmt.Errorf("running launcher answered %s", res.Status)
	}
	return nil
}
