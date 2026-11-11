package main

import (
	"bytes"
	"testing"
)

func TestEventWriterWritesOneFlushedLine(t *testing.T) {
	tests := []struct {
		name  string
		event shellEvent
		want  string
	}{
		{
			name:  "ready",
			event: readyEvent(),
			want:  "{\"event\":\"ready\"}\n",
		},
		{
			name:  "closed",
			event: closedEvent(),
			want:  "{\"event\":\"closed\"}\n",
		},
		{
			name:  "error",
			event: errorEvent("bad thing"),
			want:  "{\"event\":\"error\",\"message\":\"bad thing\"}\n",
		},
		{
			name: "download with unknown size",
			event: downloadEvent("https://example.com/file.zip", "file.zip", -1,
				downloadHeaders("a=b", "https://example.com/", "TestAgent")),
			want: "{\"event\":\"download\",\"url\":\"https://example.com/file.zip\",\"filename\":\"file.zip\",\"size\":-1,\"headers\":{\"Cookie\":\"a=b\",\"Referer\":\"https://example.com/\",\"User-Agent\":\"TestAgent\"}}\n",
		},
		{
			name: "download with known size",
			event: downloadEvent("https://example.com/file.zip", "file.zip", 123,
				downloadHeaders("", "https://example.com/", "TestAgent")),
			want: "{\"event\":\"download\",\"url\":\"https://example.com/file.zip\",\"filename\":\"file.zip\",\"size\":123,\"headers\":{\"Referer\":\"https://example.com/\",\"User-Agent\":\"TestAgent\"}}\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var out bytes.Buffer
			writer := newEventWriter(&out)
			if err := writer.write(test.event); err != nil {
				t.Fatalf("write: %v", err)
			}
			if got := out.String(); got != test.want {
				t.Fatalf("event = %q, want %q", got, test.want)
			}
		})
	}
}

func TestEventWriterFlushesEachLine(t *testing.T) {
	var out bytes.Buffer
	writer := newEventWriter(&out)

	if err := writer.write(readyEvent()); err != nil {
		t.Fatalf("write ready: %v", err)
	}
	if got, want := out.String(), "{\"event\":\"ready\"}\n"; got != want {
		t.Fatalf("after first write = %q, want %q", got, want)
	}
	if err := writer.write(closedEvent()); err != nil {
		t.Fatalf("write closed: %v", err)
	}
	want := "{\"event\":\"ready\"}\n{\"event\":\"closed\"}\n"
	if got := out.String(); got != want {
		t.Fatalf("after second write = %q, want %q", got, want)
	}
}

func TestDownloadHeadersOnlyIncludesKnownValues(t *testing.T) {
	if got := downloadHeaders("", "", ""); len(got) != 0 {
		t.Fatalf("headers = %v, want empty", got)
	}
	headers := downloadHeaders("a=b", "https://example.com/page", "TestAgent")
	for key, want := range map[string]string{
		"Cookie":     "a=b",
		"Referer":    "https://example.com/page",
		"User-Agent": "TestAgent",
	} {
		if got := headers[key]; got != want {
			t.Errorf("headers[%q] = %q, want %q", key, got, want)
		}
	}
}

func TestIsCaptureDownload(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{url: "https://example.com/file", want: true},
		{url: "http://example.com/file", want: true},
		{url: "blob:https://example.com/id", want: false},
		{url: "data:text/plain,hello", want: false},
		{url: "ftp://example.com/file", want: false},
		{url: "https:///file", want: false},
		{url: "not a url", want: false},
	}
	for _, test := range tests {
		if got := isCaptureDownload(test.url); got != test.want {
			t.Errorf("isCaptureDownload(%q) = %v, want %v", test.url, got, test.want)
		}
	}
}
