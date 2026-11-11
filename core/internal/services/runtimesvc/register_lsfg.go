package runtimesvc

import "light-launcher/core/internal/api"

func registerLsfg(r *api.Router, enabled bool, ev api.Events, svc *service) {
	r.Handle("GetLsfgProfileForGame", gated(enabled, "lsfg", typed(func(in lsfgProfileInput) (any, error) {
		return svc.getLsfgProfileForGame(in.Name, in.GamePath), nil
	})))
	r.Handle("DetectLosslessDll", gated(enabled, "lsfg", plain(func() (any, error) {
		return detectLosslessDll(), nil
	})))
	r.Handle("SaveLsfgProfile", gated(enabled, "lsfg", typed(func(in SaveLsfgProfileInput) (any, error) {
		return nil, svc.saveLsfgProfile(in)
	})))
	r.Handle("DisableLsfgProfile", gated(enabled, "lsfg", typed(func(in disableLsfgInput) (any, error) {
		return nil, svc.disableLsfgProfile(in.ProfileName, in.GamePath)
	})))
	r.Handle("RemoveProfile", gated(enabled, "lsfg", typed(func(in removeProfileInput) (any, error) {
		return nil, svc.removeProfile(in.MainExecutablePath)
	})))
	r.Handle("InstallLsfg", gated(enabled, "lsfg", plain(func() (any, error) {
		return nil, installLsfg(progressEmitter(ev, "lsfg-install-progress"))
	})))
	r.Handle("UninstallLsfg", gated(enabled, "lsfg", plain(func() (any, error) {
		return nil, uninstallLsfg()
	})))
}
