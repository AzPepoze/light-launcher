package downloads

import (
	"context"
	"database/sql"
	"errors"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"light-launcher/core/internal/contract"
)

func TestValidateSubfolderAcceptsPlainNames(t *testing.T) {
	for _, sub := range []string{"", "Demo", "Team/Demo", "Team/Demo 2 (v1)"} {
		if err := ValidateSubfolder(sub); err != nil {
			t.Errorf("ValidateSubfolder(%q) = %v, want nil", sub, err)
		}
	}
}

func TestValidateSubfolderRejectsUnsafePaths(t *testing.T) {
	cases := map[string]string{
		"traversal":        "..",
		"nested traversal": "a/../b",
		"trailing dots":    "a/..",
		"dot segment":      "./a",
		"absolute":         "/etc",
		"backslash":        `a\b`,
		"drive letter":     "C:/games",
		"stream name":      "game:stream",
		"empty segment":    "a//b",
		"trailing slash":   "a/",
		"control char":     "a\x00b",
	}
	for name, sub := range cases {
		err := ValidateSubfolder(sub)
		if !errors.Is(err, contract.ErrBadFolder) {
			t.Errorf("%s: ValidateSubfolder(%q) = %v, want ErrBadFolder", name, sub, err)
		}
	}
}

func TestResolveFolderCreatesMissingFolders(t *testing.T) {
	root := t.TempDir()
	dir, err := resolveFolder(root, "Team/Demo")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(root, "Team", "Demo"); dir != want {
		t.Fatalf("dir = %s, want %s", dir, want)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Fatalf("folder not created: %v", err)
	}
}

func TestResolveFolderRootWhenSubfolderIsEmpty(t *testing.T) {
	root := t.TempDir()
	dir, err := resolveFolder(root, "")
	if err != nil || dir != root {
		t.Fatalf("resolveFolder(root, \"\") = %q, %v", dir, err)
	}
}

func TestResolveFolderRefusesFileInTheWay(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "Demo"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := resolveFolder(root, "Demo/Sub"); !errors.Is(err, contract.ErrBadFolder) {
		t.Fatalf("file segment: err = %v, want ErrBadFolder", err)
	}
}

func TestResolveFolderRefusesLinkSegments(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Skipf("symlinks are not available: %v", err)
	}
	if _, err := resolveFolder(root, "link"); !errors.Is(err, contract.ErrBadFolder) {
		t.Fatalf("link segment: err = %v, want ErrBadFolder", err)
	}
	if _, err := resolveFolder(root, "link/deeper"); !errors.Is(err, contract.ErrBadFolder) {
		t.Fatalf("link below the root: err = %v, want ErrBadFolder", err)
	}
	if entries, _ := os.ReadDir(outside); len(entries) != 0 {
		t.Fatalf("a folder was created through the link")
	}
}

func TestValidateRequest(t *testing.T) {
	cases := []struct {
		name string
		req  contract.DownloadRequest
		want error
	}{
		{"defaults", contract.DownloadRequest{}, nil},
		{"downloads root", contract.DownloadRequest{Root: contract.RootDownloads, Subfolder: "Team/Demo"}, nil},
		{"games root", contract.DownloadRequest{Root: contract.RootGames}, nil},
		{"unknown root", contract.DownloadRequest{Root: "music"}, contract.ErrBadFolder},
		{"traversal", contract.DownloadRequest{Subfolder: "../escape"}, contract.ErrBadFolder},
		{"install without flag", contract.DownloadRequest{DeleteArchive: new(bool)}, contract.ErrBadRequest},
	}
	for _, c := range cases {
		err := ValidateRequest(c.req)
		if c.want == nil && err != nil {
			t.Errorf("%s: err = %v, want nil", c.name, err)
		}
		if c.want != nil && !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, want %v", c.name, err, c.want)
		}
	}
}

func TestCheckInstallOptions(t *testing.T) {
	cases := []struct {
		name string
		req  contract.DownloadRequest
		want error
	}{
		{"plain download", contract.DownloadRequest{}, nil},
		{"install into subfolder", contract.DownloadRequest{Install: true, InstallSubfolder: "Pack"}, nil},
		{"subfolder without install", contract.DownloadRequest{InstallSubfolder: "Pack"}, contract.ErrBadRequest},
		{"delete without install", contract.DownloadRequest{DeleteArchive: new(bool)}, contract.ErrBadRequest},
		{"bad install subfolder", contract.DownloadRequest{Install: true, InstallSubfolder: ".."}, contract.ErrBadFolder},
	}
	for _, c := range cases {
		err := checkInstallOptions(c.req)
		if c.want == nil && err != nil {
			t.Errorf("%s: err = %v, want nil", c.name, err)
		}
		if c.want != nil && !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, want %v", c.name, err, c.want)
		}
	}
}

func TestEnqueueSavesIntoSubfolder(t *testing.T) {
	file := &testFile{data: patternData(1 << 10), etag: `"v1"`}
	srv := httptest.NewServer(file)
	defer srv.Close()
	dir := t.TempDir()
	m := newTestManager(t, nil, dir, Settings{})

	id, err := m.Enqueue(context.Background(), contract.DownloadRequest{
		URL: srv.URL + "/demo.bin", Filename: "demo.bin", Subfolder: "Team/Demo",
	})
	if err != nil {
		t.Fatal(err)
	}
	view := waitForState(t, m, id, StateDone)
	if want := filepath.Join(dir, "Team", "Demo", "demo.bin"); view.Dest != want {
		t.Fatalf("dest = %s, want %s", view.Dest, want)
	}
	if !fileExists(view.Dest) {
		t.Fatal("file was not saved into the subfolder")
	}
}

func TestEnqueueRootGamesNeedsGamesFolder(t *testing.T) {
	file := &testFile{data: patternData(1 << 10), etag: `"v1"`}
	srv := httptest.NewServer(file)
	defer srv.Close()
	games := t.TempDir()

	noGames := newTestManager(t, nil, t.TempDir(), Settings{})
	_, err := noGames.Enqueue(context.Background(), contract.DownloadRequest{URL: srv.URL + "/a.bin", Root: contract.RootGames})
	if !errors.Is(err, contract.ErrBadFolder) {
		t.Fatalf("games root without a games folder: err = %v, want ErrBadFolder", err)
	}

	m, err := NewManager(Options{Store: NewMemoryStore(), Dir: t.TempDir(), GamesDir: games})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { m.Shutdown(context.Background()) })
	id, err := m.Enqueue(context.Background(), contract.DownloadRequest{
		URL: srv.URL + "/b.bin", Filename: "b.bin", Root: contract.RootGames, Subfolder: "Mods",
	})
	if err != nil {
		t.Fatal(err)
	}
	view := waitForState(t, m, id, StateDone)
	if want := filepath.Join(games, "Mods", "b.bin"); view.Dest != want {
		t.Fatalf("dest = %s, want %s", view.Dest, want)
	}
}

func TestEnqueueRejectsUnknownRootAndUnsafeSubfolder(t *testing.T) {
	m := newTestManager(t, nil, t.TempDir(), Settings{})
	ctx := context.Background()
	url := "http://127.0.0.1:1/file.bin"

	if _, err := m.Enqueue(ctx, contract.DownloadRequest{URL: url, Root: "music"}); !errors.Is(err, contract.ErrBadFolder) {
		t.Fatalf("unknown root: err = %v, want ErrBadFolder", err)
	}
	if _, err := m.Enqueue(ctx, contract.DownloadRequest{URL: url, Subfolder: "../escape"}); !errors.Is(err, contract.ErrBadFolder) {
		t.Fatalf("traversal: err = %v, want ErrBadFolder", err)
	}
}

func TestEnqueueRefusesSymlinkSubfolder(t *testing.T) {
	dir := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(dir, "link")); err != nil {
		t.Skipf("symlinks are not available: %v", err)
	}
	m := newTestManager(t, nil, dir, Settings{})
	_, err := m.Enqueue(context.Background(), contract.DownloadRequest{URL: "http://127.0.0.1:1/x.bin", Subfolder: "link"})
	if !errors.Is(err, contract.ErrBadFolder) {
		t.Fatalf("symlink subfolder: err = %v, want ErrBadFolder", err)
	}
}

func TestDownloadWithInstallSubfolderInstallsThere(t *testing.T) {
	hideSevenZip(t)
	archive := makeZip(t, []zipEntry{{"Game/" + launcherName(), "binary"}})
	data, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(&testFile{data: data, etag: `"v1"`})
	defer srv.Close()

	library := t.TempDir()
	m, err := NewManager(Options{
		Store:     NewMemoryStore(),
		Dir:       t.TempDir(),
		Installer: &FolderInstaller{Root: library},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { m.Shutdown(context.Background()) })

	id, err := m.Enqueue(context.Background(), contract.DownloadRequest{
		URL: srv.URL + "/Game.zip", Install: true, InstallSubfolder: "Pack/Two", GameID: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	waitForState(t, m, id, StateDone)
	waitFor(t, func() bool {
		_, err := os.Stat(filepath.Join(library, "Pack", "Two", "Game", launcherName()))
		return err == nil
	})
	if fileExists(filepath.Join(library, "Game")) {
		t.Fatal("game was installed outside the requested subfolder")
	}
}

func TestDeleteArchiveOverridesSettings(t *testing.T) {
	hideSevenZip(t)
	archive := makeZip(t, []zipEntry{{"Game/" + launcherName(), "binary"}})
	data, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(&testFile{data: data, etag: `"v1"`})
	defer srv.Close()

	keep := false
	dir := t.TempDir()
	library := t.TempDir()
	m, err := NewManager(Options{
		Store:     NewMemoryStore(),
		Dir:       dir,
		Installer: &FolderInstaller{Root: library},
		Settings:  Settings{KeepArchives: false, Connections: 1, MaxActive: 1, HostConnections: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { m.Shutdown(context.Background()) })

	id, err := m.Enqueue(context.Background(), contract.DownloadRequest{
		URL: srv.URL + "/Game.zip", Install: true, DeleteArchive: &keep,
	})
	if err != nil {
		t.Fatal(err)
	}
	waitForState(t, m, id, StateDone)
	waitFor(t, func() bool { return fileExists(filepath.Join(library, "Game", launcherName())) })
	m.installs.Wait()
	if !fileExists(filepath.Join(dir, "Game.zip")) {
		t.Fatal("deleteArchive false should keep the archive even though KeepArchives is false")
	}
}

func TestSQLiteStoreAddsFolderColumnsToOldDatabase(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "old.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	ctx := context.Background()
	// The table as older builds created it, without the folder columns.
	if _, err := db.ExecContext(ctx, `create table downloads (
  id integer primary key autoincrement, game_id integer not null default 0, url text not null,
  filename text not null, dest text not null, size integer not null default 0,
  bytes_done integer not null default 0, connections integer not null default 1,
  status text not null default 'queued', error text, etag text, headers text,
  install integer not null default 0, created_at text not null, started_at text, finished_at text)`); err != nil {
		t.Fatal(err)
	}
	store, err := NewSQLiteStore(ctx, db)
	if err != nil {
		t.Fatalf("upgrade: %v", err)
	}
	old, err := store.Insert(ctx, Download{URL: "http://x", Filename: "a.zip", Dest: "/d/a.zip", State: StateQueued})
	if err != nil {
		t.Fatal(err)
	}
	d, err := store.Get(ctx, old)
	if err != nil || d.InstallSubfolder != "" || d.DeleteArchive != nil {
		t.Fatalf("old row = %+v, %v", d, err)
	}

	yes := true
	id, err := store.Insert(ctx, Download{URL: "http://x", Filename: "b.zip", Dest: "/d/b.zip", State: StateQueued,
		InstallSubfolder: "Pack", DeleteArchive: &yes})
	if err != nil {
		t.Fatal(err)
	}
	d, err = store.Get(ctx, id)
	if err != nil || d.InstallSubfolder != "Pack" || d.DeleteArchive == nil || !*d.DeleteArchive {
		t.Fatalf("new row = %+v, %v", d, err)
	}
}
