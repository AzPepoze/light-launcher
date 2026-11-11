package plugins

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"light-launcher/core/internal/contract"
)

const (
	secretsFileName = "secrets.enc"
	secretsKeyName  = "secrets.key"
	secretKeyBytes  = 32
)

// secretBook maps plugin id to its key/value secrets.
type secretBook map[string]map[string]string

// trackedSecrets writes through to the secret store and records each key name in the plugin's state.
type trackedSecrets struct {
	inner contract.Secrets
	state *stateStore
}

func (s trackedSecrets) Get(ctx context.Context, plugin, key string) (string, bool, error) {
	return s.inner.Get(ctx, plugin, key)
}

func (s trackedSecrets) Set(ctx context.Context, plugin, key, value string) error {
	if err := s.inner.Set(ctx, plugin, key, value); err != nil {
		return err
	}
	return s.state.update(plugin, func(state *PluginState) {
		if !slices.Contains(state.SecretKeys, key) {
			state.SecretKeys = append(state.SecretKeys, key)
		}
	})
}

func (s trackedSecrets) Delete(ctx context.Context, plugin, key string) error {
	if err := s.inner.Delete(ctx, plugin, key); err != nil {
		return err
	}
	return s.state.update(plugin, func(state *PluginState) {
		state.SecretKeys = slices.DeleteFunc(state.SecretKeys, func(stored string) bool {
			return stored == key
		})
	})
}

// fileSecrets keeps secrets in an AES-256-GCM encrypted file. The key sits in a 0600 file beside it.
// It is the fallback when no OS keyring is available.
type fileSecrets struct {
	mu      sync.Mutex
	path    string
	keyPath string
}

func newFileSecrets(dir string) *fileSecrets {
	return &fileSecrets{
		path:    filepath.Join(dir, secretsFileName),
		keyPath: filepath.Join(dir, secretsKeyName),
	}
}

func (s *fileSecrets) Get(_ context.Context, plugin, key string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	book, err := s.load()
	if err != nil {
		return "", false, err
	}
	value, ok := book[plugin][key]
	return value, ok, nil
}

func (s *fileSecrets) Set(_ context.Context, plugin, key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	book, err := s.load()
	if err != nil {
		return err
	}
	if book[plugin] == nil {
		book[plugin] = map[string]string{}
	}
	book[plugin][key] = value
	return s.save(book)
}

func (s *fileSecrets) Delete(_ context.Context, plugin, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	book, err := s.load()
	if err != nil {
		return err
	}
	delete(book[plugin], key)
	return s.save(book)
}

// load decrypts the file. A missing file is an empty book.
func (s *fileSecrets) load() (secretBook, error) {
	book := secretBook{}
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return book, nil
	}
	if err != nil {
		return nil, err
	}
	aead, err := s.cipher(false)
	if err != nil {
		return nil, err
	}
	if len(data) < aead.NonceSize() {
		return nil, errors.New("secrets file is truncated")
	}
	nonce, sealed := data[:aead.NonceSize()], data[aead.NonceSize():]
	plain, err := aead.Open(nil, nonce, sealed, nil)
	if err != nil {
		return nil, errors.New("secrets file cannot be decrypted with the key file")
	}
	if err := json.Unmarshal(plain, &book); err != nil {
		return nil, fmt.Errorf("secrets file is corrupt: %w", err)
	}
	return book, nil
}

// save encrypts the book and replaces the file through a temp file.
func (s *fileSecrets) save(book secretBook) error {
	aead, err := s.cipher(true)
	if err != nil {
		return err
	}
	plain, err := json.Marshal(book)
	if err != nil {
		return err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	sealed := aead.Seal(nonce, nonce, plain, nil)
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, sealed, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// cipher loads the key. When create is true, a missing key file is made first.
func (s *fileSecrets) cipher(create bool) (cipher.AEAD, error) {
	key, err := s.loadKey(create)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func (s *fileSecrets) loadKey(create bool) ([]byte, error) {
	key, err := os.ReadFile(s.keyPath)
	if err == nil && len(key) == secretKeyBytes {
		return key, nil
	}
	if err == nil {
		return nil, errors.New("secrets key file has the wrong size")
	}
	if !errors.Is(err, os.ErrNotExist) || !create {
		return nil, fmt.Errorf("secrets key file: %w", err)
	}
	return createKey(s.keyPath)
}

func createKey(path string) ([]byte, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	key := make([]byte, secretKeyBytes)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, key, 0o600); err != nil {
		return nil, err
	}
	return key, nil
}
