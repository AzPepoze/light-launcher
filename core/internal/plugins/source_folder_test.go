package plugins

import (
	"strings"
	"testing"
)

// sourceWithGrants returns a manifest with one source and the given permissions object body.
func sourceWithGrants(permissions string) []byte {
	raw := string(sourceJSON(`{ "id": "demo", "label": "Demo", "capabilities": ["browse"] }`))
	return []byte(strings.Replace(raw, `"permissions": {}`, `"permissions": `+permissions, 1))
}

func TestSourceInfosResolveTheGamesFolderGrant(t *testing.T) {
	manifest, err := ParseManifest(sourceWithGrants(`{ "fs": ["$GAMES"] }`))
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	folders := map[string]string{"$GAMES": "/home/me/Games", "$DOWNLOADS": "/home/me/Downloads"}
	infos := sourceInfosWithFolders("f95", manifest, folders)
	if len(infos) != 1 {
		t.Fatalf("sources = %+v, want one", infos)
	}
	if infos[0].GamesFolder != "/home/me/Games" {
		t.Errorf("games folder = %q, want the resolved $GAMES folder", infos[0].GamesFolder)
	}
}

func TestSourceInfosLeaveTheGamesFolderEmptyWithoutTheGrant(t *testing.T) {
	manifest, err := ParseManifest(sourceWithGrants(`{ "fs": ["$DOWNLOADS"] }`))
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	folders := map[string]string{"$GAMES": "/home/me/Games", "$DOWNLOADS": "/home/me/Downloads"}
	infos := sourceInfosWithFolders("demo", manifest, folders)
	if len(infos) != 1 {
		t.Fatalf("sources = %+v, want one", infos)
	}
	if infos[0].GamesFolder != "" {
		t.Errorf("games folder = %q, want empty without the $GAMES grant", infos[0].GamesFolder)
	}
}

func TestSourceInfosWithoutHostFoldersLeaveTheGamesFolderEmpty(t *testing.T) {
	manifest, err := ParseManifest(sourceWithGrants(`{ "fs": ["$GAMES"] }`))
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	infos := sourceInfos("demo", manifest)
	if len(infos) != 1 {
		t.Fatalf("sources = %+v, want one", infos)
	}
	if infos[0].GamesFolder != "" {
		t.Errorf("games folder = %q, want empty without host folders", infos[0].GamesFolder)
	}
}
