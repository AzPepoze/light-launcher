// Command fixtureplugin is a small plugin that speaks the launcher protocol. The supervisor tests run it.
//
// Methods:
//   - initialize, shutdown, ping: handshake, exit and liveness
//   - source.parseRef, source.echo: answer with their params
//   - source.probe {service, args}: asks the host for a service and returns what the host answered
//   - source.crash: exits at once, to test restarts
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

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
		respond(msg.ID, map[string]any{
			"pluginVersion": "0.1.0",
			"capabilities": map[string]any{
				"fixture": map[string]any{"browse": true, "auth": "none", "updateCheck": false, "resolveFiles": false, "versioned": false},
			},
		})
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
	case "source.probe":
		respond(msg.ID, probe(msg.Params))
	case "source.crash":
		os.Exit(3)
	default:
		respondError(msg.ID, -32601, "method not found: "+msg.Method)
	}
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

func notify(method string, params any) {
	send(map[string]any{"jsonrpc": "2.0", "method": method, "params": params})
}

func send(value any) {
	data, _ := json.Marshal(value)
	_, _ = output.Write(append(data, '\n'))
	_ = output.Flush()
}
