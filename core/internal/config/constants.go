package config

const (
	AppName         = "light-launcher"
	ConfigDirName   = "config/executables"
	PrefixesDirName = "prefixes"
	LogsDirName     = "logs"

	DefaultMemoryValue = "4G"
	DefaultMultiplier  = "2"
	DefaultWidth       = "1280"
	DefaultHeight      = "720"
	DefaultRefreshRate = "60"
)

var DefaultExcludeNames = []string{
	"UnityCrashHandler64",
	"uninstall",
	"redist",
	"vc_redist",
	"dxsetup",
}
