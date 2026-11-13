package librarysvc

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/library"
)

// actionSources declares one source with one action and records the plugin calls.
type actionSources struct {
	fakeSources
	method string
	params contract.ActionRequest
}

func (a *actionSources) Sources() []contract.SourceInfo {
	return []contract.SourceInfo{{
		Plugin: "demo-site", SourceID: "demo",
		Actions: []contract.SourceAction{{ID: "sync", Label: "Sync", Method: "demo-site.sync"}},
	}}
}

func (a *actionSources) CallAction(_ context.Context, plugin, method string, params any, out any) error {
	a.method = plugin + ":" + method
	a.params = params.(contract.ActionRequest)
	raw, _ := json.Marshal(contract.ActionResult{Message: "Synced", Refresh: true})
	return json.Unmarshal(raw, out)
}

func TestRunActionCallsTheDeclaredMethod(t *testing.T) {
	sources := &actionSources{}
	s := &service{sources: sources}
	result, err := s.runAction(context.Background(), contract.ActionRequest{Source: "demo", Action: "sync"})
	if err != nil {
		t.Fatalf("runAction: %v", err)
	}
	if sources.method != "demo-site:demo-site.sync" {
		t.Errorf("called %q, want demo-site:demo-site.sync", sources.method)
	}
	if sources.params.Source != "demo" || sources.params.Action != "sync" || sources.params.GameID != nil {
		t.Errorf("params = %+v, want source demo, action sync, no game", sources.params)
	}
	if result.Message != "Synced" || !result.Refresh {
		t.Errorf("result = %+v, want the plugin answer", result)
	}
}

func TestRunActionRefusesUnknownSourceAndAction(t *testing.T) {
	s := &service{sources: &actionSources{}}
	if _, err := s.runAction(context.Background(), contract.ActionRequest{Source: "nope", Action: "sync"}); !errors.Is(err, library.ErrUnknownSource) {
		t.Errorf("unknown source error = %v, want ErrUnknownSource", err)
	}
	if _, err := s.runAction(context.Background(), contract.ActionRequest{Source: "demo", Action: "nope"}); !errors.Is(err, errUnknownAction) {
		t.Errorf("unknown action error = %v, want errUnknownAction", err)
	}
}
