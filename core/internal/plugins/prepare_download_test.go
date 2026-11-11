package plugins

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/logger"
)

func TestPreparedDownloadRequestOptions(t *testing.T) {
	games := t.TempDir()
	keep := false
	cases := []struct {
		name   string
		fs     []string
		folder string
		gameID int64
		result PrepareDownloadResult
		want   contract.DownloadRequest
		code   int
	}{
		{
			name:   "defaults",
			gameID: 7,
			result: PrepareDownloadResult{
				URL: "https://files.example/game.zip", Headers: map[string]string{"X-Token": "t"}, Filename: "game.zip",
			},
			want: contract.DownloadRequest{
				GameID: 7, URL: "https://files.example/game.zip", Headers: map[string]string{"X-Token": "t"},
				Filename: "game.zip", Root: contract.RootDownloads,
			},
		},
		{
			name: "downloads subfolder",
			result: PrepareDownloadResult{
				URL: "https://files.example/game.zip", Subfolder: "Team/Demo",
			},
			gameID: 7,
			want: contract.DownloadRequest{
				GameID: 7, URL: "https://files.example/game.zip", Root: contract.RootDownloads, Subfolder: "Team/Demo",
			},
		},
		{
			name: "games folder with grant",
			fs:   []string{"$GAMES"}, folder: games,
			gameID: 7,
			result: PrepareDownloadResult{
				URL: "https://files.example/game.zip", Root: contract.RootGames, Subfolder: "Mods",
			},
			want: contract.DownloadRequest{
				GameID: 7, URL: "https://files.example/game.zip", Root: contract.RootGames, Subfolder: "Mods",
			},
		},
		{
			name: "install with grant",
			fs:   []string{"$GAMES"}, folder: games,
			gameID: 7,
			result: PrepareDownloadResult{
				URL: "https://files.example/game.zip", Install: true, InstallSubfolder: "Pack", DeleteArchive: &keep,
			},
			want: contract.DownloadRequest{
				GameID: 7, URL: "https://files.example/game.zip", Root: contract.RootDownloads,
				Install: true, InstallSubfolder: "Pack", DeleteArchive: &keep,
			},
		},
		{
			name:   "games folder without grant",
			gameID: 7,
			result: PrepareDownloadResult{
				URL: "https://files.example/game.zip", Root: contract.RootGames,
			},
			code: CodePermissionDenied,
		},
		{
			name: "install without games grant",
			fs:   []string{"$DOWNLOADS"}, folder: t.TempDir(),
			gameID: 7,
			result: PrepareDownloadResult{
				URL: "https://files.example/game.zip", Install: true,
			},
			code: CodePermissionDenied,
		},
		{
			name:   "unknown root",
			gameID: 7,
			result: PrepareDownloadResult{
				URL: "https://files.example/game.zip", Root: "music",
			},
			code: CodeInvalidParams,
		},
		{
			name:   "traversal in subfolder",
			gameID: 7,
			result: PrepareDownloadResult{
				URL: "https://files.example/game.zip", Subfolder: "../escape",
			},
			code: CodeInvalidParams,
		},
		{
			name:   "absolute subfolder",
			gameID: 7,
			result: PrepareDownloadResult{
				URL: "https://files.example/game.zip", Subfolder: "/tmp/games",
			},
			code: CodeInvalidParams,
		},
		{
			name: "traversal in install subfolder",
			fs:   []string{"$GAMES"}, folder: games,
			gameID: 7,
			result: PrepareDownloadResult{
				URL: "https://files.example/game.zip", Install: true, InstallSubfolder: "..",
			},
			code: CodeInvalidParams,
		},
		{
			name:   "install subfolder without install",
			gameID: 7,
			result: PrepareDownloadResult{
				URL: "https://files.example/game.zip", InstallSubfolder: "Pack",
			},
			code: CodeInvalidParams,
		},
		{
			name:   "delete archive without install",
			gameID: 7,
			result: PrepareDownloadResult{
				URL: "https://files.example/game.zip", DeleteArchive: &keep,
			},
			code: CodeInvalidParams,
		},
		{
			name:   "negative game id",
			gameID: -1,
			result: PrepareDownloadResult{URL: "https://files.example/game.zip"},
			code:   CodeInvalidParams,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			env, _ := downloadEnv(t, c.fs, c.folder)
			got, err := env.preparedDownloadRequest(c.gameID, c.result)
			if c.code != 0 {
				expectCode(t, err, c.code)
				return
			}
			if err != nil {
				t.Fatalf("preparedDownloadRequest: %v", err)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("request = %+v, want %+v", got, c.want)
			}
		})
	}
}

func TestDecodePrepareDownloadIgnoresUnknownFields(t *testing.T) {
	result, err := decodePrepareDownload([]byte(`{"url":"https://files.example/a.zip","headers":{},"root":"games","where":"/tmp"}`), "plugin:test")
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Root != contract.RootGames {
		t.Fatalf("root = %q, want games", result.Root)
	}
}

func TestDecodePrepareDownloadLogsUnknownFields(t *testing.T) {
	dir := t.TempDir()
	if err := logger.Init(dir, "test.log", 1); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(logger.Close)

	if _, err := decodePrepareDownload([]byte(`{"url":"https://files.example/a.zip","headers":{},"where":"/tmp"}`), "plugin:test"); err != nil {
		t.Fatalf("decode: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "test.log"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `ignored unknown field "where"`) {
		t.Fatalf("unknown field was not logged: %s", data)
	}
}

func TestDecodePrepareDownloadRejectsWrongTypes(t *testing.T) {
	_, err := decodePrepareDownload([]byte(`{"url":"https://files.example/a.zip","headers":{},"root":5}`), "plugin:test")
	expectCode(t, err, CodeInvalidParams)
}

// newFixtureDownloadHost installs the fixture plugin with these permissions and a recording downloader.
func newFixtureDownloadHost(t *testing.T, perms Permissions) (*Host, *fakeDownloads) {
	t.Helper()
	base := t.TempDir()
	fake := &fakeDownloads{}
	host, err := NewHost(HostOptions{
		Layout:      Layout{Root: filepath.Join(base, "plugins"), Data: filepath.Join(base, "plugin-data")},
		HostVersion: "2.0.0",
		Secrets:     newFileSecrets(filepath.Join(base, "secrets")),
		Downloads:   fake,
		Folders:     map[string]string{"$GAMES": filepath.Join(base, "games")},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(host.Shutdown)

	source := filepath.Join(t.TempDir(), "fixture.llplugin")
	writeFixturePackageWith(t, source, "0.1.0", perms)
	if _, err := host.Install(context.Background(), InstallRequest{
		Source: source, AllowUnsigned: true, AcceptedPermissions: PermissionSummary(perms),
	}); err != nil {
		t.Fatal(err)
	}
	return host, fake
}

func TestFixturePluginDownloadHintsReachTheDownloader(t *testing.T) {
	host, fake := newFixtureDownloadHost(t, Permissions{Network: []string{"files.example"}, FS: []string{"$GAMES"}})
	id, err := NewRegistry(host).StartDownload(context.Background(), "fixture", map[string]any{"source": "fixture"}, 9)
	if err != nil {
		t.Fatalf("StartDownload: %v", err)
	}
	if id != 42 {
		t.Fatalf("download id = %d, want 42", id)
	}
	got := fake.last
	if got.GameID != 9 || got.URL != "https://files.example/game.zip" || got.Filename != "game.zip" ||
		got.Headers["X-Token"] != "fixture" || got.Root != contract.RootGames ||
		got.Subfolder != "Team/Demo" || !got.Install || got.InstallSubfolder != "Pack" ||
		got.DeleteArchive == nil || *got.DeleteArchive {
		t.Fatalf("queued request = %+v", got)
	}
}

func TestFixturePluginDownloadHintsNeedGamesGrant(t *testing.T) {
	host, fake := newFixtureDownloadHost(t, Permissions{Network: []string{"files.example"}})
	_, err := NewRegistry(host).StartDownload(context.Background(), "fixture", map[string]any{"source": "fixture"}, 9)
	expectCode(t, err, CodePermissionDenied)
	if fake.last.URL != "" {
		t.Fatalf("a denied prepare reached the downloader: %+v", fake.last)
	}
}
