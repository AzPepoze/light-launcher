package lsfg

type ConfigProfile struct {
	Name     string      `toml:"name"`
	ActiveIn interface{} `toml:"active_in"` // Can be string or []string
}

type ConfigFile struct {
	Profiles []ConfigProfile `toml:"profile"`
}
