package library

import (
	"context"
	"testing"

	"light-launcher/core/internal/contract"
)

func TestSearchRanksExactTitleFirst(t *testing.T) {
	sources := &fakeSources{
		infos: []contract.SourceInfo{{Plugin: "f95plugin", SourceID: "f95", Capabilities: []string{CapabilityBrowse}}},
		replies: map[string]func(any) (any, error){
			methodSearch: func(any) (any, error) {
				return map[string]any{"games": []map[string]string{
					{"source": "f95", "sourceId": "1", "url": "https://f95/1", "title": "Unrelated Thing"},
					{"source": "f95", "sourceId": "2", "url": "https://f95/2", "title": "Harbor Lights"},
				}}, nil
			},
		},
	}

	results := SearchSources(context.Background(), sources, "Harbor Lights")
	if len(results) != 1 || len(results[0].Games) != 2 {
		t.Fatalf("results = %+v", results)
	}
	games := results[0].Games
	if games[0].Title != "Harbor Lights" {
		t.Errorf("best match = %q, want Harbor Lights first", games[0].Title)
	}
	if games[0].Score <= games[1].Score {
		t.Errorf("scores = %v, %v; the best match must score higher", games[0].Score, games[1].Score)
	}
}
