package downloads

// Status is what the downloads page shows besides the list: running installs and tool and limit state.
type Status struct {
	Installs       []InstallProgress `json:"installs"`
	SevenZip       bool              `json:"sevenZip"`
	SpeedLimitKBps int               `json:"speedLimitKBps"`
}

// installLister is implemented by installers that can report the installs they run.
type installLister interface {
	ActiveInstalls() []InstallProgress
}

// Status returns the running installs, whether 7-Zip is on PATH, and the speed limit.
func (m *Manager) Status() Status {
	installs := []InstallProgress{}
	if lister, ok := m.installer.(installLister); ok {
		installs = append(installs, lister.ActiveInstalls()...)
	}
	m.mu.Lock()
	limit := m.settings.SpeedLimitKBps
	m.mu.Unlock()
	return Status{Installs: installs, SevenZip: HasSevenZip(), SpeedLimitKBps: limit}
}
