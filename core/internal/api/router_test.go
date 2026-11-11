package api

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestTimeoutOfReportsOnlyMethodsWithALimit(t *testing.T) {
	router := NewRouter()
	noop := func(context.Context, json.RawMessage) (any, error) { return nil, nil }
	router.HandleWithTimeout("Slow", 11*time.Minute, noop)
	router.Handle("Fast", noop)

	if got := router.TimeoutOf("Slow"); got != 11*time.Minute {
		t.Errorf("TimeoutOf(Slow) = %v, want 11m", got)
	}
	if got := router.TimeoutOf("Fast"); got != 0 {
		t.Errorf("TimeoutOf(Fast) = %v, want 0", got)
	}
	if got := router.TimeoutOf("Missing"); got != 0 {
		t.Errorf("TimeoutOf(Missing) = %v, want 0", got)
	}
}

func TestHandleWithTimeoutRegistersTheMethod(t *testing.T) {
	router := NewRouter()
	router.HandleWithTimeout("Slow", time.Minute, func(context.Context, json.RawMessage) (any, error) {
		return "done", nil
	})
	got, err := router.Dispatch(context.Background(), "Slow", nil)
	if err != nil || got != "done" {
		t.Fatalf("Dispatch = %v, %v; want done, nil", got, err)
	}
}
