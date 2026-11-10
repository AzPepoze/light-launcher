package runtimesvc

import "light-launcher/core/internal/api"

func registerPrefix(r *api.Router, enabled bool) {
	r.Handle("ListPrefixes", listOrEmpty(enabled, plain(func() (any, error) {
		return listPrefixes()
	})))
	r.Handle("CreatePrefix", gated(enabled, "prefix", typed(func(in nameInput) (any, error) {
		return nil, createPrefix(in.Name)
	})))
	r.Handle("GetPrefixBaseDir", gated(enabled, "prefix", plain(func() (any, error) {
		return prefixBaseFolder(), nil
	})))
	r.Handle("GetPrefixCreatedAt", gated(enabled, "prefix", typed(func(in prefixNameInput) (any, error) {
		return getPrefixStats(in.PrefixName).CreatedAt, nil
	})))
	r.Handle("GetPrefixStats", gated(enabled, "prefix", typed(func(in prefixNameInput) (any, error) {
		return getPrefixStats(in.PrefixName), nil
	})))
	r.Handle("RemovePrefix", gated(enabled, "prefix", typed(func(in nameInput) (any, error) {
		return nil, removePrefix(in.Name)
	})))
	r.Handle("SavePrefixConfig", gated(enabled, "prefix", typed(func(in savePrefixInput) (any, error) {
		return nil, savePrefixConfig(in.PrefixName, in.Options)
	})))
	r.Handle("LoadPrefixConfig", gated(enabled, "prefix", typed(func(in prefixNameInput) (any, error) {
		return loadPrefixConfig(in.PrefixName), nil
	})))
	r.Handle("LoadPrefixConfigWithProton", gated(enabled, "prefix", typed(func(in prefixNameInput) (any, error) {
		return loadPrefixConfigWithProton(in.PrefixName), nil
	})))
	r.Handle("RunPrefixTool", gated(enabled, "prefix", typed(func(in runPrefixToolInput) (any, error) {
		return nil, runPrefixTool(in.PrefixPath, in.ToolName, in.ProtonPath)
	})))
}
