package main

import (
	"io"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/logger"
)

// serviceRegistrars lists the service packages. Each entry adds its methods to the router.
// Append new services here, one entry per package.
var serviceRegistrars = []func(*api.Router, api.Events) error{
	// (no services yet)
}

// PluginCLI runs `light-launcher plugin ...`. It stays nil until the plugin host is wired in.
var PluginCLI func(args []string, out io.Writer) int

func registerServices(r *api.Router, hub *api.Hub) error {
	for _, register := range serviceRegistrars {
		if err := register(r, hub); err != nil {
			return err
		}
	}
	logger.Info("api", "registered %d methods", len(r.Methods()))
	return nil
}
