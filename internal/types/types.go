package types

type LsfgConfig struct {
	Enabled    bool    `json:"Enabled"`
	Multiplier string  `json:"Multiplier"`
	PerfMode   bool    `json:"PerfMode"`
	DllPath    string  `json:"DllPath"`
	Gpu        string  `json:"Gpu"`
	FlowScale  string  `json:"FlowScale"`
	Pacing     string  `json:"Pacing"`
	AllowFp16  bool    `json:"AllowFp16"`
}

type GamescopeConfig struct {
	Enabled     bool   `json:"Enabled"`
	Width       string `json:"Width"`
	Height      string `json:"Height"`
	RefreshRate string `json:"RefreshRate"`
}

type MemoryConfig struct {
	Enabled bool   `json:"Enabled"`
	Value   string `json:"Value"`
}

type ExtrasConfig struct {
	EnableMangoHud bool            `json:"EnableMangoHud"`
	EnableGamemode bool            `json:"EnableGamemode"`
	Lsfg           LsfgConfig      `json:"Lsfg"`
	Gamescope      GamescopeConfig `json:"Gamescope"`
	Memory         MemoryConfig    `json:"Memory"`
}

type LaunchOptions struct {
	Name          string       `json:"Name"`
	RunnerPath    string       `json:"RunnerPath"`
	GamePath      string       `json:"GamePath"`
	UseGamePath   bool         `json:"UseGamePath"`
	PrefixPath    string       `json:"PrefixPath"`
	ProtonPattern string       `json:"ProtonPattern"`
	ProtonPath    string       `json:"ProtonPath"`
	CustomArgs    string       `json:"CustomArgs"`
	Extras        ExtrasConfig `json:"Extras"`
}

type SystemToolsStatus struct {
	HasGamescope  bool `json:"hasGamescope"`
	HasMangoHud   bool `json:"hasMangoHud"`
	HasGameMode    bool `json:"hasGameMode"`
	HasVulkanInfo bool `json:"hasVulkanInfo"`
}

type SystemInfo struct {
	OS     string `json:"os"`
	Kernel string `json:"kernel"`
	CPU    string `json:"cpu"`
	GPU    string `json:"gpu"`
	RAM    string `json:"ram"`
	Driver string `json:"driver"`
}

type SystemUsage struct {
	CPU string `json:"cpu"`
	RAM string `json:"ram"`
	GPU string `json:"gpu"`
}

type LsfgProfileData struct {
	Name            string  `json:"name"`
	Multiplier      int     `json:"multiplier"`
	PerformanceMode bool    `json:"performanceMode"`
	GPU             string  `json:"gpu"`
	FlowScale       float32 `json:"flowScale"`
	Pacing          string  `json:"pacing"`
	DllPath         string  `json:"dllPath"`
	AllowFp16       bool    `json:"allowFp16"`
}

type GameInfo struct {
	Name     string        `json:"name"`
	Path     string        `json:"path"`
	Icon     string        `json:"icon"`
	Config   LaunchOptions `json:"config"`
	IsRecent bool          `json:"isRecent"`
}

type RunningSession struct {
	Pid      int    `json:"pid"`
	GamePath string `json:"gamePath"`
	GameName string `json:"gameName"`
}

type UtilsStatus struct {
	IsLsfgInstalled bool   `json:"isLsfgInstalled"`
	LsfgVersion     string `json:"lsfgVersion"`
}

type ProtonTool struct {
	Name        string `json:"Name"`
	Path        string `json:"Path"`
	IsSteam     bool   `json:"IsSteam"`
	DisplayName string `json:"DisplayName"`
}
