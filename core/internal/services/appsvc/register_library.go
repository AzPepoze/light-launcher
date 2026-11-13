package appsvc

import (
	"context"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/logger"
	"light-launcher/core/internal/types"
)

func registerSettingsMethods(r *api.Router, svc *service) {
	r.Handle("GetAppSettings", api.NoInput(func(context.Context) (types.AppSettings, error) {
		return svc.settings.Load(), nil
	}))
	r.Handle("SaveAppSettings", api.Typed(func(_ context.Context, in settingsPayload) (any, error) {
		return nil, svc.settings.Save(in.Settings)
	}))
	r.Handle("GetConfig", api.Typed(func(ctx context.Context, in executablePayload) (*types.LaunchOptions, error) {
		if svc.localView != nil {
			if config := svc.localView.LocalConfig(ctx, in.ExecutablePath); config != nil {
				return config, nil
			}
		}
		return svc.games.Find(in.ExecutablePath), nil
	}))
	r.Handle("SaveGameConfig", api.Typed(func(ctx context.Context, in optionsPayload) (any, error) {
		if svc.localView != nil {
			if err := svc.localView.SaveLocalConfig(ctx, in.Options); err != nil {
				return nil, err
			}
		}
		return nil, svc.games.Save(&in.Options)
	}))
}

func registerGameMethods(r *api.Router, svc *service) {
	r.Handle("GetAllGames", api.NoInput(func(ctx context.Context) ([]types.GameInfo, error) {
		if svc.localView != nil {
			return svc.localView.LocalGames(ctx)
		}
		return svc.getAllGames(), nil
	}))
	r.Handle("RemoveGame", api.Typed(func(ctx context.Context, in executablePayload) (any, error) {
		if err := svc.removeGame(in.ExecutablePath); err != nil {
			return nil, err
		}
		if svc.localView != nil {
			if err := svc.localView.RemoveLocalConfig(ctx, in.ExecutablePath); err != nil {
				logger.Warn("appsvc", "could not remove the local library row: %v", err)
			}
		}
		return nil, nil
	}))
	r.Handle("SearchExecutables", api.Typed(func(_ context.Context, in searchPayload) ([]string, error) {
		return searchExecutables(in.FolderPath, depthOrUnlimited(in.MaxDepth), in.ExcludeNames), nil
	}))
	r.Handle("GetAutoScannedGames", api.Typed(func(ctx context.Context, in forcePayload) ([]types.ScannedFolderGroup, error) {
		groups, err := svc.getAutoScannedGames(in.Force)
		if err != nil {
			return nil, err
		}
		if svc.scanSink != nil {
			if err := svc.scanSink(ctx, groups); err != nil {
				logger.Warn("appsvc", "could not save scanned games in the library: %v", err)
			}
		}
		if svc.localView != nil {
			stored, err := svc.localView.LocalGroups(ctx)
			if err == nil && len(stored) > 0 {
				return stored, nil
			}
		}
		return groups, nil
	}))
	r.Handle("AddScanFolder", api.Typed(func(_ context.Context, in folderPayload) (any, error) {
		return nil, svc.addScanFolder(in.FolderPath)
	}))
	r.Handle("RemoveScanFolder", api.Typed(func(_ context.Context, in folderPayload) (any, error) {
		return nil, svc.removeScanFolder(in.FolderPath)
	}))
	r.Handle("UpdateScanFolderConfig", api.Typed(func(_ context.Context, in folderConfigPayload) (any, error) {
		return nil, svc.updateScanFolderConfig(in.FolderPath, in.Depth, in.ExcludeNames)
	}))
	r.Handle("RenameScanFolder", api.Typed(func(_ context.Context, in renamePayload) (string, error) {
		return svc.renameScanFolder(in.OldPath, in.NewPath, in.Depth, in.ExcludeNames)
	}))
	r.Handle("GetScanFolderConfig", api.Typed(func(_ context.Context, in folderPayload) (types.ScanFolderConfig, error) {
		return svc.getScanFolderConfig(in.FolderPath), nil
	}))
	r.Handle("SetPreferredLauncher", api.Typed(func(ctx context.Context, in preferredLauncherPayload) (types.GameInfo, error) {
		if svc.localView != nil {
			if err := svc.localView.SetLocalLauncher(ctx, in.Folder, in.Path); err != nil {
				return types.GameInfo{}, err
			}
		}
		return svc.setPreferredLauncher(in.Folder, in.Path)
	}))
	r.Handle("ClearPreferredLauncher", api.Typed(func(_ context.Context, in launcherFolderPayload) (types.GameInfo, error) {
		return svc.clearPreferredLauncher(in.Folder)
	}))
	r.Handle("BlacklistGame", api.Typed(func(ctx context.Context, in executablePayload) (any, error) {
		if svc.localView != nil {
			_ = svc.localView.SetLocalIgnored(ctx, in.ExecutablePath, true)
		}
		return nil, svc.blacklistGame(in.ExecutablePath)
	}))
	r.Handle("UnblacklistGame", api.Typed(func(ctx context.Context, in executablePayload) (any, error) {
		if svc.localView != nil {
			_ = svc.localView.SetLocalIgnored(ctx, in.ExecutablePath, false)
		}
		return nil, svc.unblacklistGame(in.ExecutablePath)
	}))
}

func registerActivityMethods(r *api.Router, svc *service) {
	r.Handle("GetGameActivity", api.NoInput(func(context.Context) ([]types.GameActivity, error) {
		return svc.activity.List(), nil
	}))
}

// depthOrUnlimited maps an omitted depth to an unlimited scan.
func depthOrUnlimited(depth *int) int {
	if depth == nil {
		return unlimitedDepth
	}
	return *depth
}
