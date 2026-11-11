package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBannerScript(t *testing.T) {
	if got := bannerScript(""); got != "" {
		t.Fatalf("bannerScript(\"\") = %q, want an empty string", got)
	}

	message := "Finish the page, then press \"Download\""
	script := bannerScript(message)
	if !strings.Contains(script, "window.top !== window") {
		t.Error("banner script runs in child frames")
	}
	if !strings.Contains(script, "pointer-events:none") {
		t.Error("banner script can block clicks")
	}
	encoded, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("marshal message: %v", err)
	}
	if !strings.Contains(script, string(encoded)) {
		t.Fatalf("banner script does not contain the escaped message: %s", script)
	}
}
