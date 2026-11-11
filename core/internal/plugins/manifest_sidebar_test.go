package plugins

import (
	"strings"
	"testing"
)

// sourceWithSidebar builds a manifest whose only source is the given JSON object.
func sourceWithSidebar(source string) string {
	return strings.Replace(validManifest, `{ "id": "demo", "label": "Demo", "capabilities": ["browse", "updateCheck"] }`, source, 1)
}

// filtersSource returns a source with the browse:filters capability and the given sidebar groups.
func filtersSource(groups string) string {
	return `{ "id": "demo", "label": "Demo", "capabilities": ["browse", "browse:filters"], "sidebar": [` + groups + `] }`
}

const (
	searchGroup  = `{ "type": "search", "id": "text", "placeholder": "Search" }`
	toggleGroup  = `{ "type": "toggles", "id": "show", "options": [{ "value": "installed", "label": "Installed" }] }`
	selectGroup  = `{ "type": "select", "id": "status", "label": "Status", "options": [{ "value": "new", "label": "New" }] }`
	dynamicGroup = `{ "type": "multi", "id": "tags", "label": "Tags", "dynamic": true, "searchable": true }`
	sortGroup    = `{ "type": "sort", "id": "sort", "orderable": true, "options": [{ "value": "title", "label": "Title" }, { "value": "date", "label": "Date" }], "default": { "value": "date", "order": "desc" } }`
)

func TestSidebarAcceptsEveryGroupType(t *testing.T) {
	groups := strings.Join([]string{searchGroup, toggleGroup, selectGroup, dynamicGroup, sortGroup}, ",")
	source := filtersSource(groups)
	manifest, err := ParseManifest([]byte(sourceWithSidebar(source)))
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	infos := sourceInfos("demo-site", manifest)
	if len(infos) != 1 || len(infos[0].Sidebar) != 5 {
		t.Fatalf("sourceInfos = %+v, want one source with 5 sidebar groups", infos)
	}
	sort := infos[0].Sidebar[4]
	if sort.Default == nil || sort.Default.Value != "date" || sort.Default.Order != "desc" {
		t.Fatalf("sort default = %+v, want date desc", sort.Default)
	}
	if !infos[0].Sidebar[3].Dynamic {
		t.Fatalf("multi group should be dynamic")
	}
}

func TestSidebarRejectsBadGroups(t *testing.T) {
	cases := []struct {
		name    string
		source  string
		wantErr string
	}{
		{"no browse:filters", `{ "id": "demo", "label": "Demo", "capabilities": ["browse"], "sidebar": [` + searchGroup + `] }`, "needs the browse:filters"},
		{"unknown group type", filtersSource(`{ "type": "slider", "id": "size" }`), "unknown group type"},
		{"duplicate group id", filtersSource(searchGroup + "," + searchGroup), "duplicate group id"},
		{"bad group id", filtersSource(`{ "type": "search", "id": "Text" }`), "sidebar[0].id"},
		{"toggles without options", filtersSource(`{ "type": "toggles", "id": "show" }`), "at least one option"},
		{"select without label", filtersSource(`{ "type": "select", "id": "status", "options": [{ "value": "new", "label": "New" }] }`), "label"},
		{"select without options or dynamic", filtersSource(`{ "type": "select", "id": "status", "label": "Status" }`), "at least one option"},
		{"dynamic with options", filtersSource(`{ "type": "multi", "id": "tags", "label": "Tags", "dynamic": true, "options": [{ "value": "a", "label": "A" }] }`), "dynamic group"},
		{"duplicate option value", filtersSource(`{ "type": "toggles", "id": "show", "options": [{ "value": "a", "label": "A" }, { "value": "a", "label": "B" }] }`), "duplicate option value"},
		{"option without label", filtersSource(`{ "type": "toggles", "id": "show", "options": [{ "value": "a" }] }`), "options[0].label"},
		{"default on a toggle", filtersSource(`{ "type": "toggles", "id": "show", "options": [{ "value": "a", "label": "A" }], "default": { "value": "a", "order": "asc" } }`), "only a sort group"},
		{"sort default not an option", filtersSource(`{ "type": "sort", "id": "sort", "options": [{ "value": "title", "label": "Title" }], "default": { "value": "date", "order": "asc" } }`), "not one of the sort options"},
		{"sort default bad order", filtersSource(`{ "type": "sort", "id": "sort", "options": [{ "value": "title", "label": "Title" }], "default": { "value": "title", "order": "up" } }`), "must be asc or desc"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseManifest([]byte(sourceWithSidebar(tc.source)))
			if err == nil {
				t.Fatalf("expected an error")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error %q does not mention %q", err, tc.wantErr)
			}
		})
	}
}

func TestSourceWithoutSidebarHasNoGroups(t *testing.T) {
	manifest, err := ParseManifest([]byte(validManifest))
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	if infos := sourceInfos("demo-site", manifest); len(infos[0].Sidebar) != 0 {
		t.Fatalf("sidebar = %+v, want none", infos[0].Sidebar)
	}
}
