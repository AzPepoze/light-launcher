package plugins

import (
	"context"
	"errors"
	"time"

	"light-launcher/core/internal/logger"
)

const (
	defaultHeartbeat = 15 * time.Second
	pingTimeout      = 5 * time.Second
	// pingMethod is not in the SDK protocol yet. A plugin that does not know it still answers, which proves it is alive.
	pingMethod = "ping"
)

// heartbeat pings the plugin on an interval. A plugin that never answers is killed, and that counts as a crash.
func (s *Supervisor) heartbeat(conn *connection) {
	ticker := time.NewTicker(s.cfg.heartbeat)
	defer ticker.Stop()
	for {
		select {
		case <-conn.exited:
			return
		case <-ticker.C:
			if !s.ping(conn) {
				logger.Warn(s.cfg.env.category(), "plugin did not answer ping, stopping it")
				conn.child.kill()
				return
			}
		}
	}
}

// ping returns true when the plugin answers at all, even with an error other than a timeout or cancel.
func (s *Supervisor) ping(conn *connection) bool {
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	err := conn.peer.Call(ctx, pingMethod, nil, nil)
	if err == nil {
		return true
	}
	var rpcErr *RPCError
	if !errors.As(err, &rpcErr) {
		return false
	}
	return rpcErr.Code != CodeTimeout && rpcErr.Code != CodeCanceled
}
