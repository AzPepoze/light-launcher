package plugins

import (
	"time"

	"light-launcher/core/internal/logger"
)

// watchIdle stops the plugin when no call has touched it for the idle timeout.
func (s *Supervisor) watchIdle(conn *connection) {
	timeout := *s.cfg.idleTimeout
	if timeout <= 0 {
		return
	}
	for {
		wait, ok := s.idleWait(conn, timeout)
		if !ok {
			return
		}
		if wait <= 0 {
			if s.stopIfIdle(conn, timeout) {
				return
			}
			continue
		}
		timer := s.cfg.after(wait)
		select {
		case <-conn.exited:
			return
		case <-s.activityCh:
			continue
		case <-timer:
			if s.stopIfIdle(conn, timeout) {
				return
			}
		}
	}
}

// idleWait returns how long to wait before the next idle check, and false when the watcher should end.
func (s *Supervisor) idleWait(conn *connection, timeout time.Duration) (time.Duration, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn != conn || conn.stopping {
		return 0, false
	}
	if s.inFlight > 0 || len(s.jobs) > 0 {
		return timeout, true
	}
	remaining := timeout - s.cfg.now().Sub(s.lastActive)
	if remaining < 0 {
		remaining = 0
	}
	return remaining, true
}

// stopIfIdle shuts the connection down when it is still idle. It reports whether the watcher is done.
func (s *Supervisor) stopIfIdle(conn *connection, timeout time.Duration) bool {
	s.mu.Lock()
	if s.conn != conn || conn.stopping {
		s.mu.Unlock()
		return true
	}
	if s.inFlight > 0 || len(s.jobs) > 0 || s.cfg.now().Sub(s.lastActive) < timeout {
		s.mu.Unlock()
		return false
	}
	conn.stopping = true
	s.idle = true
	s.mu.Unlock()

	logger.Info(s.cfg.env.category(), "stopping after %s idle", timeout)
	s.stopConn(conn)
	s.cfg.onIdle()
	return true
}
