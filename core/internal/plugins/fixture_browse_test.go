package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"testing"

	"light-launcher/core/internal/contract"
)

// fixtureSidebar is the sidebar the fixture manifest declares: a search box, a tag list and a title sort.
func fixtureSidebar() []contract.SidebarGroup {
	return []contract.SidebarGroup{
		{Type: "search", ID: "text", Placeholder: "Search titles"},
		{Type: "multi", ID: "tags", Label: "Tags", Dynamic: true, Searchable: true},
		{
			Type:      "sort",
			ID:        "sort",
			Orderable: true,
			Options:   []contract.SidebarOption{{Value: "title", Label: "Title"}},
			Default:   &contract.SortDefault{Value: "title", Order: "asc"},
		},
	}
}

type browseAnswer struct {
	Games []struct {
		SourceID string `json:"sourceId"`
		Title    string `json:"title"`
	} `json:"games"`
	NextCursor string `json:"nextCursor"`
	Total      int    `json:"total"`
}

func (a browseAnswer) titles() []string {
	titles := make([]string, 0, len(a.Games))
	for _, game := range a.Games {
		titles = append(titles, game.Title)
	}
	return titles
}

type filtersAnswer struct {
	Options []struct {
		Value string `json:"value"`
		Label string `json:"label"`
		Count int    `json:"count"`
	} `json:"options"`
}

// browseRequest is a source.browse request for the fixture source with the given fields changed.
func browseRequest(overrides map[string]any) map[string]any {
	request := map[string]any{"source": "fixture", "limit": 100}
	maps.Copy(request, overrides)
	return request
}

func callBrowse(t *testing.T, sup *Supervisor, request map[string]any) (browseAnswer, error) {
	t.Helper()
	raw, err := callFixture(sup, "source.browse", request)
	if err != nil {
		return browseAnswer{}, err
	}
	var answer browseAnswer
	if err := json.Unmarshal(raw, &answer); err != nil {
		t.Fatalf("decode browse answer %s: %v", raw, err)
	}
	return answer, nil
}

func callListFilters(t *testing.T, sup *Supervisor, request map[string]any) (filtersAnswer, error) {
	t.Helper()
	raw, err := callFixture(sup, "source.listFilters", request)
	if err != nil {
		return filtersAnswer{}, err
	}
	var answer filtersAnswer
	if err := json.Unmarshal(raw, &answer); err != nil {
		t.Fatalf("decode listFilters answer %s: %v", raw, err)
	}
	return answer, nil
}

// requireRPCCode fails unless err is a plugin error with the code.
func requireRPCCode(t *testing.T, err error, code int) *RPCError {
	t.Helper()
	var rpcErr *RPCError
	if !errors.As(err, &rpcErr) || rpcErr.Code != code {
		t.Fatalf("err = %v, want code %d", err, code)
	}
	return rpcErr
}

func TestFixtureManifestDeclaresValidSidebar(t *testing.T) {
	manifest := fixtureManifest(Permissions{})
	if err := manifest.Validate(); err != nil {
		t.Fatalf("fixture manifest: %v", err)
	}
	infos := sourceInfos("fixture", manifest)
	if len(infos) != 1 || len(infos[0].Sidebar) != 3 {
		t.Fatalf("sourceInfos = %+v, want one source with 3 sidebar groups", infos)
	}
}

func TestBrowsePagesThroughTheCatalog(t *testing.T) {
	sup, _, _ := newFixtureSupervisor(t, Permissions{})
	var pages []int
	var titles []string
	cursor := ""
	for {
		answer, err := callBrowse(t, sup, browseRequest(map[string]any{"sort": "title", "order": "asc", "limit": 3, "cursor": cursor}))
		if err != nil {
			t.Fatalf("browse after cursor %q: %v", cursor, err)
		}
		if answer.Total != 8 {
			t.Fatalf("total = %d, want 8", answer.Total)
		}
		pages = append(pages, len(answer.Games))
		titles = append(titles, answer.titles()...)
		if answer.NextCursor == "" {
			break
		}
		if len(pages) > 10 {
			t.Fatalf("browse did not end: pages %v", pages)
		}
		cursor = answer.NextCursor
	}
	if want := []int{3, 3, 2}; fmt.Sprint(pages) != fmt.Sprint(want) {
		t.Fatalf("page sizes = %v, want %v", pages, want)
	}
	want := []string{"Alpha Quest", "Blue Harbor", "Cinder Keep", "Delta Puzzle", "Emerald Road", "Foggy Lantern", "Glass Orchard", "Harbor Lights"}
	if !slices.Equal(titles, want) {
		t.Fatalf("titles = %v, want %v", titles, want)
	}
}

func TestBrowseFiltersQueryTagsAndOrder(t *testing.T) {
	cases := []struct {
		name      string
		overrides map[string]any
		want      []string
	}{
		{"query ignores case", map[string]any{"query": "HARBOR", "sort": "title"}, []string{"Blue Harbor", "Harbor Lights"}},
		{"one tag", map[string]any{"filters": map[string]any{"tags": []string{"rpg"}}, "sort": "title", "order": "asc"}, []string{"Alpha Quest", "Cinder Keep", "Emerald Road"}},
		{"tag as a string, sorted descending", map[string]any{"filters": map[string]any{"tags": "puzzle"}, "sort": "title", "order": "desc"}, []string{"Harbor Lights", "Foggy Lantern", "Delta Puzzle"}},
		{"unknown group is ignored", map[string]any{"filters": map[string]any{"installed": true}, "sort": "title"}, []string{"Alpha Quest", "Blue Harbor", "Cinder Keep", "Delta Puzzle", "Emerald Road", "Foggy Lantern", "Glass Orchard", "Harbor Lights"}},
		{"no match", map[string]any{"query": "nothing here"}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sup, _, _ := newFixtureSupervisor(t, Permissions{})
			answer, err := callBrowse(t, sup, browseRequest(tc.overrides))
			if err != nil {
				t.Fatalf("source.browse: %v", err)
			}
			if !slices.Equal(answer.titles(), tc.want) || answer.Total != len(tc.want) {
				t.Fatalf("titles = %v (total %d), want %v", answer.titles(), answer.Total, tc.want)
			}
			if answer.NextCursor != "" {
				t.Fatalf("nextCursor = %q, want none for a short list", answer.NextCursor)
			}
		})
	}
}

func TestBrowseRejectsAForeignCursor(t *testing.T) {
	sup, _, _ := newFixtureSupervisor(t, Permissions{})
	first, err := callBrowse(t, sup, browseRequest(map[string]any{"sort": "title", "limit": 3}))
	if err != nil || first.NextCursor == "" {
		t.Fatalf("first page: cursor %q, err %v", first.NextCursor, err)
	}

	_, err = callBrowse(t, sup, browseRequest(map[string]any{"query": "harbor", "sort": "title", "limit": 3, "cursor": first.NextCursor}))
	rpcErr := requireRPCCode(t, err, CodeInvalidParams)
	if data, ok := rpcErr.Data.(map[string]any); !ok || data["code"] != "invalid_cursor" {
		t.Fatalf("error data = %v, want code invalid_cursor", rpcErr.Data)
	}

	if _, err := callBrowse(t, sup, browseRequest(map[string]any{"sort": "title", "limit": 3, "cursor": first.NextCursor})); err != nil {
		t.Fatalf("the same request with its own cursor: %v", err)
	}
	_, err = callBrowse(t, sup, browseRequest(map[string]any{"sort": "title", "limit": 3, "cursor": "not-a-cursor"}))
	requireRPCCode(t, err, CodeInvalidParams)
}

func TestBrowseRejectsBadParams(t *testing.T) {
	cases := []struct {
		name      string
		overrides map[string]any
	}{
		{"limit zero", map[string]any{"limit": 0}},
		{"limit above 100", map[string]any{"limit": 101}},
		{"order not asc or desc", map[string]any{"order": "up"}},
		{"unknown sort", map[string]any{"sort": "date"}},
		{"unknown source", map[string]any{"source": "other"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sup, _, _ := newFixtureSupervisor(t, Permissions{})
			_, err := callBrowse(t, sup, browseRequest(tc.overrides))
			requireRPCCode(t, err, CodeInvalidParams)
		})
	}
}

func TestListFiltersCountsTheTags(t *testing.T) {
	cases := []struct {
		name      string
		overrides map[string]any
		want      []string
		wantErr   bool
	}{
		{"every tag with its count", nil, []string{"fantasy:2", "horror:2", "puzzle:3", "rpg:3", "visual-novel:3"}, false},
		{"query matches part of a tag", map[string]any{"query": "R"}, []string{"horror:2", "rpg:3"}, false},
		{"limit cuts the list", map[string]any{"limit": 1}, []string{"fantasy:2"}, false},
		{"unknown group", map[string]any{"group": "genre"}, nil, true},
		{"limit zero", map[string]any{"limit": 0}, nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sup, _, _ := newFixtureSupervisor(t, Permissions{})
			request := map[string]any{"source": "fixture", "group": "tags", "limit": 10}
			maps.Copy(request, tc.overrides)
			answer, err := callListFilters(t, sup, request)
			if tc.wantErr {
				requireRPCCode(t, err, CodeInvalidParams)
				return
			}
			if err != nil {
				t.Fatalf("source.listFilters: %v", err)
			}
			var got []string
			for _, option := range answer.Options {
				got = append(got, fmt.Sprintf("%s:%d", option.Value, option.Count))
			}
			if !slices.Equal(got, tc.want) {
				t.Fatalf("options = %v, want %v", got, tc.want)
			}
		})
	}
}
