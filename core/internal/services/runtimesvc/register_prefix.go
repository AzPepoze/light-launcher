package runtimesvc

import "light-launcher/core/internal/api"

func registerPrefix(r *api.Router, enabled bool, svc *service) {
	r.Handle("ListPrefixes", listOrEmpty(enabled, plain(func() (any, error) {
		return svc.listPrefixes()
	})))
	r.Handle("CreatePrefix", gated(enabled, "prefix", typed(func(in nameInput) (any, error) {
		return nil, svc.createPrefix(in.Name)
	})))
	r.Handle("GetPrefixBaseDir", gated(enabled, "prefix", plain(func() (any, error) {
		return svc.prefixes.BaseDir(), nil
	})))
	r.Handle("GetPrefixCreatedAt", gated(enabled, "prefix", typed(func(in prefixNameInput) (any, error) {
		return svc.getPrefixStats(in.PrefixName).CreatedAt, nil
	})))
	r.Handle("GetPrefixStats", gated(enabled, "prefix", typed(func(in prefixNameInput) (any, error) {
		return svc.getPrefixStats(in.PrefixName), nil
	})))
	r.Handle("RemovePrefix", gated(enabled, "prefix", typed(func(in nameInput) (any, error) {
		return nil, svc.removePrefix(in.Name)
	})))
	r.Handle("SavePrefixConfig", gated(enabled, "prefix", typed(func(in savePrefixInput) (any, error) {
		return nil, svc.savePrefixConfig(in.PrefixName, in.Options)
	})))
	r.Handle("LoadPrefixConfig", gated(enabled, "prefix", typed(func(in prefixNameInput) (any, error) {
		return svc.loadPrefixConfig(in.PrefixName), nil
	})))
	r.Handle("LoadPrefixConfigWithProton", gated(enabled, "prefix", typed(func(in prefixNameInput) (any, error) {
		return svc.loadPrefixConfigWithProton(in.PrefixName), nil
	})))
	r.Handle("RunPrefixTool", gated(enabled, "prefix", typed(func(in runPrefixToolInput) (any, error) {
		return nil, runPrefixTool(in.PrefixPath, in.ToolName, in.ProtonPath)
	})))
}
