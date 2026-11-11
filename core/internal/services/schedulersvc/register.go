// Package schedulersvc exposes the background scheduler to the UI.
package schedulersvc

import (
	"context"

	"light-launcher/core/internal/api"
	"light-launcher/core/internal/scheduler"
)

type runInput struct {
	Job string `json:"job"`
}

type runResult struct {
	Started bool `json:"started"`
}

// Register adds the scheduler methods to the router. Jobs push their own library events.
func Register(r *api.Router, _ api.Events, sched *scheduler.Scheduler) {
	r.Handle("SchedulerStatus", api.NoInput(func(context.Context) (scheduler.Status, error) {
		return sched.Status(), nil
	}))
	r.Handle("SchedulerRunNow", api.Typed(func(_ context.Context, in runInput) (runResult, error) {
		started, err := sched.RunNow(in.Job)
		if err != nil {
			return runResult{}, err
		}
		return runResult{Started: started}, nil
	}))
}
