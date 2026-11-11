package assist

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestWriterWritesOneFlushedLine(t *testing.T) {
	tests := []struct {
		name  string
		event Event
		want  string
	}{
		{
			name:  "ready",
			event: Ready(),
			want:  "{\"event\":\"ready\"}\n",
		},
		{
			name:  "closed",
			event: Closed(),
			want:  "{\"event\":\"closed\"}\n",
		},
		{
			name:  "error",
			event: Error("bad thing"),
			want:  "{\"event\":\"error\",\"message\":\"bad thing\"}\n",
		},
		{
			name: "download with unknown size",
			event: Download("https://example.com/file.zip", "file.zip", -1,
				DownloadHeaders("a=b", "https://example.com/", "TestAgent")),
			want: "{\"event\":\"download\",\"url\":\"https://example.com/file.zip\",\"filename\":\"file.zip\",\"size\":-1,\"headers\":{\"Cookie\":\"a=b\",\"Referer\":\"https://example.com/\",\"User-Agent\":\"TestAgent\"}}\n",
		},
		{
			name: "download with known size",
			event: Download("https://example.com/file.zip", "file.zip", 123,
				DownloadHeaders("", "https://example.com/", "TestAgent")),
			want: "{\"event\":\"download\",\"url\":\"https://example.com/file.zip\",\"filename\":\"file.zip\",\"size\":123,\"headers\":{\"Referer\":\"https://example.com/\",\"User-Agent\":\"TestAgent\"}}\n",
		},
		{
			name: "download without headers",
			event: Download("https://example.com/file.zip", "file.zip", 0,
				DownloadHeaders("", "", "")),
			want: "{\"event\":\"download\",\"url\":\"https://example.com/file.zip\",\"filename\":\"file.zip\",\"size\":0}\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var out bytes.Buffer
			writer := NewWriter(&out)
			if err := writer.Write(test.event); err != nil {
				t.Fatalf("write: %v", err)
			}
			if got := out.String(); got != test.want {
				t.Fatalf("event = %q, want %q", got, test.want)
			}
		})
	}
}

func TestDownloadWithHeadersWritesAnEmptyObject(t *testing.T) {
	event := DownloadWithHeaders("https://example.com/empty", "empty", 0, nil)
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(data), `"size":0`) {
		t.Fatalf("download event without size: %s", data)
	}
	if !strings.Contains(string(data), `"headers":{}`) {
		t.Fatalf("download event without headers: %s", data)
	}
}

func TestErrorEventCarriesMessage(t *testing.T) {
	data, err := json.Marshal(Error("cannot start"))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var event map[string]any
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if event["event"] != "error" || event["message"] != "cannot start" {
		t.Fatalf("error event = %#v", event)
	}
	if len(event) != 2 {
		t.Fatalf("error event has extra fields: %#v", event)
	}
}

func TestWriterFlushesEachLine(t *testing.T) {
	var out bytes.Buffer
	writer := NewWriter(&out)

	if err := writer.Write(Ready()); err != nil {
		t.Fatalf("write ready: %v", err)
	}
	if got, want := out.String(), "{\"event\":\"ready\"}\n"; got != want {
		t.Fatalf("after first write = %q, want %q", got, want)
	}
	if err := writer.Write(Closed()); err != nil {
		t.Fatalf("write closed: %v", err)
	}
	want := "{\"event\":\"ready\"}\n{\"event\":\"closed\"}\n"
	if got := out.String(); got != want {
		t.Fatalf("after second write = %q, want %q", got, want)
	}
}

func TestWriterKeepsFirstError(t *testing.T) {
	writer := NewWriter(failingWriter{})
	if err := writer.Write(Ready()); err == nil {
		t.Fatal("first write succeeded, want a write error")
	}
	if err := writer.Write(Closed()); err == nil {
		t.Fatal("second write succeeded, want the first error")
	}
	if writer.Err() == nil {
		t.Fatal("Err() = nil, want the first error")
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}
