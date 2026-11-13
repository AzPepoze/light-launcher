package library

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"light-launcher/core/internal/contract"
)

// The shapes below are copied from the plugin SDK protocol doc. They are separate from the
// launcher's own types, so a field the SDK does not know fails the strict decode in strictSources.

type sdkRef struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
	URL  string `json:"url"`
}

type sdkGameRef struct {
	ID        *int64         `json:"id,omitempty"`
	Source    string         `json:"source"`
	SourceID  string         `json:"sourceId"`
	URL       string         `json:"url"`
	Title     string         `json:"title"`
	Version   *string        `json:"version,omitempty"`
	UpdatedAt *string        `json:"updatedAt,omitempty"`
	Extra     map[string]any `json:"extra,omitempty"`
}

type sdkParseRefParams struct {
	URL string `json:"url"`
}

type sdkFetchGameParams struct {
	Source string `json:"source"`
	Ref    sdkRef `json:"ref"`
}

type sdkSearchParams struct {
	Source string  `json:"source"`
	URL    *string `json:"url,omitempty"`
	Query  *string `json:"query,omitempty"`
	Page   *int    `json:"page,omitempty"`
}

type sdkGameParams struct {
	Source string     `json:"source"`
	Game   sdkGameRef `json:"game"`
}

// strictSources answers source.* calls with SDK replies. It rejects requests with unknown fields
// and requests without a source id.
type strictSources struct {
	infos []contract.SourceInfo
}

func (f *strictSources) Sources() []contract.SourceInfo {
	return f.infos
}

func (f *strictSources) StartDownload(context.Context, string, any, int64) (int64, error) {
	return 0, errors.New("StartDownload is not faked here")
}

func (f *strictSources) Browse(context.Context, string, contract.BrowseRequest) (contract.BrowsePage, error) {
	return contract.BrowsePage{}, errors.New("Browse is not faked here")
}

func (f *strictSources) CallAction(context.Context, string, string, any, any) error {
	return nil
}

func (f *strictSources) ListFilters(context.Context, string, contract.FiltersRequest) (contract.FiltersPage, error) {
	return contract.FiltersPage{}, errors.New("ListFilters is not faked here")
}

func (f *strictSources) Call(_ context.Context, _, method string, params, out any) error {
	raw, err := json.Marshal(params)
	if err != nil {
		return err
	}
	reply, err := strictReply(method, raw)
	if err != nil {
		return err
	}
	data, err := json.Marshal(reply)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

func strictReply(method string, raw []byte) (any, error) {
	switch method {
	case methodParseRef:
		var in sdkParseRefParams
		if err := decodeStrict(raw, &in); err != nil {
			return nil, err
		}
		if !strings.HasPrefix(in.URL, "https://f95/") {
			return nil, nil
		}
		return map[string]string{"kind": "game", "id": "42", "url": in.URL}, nil
	case methodFetchGame:
		var in sdkFetchGameParams
		if err := decodeStrict(raw, &in); err != nil {
			return nil, err
		}
		if in.Source == "" {
			return nil, errors.New("fetchGame: source is required")
		}
		return map[string]any{"game": map[string]string{
			"source": in.Source, "sourceId": "42", "url": "https://f95/42", "title": "Strict Game",
		}}, nil
	case methodSearch:
		var in sdkSearchParams
		if err := decodeStrict(raw, &in); err != nil {
			return nil, err
		}
		if in.Source == "" {
			return nil, errors.New("search: source is required")
		}
		query := ""
		if in.Query != nil {
			query = *in.Query
		}
		return map[string]any{"games": []map[string]string{{
			"source": in.Source, "sourceId": "7", "url": "https://f95/7", "title": "Found " + query,
		}}}, nil
	case methodResolveFiles:
		var in sdkGameParams
		if err := decodeStrict(raw, &in); err != nil {
			return nil, err
		}
		if in.Source == "" {
			return nil, errors.New("resolveFiles: source is required")
		}
		return map[string]any{"files": []map[string]string{{
			"id": "dl", "label": "Windows", "url": "https://f95/dl.zip", "kind": "direct", "platform": "windows",
		}}}, nil
	case methodCheckUpdate:
		var in sdkGameParams
		if err := decodeStrict(raw, &in); err != nil {
			return nil, err
		}
		if in.Source == "" {
			return nil, errors.New("checkUpdate: source is required")
		}
		return map[string]any{"update": map[string]string{"version": "2.0"}}, nil
	}
	return nil, errors.New("unexpected method " + method)
}

func decodeStrict(raw []byte, out any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	return decoder.Decode(out)
}

func strictF95Sources() *strictSources {
	return &strictSources{infos: []contract.SourceInfo{{
		Plugin: "f95plugin", SourceID: "f95", Label: "F95", Capabilities: []string{CapabilityBrowse, CapabilityUpdateCheck, CapabilityResolveFiles},
	}}}
}

func TestStrictPreviewSendsSDKShapes(t *testing.T) {
	lib := openTestLibrary(t, Options{Sources: strictF95Sources()})
	preview, err := lib.Preview(context.Background(), "https://f95/42")
	if err != nil {
		t.Fatal(err)
	}
	if preview.Game.Title != "Strict Game" {
		t.Errorf("title = %q", preview.Game.Title)
	}
}

func TestStrictAddFromURLResolvesFiles(t *testing.T) {
	lib := openTestLibrary(t, Options{Sources: strictF95Sources()})
	ctx := context.Background()
	game, err := lib.AddFromURL(ctx, "https://f95/42")
	if err != nil {
		t.Fatal(err)
	}
	files, err := lib.ListFiles(ctx, game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Label != "Windows" {
		t.Errorf("files = %+v, want the resolved Windows file", files)
	}
}

func TestStrictSearchCallsEverySourceOnce(t *testing.T) {
	sources := &strictSources{infos: []contract.SourceInfo{
		{Plugin: "multi", SourceID: "f95", Capabilities: []string{CapabilityBrowse}},
		{Plugin: "multi", SourceID: "lewd", Capabilities: []string{CapabilityBrowse}},
	}}
	results := SearchSources(context.Background(), sources, "harbor")
	if len(results) != 2 {
		t.Fatalf("results = %d, want one per source", len(results))
	}
	for _, result := range results {
		if result.Error != "" || len(result.Games) != 1 || result.Games[0].Title != "Found harbor" {
			t.Errorf("result for %s = %+v", result.Source, result)
		}
		if result.Games[0].Source != result.Source {
			t.Errorf("game source = %q, want %q", result.Games[0].Source, result.Source)
		}
	}
}

func TestStrictCheckUpdateSendsGameRef(t *testing.T) {
	lib := openTestLibrary(t, Options{Sources: strictF95Sources()})
	game := f95Game(t, lib, "1.0")
	result, err := lib.CheckGame(context.Background(), game.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !result.HasUpdate || result.To == nil || *result.To != "2.0" {
		t.Errorf("result = %+v, want an update to 2.0", result)
	}
}

func TestDecodeStrictRejectsUnknownFields(t *testing.T) {
	var in sdkParseRefParams
	if err := decodeStrict([]byte(`{"url":"x","extra":1}`), &in); err == nil {
		t.Error("unknown fields must fail the strict decode")
	}
}
