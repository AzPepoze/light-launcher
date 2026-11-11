package plugins

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// pluginReposEnv overrides the folder that holds the sibling plugin repositories.
const pluginReposEnv = "LL_PLUGIN_REPOS"

// pluginRepoParent is the folder that holds the sibling plugin repositories, which are only read.
// By default it is the folder next to this repository (the tests run in core/internal/plugins).
func pluginRepoParent() string {
	if dir := os.Getenv(pluginReposEnv); dir != "" {
		return dir
	}
	return filepath.Join("..", "..", "..", "..")
}

// TestRealPluginsHandshake compiles each real plugin with bun into a temp folder and runs its handshake.
// The compile writes to the temp folder, so the plugin repositories stay unchanged.
func TestRealPluginsHandshake(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("the real plugins are built for linux here")
	}
	bun, err := exec.LookPath("bun")
	if err != nil {
		t.Skip("bun is not installed")
	}
	for _, name := range []string{"itch", "f95"} {
		t.Run(name, func(t *testing.T) {
			source := filepath.Join(pluginRepoParent(), "light-launcher-plugin-"+name)
			if _, err := os.Stat(filepath.Join(source, "src", "main.ts")); err != nil {
				t.Skipf("plugin repo not present: %v", err)
			}
			program := compileBunPlugin(t, bun, source, name)
			sup := supervisorForProgram(t, name, program)
			ctx := t.Context()
			if _, err := sup.connect(ctx); err != nil {
				t.Fatalf("handshake with the real %s plugin: %v", name, err)
			}
			if !sup.Running() {
				t.Fatalf("%s plugin is not running after the handshake", name)
			}
			sup.Shutdown()
		})
	}
}

// compileBunPlugin runs "bun build --compile" for linux-x64 and writes the program into a temp folder.
func compileBunPlugin(t *testing.T, bun, source, name string) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), name)
	build := exec.Command(bun, "build", "--compile", "--target=bun-linux-x64", "src/main.ts", "--outfile", out)
	build.Dir = source
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("bun build %s: %v\n%s", name, err, output)
	}
	return out
}

// supervisorForProgram starts a supervisor for a compiled program placed at its manifest entry path.
func supervisorForProgram(t *testing.T, name, program string) *Supervisor {
	t.Helper()
	dir := t.TempDir()
	entry := "bin/" + PlatformKey() + "/" + name
	copyExecutable(t, program, filepath.Join(dir, filepath.FromSlash(entry)))
	manifest := Manifest{
		ID:         name,
		Name:       name,
		Version:    "0.0.0",
		APIVersion: SupportedAPIVersion,
		MinHost:    "2.0.0",
		Entry:      map[string]string{PlatformKey(): entry},
	}
	deps := &serviceDeps{limiter: newHostLimiter(0), cache: newETagCache()}
	env := newPluginEnv(manifest, t.TempDir(), deps, newLineRing(100))
	return newSupervisor(supervisorConfig{
		manifest:    manifest,
		dir:         dir,
		dataDir:     env.dataDir,
		hostVersion: "2.0.0",
		env:         env,
	})
}
