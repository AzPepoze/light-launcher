package runtimesvc

import "light-launcher/core/internal/api"

func registerLsfg(r *api.Router, enabled bool, ev api.Events) {
	r.Handle("GetLsfgProfileForGame", gated(enabled, "lsfg", typed(func(in lsfgProfileInput) (any, error) {
		return getLsfgProfileForGame(in.Name, in.GamePath), nil
	})))
	r.Handle("DetectLosslessDll", gated(enabled, "lsfg", plain(func() (any, error) {
		return detectLosslessDll(), nil
	})))
	r.Handle("SaveLsfgProfile", gated(enabled, "lsfg", typed(func(in SaveLsfgProfileInput) (any, error) {
		return nil, saveLsfgProfile(in)
	})))
	r.Handle("DisableLsfgProfile", gated(enabled, "lsfg", typed(func(in disableLsfgInput) (any, error) {
		return nil, disableLsfgProfile(in.ProfileName, in.GamePath)
	})))
	r.Handle("RemoveProfile", gated(enabled, "lsfg", typed(func(in removeProfileInput) (any, error) {
		return nil, removeProfile(in.MainExecutablePath)
	})))
	r.Handle("InstallLsfg", gated(enabled, "lsfg", plain(func() (any, error) {
		return nil, installLsfg(progressEmitter(ev, "lsfg-install-progress"))
	})))
	r.Handle("UninstallLsfg", gated(enabled, "lsfg", plain(func() (any, error) {
		return nil, uninstallLsfg()
	})))
}
