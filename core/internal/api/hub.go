package api

import "sync"

// Event is one message pushed to the UI.
type Event struct {
	Channel string `json:"channel"`
	Data    any    `json:"data"`
}

// Hub fans events out to every subscriber. It implements Events.
type Hub struct {
	mu   sync.Mutex
	subs map[chan Event]struct{}
}

func NewHub() *Hub {
	return &Hub{subs: map[chan Event]struct{}{}}
}

// Subscribe returns a channel of events and a function that stops the subscription.
func (h *Hub) Subscribe() (<-chan Event, func()) {
	ch := make(chan Event, 64)
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()
	cancel := func() {
		h.mu.Lock()
		delete(h.subs, ch)
		h.mu.Unlock()
	}
	return ch, cancel
}

// Emit never blocks: a slow subscriber drops events.
func (h *Hub) Emit(channel string, data any) {
	event := Event{Channel: channel, Data: data}
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subs {
		select {
		case ch <- event:
		default:
		}
	}
}
