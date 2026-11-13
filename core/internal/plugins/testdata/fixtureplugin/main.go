// Command fixtureplugin is a small plugin that speaks the launcher protocol. The supervisor tests run it.
//
// Methods:
//   - initialize, shutdown, ping: handshake, exit and liveness
//   - source.parseRef, source.echo: answer with their params
//   - source.prepareDownload: returns a URL with download hints
//   - source.probe {service, args}: asks the host for a service and returns what the host answered
//   - source.initializeProbe: returns what a settings.get during initialize saw, when the test asked for it
//   - source.wait {ms}: answers after a delay, to keep a call in flight
//   - source.progress {id, done, total}: reports a progress job
//   - source.crash: exits at once, to test restarts
//   - source.listGames {source, scope, cursor, limit}: pages through 250 fixed games
//   - source.browse {source, query, filters, sort, order, cursor, limit}: pages through a fixed catalog of 8 games
//   - source.listFilters {source, group, query, limit}: counts the tags of the catalog
//
// When the plugin data folder holds initializeSettingsMarker, initialize asks the host for its
// settings before answering, so a test can prove that host services run during the handshake.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	codeInvalidParams = -32602
	listGamesTotal    = 250
	listGamesMaxLimit = 200
	fixtureSource     = "fixture"
	fixtureScope      = "mine"
	// initializeSettingsMarker is a file name in the data folder that turns on the initialize probe.
	initializeSettingsMarker = "initialize-probes-settings"
)

// initializeProbe records the settings.get a probed initialize made. It is null unless the
// plugin data folder holds initializeSettingsMarker.
var initializeProbe map[string]any

type message struct {
	ID     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	Result json.RawMessage `json:"result"`
	Error  json.RawMessage `json:"error"`
}

type probeParams struct {
	Service string          `json:"service"`
	Args    json.RawMessage `json:"args"`
}

var (
	input       = bufio.NewReader(os.Stdin)
	output      = bufio.NewWriter(os.Stdout)
	hostCallIDs = 1000
)

func main() {
	fmt.Fprintln(os.Stderr, "fixture plugin started")
	for {
		msg, ok := readMessage()
		if !ok {
			return
		}
		if msg.Method != "" && len(msg.ID) > 0 {
			handle(msg)
		}
	}
}

func handle(msg message) {
	switch msg.Method {
	case "initialize":
		probeSettingsDuringInitialize(msg.Params)
		respond(msg.ID, map[string]any{
			"pluginVersion": "0.1.0",
			"capabilities": map[string]any{
				"fixture": map[string]any{"browse": true, "auth": "none", "updateCheck": false, "resolveFiles": false, "versioned": false},
			},
		})
	case "source.initializeProbe":
		respond(msg.ID, initializeProbe)
	case "shutdown":
		respond(msg.ID, nil)
		os.Exit(0)
	case "ping":
		respondError(msg.ID, -32601, "method not found")
	case "source.parseRef":
		respond(msg.ID, map[string]any{"kind": "game", "id": "42", "url": "https://example.test/g/42"})
	case "source.echo":
		notify("log", map[string]string{"level": "info", "message": "echo called"})
		respond(msg.ID, msg.Params)
	case "source.prepareDownload":
		respond(msg.ID, prepareDownload())
	case "source.probe":
		respond(msg.ID, probe(msg.Params))
	case "source.wait":
		wait(msg.Params)
		respond(msg.ID, nil)
	case "source.progress":
		progress(msg.Params)
		respond(msg.ID, nil)
	case "source.crash":
		os.Exit(3)
	case "source.listGames":
		answerListGames(msg)
	case "source.browse":
		answerBrowse(msg)
	case "source.listFilters":
		answerListFilters(msg)
	default:
		respondError(msg.ID, -32601, "method not found: "+msg.Method)
	}
}

// prepareDownload returns a fixed answer with every optional download hint.
func prepareDownload() map[string]any {
	return map[string]any{
		"url":              "https://files.example/game.zip",
		"headers":          map[string]string{"X-Token": "fixture"},
		"filename":         "game.zip",
		"size":             1024,
		"root":             "games",
		"subfolder":        "Team/Demo",
		"install":          true,
		"installSubfolder": "Pack",
		"deleteArchive":    false,
		"futureField":      "ignored",
	}
}

type listGamesParams struct {
	Source string `json:"source"`
	Scope  string `json:"scope"`
	Cursor string `json:"cursor"`
	Limit  int    `json:"limit"`
}

type fixtureGame struct {
	Source   string `json:"source"`
	SourceID string `json:"sourceId"`
	URL      string `json:"url"`
	Title    string `json:"title"`
}

type listGamesResult struct {
	Games      []fixtureGame `json:"games"`
	NextCursor string        `json:"nextCursor,omitempty"`
	Total      int           `json:"total"`
}

func answerListGames(msg message) {
	result, err := listGames(msg.Params)
	if err != nil {
		respondError(msg.ID, codeInvalidParams, err.Error())
		return
	}
	respond(msg.ID, result)
}

// listGames returns one page of the fixed game list. The cursor is the offset of the next game.
func listGames(raw json.RawMessage) (listGamesResult, error) {
	var params listGamesParams
	if err := json.Unmarshal(raw, &params); err != nil {
		return listGamesResult{}, fmt.Errorf("params: %w", err)
	}
	if params.Source != fixtureSource {
		return listGamesResult{}, fmt.Errorf("unknown source %q", params.Source)
	}
	if params.Scope != fixtureScope {
		return listGamesResult{}, fmt.Errorf("unknown scope %q", params.Scope)
	}
	if params.Limit < 1 || params.Limit > listGamesMaxLimit {
		return listGamesResult{}, fmt.Errorf("limit must be 1 to %d, not %d", listGamesMaxLimit, params.Limit)
	}
	offset, err := parseCursor(params.Cursor)
	if err != nil {
		return listGamesResult{}, err
	}
	end := min(offset+params.Limit, listGamesTotal)
	result := listGamesResult{Games: make([]fixtureGame, 0, end-offset), Total: listGamesTotal}
	for i := offset; i < end; i++ {
		result.Games = append(result.Games, fixtureGameAt(i))
	}
	if end < listGamesTotal {
		result.NextCursor = strconv.Itoa(end)
	}
	return result, nil
}

func parseCursor(cursor string) (int, error) {
	if cursor == "" {
		return 0, nil
	}
	offset, err := strconv.Atoi(cursor)
	if err != nil || offset < 0 || offset >= listGamesTotal {
		return 0, fmt.Errorf("bad cursor %q", cursor)
	}
	return offset, nil
}

// fixtureGameAt returns game number i (starting at 0). Its ids are stable across runs.
func fixtureGameAt(i int) fixtureGame {
	sourceID := fmt.Sprintf("game-%03d", i+1)
	return fixtureGame{
		Source:   fixtureSource,
		SourceID: sourceID,
		URL:      "https://example.test/g/" + sourceID,
		Title:    fmt.Sprintf("Fixture game %03d", i+1),
	}
}

// wait sleeps for the number of milliseconds in params.
func wait(raw json.RawMessage) {
	var params struct {
		MS int `json:"ms"`
	}
	_ = json.Unmarshal(raw, &params)
	time.Sleep(time.Duration(params.MS) * time.Millisecond)
}

// progress reports one progress job update.
func progress(raw json.RawMessage) {
	var params struct {
		ID    string `json:"id"`
		Done  int64  `json:"done"`
		Total int64  `json:"total"`
	}
	_ = json.Unmarshal(raw, &params)
	notify("progress", params)
}

// probeSettingsDuringInitialize asks the host for settings before initialize answers. It runs only
// when the test left initializeSettingsMarker in the plugin data folder.
func probeSettingsDuringInitialize(raw json.RawMessage) {
	var params struct {
		DataDir string `json:"dataDir"`
	}
	if err := json.Unmarshal(raw, &params); err != nil || params.DataDir == "" {
		return
	}
	if _, err := os.Stat(filepath.Join(params.DataDir, initializeSettingsMarker)); err != nil {
		return
	}
	result, hostErr := hostCall("settings.get", json.RawMessage(`{}`))
	if hostErr != nil {
		initializeProbe = map[string]any{"received": false, "error": string(hostErr)}
		return
	}
	var answer struct {
		Settings map[string]any `json:"settings"`
	}
	_ = json.Unmarshal(result, &answer)
	initializeProbe = map[string]any{"received": true, "settings": answer.Settings}
}

// probe asks the host for a service and returns the host's answer, so the test can check it.
func probe(raw json.RawMessage) map[string]json.RawMessage {
	var params probeParams
	_ = json.Unmarshal(raw, &params)
	result, hostErr := hostCall(params.Service, params.Args)
	return map[string]json.RawMessage{"result": result, "error": hostErr}
}

// hostCall sends a request to the host and waits for its answer. Other messages are ignored.
func hostCall(method string, params json.RawMessage) (json.RawMessage, json.RawMessage) {
	hostCallIDs++
	id := strconv.Itoa(hostCallIDs)
	send(map[string]any{"jsonrpc": "2.0", "id": json.RawMessage(id), "method": method, "params": params})
	for {
		msg, ok := readMessage()
		if !ok {
			os.Exit(4)
		}
		if string(msg.ID) == id {
			return msg.Result, msg.Error
		}
	}
}

func readMessage() (message, bool) {
	line, err := input.ReadBytes('\n')
	if err != nil {
		return message{}, false
	}
	var msg message
	if err := json.Unmarshal(line, &msg); err != nil {
		return message{}, true
	}
	return msg, true
}

func respond(id json.RawMessage, result any) {
	send(map[string]any{"jsonrpc": "2.0", "id": id, "result": result})
}

func respondError(id json.RawMessage, code int, text string) {
	send(map[string]any{"jsonrpc": "2.0", "id": id, "error": map[string]any{"code": code, "message": text}})
}

func respondErrorData(id json.RawMessage, code int, text string, data any) {
	send(map[string]any{"jsonrpc": "2.0", "id": id, "error": map[string]any{"code": code, "message": text, "data": data}})
}

func notify(method string, params any) {
	send(map[string]any{"jsonrpc": "2.0", "method": method, "params": params})
}

func send(value any) {
	data, _ := json.Marshal(value)
	_, _ = output.Write(append(data, '\n'))
	_ = output.Flush()
}
