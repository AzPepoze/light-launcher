package platform

// Current returns the features available on Linux.
func Current() Features {
	return Features{
		Proton:           true,
		Prefix:           true,
		Lsfg:             true,
		Gamescope:        true,
		MangoHud:         true,
		GameMode:         true,
		LiveTerminalLogs: true,
	}
}
