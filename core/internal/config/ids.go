package config

import (
	"crypto/rand"
	"encoding/hex"
)

// NewID returns a 16-character hex id, as the Electron version did.
func NewID() string {
	buffer := make([]byte, 8)
	_, _ = rand.Read(buffer)
	return hex.EncodeToString(buffer)
}
