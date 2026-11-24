package runtimesvc

import (
	"light-launcher/core/internal/api"
	"light-launcher/core/internal/config"
	"light-launcher/core/internal/drives"
)

func registerDrives(r *api.Router, enabled bool) {
	r.Handle("ListPartitions", listOrEmpty(enabled, plain(func() (any, error) {
		return drives.ListPartitions(), nil
	})))
	r.Handle("GetDriveSetup", listOrEmpty(enabled, plain(func() (any, error) {
		return drives.LoadSetup(config.DriveSetupFilePath())
	})))
	r.Handle("SaveDriveSetup", gated(enabled, "prefix", typed(func(in saveDriveSetupInput) (any, error) {
		return nil, drives.SaveSetup(config.DriveSetupFilePath(), in.Entries)
	})))
	r.Handle("ListPrefixDrives", gated(enabled, "prefix", typed(func(in prefixPathInput) (any, error) {
		return listPrefixDrives(in.PrefixPath)
	})))
}
