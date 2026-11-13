package appsvc

import (
	"context"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/types"
)

// ScanSink receives the result of every folder scan. The launcher wires it to the library,
// so scanned games become library rows without appsvc knowing about the library.
type ScanSink func(context.Context, []types.ScannedFolderGroup) error

// Register adds the app, settings, game and activity methods to the router.
// ev receives "window-close" and "restart-app" and must not be nil.
// RunGame, prefix, Proton, LSFG, session and system methods are registered elsewhere.
func Register(r *api.Router, ev api.Events) {
	RegisterWithScanSink(r, ev, nil)
}

// LocalView is the library view of local games. appsvc reads through it after the migration,
// so Home and the old config methods show the same rows as the Library page.
type LocalView interface {
	LocalGames(context.Context) ([]types.GameInfo, error)
	LocalGroups(context.Context) ([]types.ScannedFolderGroup, error)
	LocalConfig(context.Context, string) *types.LaunchOptions
	SaveLocalConfig(context.Context, types.LaunchOptions) error
	RemoveLocalConfig(context.Context, string) error
	SetLocalLauncher(context.Context, string, string) error
	SetLocalIgnored(context.Context, string, bool) error
}

// RegisterWithScanSink is Register with a sink that saves scanned games into the library.
func RegisterWithScanSink(r *api.Router, ev api.Events, sink ScanSink) {
	RegisterWithLocalView(r, ev, sink, nil)
}

// RegisterWithLocalView also makes the old GetAllGames, GetAutoScannedGames and GetConfig
// methods read the canonical library rows.
func RegisterWithLocalView(r *api.Router, ev api.Events, sink ScanSink, view LocalView) {
	svc := newDefaultService()
	svc.events = ev
	svc.scanSink = sink
	svc.localView = view
	registerAppMethods(r, ev)
	registerPickerMethods(r)
	registerSettingsMethods(r, svc)
	registerGameMethods(r, svc)
	registerActivityMethods(r, svc)
}

func registerAppMethods(r *api.Router, ev api.Events) {
	r.Handle("GetInitialLauncherPath", api.NoInput(func(context.Context) (string, error) {
		return initialLauncherPath(), nil
	}))
	r.Handle("GetInitialGamePath", api.NoInput(func(context.Context) (string, error) {
		return initialGamePath(), nil
	}))
	r.Handle("GetShouldEditLsfg", api.NoInput(func(context.Context) (bool, error) {
		return shouldEditLsfg(), nil
	}))
	r.Handle("IsDir", api.Typed(func(_ context.Context, in pathPayload) (bool, error) {
		return isDirectory(in.Path), nil
	}))
	r.Handle("GetExeIcon", api.Typed(func(ctx context.Context, in executablePayload) (string, error) {
		return extractExeIcon(ctx, in.ExecutablePath)
	}))
	r.Handle("GetTotalRam", api.NoInput(func(context.Context) (int, error) {
		return totalRAMGiB()
	}))
	r.Handle("GetImageBase64", api.Typed(func(_ context.Context, in imagePayload) (string, error) {
		return imageDataURL(in.ImagePath), nil
	}))
	r.Handle("OpenURL", api.Typed(func(_ context.Context, in urlPayload) (any, error) {
		return nil, openURL(in.URL)
	}))
	r.Handle("OpenFileLocation", api.Typed(func(_ context.Context, in pathPayload) (any, error) {
		return nil, openFileLocation(in.Path)
	}))
	r.Handle("CloseWindow", api.NoInput(func(context.Context) (any, error) {
		ev.Emit("window-close", nil)
		return nil, nil
	}))
	r.Handle("RestartApp", api.NoInput(func(context.Context) (any, error) {
		ev.Emit("restart-app", nil)
		return nil, nil
	}))
}

func registerPickerMethods(r *api.Router) {
	r.Handle("PickFile", api.NoInput(pickFile))
	r.Handle("PickFolder", api.NoInput(pickFolder))
	r.Handle("PickFileCustom", api.Typed(func(ctx context.Context, in customFilePayload) (string, error) {
		return pickFileCustom(ctx, in.Title, in.Filters)
	}))
}
