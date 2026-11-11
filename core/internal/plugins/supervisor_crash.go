package plugins

import (
	"fmt"
	"time"

	"light-launcher/core/internal/logger"
)

const (
	defaultBackoff     = time.Second
	maxBackoff         = 30 * time.Second
	defaultCrashWindow = 5 * time.Minute
	defaultCrashLimit  = 3
)

// watch reaps the process once its output has ended, then records the exit.
func (s *Supervisor) watch(conn *connection) {
	<-conn.peer.Done()
	<-conn.child.stderrDone
	_ = conn.child.wait(exitGrace)
	reason := s.onExit(conn)
	close(conn.exited)
	if reason != "" {
		s.notifyDisabled(reason)
	}
}

// onExit clears the connection and counts the exit as a crash, unless Shutdown asked for it.
func (s *Supervisor) onExit(conn *connection) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn == conn {
		s.conn = nil
		s.jobs = map[string]bool{}
		s.cfg.env.clearPeer(conn.peer)
	}
	if conn.stopping {
		return ""
	}
	logger.Warn(s.cfg.env.category(), "plugin process ended unexpectedly")
	return s.recordCrashLocked(time.Now())
}

// recordCrashLocked counts a crash. It returns a disable reason when the crash limit is reached.
func (s *Supervisor) recordCrashLocked(now time.Time) string {
	s.crashes = append(recentTimes(s.crashes, now.Add(-s.cfg.crashWindow)), now)
	if len(s.crashes) >= s.cfg.crashLimit {
		s.disabled = true
		return fmt.Sprintf("crashed %d times within %s", len(s.crashes), s.cfg.crashWindow)
	}
	delay := backoffFor(len(s.crashes), s.cfg.backoff)
	s.nextStart = now.Add(delay)
	logger.Warn(s.cfg.env.category(), "restart in %s", delay)
	return ""
}

func (s *Supervisor) notifyDisabled(reason string) {
	logger.Error(s.cfg.env.category(), "plugin disabled: %s", reason)
	s.cfg.onDisable(reason)
}

// recentTimes keeps the times after since.
func recentTimes(times []time.Time, since time.Time) []time.Time {
	var kept []time.Time
	for _, at := range times {
		if at.After(since) {
			kept = append(kept, at)
		}
	}
	return kept
}

// backoffFor doubles the base delay for each crash in the window, up to maxBackoff.
func backoffFor(crashes int, base time.Duration) time.Duration {
	delay := base
	for i := 1; i < crashes && delay < maxBackoff; i++ {
		delay *= 2
	}
	return min(delay, maxBackoff)
}
