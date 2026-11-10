package platform

import (
	"encoding/json"
	"testing"
)

func TestFeaturesUseCamelCaseJSONKeys(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"proton", "proton"},
		{"prefix", "prefix"},
		{"lsfg", "lsfg"},
		{"gamescope", "gamescope"},
		{"mangohud", "mangoHud"},
		{"gamemode", "gameMode"},
		{"live terminal logs", "liveTerminalLogs"},
	}

	data, err := json.Marshal(Current())
	if err != nil {
		t.Fatalf("marshal features: %v", err)
	}
	var keys map[string]any
	if err := json.Unmarshal(data, &keys); err != nil {
		t.Fatalf("unmarshal features: %v", err)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, ok := keys[test.want]; !ok {
				t.Errorf("JSON key %q missing from %s", test.want, data)
			}
		})
	}
}
