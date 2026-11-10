package downloads

import "sync"

// hostGates caps connections to one host across every running download.
type hostGates struct {
	mu    sync.Mutex
	limit int
	gates map[string]*Gate
}

func newHostGates(limit int) *hostGates {
	return &hostGates{limit: limit, gates: map[string]*Gate{}}
}

// setLimit applies a new per-host limit to existing and future hosts.
func (h *hostGates) setLimit(limit int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.limit = limit
	for _, gate := range h.gates {
		gate.SetLimit(limit)
	}
}

// forURL returns the gate for the host of rawURL.
func (h *hostGates) forURL(rawURL string) *Gate {
	host := hostOf(rawURL)
	h.mu.Lock()
	defer h.mu.Unlock()
	gate, ok := h.gates[host]
	if !ok {
		gate = NewGate(h.limit)
		h.gates[host] = gate
	}
	return gate
}
