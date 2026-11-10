package appsvc

import (
	"crypto/rand"
	"encoding/hex"
)

// generateID returns a 16-character hex id, as the Electron version did.
func generateID() string {
	return randomHex(8)
}

// randomHex returns n random bytes as hex. crypto/rand does not fail on supported platforms.
func randomHex(n int) string {
	buffer := make([]byte, n)
	_, _ = rand.Read(buffer)
	return hex.EncodeToString(buffer)
}
