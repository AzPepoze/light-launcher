package appsvc

import (
	"encoding/base64"
)

// dataURL encodes bytes as a base64 data URL for the UI.
func dataURL(mime string, data []byte) string {
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data)
}
