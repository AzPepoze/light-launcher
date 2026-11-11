// Package api is the single seam between the UI and the Go services.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Handler serves one UI method. The payload is the raw JSON object sent by the UI.
type Handler func(ctx context.Context, payload json.RawMessage) (any, error)

// Events pushes a named event to every connected UI (progress bars, running sessions).
type Events interface {
	Emit(channel string, data any)
}

// Router maps method names such as "GetAllGames" to handlers.
type Router struct {
	mu       sync.RWMutex
	handlers map[string]Handler
	timeouts map[string]time.Duration
}

func NewRouter() *Router {
	return &Router{handlers: map[string]Handler{}, timeouts: map[string]time.Duration{}}
}

// Handle registers a handler. Registering the same method twice is a programming error.
func (r *Router) Handle(method string, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.handlers[method]; exists {
		panic("api: duplicate method " + method)
	}
	r.handlers[method] = handler
}

// HandleWithTimeout registers a handler that may run longer than the default invoke timeout.
func (r *Router) HandleWithTimeout(method string, timeout time.Duration, handler Handler) {
	r.Handle(method, handler)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.timeouts[method] = timeout
}

// TimeoutOf returns the limit set by HandleWithTimeout, or zero when the method has none.
func (r *Router) TimeoutOf(method string) time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.timeouts[method]
}

func (r *Router) Dispatch(ctx context.Context, method string, payload json.RawMessage) (any, error) {
	r.mu.RLock()
	handler, ok := r.handlers[method]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown method %s", method)
	}
	return handler(ctx, payload)
}

// Methods lists the registered method names, for tests and diagnostics.
func (r *Router) Methods() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.handlers))
	for name := range r.handlers {
		names = append(names, name)
	}
	return names
}

// Typed wraps a function with a decoded input so handlers stay one-liners.
func Typed[In any, Out any](fn func(ctx context.Context, in In) (Out, error)) Handler {
	return func(ctx context.Context, payload json.RawMessage) (any, error) {
		var in In
		if len(payload) > 0 && string(payload) != "null" {
			if err := json.Unmarshal(payload, &in); err != nil {
				return nil, fmt.Errorf("bad payload: %w", err)
			}
		}
		return fn(ctx, in)
	}
}

// NoInput wraps a function that takes no payload.
func NoInput[Out any](fn func(ctx context.Context) (Out, error)) Handler {
	return func(ctx context.Context, _ json.RawMessage) (any, error) {
		return fn(ctx)
	}
}
