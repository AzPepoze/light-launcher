package plugins

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

// pipePeers connects a host peer and a plugin peer with two in-memory pipes.
func pipePeers(hostHandler, pluginHandler requestHandler, pluginNotify notifyHandler) (*Peer, *Peer) {
	pluginReads, hostToPlugin := io.Pipe()
	hostReads, pluginToHost := io.Pipe()
	host := NewPeer(hostReads, hostToPlugin, hostHandler, nil)
	plugin := NewPeer(pluginReads, pluginToHost, pluginHandler, pluginNotify)
	return host, plugin
}

func echoHandler(_ context.Context, _ string, params json.RawMessage) (any, error) {
	return params, nil
}

func TestCallRoundTrip(t *testing.T) {
	host, _ := pipePeers(nil, echoHandler, nil)
	var out map[string]string
	if err := host.Call(context.Background(), "echo", map[string]string{"query": "space"}, &out); err != nil {
		t.Fatalf("Call: %v", err)
	}
	if out["query"] != "space" {
		t.Fatalf("out = %v", out)
	}
}

func TestCallReturnsRPCErrorCode(t *testing.T) {
	deny := func(context.Context, string, json.RawMessage) (any, error) {
		return nil, &RPCError{Code: CodePermissionDenied, Message: "network denied"}
	}
	host, _ := pipePeers(nil, deny, nil)
	err := host.Call(context.Background(), "anything", nil, nil)
	var rpcErr *RPCError
	if !errors.As(err, &rpcErr) || rpcErr.Code != CodePermissionDenied {
		t.Fatalf("err = %v, want permission denied", err)
	}
}

func TestRequestsCanGoBothWays(t *testing.T) {
	var plugin *Peer
	pluginHandler := func(ctx context.Context, _ string, _ json.RawMessage) (any, error) {
		var fromHost map[string]string
		if err := plugin.Call(ctx, "host.lookup", map[string]string{"key": "a"}, &fromHost); err != nil {
			return nil, err
		}
		return map[string]string{"got": fromHost["value"]}, nil
	}
	hostHandler := func(context.Context, string, json.RawMessage) (any, error) {
		return map[string]string{"value": "stored"}, nil
	}
	var host *Peer
	host, plugin = pipePeers(hostHandler, pluginHandler, nil)

	var out map[string]string
	if err := host.Call(context.Background(), "plugin.run", nil, &out); err != nil {
		t.Fatalf("Call: %v", err)
	}
	if out["got"] != "stored" {
		t.Fatalf("out = %v", out)
	}
}

func TestNotificationsReachTheHandler(t *testing.T) {
	received := make(chan string, 1)
	host, _ := pipePeers(nil, nil, func(method string, _ json.RawMessage) {
		received <- method
	})
	if err := host.Notify("progress", map[string]int{"done": 1}); err != nil {
		t.Fatal(err)
	}
	select {
	case method := <-received:
		if method != "progress" {
			t.Fatalf("method = %q", method)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("notification was not delivered")
	}
}

func TestCancelSendsCancelNotification(t *testing.T) {
	started := make(chan struct{})
	canceled := make(chan struct{})
	slow := func(ctx context.Context, _ string, _ json.RawMessage) (any, error) {
		close(started)
		<-ctx.Done()
		return nil, ctx.Err()
	}
	notify := func(method string, _ json.RawMessage) {
		if method == cancelMethod {
			close(canceled)
		}
	}
	host, _ := pipePeers(nil, slow, notify)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-started
		cancel()
	}()
	err := host.Call(ctx, "slow", nil, nil)
	var rpcErr *RPCError
	if !errors.As(err, &rpcErr) || rpcErr.Code != CodeCanceled {
		t.Fatalf("err = %v, want canceled", err)
	}
	select {
	case <-canceled:
	case <-time.After(2 * time.Second):
		t.Fatal("cancel notification was not sent")
	}
}

func TestDeadlineReturnsTimeoutCode(t *testing.T) {
	block := func(ctx context.Context, _ string, _ json.RawMessage) (any, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	host, _ := pipePeers(nil, block, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err := host.Call(ctx, "slow", nil, nil)
	var rpcErr *RPCError
	if !errors.As(err, &rpcErr) || rpcErr.Code != CodeTimeout {
		t.Fatalf("err = %v, want timeout", err)
	}
}

func TestPendingCallFailsWhenStreamCloses(t *testing.T) {
	hostReads, pluginToHost := io.Pipe()
	host := NewPeer(hostReads, io.Discard, nil, nil)
	go func() {
		time.Sleep(20 * time.Millisecond)
		_ = pluginToHost.Close()
	}()
	err := host.Call(context.Background(), "wait", nil, nil)
	if !errors.Is(err, ErrPeerClosed) {
		t.Fatalf("err = %v, want ErrPeerClosed", err)
	}
	<-host.Done()
}

func TestBadLineGetsParseError(t *testing.T) {
	var output strings.Builder
	peer := NewPeer(strings.NewReader("this is not json\n"), &output, nil, nil)
	<-peer.Done()

	line, err := bufio.NewReader(strings.NewReader(output.String())).ReadString('\n')
	if err != nil {
		t.Fatalf("no response written: %v", err)
	}
	if !strings.Contains(line, `"id":null`) || !strings.Contains(line, "-32700") {
		t.Fatalf("response = %s", line)
	}
}
