//go:build windows

package main

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/jchv/go-webview2"
)

func TestBindNativeFunctionsOnlyAcceptsTheLauncherOrigin(t *testing.T) {
	const (
		launcher = "http://127.0.0.1:4000"
		secret   = "test-secret"
	)
	view := &fakeWebView{bindings: map[string]any{}}
	var picked, opened []string
	err := bindNativeFunctions(view, launcher, secret, nativeActions{
		pick: func(method string, id int) { picked = append(picked, method) },
		open: func(target string) error {
			opened = append(opened, target)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("bindNativeFunctions: %v", err)
	}

	pick, ok := view.bindings["llRequestPick"].(func(string, string, string, int))
	if !ok {
		t.Fatalf("llRequestPick bound as %T", view.bindings["llRequestPick"])
	}
	open, ok := view.bindings["llOpenExternal"].(func(string, string, string) error)
	if !ok {
		t.Fatalf("llOpenExternal bound as %T", view.bindings["llOpenExternal"])
	}

	pick("wrong-secret", launcher, "pickFile", 1)
	pick(secret, "https://example.com", "pickFile", 2)
	if err := open("wrong-secret", launcher, "https://example.com/file"); err != nil {
		t.Fatalf("open with a wrong secret: %v", err)
	}
	if err := open(secret, "https://example.com", "https://example.com/file"); err != nil {
		t.Fatalf("open from another origin: %v", err)
	}
	if len(picked) != 0 || len(opened) != 0 {
		t.Fatalf("another origin called the bridge: picked=%v opened=%v", picked, opened)
	}

	pick(secret, launcher, "pickFile", 3)
	if len(picked) != 1 || picked[0] != "pickFile" {
		t.Fatalf("picked = %v, want [pickFile]", picked)
	}
	if err := open(secret, launcher, "https://example.org/file"); err != nil {
		t.Fatalf("open from the launcher origin: %v", err)
	}
	if len(opened) != 1 || opened[0] != "https://example.org/file" {
		t.Fatalf("opened = %v, want the external address", opened)
	}
}

func TestBridgeScriptForFillsTheOriginAndSecret(t *testing.T) {
	script, err := bridgeScriptFor("http://127.0.0.1:4000", "test-secret")
	if err != nil {
		t.Fatalf("bridgeScriptFor: %v", err)
	}
	for _, placeholder := range []string{"__LL_ORIGIN__", "__LL_SECRET__"} {
		if strings.Contains(script, placeholder) {
			t.Errorf("script still contains %s", placeholder)
		}
	}
	if !strings.Contains(script, `"http://127.0.0.1:4000"`) {
		t.Error("script does not contain the launcher origin")
	}
	if !strings.Contains(script, `"test-secret"`) {
		t.Error("script does not contain the bridge secret")
	}
}

func TestNewBridgeSecretIsRandomHex(t *testing.T) {
	first, err := newBridgeSecret()
	if err != nil {
		t.Fatalf("newBridgeSecret: %v", err)
	}
	second, err := newBridgeSecret()
	if err != nil {
		t.Fatalf("newBridgeSecret: %v", err)
	}
	if len(first) != 64 {
		t.Fatalf("secret length = %d, want 64", len(first))
	}
	if first == second {
		t.Fatal("two secrets are equal")
	}
	if _, err := hex.DecodeString(first); err != nil {
		t.Fatalf("secret is not hex: %v", err)
	}
}

type fakeWebView struct {
	webview2.WebView
	bindings map[string]any
}

func (f *fakeWebView) Bind(name string, fn any) error {
	f.bindings[name] = fn
	return nil
}
