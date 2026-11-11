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
		deadline, ok := s.idleDeadline(conn, timeout)
		if !ok {
			return
		}
		if !deadline.After(s.cfg.now()) {
			if s.stopIfIdle(conn, timeout) {
				return
			}
			continue
		}
		timer := s.cfg.until(deadline)
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

// idleDeadline returns when the next idle check is due, and false when the watcher should end.
// It is an absolute time, so the delay between computing it and arming the timer cannot push it back.
func (s *Supervisor) idleDeadline(conn *connection, timeout time.Duration) (time.Time, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn != conn || conn.stopping {
		return time.Time{}, false
	}
	if s.inFlight > 0 || len(s.jobs) > 0 {
		return s.cfg.now().Add(timeout), true
	}
	return s.lastActive.Add(timeout), true
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
