package plugins

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// PluginState is what the host remembers about one plugin, besides its files.
type PluginState struct {
	Enabled             bool     `json:"enabled"`
	Source              string   `json:"source,omitempty"`
	Unsigned            bool     `json:"unsigned,omitempty"`
	DisabledReason      string   `json:"disabledReason,omitempty"`
	AcceptedPermissions []string `json:"acceptedPermissions,omitempty"`
	// SecretKeys lists the secrets the plugin stored. Keyrings cannot list keys, so uninstall reads them here.
	SecretKeys []string `json:"secretKeys,omitempty"`
}

// stateStore keeps state.json in step with memory. It is safe for concurrent use.
type stateStore struct {
	mu      sync.Mutex
	path    string
	entries map[string]PluginState
}

func loadState(path string) (*stateStore, error) {
	store := &stateStore{path: path, entries: map[string]PluginState{}}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return store, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &store.entries); err != nil {
		return nil, errors.New("state.json is not valid: " + err.Error())
	}
	return store, nil
}

// get returns the state of a plugin. A plugin without an entry counts as enabled.
func (s *stateStore) get(id string) PluginState {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.entries[id]
	if !ok {
		return PluginState{Enabled: true}
	}
	return state
}

// update changes one plugin's entry and saves the file.
func (s *stateStore) update(id string, change func(*PluginState)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.entries[id]
	if !ok {
		state = PluginState{Enabled: true}
	}
	change(&state)
	s.entries[id] = state
	return s.saveLocked()
}

func (s *stateStore) remove(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, id)
	return s.saveLocked()
}

func (s *stateStore) saveLocked() error {
	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
