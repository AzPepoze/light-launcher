package runtimesvc

import "light-launcher/core/internal/api"

func registerProton(r *api.Router, enabled bool, ev api.Events) {
	r.Handle("ScanProtonVersions", listOrEmpty(enabled, plain(func() (any, error) {
		return scanProtonVersions(), nil
	})))
	r.Handle("GetProtonVariants", listOrEmpty(enabled, plain(func() (any, error) {
		return getProtonVariants(), nil
	})))
	r.Handle("GetProtonReleases", listOrEmpty(enabled, typed(func(in variantInput) (any, error) {
		return getProtonReleases(in.VariantID)
	})))
	r.Handle("InstallProtonVersion", gated(enabled, "proton", typed(func(in installProtonInput) (any, error) {
		return nil, installProtonVersion(in.URL, in.Version, progressEmitter(ev, "install-proton-progress"))
	})))
}
