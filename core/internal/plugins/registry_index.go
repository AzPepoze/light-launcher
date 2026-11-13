package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	// RegistryIndexVersion is the only community plugin index version this launcher reads.
	RegistryIndexVersion = 1

	registryIndexTimeout        = 10 * time.Second
	maxRegistryIndexBytes       = 512 << 10
	maxRegistryEntries          = 500
	maxRegistryRedirects        = 5
	maxRegistryNameBytes        = 80
	maxRegistryDescriptionBytes = 1000
	maxRegistryAuthorBytes      = 80
)

var (
	// ErrRegistryURL means the configured index URL, or a redirect from it, is not https.
	ErrRegistryURL = errors.New("community plugin index URL must use https")
	// ErrRegistryFetch means the index could not be downloaded.
	ErrRegistryFetch = errors.New("could not fetch the community plugin index")
	// ErrRegistryTooLarge means the index reply is bigger than the launcher accepts.
	ErrRegistryTooLarge = errors.New("community plugin index is too large")
	// ErrRegistryVersion means the index declares a version this launcher does not read.
	ErrRegistryVersion = errors.New("unsupported community plugin index version")
	// ErrRegistryInvalid means the index, or one of its entries, failed validation.
	ErrRegistryInvalid = errors.New("invalid community plugin index")

	registrySHA256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// RegistryIndex is the community plugin list published at PluginRegistryURL.
// Every entry describes an installable package, so sha256 is always required.
type RegistryIndex struct {
	Version int                  `json:"version"`
	Plugins []RegistryIndexEntry `json:"plugins"`
}

// RegistryIndexEntry is one plugin a community index offers.
type RegistryIndexEntry struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Repository  string `json:"repository"`
	Version     string `json:"version"`
	DownloadURL string `json:"downloadUrl"`
	// SHA256 is the digest of the package at DownloadURL. An install is refused without it.
	SHA256   string `json:"sha256"`
	Homepage string `json:"homepage"`
}

// RegistryIndexResult is the reply of PluginRegistry. Error is set when a configured index could
// not be loaded; the UI shows it with a retry instead of failing the whole call.
type RegistryIndexResult struct {
	Configured bool                 `json:"configured"`
	URL        string               `json:"url"`
	Entries    []RegistryIndexEntry `json:"entries"`
	Error      string               `json:"error,omitempty"`
}

// HTTPDoer is the part of http.Client the registry index client uses.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// RegistryIndexClient fetches a community plugin index. The zero value uses an https-only client
// and a short timeout. It never caches, installs or polls.
type RegistryIndexClient struct {
	Client  HTTPDoer
	Timeout time.Duration
}

// FetchRegistryIndex fetches and validates an index with the default client.
func FetchRegistryIndex(ctx context.Context, rawURL string) (RegistryIndex, error) {
	return RegistryIndexClient{}.Fetch(ctx, rawURL)
}

// Fetch downloads one index and returns it after version, entry and limit checks.
func (c RegistryIndexClient) Fetch(ctx context.Context, rawURL string) (RegistryIndex, error) {
	if err := checkRegistryIndexURL(rawURL); err != nil {
		return RegistryIndex{}, err
	}
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = registryIndexTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	body, err := fetchRegistryIndexBody(ctx, c.httpClient(), rawURL)
	if err != nil {
		return RegistryIndex{}, err
	}
	return decodeRegistryIndex(body)
}

// httpClient returns a client that lets Fetch follow redirects itself, so every hop stays on https.
func (c RegistryIndexClient) httpClient() HTTPDoer {
	client := c.Client
	if client == nil {
		return &http.Client{}
	}
	if std, ok := client.(*http.Client); ok {
		clone := *std
		clone.CheckRedirect = func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}
		return &clone
	}
	return client
}

// fetchRegistryIndexBody reads the reply body, with an https check on every redirect.
func fetchRegistryIndexBody(ctx context.Context, client HTTPDoer, rawURL string) ([]byte, error) {
	current := rawURL
	for redirects := 0; ; redirects++ {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, current, nil)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrRegistryFetch, err)
		}
		request.Header.Set("Accept", "application/json")
		request.Header.Set("User-Agent", "light-launcher-plugin-registry")

		response, err := client.Do(request)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrRegistryFetch, err)
		}
		if isRegistryRedirect(response.StatusCode) {
			location := response.Header.Get("Location")
			response.Body.Close()
			if redirects >= maxRegistryRedirects {
				return nil, fmt.Errorf("%w: too many redirects", ErrRegistryFetch)
			}
			next, err := request.URL.Parse(location)
			if err != nil {
				return nil, fmt.Errorf("%w: bad redirect: %v", ErrRegistryFetch, err)
			}
			if err := checkRegistryIndexURL(next.String()); err != nil {
				return nil, err
			}
			current = next.String()
			continue
		}
		return readRegistryIndexBody(response)
	}
}

// readRegistryIndexBody enforces the status code and the size limit.
func readRegistryIndexBody(response *http.Response) ([]byte, error) {
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d", ErrRegistryFetch, response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxRegistryIndexBytes+1))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRegistryFetch, err)
	}
	if len(body) > maxRegistryIndexBytes {
		return nil, ErrRegistryTooLarge
	}
	return body, nil
}

func isRegistryRedirect(status int) bool {
	switch status {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther,
		http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		return true
	default:
		return false
	}
}

// checkRegistryIndexURL accepts only an absolute https URL with a host.
func checkRegistryIndexURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" {
		return fmt.Errorf("%w: %s", ErrRegistryURL, raw)
	}
	return nil
}

func isHTTPSURL(raw string) bool {
	parsed, err := url.Parse(raw)
	return err == nil && parsed.Scheme == "https" && parsed.Hostname() != ""
}

// decodeRegistryIndex parses the index and applies the validation rules to every entry.
func decodeRegistryIndex(data []byte) (RegistryIndex, error) {
	var index RegistryIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return RegistryIndex{}, fmt.Errorf("%w: %v", ErrRegistryInvalid, err)
	}
	for i := range index.Plugins {
		index.Plugins[i].SHA256 = strings.ToLower(strings.TrimSpace(index.Plugins[i].SHA256))
	}
	if err := index.validate(); err != nil {
		return RegistryIndex{}, err
	}
	return index, nil
}

// validate checks the version, the entry count and each entry, including duplicate ids.
func (index RegistryIndex) validate() error {
	if index.Version != RegistryIndexVersion {
		return fmt.Errorf("%w: got version %d, want %d", ErrRegistryVersion, index.Version, RegistryIndexVersion)
	}
	if len(index.Plugins) > maxRegistryEntries {
		return fmt.Errorf("%w: %d entries, at most %d are allowed", ErrRegistryInvalid, len(index.Plugins), maxRegistryEntries)
	}
	seen := make(map[string]struct{}, len(index.Plugins))
	for i, entry := range index.Plugins {
		if err := entry.validate(); err != nil {
			return fmt.Errorf("%w: plugin %d: %v", ErrRegistryInvalid, i+1, err)
		}
		if _, duplicate := seen[entry.ID]; duplicate {
			return fmt.Errorf("%w: duplicate id %q", ErrRegistryInvalid, entry.ID)
		}
		seen[entry.ID] = struct{}{}
	}
	return nil
}

// validate checks one entry. URLs must be https, and every install entry needs a SHA256.
func (entry RegistryIndexEntry) validate() error {
	if !idPattern.MatchString(entry.ID) {
		return fmt.Errorf("id %q must be 2 to 40 lowercase letters, digits or dashes, starting with a letter", entry.ID)
	}
	if strings.TrimSpace(entry.Name) == "" || len(entry.Name) > maxRegistryNameBytes {
		return fmt.Errorf("name must be 1 to %d characters", maxRegistryNameBytes)
	}
	if len(entry.Description) > maxRegistryDescriptionBytes {
		return fmt.Errorf("description must be at most %d characters", maxRegistryDescriptionBytes)
	}
	if len(entry.Author) > maxRegistryAuthorBytes {
		return fmt.Errorf("author must be at most %d characters", maxRegistryAuthorBytes)
	}
	if !versionPattern.MatchString(entry.Version) {
		return fmt.Errorf("version %q must be a semver version such as 1.0.0", entry.Version)
	}
	if !isHTTPSURL(entry.DownloadURL) {
		return fmt.Errorf("downloadUrl %q must be an https URL", entry.DownloadURL)
	}
	if !registrySHA256Pattern.MatchString(entry.SHA256) {
		return fmt.Errorf("sha256 %q must be 64 hex digits", entry.SHA256)
	}
	if entry.Homepage != "" && !isHTTPSURL(entry.Homepage) {
		return fmt.Errorf("homepage %q must be an https URL", entry.Homepage)
	}
	if entry.Repository != "" && !repositoryPattern.MatchString(entry.Repository) && !isHTTPSURL(entry.Repository) {
		return fmt.Errorf("repository %q must look like github:owner/repo or be an https URL", entry.Repository)
	}
	return nil
}
