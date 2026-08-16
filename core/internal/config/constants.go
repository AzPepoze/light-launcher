package config

const (
	AppName = "LightLauncher"

	ConfigDirName   = "config/executables"
	PrefixesDirName = "prefixes"
	ProtonsDirName  = "protons"
	LogsDirName     = "logs"

	DefaultMemoryValue = "4G"
	DefaultMultiplier  = "2"
	DefaultWidth       = "1920"
	DefaultHeight      = "1080"
	DefaultRefreshRate = "60"
)

var DefaultExcludeNames = []string{
	"UnityCrashHandler64",
	"uninstall",
	"redist",
	"vc_redist",
	"dxsetup",
}
