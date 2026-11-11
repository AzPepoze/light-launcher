package bridge

import (
	"encoding/json"
	"testing"
)

func TestAllowsOnlyForTheLauncherOrigin(t *testing.T) {
	const launcher = "http://127.0.0.1:4000"
	tests := []struct {
		page string
		want bool
	}{
		{page: "http://127.0.0.1:4000/", want: true},
		{page: "http://127.0.0.1:4000/game/1?tab=files#top", want: true},
		{page: "http://127.0.0.1:4001/", want: false},
		{page: "https://127.0.0.1:4000/", want: false},
		{page: "https://example.com/", want: false},
		{page: "about:blank", want: false},
		{page: "not a url", want: false},
	}
	for _, test := range tests {
		if got := Allows(test.page, launcher); got != test.want {
			t.Errorf("Allows(%q) = %v, want %v", test.page, got, test.want)
		}
	}
	if Allows("http://127.0.0.1:4000/", "") {
		t.Error("an empty launcher origin allowed the bridge")
	}
	if !Allows("http://EXAMPLE.com/", "http://example.com") {
		t.Error("a host with different case did not match the launcher origin")
	}
}

func TestParseRequest(t *testing.T) {
	request, ok := ParseRequest(`{"id":7,"method":"pickFolder"}`)
	if !ok {
		t.Fatal("valid request was rejected")
	}
	if request.ID != 7 || request.Method != "pickFolder" {
		t.Fatalf("request = %+v", request)
	}
	for _, body := range []string{"", "not json", `{"id":0,"method":"pickFile"}`, `{"method":"pickFile"}`} {
		if _, ok := ParseRequest(body); ok {
			t.Errorf("ParseRequest(%q) succeeded, want an error", body)
		}
	}
}

func TestResolveScript(t *testing.T) {
	script, err := ResolveScript(3, "")
	if err != nil {
		t.Fatalf("ResolveScript: %v", err)
	}
	if want := "window.llShell._resolve(3, null)"; script != want {
		t.Fatalf("ResolveScript(cancel) = %q, want %q", script, want)
	}

	path := `C:\Games\a"b`
	literal, err := json.Marshal(path)
	if err != nil {
		t.Fatalf("marshal path: %v", err)
	}
	script, err = ResolveScript(4, path)
	if err != nil {
		t.Fatalf("ResolveScript: %v", err)
	}
	want := "window.llShell._resolve(4, " + string(literal) + ")"
	if script != want {
		t.Fatalf("ResolveScript(path) = %q, want %q", script, want)
	}
}
