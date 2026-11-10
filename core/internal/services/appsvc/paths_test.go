package appsvc

import (
	"os"
	"path/filepath"
	"testing"

	"light-launcher/core/internal/config"
)

func TestPrefixBaseDirectoryUsesCustomDirWhenSet(t *testing.T) {
	isolateConfigHome(t)
	custom := t.TempDir()
	if got := prefixBaseDirectory(custom); got != custom {
		t.Fatalf("prefixBaseDirectory(custom) = %q, want %q", got, custom)
	}
	if got := prefixBaseDirectory("   "); got != config.GetPrefixBaseDirectory() {
		t.Fatalf("blank custom dir must fall back to the default, got %q", got)
	}
}

func TestDirectoriesHangOffBaseDirectory(t *testing.T) {
	isolateConfigHome(t)
	base := config.GetBaseDirectory()
	if got := config.GetConfigDirectory(); got != filepath.Join(base, "config", "executables") {
		t.Errorf("config directory = %q", got)
	}
	if got := config.GetPrefixBaseDirectory(); got != filepath.Join(base, "prefixes") {
		t.Errorf("prefix directory = %q", got)
	}
	if got := config.GetLogsDirectory(); got != filepath.Join(base, "logs") {
		t.Errorf("logs directory = %q", got)
	}
	if got := config.GetExecutableConfigPath("MyGame", "id123"); got != filepath.Join(config.GetConfigDirectory(), "id123") {
		t.Errorf("executable config path by id = %q", got)
	}
	if got := config.GetExecutableConfigPath("MyGame", ""); got != filepath.Join(config.GetConfigDirectory(), "MyGame") {
		t.Errorf("executable config path by name = %q", got)
	}
}

func TestExpandPathHandlesTildeAndUnicode(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home directory")
	}
	if got := expandPath("~"); got != home {
		t.Errorf("expandPath(~) = %q, want %q", got, home)
	}
	if got := expandPath("~/games/steam"); got != filepath.Join(home, "games", "steam") {
		t.Errorf("expandPath(~/games/steam) = %q", got)
	}
	unicodePath := filepath.Join(string(filepath.Separator), "mnt", "Drive1", "Game♥", "OK", "Lustful Spirit Hunt.exe")
	if got := expandPath(unicodePath); got != unicodePath {
		t.Errorf("expandPath kept unicode path = %q, want %q", got, unicodePath)
	}
	if expandPath("") != "" {
		t.Error("expandPath of empty input must stay empty")
	}
}

func TestSafeExistsNeverFails(t *testing.T) {
	if safeExists("") {
		t.Error("empty path must not exist")
	}
	if safeExists(filepath.Join(t.TempDir(), "missing", "file.exe")) {
		t.Error("missing path must not exist")
	}
	home, err := os.UserHomeDir()
	if err == nil && !safeExists(home) {
		t.Error("home directory must exist")
	}
}
