package plugins

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// ErrPeerClosed is returned for calls that were pending when the connection ended.
var ErrPeerClosed = errors.New("plugin connection closed")

// serveTimeout bounds how long one plugin request to the host may run.
const serveTimeout = 2 * time.Minute

// requestHandler answers a request from the other side. Its error becomes the response error.
type requestHandler func(ctx context.Context, method string, params json.RawMessage) (any, error)

// notifyHandler receives a notification from the other side.
type notifyHandler func(method string, params json.RawMessage)

// Peer is one end of a newline-delimited JSON-RPC 2.0 connection. Both ends can send requests.
type Peer struct {
	writer    io.Writer
	writeMu   sync.Mutex
	nextID    atomic.Int64
	mu        sync.Mutex
	pending   map[string]chan response
	closed    bool
	onRequest requestHandler
	onNotify  notifyHandler
	done      chan struct{}
}

type response struct {
	result json.RawMessage
	err    *RPCError
}

// NewPeer starts reading lines from reader. Writes go to writer.
func NewPeer(reader io.Reader, writer io.Writer, onRequest requestHandler, onNotify notifyHandler) *Peer {
	peer := &Peer{
		writer:    writer,
		pending:   map[string]chan response{},
		onRequest: onRequest,
		onNotify:  onNotify,
		done:      make(chan struct{}),
	}
	go peer.readLoop(reader)
	return peer
}

// Done is closed when the reading side ends, because the stream closed or a line was too long.
func (p *Peer) Done() <-chan struct{} {
	return p.done
}

// Call sends a request and waits for its answer. params and out are JSON-encoded and decoded.
// When ctx ends first, Call sends a cancel notification and returns a timeout or canceled RPCError.
func (p *Peer) Call(ctx context.Context, method string, params, out any) error {
	rawParams, err := marshalParams(params)
	if err != nil {
		return err
	}
	id := strconv.FormatInt(p.nextID.Add(1), 10)
	answer := make(chan response, 1)
	if err := p.register(id, answer); err != nil {
		return err
	}
	defer p.unregister(id)

	request := envelope{JSONRPC: "2.0", ID: json.RawMessage(id), Method: method, Params: rawParams}
	if err := p.send(request); err != nil {
		return err
	}
	select {
	case resp := <-answer:
		return decodeResponse(resp, out)
	case <-ctx.Done():
		_ = p.Notify(cancelMethod, map[string]string{"id": id})
		return contextError(ctx.Err())
	case <-p.done:
		return ErrPeerClosed
	}
}

// Notify sends a notification. It gets no answer.
func (p *Peer) Notify(method string, params any) error {
	rawParams, err := marshalParams(params)
	if err != nil {
		return err
	}
	return p.send(envelope{JSONRPC: "2.0", Method: method, Params: rawParams})
}

func (p *Peer) readLoop(reader io.Reader) {
	defer p.finish()
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64<<10), maxLineBytes)
	for scanner.Scan() {
		p.handleLine(scanner.Bytes())
	}
}

func (p *Peer) finish() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	p.pending = map[string]chan response{}
	close(p.done)
}

func (p *Peer) handleLine(line []byte) {
	if len(bytes.TrimSpace(line)) == 0 {
		return
	}
	var msg envelope
	if err := json.Unmarshal(line, &msg); err != nil || msg.JSONRPC != "2.0" {
		_ = p.writeResponse(json.RawMessage("null"), nil, &RPCError{Code: CodeParseError, Message: "line is not a JSON-RPC 2.0 message"})
		return
	}
	switch {
	case msg.Method != "" && len(msg.ID) == 0:
		p.dispatchNotify(msg)
	case msg.Method != "":
		go p.serve(msg)
	default:
		p.deliver(msg)
	}
}

func (p *Peer) dispatchNotify(msg envelope) {
	if p.onNotify != nil {
		p.onNotify(msg.Method, msg.Params)
	}
}

// serve answers one request from the other side. The response is written even when the handler fails.
func (p *Peer) serve(msg envelope) {
	ctx, cancel := context.WithTimeout(context.Background(), serveTimeout)
	defer cancel()
	ctx = context.WithValue(ctx, requestIDKey{}, requestIDText(msg.ID))
	result, err := p.handle(ctx, msg)
	if err != nil {
		_ = p.writeResponse(msg.ID, nil, toRPCError(err))
		return
	}
	raw, err := json.Marshal(result)
	if err != nil {
		_ = p.writeResponse(msg.ID, nil, &RPCError{Code: CodeInternal, Message: err.Error()})
		return
	}
	_ = p.writeResponse(msg.ID, raw, nil)
}

func (p *Peer) handle(ctx context.Context, msg envelope) (any, error) {
	if p.onRequest == nil {
		return nil, &RPCError{Code: CodeMethodNotFound, Message: "method not found: " + msg.Method}
	}
	return p.onRequest(ctx, msg.Method, msg.Params)
}

func (p *Peer) deliver(msg envelope) {
	key := string(msg.ID)
	p.mu.Lock()
	answer, ok := p.pending[key]
	delete(p.pending, key)
	p.mu.Unlock()
	if ok {
		answer <- response{result: msg.Result, err: msg.Error}
	}
}

func (p *Peer) writeResponse(id json.RawMessage, result json.RawMessage, rpcErr *RPCError) error {
	reply := envelope{JSONRPC: "2.0", ID: id, Error: rpcErr}
	if rpcErr == nil {
		reply.Result = result
		if len(reply.Result) == 0 {
			reply.Result = json.RawMessage("null")
		}
	}
	return p.send(reply)
}

// send writes one message and its newline in a single write, under a lock, so lines never interleave.
func (p *Peer) send(msg envelope) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	p.writeMu.Lock()
	defer p.writeMu.Unlock()
	_, err = p.writer.Write(data)
	return err
}

func (p *Peer) register(id string, answer chan response) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return ErrPeerClosed
	}
	p.pending[id] = answer
	return nil
}

func (p *Peer) unregister(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.pending, id)
}

func marshalParams(params any) (json.RawMessage, error) {
	if params == nil {
		return nil, nil
	}
	return json.Marshal(params)
}

func decodeResponse(resp response, out any) error {
	if resp.err != nil {
		return resp.err
	}
	if out == nil || len(resp.result) == 0 {
		return nil
	}
	return json.Unmarshal(resp.result, out)
}

func contextError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return &RPCError{Code: CodeTimeout, Message: "request timed out"}
	}
	return &RPCError{Code: CodeCanceled, Message: "request canceled"}
}

// requestIDKey is the context key for the id of the request being answered.
type requestIDKey struct{}

// requestIDFrom returns the id of the plugin request being answered, as text.
func requestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// requestIDText returns a request id as plain text: 7 stays 7, and "abc" becomes abc.
func requestIDText(raw json.RawMessage) string {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text
	}
	return string(raw)
}

// toRPCError keeps an RPCError as it is. Any other error becomes an internal error.
func toRPCError(err error) *RPCError {
	var rpcErr *RPCError
	if errors.As(err, &rpcErr) {
		return rpcErr
	}
	return &RPCError{Code: CodeInternal, Message: err.Error()}
}
