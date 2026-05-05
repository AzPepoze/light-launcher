package app

import (
	"light-launcher/internal/config"
)

func (app *App) ListPrefixes() ([]string, error) {
	return config.ListPrefixes()
}

func (app *App) CreatePrefix(name string) error {
	return config.CreatePrefix(name)
}

func (app *App) GetPrefixBaseDir() string {
	return config.GetPrefixBaseDirectory()
}

func (app *App) RemovePrefix(name string) error {
	return config.RemovePrefix(name)
}
