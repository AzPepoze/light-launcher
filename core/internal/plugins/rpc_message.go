package plugins

import (
	"encoding/json"
	"fmt"
)

// Error codes from schema/protocol.schema.json. The -32000 range belongs to the launcher.
const (
	CodeParseError       = -32700
	CodeInvalidRequest   = -32600
	CodeMethodNotFound   = -32601
	CodeInvalidParams    = -32602
	CodeInternal         = -32603
	CodePermissionDenied = -32001
	CodeCanceled         = -32002
	CodeTimeout          = -32003
	CodeRateLimited      = -32004
	CodeRemoteError      = -32005
	CodeNotModified      = -32006
)

// cancelMethod is the notification that asks the other side to stop a request.
const cancelMethod = "cancel"

// maxLineBytes bounds one protocol line, so a broken plugin cannot exhaust memory.
const maxLineBytes = 32 << 20

// RPCError is a JSON-RPC error object. Its code is stable and safe to show to the UI.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("%s (code %d)", e.Message, e.Code)
}

// envelope is one JSON-RPC message. Requests have a method and an id, notifications have no id.
type envelope struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}
