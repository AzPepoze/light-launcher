package plugins

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/logger"
)

// serviceDeps holds the launcher services that plugins reach through permission checks.
type serviceDeps struct {
	secrets   contract.Secrets
	downloads contract.Downloads
	library   contract.Library
	browser   BrowserAssist
	folders   map[string]string
	transport http.RoundTripper
	limiter   *hostLimiter
	cache     *etagCache
	events    api.Events
}

// pluginEnv is what the host gives one running plugin: its identity, grants, data and logs.
type pluginEnv struct {
	id             string
	manifest       Manifest
	dataDir        string
	settingsSchema map[string]any // nil when the plugin has no settings
	deps           *serviceDeps
	lines          *lineRing
	storageMu      sync.Mutex
	peerMu         sync.Mutex
	peer           *Peer // the connection of the running plugin, set on each start
}

// setPeer records the connection of a newly started plugin.
func (env *pluginEnv) setPeer(peer *Peer) {
	env.peerMu.Lock()
	defer env.peerMu.Unlock()
	env.peer = peer
}

// notifyPlugin sends a notification to the running plugin. It is dropped when no plugin is connected.
func (env *pluginEnv) notifyPlugin(method string, params any) {
	env.peerMu.Lock()
	peer := env.peer
	env.peerMu.Unlock()
	if peer == nil {
		return
	}
	if err := peer.Notify(method, params); err != nil {
		logger.Debug(env.category(), "could not send %s: %v", method, err)
	}
}

func newPluginEnv(manifest Manifest, dataDir string, deps *serviceDeps, lines *lineRing) *pluginEnv {
	return &pluginEnv{id: manifest.ID, manifest: manifest, dataDir: dataDir, deps: deps, lines: lines}
}

// call runs a plugin request to the host. It checks the manifest grant before the service runs.
func (env *pluginEnv) call(ctx context.Context, method string, params json.RawMessage) (any, error) {
	svc, ok := services[method]
	if !ok {
		return nil, &RPCError{Code: CodeMethodNotFound, Message: "unknown host service " + method}
	}
	if svc.needs != nil && !svc.needs(env.manifest.Permissions) {
		return nil, permissionDenied(method, svc.grantName)
	}
	return svc.run(ctx, env, params)
}

// notify handles log, notify and progress notifications from the plugin.
func (env *pluginEnv) notify(method string, params json.RawMessage) {
	switch method {
	case "log":
		env.handleLog(params)
	case "notify":
		env.handleNotify(params)
	case "progress":
		env.handleProgress(params)
	default:
		logger.Debug(env.category(), "ignored notification %s", method)
	}
}

type logParams struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

func (env *pluginEnv) handleLog(params json.RawMessage) {
	var entry logParams
	if err := json.Unmarshal(params, &entry); err != nil {
		env.log("warn", "log notification with bad params")
		return
	}
	env.log(entry.Level, entry.Message)
}

func (env *pluginEnv) handleNotify(params json.RawMessage) {
	var entry logParams
	if err := json.Unmarshal(params, &entry); err != nil {
		return
	}
	env.log(entry.Level, entry.Message)
	env.emit("plugin-notify", map[string]string{"plugin": env.id, "level": entry.Level, "message": entry.Message})
}

type progressParams struct {
	ID      string `json:"id"`
	Done    int64  `json:"done"`
	Total   int64  `json:"total"`
	Message string `json:"message"`
}

func (env *pluginEnv) handleProgress(params json.RawMessage) {
	var progress progressParams
	if err := json.Unmarshal(params, &progress); err != nil {
		return
	}
	env.emit("plugin-progress", map[string]any{
		"plugin":  env.id,
		"id":      progress.ID,
		"done":    progress.Done,
		"total":   progress.Total,
		"message": progress.Message,
	})
}

// log writes one line to the launcher log and to the plugin's log buffer.
func (env *pluginEnv) log(level, message string) {
	env.lines.add(LogLine{Time: time.Now().UTC().Format(time.RFC3339Nano), Level: level, Text: message})
	switch strings.ToLower(level) {
	case "debug":
		logger.Debug(env.category(), "%s", message)
	case "warn", "warning":
		logger.Warn(env.category(), "%s", message)
	case "error":
		logger.Error(env.category(), "%s", message)
	default:
		logger.Info(env.category(), "%s", message)
	}
}

func (env *pluginEnv) category() string {
	return "plugin:" + env.id
}

func (env *pluginEnv) emit(channel string, data any) {
	if env.deps.events != nil {
		env.deps.events.Emit(channel, data)
	}
}

// sourceIDs returns the source ids this plugin contributes.
func (env *pluginEnv) sourceIDs() map[string]bool {
	ids := map[string]bool{}
	for _, source := range env.manifest.Contributes.Sources {
		ids[source.ID] = true
	}
	return ids
}

// LogLine is one line a plugin logged. Time is when the launcher received it, in RFC 3339 with UTC.
type LogLine struct {
	Time  string `json:"time"`
	Level string `json:"level,omitempty"`
	Text  string `json:"text"`
}

// lineRing keeps the last lines a plugin wrote, for the UI log view.
type lineRing struct {
	mu    sync.Mutex
	max   int
	lines []LogLine
}

func newLineRing(max int) *lineRing {
	return &lineRing{max: max}
}

func (r *lineRing) add(line LogLine) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lines = append(r.lines, line)
	if over := len(r.lines) - r.max; over > 0 {
		r.lines = append([]LogLine(nil), r.lines[over:]...)
	}
}

// tail returns the last n lines, or all of them when n is 0 or larger.
func (r *lineRing) tail(n int) []LogLine {
	r.mu.Lock()
	defer r.mu.Unlock()
	if n <= 0 || n > len(r.lines) {
		n = len(r.lines)
	}
	return append([]LogLine(nil), r.lines[len(r.lines)-n:]...)
}
