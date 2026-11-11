package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirectoriesHangOffBaseDirectory(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("APPDATA", dir)

	base := GetBaseDirectory()
	if got := GetConfigDirectory(); got != filepath.Join(base, ConfigDirName) {
		t.Errorf("config directory = %q", got)
	}
	if got := GetPrefixBaseDirectory(); got != filepath.Join(base, PrefixesDirName) {
		t.Errorf("prefix directory = %q", got)
	}
	if got := GetLogsDirectory(); got != filepath.Join(base, LogsDirName) {
		t.Errorf("logs directory = %q", got)
	}
	if got := SettingsFilePath(); got != filepath.Join(base, "settings.json") {
		t.Errorf("settings path = %q", got)
	}
	if got := ActivityFilePath(); got != filepath.Join(base, "activity.json") {
		t.Errorf("activity path = %q", got)
	}
}

func TestPrefixBaseDirUsesCustomDirWhenSet(t *testing.T) {
	custom := t.TempDir()
	if got := PrefixBaseDir(custom); got != custom {
		t.Fatalf("PrefixBaseDir(custom) = %q, want %q", got, custom)
	}
	if got := PrefixBaseDir("   "); got != GetPrefixBaseDirectory() {
		t.Fatalf("blank custom dir must fall back to the default, got %q", got)
	}
}

func TestPathExistsOnTempDir(t *testing.T) {
	dir := t.TempDir()
	if !PathExists(dir) {
		t.Fatal("temporary directory must exist")
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	if PathExists(dir) {
		t.Fatal("removed directory must not exist")
	}
}
