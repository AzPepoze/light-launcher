package downloads

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"light-launcher/core/internal/contract"
	"light-launcher/core/internal/logger"
)

// encryptCTR encrypts src with a fresh AES-128-CTR key and returns the bytes and the host spec.
func encryptCTR(t *testing.T, src []byte) ([]byte, *contract.DecryptSpec) {
	t.Helper()
	key := patternData(16)
	nonce := patternData(8)
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	iv := make([]byte, aes.BlockSize)
	copy(iv, nonce)
	out := make([]byte, len(src))
	cipher.NewCTR(block, iv).XORKeyStream(out, src)
	return out, &contract.DecryptSpec{
		Algorithm: contract.DecryptAlgorithmAES128CTR,
		Key:       base64.StdEncoding.EncodeToString(key),
		Nonce:     base64.StdEncoding.EncodeToString(nonce),
	}
}

func TestTransferDecryptsWithAESCTRAndKeepsByteCount(t *testing.T) {
	plain := patternData(300_000)
	ciphertext, spec := encryptCTR(t, plain)
	file := &testFile{data: ciphertext, etag: `"v1"`}
	srv := httptest.NewServer(file)
	defer srv.Close()

	row := newTestRow(t, srv.URL+"/game.bin", 2)
	row.Decrypt = spec
	_, errs := startJob(newTestTransferer(1), row, nil)
	if err := <-errs; err != nil {
		t.Fatalf("transfer: %v", err)
	}
	mustMatchFile(t, row.Dest, plain)
	info, err := os.Stat(row.Dest)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() != int64(len(plain)) {
		t.Fatalf("decrypted size = %d, want %d", info.Size(), len(plain))
	}
	if _, err := os.Stat(row.Dest + ".part"); !os.IsNotExist(err) {
		t.Fatal("the encrypted part file should be gone after the decrypted copy is in place")
	}
	assertNoDecryptTemp(t, filepath.Dir(row.Dest))
}

func TestDownloadWithoutDecryptKeepsBytesUnchanged(t *testing.T) {
	plain := patternData(64_000)
	file := &testFile{data: plain, etag: `"v1"`}
	srv := httptest.NewServer(file)
	defer srv.Close()

	row := newTestRow(t, srv.URL+"/plain.bin", 1)
	_, errs := startJob(newTestTransferer(1), row, nil)
	if err := <-errs; err != nil {
		t.Fatalf("transfer: %v", err)
	}
	mustMatchFile(t, row.Dest, plain)
	assertNoDecryptTemp(t, filepath.Dir(row.Dest))
}

func TestDecryptedDownloadInstallsTheArchive(t *testing.T) {
	archive := makeZip(t, []zipEntry{{"Game/" + launcherName(), "binary"}})
	data, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, spec := encryptCTR(t, data)
	file := &testFile{data: ciphertext, etag: `"v1"`}
	srv := httptest.NewServer(file)
	defer srv.Close()

	library := t.TempDir()
	dir := t.TempDir()
	m, err := NewManager(Options{
		Store:     NewMemoryStore(),
		Dir:       dir,
		Installer: &FolderInstaller{Root: library},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { m.Shutdown(context.Background()) })

	id, err := m.Enqueue(context.Background(), contract.DownloadRequest{
		URL: srv.URL + "/Game.zip", Install: true, GameID: 4, Decrypt: spec,
	})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	waitForState(t, m, id, StateDone)
	waitFor(t, func() bool {
		_, err := os.Stat(filepath.Join(library, "Game", launcherName()))
		return err == nil
	})
	waitFor(t, func() bool { return !fileExists(filepath.Join(dir, "Game.zip")) })
	assertNoDecryptTemp(t, dir)
}

func TestWrongDecryptKeyFailsAndDoesNotInstall(t *testing.T) {
	archive := makeZip(t, []zipEntry{{"Game/" + launcherName(), "binary"}})
	data, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, spec := encryptCTR(t, data)
	wrongKey := patternData(16)
	wrongKey[0] ^= 0xff
	wrong := *spec
	wrong.Key = base64.StdEncoding.EncodeToString(wrongKey)
	file := &testFile{data: ciphertext, etag: `"v1"`}
	srv := httptest.NewServer(file)
	defer srv.Close()

	library := t.TempDir()
	dir := t.TempDir()
	m, err := NewManager(Options{
		Store:     NewMemoryStore(),
		Dir:       dir,
		Installer: &FolderInstaller{Root: library},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { m.Shutdown(context.Background()) })

	id, err := m.Enqueue(context.Background(), contract.DownloadRequest{
		URL: srv.URL + "/Game.zip", Install: true, GameID: 4, Decrypt: &wrong,
	})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	view := waitForState(t, m, id, StateError)
	if !strings.Contains(view.Error, "decrypt") {
		t.Fatalf("error = %q, want it to mention decryption", view.Error)
	}
	if strings.Contains(view.Error, wrong.Key) {
		t.Fatal("the decrypt error must not contain the key")
	}
	if entries, err := os.ReadDir(library); err != nil || len(entries) != 0 {
		t.Fatalf("a failed decrypt installed something: entries=%v err=%v", entries, err)
	}
	if _, err := os.Stat(view.Dest); !os.IsNotExist(err) {
		t.Fatalf("final path exists after a failed decrypt: %v", err)
	}
	assertNoDecryptTemp(t, dir)
}

func TestEnqueueRejectsBadDecryptSpecs(t *testing.T) {
	goodKey := base64.StdEncoding.EncodeToString(patternData(16))
	goodNonce := base64.StdEncoding.EncodeToString(patternData(8))
	cases := map[string]*contract.DecryptSpec{
		"wrong algorithm":  {Algorithm: "aes-256-ctr", Key: goodKey, Nonce: goodNonce},
		"empty algorithm":  {Key: goodKey, Nonce: goodNonce},
		"short key":        {Algorithm: contract.DecryptAlgorithmAES128CTR, Key: base64.StdEncoding.EncodeToString(patternData(15)), Nonce: goodNonce},
		"long key":         {Algorithm: contract.DecryptAlgorithmAES128CTR, Key: base64.StdEncoding.EncodeToString(patternData(17)), Nonce: goodNonce},
		"bad key base64":   {Algorithm: contract.DecryptAlgorithmAES128CTR, Key: "not base64!", Nonce: goodNonce},
		"short nonce":      {Algorithm: contract.DecryptAlgorithmAES128CTR, Key: goodKey, Nonce: base64.StdEncoding.EncodeToString(patternData(7))},
		"long nonce":       {Algorithm: contract.DecryptAlgorithmAES128CTR, Key: goodKey, Nonce: base64.StdEncoding.EncodeToString(patternData(9))},
		"bad nonce base64": {Algorithm: contract.DecryptAlgorithmAES128CTR, Key: goodKey, Nonce: "?"},
	}
	file := &testFile{data: patternData(1024), etag: `"v1"`}
	srv := httptest.NewServer(file)
	defer srv.Close()

	for name, spec := range cases {
		t.Run(name, func(t *testing.T) {
			store := NewMemoryStore()
			m := newTestManager(t, store, t.TempDir(), Settings{})
			_, err := m.Enqueue(context.Background(), contract.DownloadRequest{
				URL: srv.URL + "/game.zip", Filename: "game.zip", Decrypt: spec,
			})
			if err == nil {
				t.Fatal("a bad decrypt spec was accepted")
			}
			if spec.Key != "" && strings.Contains(err.Error(), spec.Key) {
				t.Fatal("the validation error must not contain the key")
			}
			rows, listErr := store.List(context.Background())
			if listErr != nil {
				t.Fatal(listErr)
			}
			if len(rows) != 0 {
				t.Fatalf("a rejected request left %d download rows", len(rows))
			}
		})
	}
}

func TestStoreDecryptRoundTrip(t *testing.T) {
	for name, store := range storesUnderTest(t) {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			_, spec := encryptCTR(t, []byte("payload"))
			id, err := store.Insert(ctx, Download{
				URL: "https://example.test/a.zip", Filename: "a.zip", Dest: "/tmp/a.zip",
				State: StateQueued, Decrypt: spec,
			})
			if err != nil {
				t.Fatal(err)
			}
			got, err := store.Get(ctx, id)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.Decrypt, spec) {
				t.Fatalf("decrypt spec = %+v, want %+v", got.Decrypt, spec)
			}
			rows, err := store.List(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) != 1 || !reflect.DeepEqual(rows[0].Decrypt, spec) {
				t.Fatalf("decrypt spec lost from the list: %+v", rows)
			}
		})
	}
}

func TestSQLiteDecryptSpecSurvivesManagerRestart(t *testing.T) {
	plain := patternData(50_000)
	ciphertext, spec := encryptCTR(t, plain)
	file := &testFile{data: ciphertext, etag: `"v1"`}
	srv := httptest.NewServer(file)
	defer srv.Close()

	store := newSQLiteTestStore(t)
	dir := t.TempDir()
	dest := filepath.Join(dir, "restart.bin")
	id, err := store.Insert(context.Background(), Download{
		URL: srv.URL + "/restart.bin", Filename: "restart.bin", Dest: dest,
		Connections: 1, State: StateQueued, Decrypt: spec,
	})
	if err != nil {
		t.Fatal(err)
	}
	// A new manager reads the row from SQLite, as it would after a restart.
	m := newTestManager(t, store, dir, Settings{})
	if err := m.Restore(context.Background()); err != nil {
		t.Fatal(err)
	}
	done := waitForState(t, m, id, StateDone)
	mustMatchFile(t, done.Dest, plain)
}

// assertNoDecryptTemp fails when a decryption temp file was left in dir.
func assertNoDecryptTemp(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), ".decrypt-") {
			t.Fatalf("decryption temp file left behind: %s", e.Name())
		}
	}
}

func TestDecryptKeyIsNotLogged(t *testing.T) {
	logDir := t.TempDir()
	if err := logger.Init(logDir, "test.log", 1); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(logger.Close)

	plain := patternData(20_000)
	ciphertext, spec := encryptCTR(t, plain)
	file := &testFile{data: ciphertext, etag: `"v1"`}
	srv := httptest.NewServer(file)
	defer srv.Close()

	m := newTestManager(t, nil, t.TempDir(), Settings{})
	id, err := m.Enqueue(context.Background(), contract.DownloadRequest{
		URL: srv.URL + "/ok.bin", Filename: "ok.bin", Decrypt: spec,
	})
	if err != nil {
		t.Fatal(err)
	}
	waitForState(t, m, id, StateDone)

	wrongKey := patternData(16)
	wrongKey[0] ^= 0xff
	wrong := *spec
	wrong.Key = base64.StdEncoding.EncodeToString(wrongKey)
	id, err = m.Enqueue(context.Background(), contract.DownloadRequest{
		URL: srv.URL + "/bad.bin", Filename: "bad.bin", Decrypt: &wrong, Install: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	waitForState(t, m, id, StateError)

	data, err := os.ReadFile(filepath.Join(logDir, "test.log"))
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{spec.Key, wrong.Key} {
		if strings.Contains(string(data), key) {
			t.Fatalf("a decrypt key reached the log: %s", key)
		}
	}
	// Printing a spec by accident must redact it too.
	if printed := fmt.Sprintf("%v %+v", spec, *spec); strings.Contains(printed, spec.Key) || strings.Contains(printed, spec.Nonce) {
		t.Fatalf("printing a decrypt spec revealed its secrets: %s", printed)
	}
}
