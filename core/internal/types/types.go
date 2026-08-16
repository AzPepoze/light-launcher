package types

type LsfgConfig struct {
	Enabled    bool   `json:"Enabled"`
	Multiplier string `json:"Multiplier"`
	PerfMode   bool   `json:"PerfMode"`
	DllPath    string `json:"DllPath"`
	Gpu        string `json:"Gpu"`
	FlowScale  string `json:"FlowScale"`
	Pacing     string `json:"Pacing"`
	AllowFp16  bool   `json:"AllowFp16"`
}

type GamescopeConfig struct {
	Enabled        bool   `json:"Enabled"`
	Width          string `json:"Width"`          // Nested width (-w)
	Height         string `json:"Height"`         // Nested height (-h)
	OutputWidth    string `json:"OutputWidth"`    // Output width (-W)
	OutputHeight   string `json:"OutputHeight"`   // Output height (-H)
	RefreshRate    string `json:"RefreshRate"`    // Nested refresh (-r)
	FramerateLimit string `json:"FramerateLimit"` // Frame limit (--framerate-limit)
	WindowMode     string `json:"WindowMode"`     // "fullscreen", "borderless", "windowed"
	Scaler         string `json:"Scaler"`         // auto, integer, fit, fill, stretch
	Filter         string `json:"Filter"`         // linear, nearest, fsr, nis, pixel
	Sharpness      string `json:"Sharpness"`      // 0-20
	HDR            bool   `json:"HDR"`
	AdaptiveSync   bool   `json:"AdaptiveSync"`
	Mangoapp       bool   `json:"Mangoapp"`
	CustomArgs     string `json:"CustomArgs"`
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
	ID              string       `json:"ID"`
	Name            string       `json:"Name"`
	LauncherPath    string       `json:"LauncherPath"`
	GamePath        string       `json:"GamePath"`
	UseGamePath     bool         `json:"UseGamePath"`
	PrefixPath      string       `json:"PrefixPath"`
	ProtonPath      string       `json:"ProtonPath"`
	UseCustomProton bool         `json:"UseCustomProton"`
	CustomArgs      string       `json:"CustomArgs"`
	Extras          ExtrasConfig `json:"Extras"`
}

type SystemToolsStatus struct {
	HasGamescope  bool `json:"hasGamescope"`
	HasMangoHud   bool `json:"hasMangoHud"`
	HasGameMode   bool `json:"hasGameMode"`
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
	Name          string        `json:"name"`
	Path          string        `json:"path"`
	Icon          string        `json:"icon"`
	Config        LaunchOptions `json:"config"`
	IsRecent      bool          `json:"isRecent"`
	IsAutoScanned bool          `json:"isAutoScanned"`
}

type ScannedFolderGroup struct {
	FolderPath string     `json:"folderPath"`
	FolderName string     `json:"folderName"`
	Games      []GameInfo `json:"games"`
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

type PrefixConfigWithProton struct {
	Config            LaunchOptions `json:"config"`
	ProtonDisplayName string        `json:"protonDisplayName"`
	ProtonName        string        `json:"protonName"`
	ProtonPath        string        `json:"protonPath"`
	ProtonIsSteam     bool          `json:"protonIsSteam"`
}

type ScanFolderConfig struct {
	Path         string   `json:"Path"`
	Depth        int      `json:"Depth"`
	ExcludeNames []string `json:"ExcludeNames"`
}

type AppSettings struct {
	TransparentMode   bool               `json:"TransparentMode"`
	NativeWayland     bool               `json:"NativeWayland"`
	ScanFolders       []string           `json:"ScanFolders"`
	ScanFolderConfigs []ScanFolderConfig `json:"ScanFolderConfigs"`
	Blacklist         []string           `json:"Blacklist"`
}
