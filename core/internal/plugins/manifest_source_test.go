package plugins

import (
	"strings"
	"testing"
)

// sourceJSON returns a valid manifest whose only source is the given JSON object.
func sourceJSON(source string) []byte {
	return []byte(`{
		"id": "demo", "name": "Demo", "version": "1.0.0", "apiVersion": "1", "minHost": "2.0.0",
		"entry": { "linux-x64": "bin/demo" },
		"contributes": { "sources": [` + source + `] },
		"permissions": {}
	}`)
}

func TestOldTileManifestStillLoads(t *testing.T) {
	cases := []struct {
		tile      string
		wantCover string
	}{
		{`"poster"`, "3:4"},
		{`"wide"`, "16:9"},
		{`"square"`, "1:1"},
		{``, "3:4"},
	}
	for _, c := range cases {
		source := `{ "id": "demo", "label": "Demo", "capabilities": ["browse"] }`
		if c.tile != "" {
			source = strings.Replace(source, `"capabilities"`, `"tile": `+c.tile+`, "capabilities"`, 1)
		}
		manifest, err := ParseManifest(sourceJSON(source))
		if err != nil {
			t.Fatalf("tile %s: ParseManifest: %v", c.tile, err)
		}
		info := sourceInfos("demo", manifest)[0]
		if info.Cover != c.wantCover {
			t.Errorf("tile %s: cover = %q, want %q", c.tile, info.Cover, c.wantCover)
		}
	}
}

func TestCoverWinsOverTile(t *testing.T) {
	source := `{ "id": "demo", "label": "Demo", "tile": "wide", "cover": { "ratio": "4:3" }, "capabilities": ["browse"] }`
	manifest, err := ParseManifest(sourceJSON(source))
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	info := sourceInfos("demo", manifest)[0]
	if info.Cover != "4:3" {
		t.Fatalf("cover = %q, want 4:3", info.Cover)
	}
	if info.Tile != "wide" {
		t.Fatalf("tile = %q, want wide kept for older consumers", info.Tile)
	}
}

func TestActionsGetLibraryScopeByDefault(t *testing.T) {
	source := `{ "id": "demo", "label": "Demo", "capabilities": ["browse"],
		"actions": [
			{ "id": "sync", "label": "Sync", "icon": "sync", "method": "demo.sync" },
			{ "id": "open", "label": "Open", "method": "demo.open", "scope": "game" }
		],
		"card": { "title": "title", "subtitle": "developer", "badges": ["favorite", "installed"] } }`
	manifest, err := ParseManifest(sourceJSON(source))
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	info := sourceInfos("demo", manifest)[0]
	if len(info.Actions) != 2 {
		t.Fatalf("actions = %+v, want 2", info.Actions)
	}
	if info.Actions[0].Scope != "library" || info.Actions[1].Scope != "game" {
		t.Fatalf("scopes = %q, %q, want library and game", info.Actions[0].Scope, info.Actions[1].Scope)
	}
	if info.Card == nil || info.Card.Subtitle != "developer" || len(info.Card.Badges) != 2 {
		t.Fatalf("card = %+v, want the declared card", info.Card)
	}
}

func TestRejectsBadCoverRatios(t *testing.T) {
	ratios := []string{"", "0:3", "3:0", "33:1", "1:33", "1:5", "5:1", "4/3", "3:04", "4:3:2", "a:b"}
	for _, ratio := range ratios {
		source := `{ "id": "demo", "label": "Demo", "cover": { "ratio": "` + ratio + `" }, "capabilities": ["browse"] }`
		if _, err := ParseManifest(sourceJSON(source)); err == nil {
			t.Errorf("ratio %q: ParseManifest accepted it", ratio)
		}
	}
}

func TestAcceptsCoverRatioEdges(t *testing.T) {
	for _, ratio := range []string{"1:4", "4:1", "32:8", "1:1", "2:3"} {
		source := `{ "id": "demo", "label": "Demo", "cover": { "ratio": "` + ratio + `" }, "capabilities": ["browse"] }`
		if _, err := ParseManifest(sourceJSON(source)); err != nil {
			t.Errorf("ratio %q: ParseManifest: %v", ratio, err)
		}
	}
}

func TestRejectsUnknownCardFields(t *testing.T) {
	cases := map[string]string{
		"unknown key":      `"card": { "color": "red" }`,
		"unknown title":    `"card": { "title": "nickname" }`,
		"unknown subtitle": `"card": { "subtitle": "rating" }`,
		"unknown badge":    `"card": { "badges": ["starred"] }`,
		"duplicate badge":  `"card": { "badges": ["installed", "installed"] }`,
	}
	for name, card := range cases {
		source := `{ "id": "demo", "label": "Demo", "capabilities": ["browse"], ` + card + ` }`
		if _, err := ParseManifest(sourceJSON(source)); err == nil {
			t.Errorf("%s: ParseManifest accepted %s", name, card)
		}
	}
}

func TestAcceptsCardWithGameFields(t *testing.T) {
	source := `{ "id": "demo", "label": "Demo", "capabilities": ["browse"],
		"card": { "title": "title", "subtitle": "version", "badges": ["favorite", "updateAvailable", "installed"] } }`
	if _, err := ParseManifest(sourceJSON(source)); err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
}

func TestRejectsDuplicateActionIDs(t *testing.T) {
	source := `{ "id": "demo", "label": "Demo", "capabilities": ["browse"], "actions": [
		{ "id": "sync", "label": "Sync", "method": "demo.sync" },
		{ "id": "sync", "label": "Sync again", "method": "demo.other" }
	] }`
	_, err := ParseManifest(sourceJSON(source))
	if err == nil || !strings.Contains(err.Error(), "duplicate action id") {
		t.Fatalf("ParseManifest error = %v, want duplicate action id", err)
	}
}

func TestRejectsBadActions(t *testing.T) {
	action := func(fields string) string {
		return `{ "id": "sync", "label": "Sync", "method": "demo.sync"` + fields + ` }`
	}
	cases := map[string]string{
		"bad method": `{ "id": "sync", "label": "Sync", "method": "sync" }`,
		"bad scope":  action(`, "scope": "tray"`),
		"bad icon":   action(`, "icon": "Sync Now"`),
		"no label":   `{ "id": "sync", "label": "", "method": "demo.sync" }`,
		"bad id":     `{ "id": "Sync", "label": "Sync", "method": "demo.sync" }`,
		"too many":   strings.Repeat(action("")+",", 8) + action(""),
	}
	for name, actions := range cases {
		source := `{ "id": "demo", "label": "Demo", "capabilities": ["browse"], "actions": [` + actions + `] }`
		if _, err := ParseManifest(sourceJSON(source)); err == nil {
			t.Errorf("%s: ParseManifest accepted the actions", name)
		}
	}
}

func TestAcceptsEightActions(t *testing.T) {
	var actions []string
	for _, id := range []string{"a", "b", "c", "d", "e", "f", "g", "h"} {
		actions = append(actions, `{ "id": "`+id+`", "label": "Run `+id+`", "method": "demo.`+id+`" }`)
	}
	source := `{ "id": "demo", "label": "Demo", "capabilities": ["browse"], "actions": [` + strings.Join(actions, ",") + `] }`
	if _, err := ParseManifest(sourceJSON(source)); err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
}
