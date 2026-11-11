package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"light-launcher/core/internal/contract"
)

func TestPrepareContextUsesPrepareDownloadTimeout(t *testing.T) {
	before := time.Now()
	ctx, cancel := prepareContext(context.Background())
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("prepare context has no deadline")
	}
	want := before.Add(contract.PrepareDownloadTimeout)
	if drift := deadline.Sub(want); drift < 0 || drift > time.Second {
		t.Fatalf("deadline is %v from the expected time, want about %v after start", drift, contract.PrepareDownloadTimeout)
	}
}

func TestServeTimeoutForMethods(t *testing.T) {
	cases := []struct {
		method string
		want   time.Duration
	}{
		{method: "browser.assist", want: contract.PrepareDownloadTimeout},
		{method: "downloads.enqueue", want: serveTimeout},
		{method: "http.request", want: serveTimeout},
		{method: "source.search", want: serveTimeout},
	}
	for _, c := range cases {
		if got := serveTimeoutFor(c.method); got != c.want {
			t.Errorf("serveTimeoutFor(%q) = %v, want %v", c.method, got, c.want)
		}
	}
}

// failingBrowser stands in for a browser window the person closed before a download was captured.
type failingBrowser struct {
	err error
}

func (f *failingBrowser) Assist(context.Context, BrowserAssistRequest) (BrowserCapture, error) {
	return BrowserCapture{}, f.err
}

func TestBrowserAssistClosedWindowSaysWhy(t *testing.T) {
	env := testEnv(t, Permissions{Browser: true}, nil)
	env.deps.browser = &failingBrowser{err: errors.New("the window stopped without a result")}

	_, err := callService(t, env, methodBrowserAssist, map[string]string{"url": "https://example.com"})
	expectCode(t, err, CodeInternal)
	if !strings.Contains(err.Error(), "the window stopped without a result") {
		t.Fatalf("err = %q, want the reason the window stopped", err)
	}
}

func TestBrowserAssistErrorReachesPrepareCaller(t *testing.T) {
	raw := json.RawMessage(`{"url":"https://example.com"}`)
	env := testEnv(t, Permissions{Browser: true}, nil)
	env.deps.browser = &failingBrowser{err: errors.New("browser assist timed out")}

	_, err := serviceBrowserAssist(context.Background(), env, raw)
	if err == nil || !strings.Contains(err.Error(), "browser assist timed out") {
		t.Fatalf("serviceBrowserAssist err = %v, want the timeout reason", err)
	}
}
