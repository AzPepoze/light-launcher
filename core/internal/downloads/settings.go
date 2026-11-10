package downloads

// Settings are the limits the user can change while the app runs.
type Settings struct {
	Connections     int  `json:"connections"`     // connections per download
	MaxActive       int  `json:"maxActive"`       // downloads running at once
	HostConnections int  `json:"hostConnections"` // connections to one host across all downloads
	Retries         int  `json:"retries"`         // retries per failed request
	SpeedLimitKBps  int  `json:"speedLimitKBps"`  // total speed in KiB/s; 0 is unlimited
	KeepArchives    bool `json:"keepArchives"`    // keep the archive after a successful install
}

const (
	maxActiveLimit  = 10
	maxRetriesLimit = 10
)

// DefaultSettings are used until the user changes them.
func DefaultSettings() Settings {
	return Settings{Connections: 8, MaxActive: 3, HostConnections: 8, Retries: 3}
}

// normalized clamps every field into its allowed range.
func (s Settings) normalized() Settings {
	s.Connections = ClampConnections(s.Connections)
	s.MaxActive = min(max(s.MaxActive, 1), maxActiveLimit)
	s.HostConnections = ClampConnections(s.HostConnections)
	s.Retries = min(max(s.Retries, 0), maxRetriesLimit)
	s.SpeedLimitKBps = max(s.SpeedLimitKBps, 0)
	return s
}

func bytesPerSecond(s Settings) float64 {
	return float64(s.SpeedLimitKBps) * 1024
}

// Settings returns the current settings.
func (m *Manager) Settings() Settings {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.settings
}

// SetSettings applies new settings at once and returns the normalized values.
func (m *Manager) SetSettings(s Settings) Settings {
	s = s.normalized()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings = s
	m.limiter.SetRate(bytesPerSecond(s))
	m.hosts.setLimit(s.HostConnections)
	m.pumpLocked()
	return s
}

// SetSpeedLimit changes only the total speed limit, in KiB/s, and returns the settings.
func (m *Manager) SetSpeedLimit(kbps int) Settings {
	s := m.Settings()
	s.SpeedLimitKBps = kbps
	return m.SetSettings(s)
}
