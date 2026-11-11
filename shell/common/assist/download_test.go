package assist

import "testing"

func TestCookieHeader(t *testing.T) {
	tests := []struct {
		name    string
		cookies []CookiePair
		want    string
	}{
		{name: "none"},
		{name: "one", cookies: []CookiePair{{Name: "session", Value: "abc"}}, want: "session=abc"},
		{
			name:    "many keep order",
			cookies: []CookiePair{{Name: "a", Value: "1"}, {Name: "b", Value: "2"}},
			want:    "a=1; b=2",
		},
		{
			name:    "empty name skipped",
			cookies: []CookiePair{{Name: "", Value: "x"}, {Name: "b", Value: "2"}},
			want:    "b=2",
		},
		{name: "empty value kept", cookies: []CookiePair{{Name: "a", Value: ""}}, want: "a="},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := CookieHeader(test.cookies); got != test.want {
				t.Fatalf("CookieHeader(%q) = %q, want %q", test.cookies, got, test.want)
			}
		})
	}
}

func TestDownloadHeadersOnlyIncludesKnownValues(t *testing.T) {
	if got := DownloadHeaders("", "", ""); len(got) != 0 {
		t.Fatalf("headers = %v, want empty", got)
	}
	headers := DownloadHeaders("a=b", "https://example.com/page", "TestAgent")
	want := map[string]string{
		"Cookie":     "a=b",
		"Referer":    "https://example.com/page",
		"User-Agent": "TestAgent",
	}
	if len(headers) != len(want) {
		t.Fatalf("headers = %#v, want %#v", headers, want)
	}
	for key, value := range want {
		if got := headers[key]; got != value {
			t.Errorf("headers[%q] = %q, want %q", key, got, value)
		}
	}
}

func TestDownloadFileName(t *testing.T) {
	tests := []struct {
		suggested string
		rawURL    string
		want      string
	}{
		{suggested: "game.zip", rawURL: "https://example.com/other.zip", want: "game.zip"},
		{suggested: "", rawURL: "https://example.com/files/game.zip", want: "game.zip"},
		{suggested: "", rawURL: "https://example.com/", want: "download"},
		{suggested: "", rawURL: "", want: "download"},
	}
	for _, test := range tests {
		if got := DownloadFileName(test.suggested, test.rawURL); got != test.want {
			t.Errorf("DownloadFileName(%q, %q) = %q, want %q",
				test.suggested, test.rawURL, got, test.want)
		}
	}
}

func TestResultFileName(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{path: `C:\Users\example\Downloads\file.zip`, want: "file.zip"},
		{path: "/tmp/downloads/file.zip", want: "file.zip"},
		{path: "file.zip", want: "file.zip"},
		{path: "", want: ""},
	}
	for _, test := range tests {
		if got := ResultFileName(test.path); got != test.want {
			t.Errorf("ResultFileName(%q) = %q, want %q", test.path, got, test.want)
		}
	}
}
