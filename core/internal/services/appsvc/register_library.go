package appsvc

import (
	"context"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/types"
)

func registerSettingsMethods(r *api.Router) {
	r.Handle("GetAppSettings", api.NoInput(func(context.Context) (types.AppSettings, error) {
		return loadAppSettings(), nil
	}))
	r.Handle("SaveAppSettings", api.Typed(func(_ context.Context, in settingsPayload) (any, error) {
		return nil, saveAppSettings(in.Settings)
	}))
	r.Handle("GetConfig", api.Typed(func(_ context.Context, in executablePayload) (*types.LaunchOptions, error) {
		return loadGameConfig(in.ExecutablePath), nil
	}))
	r.Handle("SaveGameConfig", api.Typed(func(_ context.Context, in optionsPayload) (any, error) {
		return nil, saveGameConfig(&in.Options)
	}))
}

func registerGameMethods(r *api.Router) {
	r.Handle("GetAllGames", api.NoInput(func(context.Context) ([]types.GameInfo, error) {
		return getAllGames(), nil
	}))
	r.Handle("RemoveGame", api.Typed(func(_ context.Context, in executablePayload) (any, error) {
		return nil, removeGame(in.ExecutablePath)
	}))
	r.Handle("SearchExecutables", api.Typed(func(_ context.Context, in searchPayload) ([]string, error) {
		return searchExecutables(in.FolderPath, depthOrUnlimited(in.MaxDepth), in.ExcludeNames), nil
	}))
	r.Handle("GetAutoScannedGames", api.Typed(func(_ context.Context, in forcePayload) ([]types.ScannedFolderGroup, error) {
		return getAutoScannedGames(in.Force)
	}))
	r.Handle("AddScanFolder", api.Typed(func(_ context.Context, in folderPayload) (any, error) {
		return nil, addScanFolder(in.FolderPath)
	}))
	r.Handle("RemoveScanFolder", api.Typed(func(_ context.Context, in folderPayload) (any, error) {
		return nil, removeScanFolder(in.FolderPath)
	}))
	r.Handle("UpdateScanFolderConfig", api.Typed(func(_ context.Context, in folderConfigPayload) (any, error) {
		return nil, updateScanFolderConfig(in.FolderPath, in.Depth, in.ExcludeNames)
	}))
	r.Handle("RenameScanFolder", api.Typed(func(_ context.Context, in renamePayload) (string, error) {
		return renameScanFolder(in.OldPath, in.NewPath, in.Depth, in.ExcludeNames)
	}))
	r.Handle("GetScanFolderConfig", api.Typed(func(_ context.Context, in folderPayload) (types.ScanFolderConfig, error) {
		return getScanFolderConfig(in.FolderPath), nil
	}))
	r.Handle("BlacklistGame", api.Typed(func(_ context.Context, in executablePayload) (any, error) {
		return nil, blacklistGame(in.ExecutablePath)
	}))
	r.Handle("UnblacklistGame", api.Typed(func(_ context.Context, in executablePayload) (any, error) {
		return nil, unblacklistGame(in.ExecutablePath)
	}))
}

func registerActivityMethods(r *api.Router) {
	r.Handle("GetGameActivity", api.NoInput(func(context.Context) ([]types.GameActivity, error) {
		return getActivities(), nil
	}))
}

// depthOrUnlimited maps an omitted depth to an unlimited scan.
func depthOrUnlimited(depth *int) int {
	if depth == nil {
		return unlimitedDepth
	}
	return *depth
}
