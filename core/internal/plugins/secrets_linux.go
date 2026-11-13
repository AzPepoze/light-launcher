//go:build linux

package plugins

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"light-launcher/core/internal/contract"
)

// newPlatformSecrets uses the secret-tool keyring when it is installed, otherwise the encrypted file.
func newPlatformSecrets(dir string) contract.Secrets {
	if _, err := exec.LookPath("secret-tool"); err != nil {
		return newFileSecrets(dir)
	}
	return keyringSecrets{}
}

// keyringSecrets stores secrets in the desktop keyring through secret-tool (libsecret).
type keyringSecrets struct{}

const keyringService = "light-launcher"

func keyringAttributes(plugin, key string) []string {
	return []string{"service", keyringService, "plugin", plugin, "key", key}
}

func (keyringSecrets) Get(ctx context.Context, plugin, key string) (string, bool, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "secret-tool", append([]string{"lookup"}, keyringAttributes(plugin, key)...)...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		return stdout.String(), true, nil
	}
	if isKeyringNoMatch(err, stderr.String()) {
		return "", false, nil
	}
	return "", false, fmt.Errorf("keyring lookup: %s", strings.TrimSpace(stderr.String()))
}

func (keyringSecrets) Set(ctx context.Context, plugin, key, value string) error {
	label := fmt.Sprintf("light-launcher %s %s", plugin, key)
	args := append([]string{"store", "--label=" + label}, keyringAttributes(plugin, key)...)
	cmd := exec.CommandContext(ctx, "secret-tool", args...)
	cmd.Stdin = strings.NewReader(value)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("keyring store: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

func (keyringSecrets) Delete(ctx context.Context, plugin, key string) error {
	return clearKeyring(ctx, keyringAttributes(plugin, key)...)
}

// DeleteAll removes every key of one plugin. secret-tool clears all unlocked items that match the
// attributes it is given, so passing the service and plugin (without a key) scopes the delete to
// this plugin and leaves other plugins alone.
func (keyringSecrets) DeleteAll(ctx context.Context, plugin string) error {
	if err := clearKeyring(ctx, "service", keyringService, "plugin", plugin); err != nil {
		return err
	}
	return nil
}

// clearKeyring deletes every item that matches these attribute pairs. secret-tool exits with
// status 1 and no message when nothing matched, which is not an error here.
func clearKeyring(ctx context.Context, attributes ...string) error {
	args := append([]string{"clear"}, attributes...)
	output, err := exec.CommandContext(ctx, "secret-tool", args...).CombinedOutput()
	if err == nil || isKeyringNoMatch(err, string(output)) {
		return nil
	}
	return fmt.Errorf("keyring clear: %s", strings.TrimSpace(string(output)))
}

// isKeyringNoMatch reports the secret-tool result for "nothing matched": status 1 with no message.
func isKeyringNoMatch(err error, message string) bool {
	var exitErr *exec.ExitError
	return errors.As(err, &exitErr) && exitErr.ExitCode() == 1 && strings.TrimSpace(message) == ""
}
