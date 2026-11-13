package plugins

import (
	"strings"
	"testing"

	"light-launcher/core/internal/contract"
)

func TestMarkSourceCollisionsKeepsTheFirstOwner(t *testing.T) {
	infos := []PluginInfo{
		{ID: "alpha", Enabled: true, Status: "running", Sources: []contract.SourceInfo{{SourceID: "shared"}}},
		{ID: "beta", Enabled: true, Status: "running", Sources: []contract.SourceInfo{{SourceID: "shared"}, {SourceID: "own"}}},
	}
	got := markSourceCollisions(infos)
	if got[0].Status == "broken" {
		t.Fatalf("first plugin was marked broken: %+v", got[0])
	}
	if got[1].Status != "broken" || !strings.Contains(got[1].Error, "already provided by plugin alpha") {
		t.Fatalf("second plugin = %+v, want broken with the owner in the error", got[1])
	}
}
