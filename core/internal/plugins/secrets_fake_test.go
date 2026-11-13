package plugins

import (
	"context"
	"sort"
	"sync"

	"light-launcher/core/internal/contract"
)

// fakeSecrets is an in-memory contract.Secrets. It can pretend to be a store that cannot list
// keys, and can inject failures, so removal tests cover the full and the per-key paths.
type fakeSecrets struct {
	mu             sync.Mutex
	book           map[string]map[string]string
	notListable    bool
	deleteAllErr   error
	deleteErr      error
	deleteAllCalls int
	deleteCalls    int
}

// newFakeSecrets returns a fake store holding the given plugin secrets, without deep copying
// the inner maps beyond the entries passed in.
func newFakeSecrets(values map[string]map[string]string) *fakeSecrets {
	fake := &fakeSecrets{book: map[string]map[string]string{}}
	for plugin, secrets := range values {
		fake.book[plugin] = map[string]string{}
		for key, value := range secrets {
			fake.book[plugin][key] = value
		}
	}
	return fake
}

func (f *fakeSecrets) Get(_ context.Context, plugin, key string) (string, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	value, found := f.book[plugin][key]
	return value, found, nil
}

func (f *fakeSecrets) Set(_ context.Context, plugin, key, value string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.book[plugin] == nil {
		f.book[plugin] = map[string]string{}
	}
	f.book[plugin][key] = value
	return nil
}

func (f *fakeSecrets) Delete(_ context.Context, plugin, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deleteCalls++
	if f.deleteErr != nil {
		return f.deleteErr
	}
	delete(f.book[plugin], key)
	return nil
}

func (f *fakeSecrets) DeleteAll(_ context.Context, plugin string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deleteAllCalls++
	if f.deleteAllErr != nil {
		return f.deleteAllErr
	}
	if f.notListable {
		return contract.ErrSecretsNotListable
	}
	delete(f.book, plugin)
	return nil
}

// keys returns the sorted secret keys stored for a plugin.
func (f *fakeSecrets) keys(plugin string) []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	keys := make([]string, 0, len(f.book[plugin]))
	for key := range f.book[plugin] {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
