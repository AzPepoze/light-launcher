//go:build !linux && !windows

package plugins

import "light-launcher/core/internal/contract"

// newPlatformSecrets uses the encrypted file, because no keyring is wired for this OS yet.
func newPlatformSecrets(dir string) contract.Secrets {
	return newFileSecrets(dir)
}
