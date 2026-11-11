package schedulersvc

import (
	"context"
	"encoding/json"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/scheduler"
)

func newRouter(t *testing.T) (*api.Router, *scheduler.Scheduler) {
	t.Helper()
	sched, err := scheduler.New(context.Background(), scheduler.Options{
		Jobs: []scheduler.Job{{
			Name:     "sync",
			Interval: time.Hour,
			Run:      func(context.Context) error { return nil },
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	r := api.NewRouter()
	Register(r, api.NewHub(), sched)
	return r, sched
}

func TestRegisterAddsSchedulerMethods(t *testing.T) {
	r, _ := newRouter(t)
	want := []string{"SchedulerRunNow", "SchedulerStatus"}
	got := r.Methods()
	for _, name := range want {
		if !slices.Contains(got, name) {
			t.Errorf("method %s is not registered", name)
		}
	}
	if len(got) != len(want) {
		t.Errorf("registered %d methods, want %d", len(got), len(want))
	}
}

func TestSchedulerStatusHasJobFields(t *testing.T) {
	r, _ := newRouter(t)
	result, err := r.Dispatch(context.Background(), "SchedulerStatus", nil)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Jobs []map[string]any `json:"jobs"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Jobs) != 1 || decoded.Jobs[0]["name"] != "sync" {
		t.Fatalf("status = %s", raw)
	}
	for _, field := range []string{"intervalSeconds", "lastRunAt", "lastDurationMs", "nextRunAt", "running"} {
		if _, ok := decoded.Jobs[0][field]; !ok {
			t.Errorf("status is missing %s: %s", field, raw)
		}
	}
}

func TestSchedulerRunNowStartsJob(t *testing.T) {
	var ran atomic.Bool
	sched, err := scheduler.New(context.Background(), scheduler.Options{
		Jobs: []scheduler.Job{{
			Name:     "sync",
			Interval: time.Hour,
			Run: func(context.Context) error {
				ran.Store(true)
				return nil
			},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	r := api.NewRouter()
	Register(r, api.NewHub(), sched)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		sched.Run(ctx)
	}()
	<-sched.Ready()
	defer func() {
		cancel()
		<-done
	}()

	result, err := r.Dispatch(ctx, "SchedulerRunNow", json.RawMessage(`{"job":"sync"}`))
	if err != nil {
		t.Fatal(err)
	}
	if !result.(runResult).Started {
		t.Fatal("RunNow did not start the job")
	}
	deadline := time.Now().Add(time.Second)
	for !ran.Load() && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if !ran.Load() {
		t.Fatal("job did not run")
	}
}

func TestSchedulerRunNowRejectsUnknownJob(t *testing.T) {
	r, _ := newRouter(t)
	if _, err := r.Dispatch(context.Background(), "SchedulerRunNow", json.RawMessage(`{"job":"missing"}`)); err == nil {
		t.Fatal("an unknown job must fail")
	}
}
