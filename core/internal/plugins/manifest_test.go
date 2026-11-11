package plugins

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const validManifest = `{
  "id": "demo-site",
  "name": "Demo Site",
  "version": "1.2.3",
  "apiVersion": "1",
  "minHost": "2.0.0",
  "entry": { "linux-x64": "bin/linux-x64/demo", "windows-x64": "bin/windows-x64/demo.exe" },
  "contributes": {
    "sources": [{ "id": "demo", "label": "Demo", "capabilities": ["browse", "updateCheck"] }],
    "settings": { "$ref": "settings.schema.json" },
    "hooks": ["onGameLaunch"],
    "commands": [{ "id": "import-key", "label": "Import key", "method": "demo.importKey" }]
  },
  "permissions": { "network": ["demo.example", "*.pixeldrain.com"], "fs": ["$GAMES"], "secrets": true }
}`

func TestParseManifestAcceptsValidFile(t *testing.T) {
	manifest, err := ParseManifest([]byte(validManifest))
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	if manifest.ID != "demo-site" || manifest.Version != "1.2.3" {
		t.Fatalf("unexpected identity: %+v", manifest)
	}
	if entry, ok := manifest.EntryFor("linux-x64"); !ok || entry != "bin/linux-x64/demo" {
		t.Fatalf("EntryFor(linux-x64) = %q, %v", entry, ok)
	}
	if !manifest.Permissions.Secrets || manifest.Permissions.Browser {
		t.Fatalf("unexpected permissions: %+v", manifest.Permissions)
	}
}

func TestParseManifestRejectsBadFiles(t *testing.T) {
	replacements := map[string]struct{ from, to, wantErr string }{
		"bad id":              {`"id": "demo-site"`, `"id": "Demo_Site"`, "id:"},
		"bad semver":          {`"version": "1.2.3"`, `"version": "1.2"`, "version:"},
		"bad api":             {`"apiVersion": "1"`, `"apiVersion": "v1"`, "apiVersion:"},
		"bad minHost":         {`"minHost": "2.0.0"`, `"minHost": "2.x"`, "minHost:"},
		"absolute entry":      {`"bin/linux-x64/demo"`, `"/usr/bin/demo"`, "entry.linux-x64"},
		"parent entry":        {`"bin/linux-x64/demo"`, `"../demo"`, "entry.linux-x64"},
		"unknown platform":    {`"linux-x64": "bin/linux-x64/demo"`, `"plan9-x64": "bin/demo"`, "entry.plan9-x64"},
		"unknown capability":  {`"updateCheck"`, `"magic"`, "unknown capability"},
		"unknown tile":        {`"capabilities": ["browse", "updateCheck"]`, `"capabilities": ["browse", "updateCheck"], "tile": "tall"`, "tile"},
		"old sync capability": {`"updateCheck"`, `"sync"`, "unknown capability"},
		"both sync":           {`"browse", "updateCheck"`, `"browse", "sync:games", "sync:index"`, "not both"},
		"unknown hook":        {`"onGameLaunch"`, `"onGameWhatever"`, "unknown hook"},
		"bad host":            {`"demo.example"`, `"demo"`, "invalid host"},
		"bad folder":          {`"$GAMES"`, `"GAMES"`, "invalid folder"},
		"old folder name":     {`"$GAMES"`, `"$ITCH_APP_DATA"`, "invalid folder"},
		"bad command method":  {`"demo.importKey"`, `"importKey"`, "method"},
		"bad repository":      {`"secrets": true`, `"repository": "gitlab:owner/repo", "secrets": true`, "repository"},
		"unknown field":       {`"secrets": true`, `"secrets": true, "typo": 1`, "unknown field"},
	}
	for name, tc := range replacements {
		t.Run(name, func(t *testing.T) {
			source := strings.Replace(validManifest, tc.from, tc.to, 1)
			_, err := ParseManifest([]byte(source))
			if err == nil {
				t.Fatalf("expected an error")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error %q does not mention %q", err, tc.wantErr)
			}
		})
	}
}

func TestIsRelativePath(t *testing.T) {
	cases := map[string]bool{
		"bin/app":       true,
		"":              false,
		"/bin/app":      false,
		"..":            false,
		"a/../b":        false,
		"a\\b":          false,
		"C:/app":        false,
		"ui/index.html": true,
	}
	for path, want := range cases {
		if got := isRelativePath(path); got != want {
			t.Errorf("isRelativePath(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestPlatformKeyFor(t *testing.T) {
	cases := map[[2]string]string{
		{"linux", "amd64"}:   "linux-x64",
		{"windows", "amd64"}: "windows-x64",
		{"darwin", "arm64"}:  "mac-arm64",
	}
	for input, want := range cases {
		if got := platformKeyFor(input[0], input[1]); got != want {
			t.Errorf("platformKeyFor(%s, %s) = %q, want %q", input[0], input[1], got, want)
		}
	}
}

// pendingFolderRenames lists the one folder name a real plugin still uses that the SDK renames.
// The rename is in progress, so this case only logs. Remove the entry once the rename lands.
var pendingFolderRenames = map[string]string{
	"light-launcher-plugin-itch": "$ITCH_APP_DATA",
}

func TestParseManifestAcceptsRealPluginFiles(t *testing.T) {
	for _, repo := range []string{"light-launcher-plugin-itch", "light-launcher-plugin-f95"} {
		t.Run(repo, func(t *testing.T) {
			path := filepath.Join(pluginRepoParent(), repo, "plugin.json")
			data, err := os.ReadFile(path)
			if os.IsNotExist(err) {
				t.Skipf("%s not present", path)
			}
			if err != nil {
				t.Fatal(err)
			}
			_, err = ParseManifest(data)
			if err == nil {
				return
			}
			if old, pending := pendingFolderRenames[repo]; pending && strings.Contains(err.Error(), old) {
				t.Logf("%s still uses %s, which is being renamed to $APPDATA_<NAME>: %v", repo, old, err)
				return
			}
			t.Fatal(err)
		})
	}
}

func TestSourceTileAndListGamesCapability(t *testing.T) {
	cases := []struct {
		name     string
		source   string
		wantTile string
	}{
		{"no tile", `{ "id": "demo", "label": "Demo", "capabilities": ["listGames"] }`, TilePoster},
		{"wide", `{ "id": "demo", "label": "Demo", "capabilities": ["listGames"], "tile": "wide" }`, TileWide},
		{"square", `{ "id": "demo", "label": "Demo", "capabilities": ["listGames"], "tile": "square" }`, TileSquare},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := strings.Replace(validManifest, `{ "id": "demo", "label": "Demo", "capabilities": ["browse", "updateCheck"] }`, tc.source, 1)
			manifest, err := ParseManifest([]byte(source))
			if err != nil {
				t.Fatalf("ParseManifest: %v", err)
			}
			infos := sourceInfos("demo-site", manifest)
			if len(infos) != 1 || infos[0].Tile != tc.wantTile {
				t.Fatalf("sourceInfos = %+v, want tile %q", infos, tc.wantTile)
			}
			if !slices.Contains(infos[0].Capabilities, "listGames") {
				t.Fatalf("capabilities = %v, want listGames", infos[0].Capabilities)
			}
		})
	}
}

func TestParseManifestAcceptsFolderNamesAndCommandsWithoutMethod(t *testing.T) {
	source := strings.Replace(validManifest, `"$GAMES"`, `"$GAMES", "$DOWNLOADS", "$DATA", "$APPDATA_ITCH"`, 1)
	source = strings.Replace(source, `, "method": "demo.importKey"`, "", 1)
	manifest, err := ParseManifest([]byte(source))
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	if len(manifest.Contributes.Commands) != 1 || manifest.Contributes.Commands[0].Method != "" {
		t.Fatalf("commands = %+v, want one command without a method", manifest.Contributes.Commands)
	}
}
