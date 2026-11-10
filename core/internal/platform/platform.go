package platform

import "runtime"

// Features lists which launch tools and UI parts the current OS supports.
type Features struct {
	Proton           bool `json:"proton"`
	Prefix           bool `json:"prefix"`
	Lsfg             bool `json:"lsfg"`
	Gamescope        bool `json:"gamescope"`
	MangoHud         bool `json:"mangoHud"`
	GameMode         bool `json:"gameMode"`
	LiveTerminalLogs bool `json:"liveTerminalLogs"`
}

// OS returns the target operating system, "linux" or "windows".
func OS() string {
	return runtime.GOOS
}
