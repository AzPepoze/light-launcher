package config

import "testing"

func TestDefaultSettingsDoNotOpenInBrowser(t *testing.T) {
	if DefaultSettings().OpenInBrowser {
		t.Fatal("OpenInBrowser defaults to true, want false")
	}
}
