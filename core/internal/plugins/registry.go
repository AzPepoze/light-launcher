package plugins

import (
	"context"
	"errors"
	"strings"

	"light-launcher/core/internal/contract"
)

var errNotSourceMethod = errors.New("only source methods can be called on a plugin source")

// Registry implements contract.Sources over the enabled plugins.
type Registry struct {
	host *Host
}

var _ contract.Sources = (*Registry)(nil)

// NewRegistry returns the source view of a host.
func NewRegistry(host *Host) *Registry {
	return &Registry{host: host}
}

// Sources lists the sources of every enabled plugin that installed cleanly.
func (r *Registry) Sources() []contract.SourceInfo {
	var sources []contract.SourceInfo
	for _, info := range r.host.List() {
		if info.Enabled && info.Status != "broken" {
			sources = append(sources, info.Sources...)
		}
	}
	return sources
}

// Call runs a "source." method on a plugin and decodes the answer into out.
func (r *Registry) Call(ctx context.Context, plugin, method string, params any, out any) error {
	if !strings.HasPrefix(method, "source.") {
		return errNotSourceMethod
	}
	sup, err := r.host.supervisorFor(plugin)
	if err != nil {
		return err
	}
	return sup.Call(ctx, method, params, out)
}
