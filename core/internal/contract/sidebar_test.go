package contract

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestSourceInfoCarriesTheSidebar pins the wire shape the renderer maps into its Library sidebar.
func TestSourceInfoCarriesTheSidebar(t *testing.T) {
	info := SourceInfo{
		Plugin: "shop", SourceID: "shop", Label: "Shop",
		Sidebar: []SidebarGroup{{
			Type: "sort", ID: "sort", Label: "Sort", Orderable: true,
			Options: []SidebarOption{{Value: "title", Label: "Title"}},
			Default: &SortDefault{Value: "title", Order: "asc"},
		}},
	}
	raw, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var wire struct {
		Sidebar []SidebarGroup `json:"sidebar"`
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(wire.Sidebar) != 1 {
		t.Fatalf("sidebar = %s, want one group", raw)
	}
	group := wire.Sidebar[0]
	if group.Type != "sort" || group.ID != "sort" || len(group.Options) != 1 || group.Options[0].Value != "title" {
		t.Errorf("group = %+v, want the sort group back", group)
	}
	if group.Default == nil || group.Default.Order != "asc" {
		t.Errorf("default = %+v, want asc", group.Default)
	}
}

func TestSourceInfoLeavesAnEmptySidebarOut(t *testing.T) {
	raw, err := json.Marshal(SourceInfo{Plugin: "p", SourceID: "p"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(raw), `"sidebar"`) {
		t.Errorf("json = %s, want no sidebar field", raw)
	}
}
