// Package updater checks GitHub Releases for newer plugin versions.
package updater

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"light-launcher/core/internal/semver"
)

const (
	defaultAPIBase     = "https://api.github.com"
	defaultUserAgent   = "light-launcher-plugin-updater"
	defaultBackoff     = time.Minute
	maxBackoff         = time.Hour
	maxResponseBytes   = 4 << 20
	maxChecksumBytes   = 1 << 20
	maxRedirects       = 5
	checksumsAssetName = "SHA256SUMS"
)

var (
	// ErrInvalidRepository means the manifest field is not github:owner/repo.
	ErrInvalidRepository = errors.New("repository must look like github:owner/repo")
	// ErrRateLimited means GitHub asked the launcher to wait before asking again.
	ErrRateLimited = errors.New("GitHub rate limit reached")
	// ErrUntrustedHost means a request or redirect pointed away from GitHub.
	ErrUntrustedHost = errors.New("refusing to contact a host that is not GitHub")
	// ErrResponseTooLarge means a GitHub reply was larger than the updater accepts.
	ErrResponseTooLarge = errors.New("GitHub response is too large")

	repositoryPattern = regexp.MustCompile(`^github:([A-Za-z0-9._-]+)/([A-Za-z0-9._-]+)$`)
)

// HTTPDoer is the part of http.Client the checker uses.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// Checker looks up the newest release of a plugin repository.
// It is safe for concurrent use.
type Checker struct {
	client           HTTPDoer
	now              func() time.Time
	allowPrereleases bool
	userAgent        string
	apiBase          string
	assetHosts       []string
	backoff          time.Duration

	mu           sync.Mutex
	cache        map[string]cachedRelease
	limitedUntil time.Time
}

type cachedRelease struct {
	etag string
	body []byte
}

// Options wires a Checker. Nil fields take safe defaults.
type Options struct {
	Client           HTTPDoer
	Now              func() time.Time
	AllowPrereleases bool
	UserAgent        string
	// APIBaseURL is for tests. It defaults to https://api.github.com.
	APIBaseURL string
	// AssetHosts lists extra hosts that downloads may come from. GitHub's own hosts are always allowed.
	AssetHosts []string
	// Backoff is how long a rate limit blocks further checks when GitHub gives no reset time.
	Backoff time.Duration
}

// Request names the installed plugin and the repository its manifest declares.
type Request struct {
	ID         string
	Repository string
	Version    string
}

// Result is the release the checker found, if it is newer than the installed version.
type Result struct {
	Available   bool   `json:"available"`
	Version     string `json:"version,omitempty"`
	AssetURL    string `json:"assetUrl,omitempty"`
	SHA256URL   string `json:"sha256Url,omitempty"`
	SHA256      string `json:"sha256,omitempty"`
	Notes       string `json:"notes,omitempty"`
	PublishedAt string `json:"publishedAt,omitempty"`
}

// RateLimitError reports when GitHub may be asked again.
type RateLimitError struct {
	Reset time.Time
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("GitHub rate limit reached, try again at %s", e.Reset.UTC().Format(time.RFC3339))
}

func (e *RateLimitError) Unwrap() error { return ErrRateLimited }

// NewChecker returns a checker. The default client has a timeout and never follows redirects itself;
// the checker follows them only to hosts it trusts.
func NewChecker(opts Options) *Checker {
	client := opts.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	if std, ok := client.(*http.Client); ok {
		// Redirects are followed by hand, so the checker can refuse a host that is not GitHub.
		clone := *std
		clone.CheckRedirect = func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}
		client = &clone
	}
	now := opts.Now
	if now == nil {
		now = time.Now
	}
	userAgent := opts.UserAgent
	if userAgent == "" {
		userAgent = defaultUserAgent
	}
	apiBase := strings.TrimRight(opts.APIBaseURL, "/")
	if apiBase == "" {
		apiBase = defaultAPIBase
	}
	backoff := opts.Backoff
	if backoff <= 0 {
		backoff = defaultBackoff
	}
	checker := &Checker{
		client:           client,
		now:              now,
		allowPrereleases: opts.AllowPrereleases,
		userAgent:        userAgent,
		apiBase:          apiBase,
		backoff:          backoff,
		cache:            map[string]cachedRelease{},
	}
	checker.assetHosts = defaultAssetHosts()
	if parsed, err := url.Parse(apiBase); err == nil && parsed.Hostname() != "" {
		checker.assetHosts = append(checker.assetHosts, parsed.Hostname())
	}
	for _, host := range opts.AssetHosts {
		checker.assetHosts = append(checker.assetHosts, strings.ToLower(host))
	}
	return checker
}

// defaultAssetHosts holds GitHub's release hosts. Release downloads redirect to the CDN hosts.
func defaultAssetHosts() []string {
	return []string{
		"api.github.com",
		"github.com",
		"objects.githubusercontent.com",
		"release-assets.githubusercontent.com",
	}
}

type release struct {
	TagName     string         `json:"tag_name"`
	Draft       bool           `json:"draft"`
	Prerelease  bool           `json:"prerelease"`
	Body        string         `json:"body"`
	PublishedAt string         `json:"published_at"`
	Assets      []releaseAsset `json:"assets"`
}

type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Check finds a newer release for req and picks its .llplugin asset.
func (c *Checker) Check(ctx context.Context, req Request) (Result, error) {
	owner, repo, err := parseRepository(req.Repository)
	if err != nil {
		return Result{}, err
	}
	installed, err := semver.Parse(req.Version)
	if err != nil {
		return Result{}, fmt.Errorf("installed version: %w", err)
	}
	latest, err := c.latestRelease(ctx, owner, repo)
	if err != nil {
		return Result{}, fmt.Errorf("check %s: %w", req.Repository, err)
	}
	versionText := strings.TrimPrefix(strings.TrimSpace(latest.TagName), "v")
	version, err := semver.Parse(versionText)
	if err != nil {
		return Result{}, fmt.Errorf("release tag %q is not a semantic version", latest.TagName)
	}
	if semver.Compare(version, installed) <= 0 {
		return Result{}, nil
	}
	asset, found, err := pickAsset(req.ID, versionText, latest.Assets)
	if err != nil {
		return Result{}, fmt.Errorf("check %s: %w", req.Repository, err)
	}
	if !found {
		return Result{Version: versionText}, nil
	}
	if err := c.checkURL(asset.BrowserDownloadURL); err != nil {
		return Result{}, err
	}
	result := Result{
		Available:   true,
		Version:     versionText,
		AssetURL:    asset.BrowserDownloadURL,
		Notes:       latest.Body,
		PublishedAt: latest.PublishedAt,
	}
	digest, checksumURL, err := c.releaseChecksum(ctx, latest.Assets, asset.Name)
	if err != nil {
		return Result{}, fmt.Errorf("check %s: %w", req.Repository, err)
	}
	result.SHA256URL = checksumURL
	result.SHA256 = digest
	return result, nil
}

func parseRepository(repository string) (string, string, error) {
	match := repositoryPattern.FindStringSubmatch(repository)
	if match == nil {
		return "", "", ErrInvalidRepository
	}
	return match[1], match[2], nil
}

// latestRelease asks the latest endpoint, or the release list when pre-releases are allowed.
func (c *Checker) latestRelease(ctx context.Context, owner, repo string) (release, error) {
	if c.allowPrereleases {
		releases, err := c.releaseList(ctx, owner, repo)
		if err != nil {
			return release{}, err
		}
		for _, candidate := range releases {
			if !candidate.Draft {
				return candidate, nil
			}
		}
		return release{}, errors.New("the repository has no published release")
	}
	var latest release
	target := fmt.Sprintf("%s/repos/%s/%s/releases/latest", c.apiBase, owner, repo)
	if err := c.getJSON(ctx, target, &latest); err != nil {
		return release{}, err
	}
	if latest.Draft {
		return release{}, errors.New("the latest release is a draft")
	}
	return latest, nil
}

func (c *Checker) releaseList(ctx context.Context, owner, repo string) ([]release, error) {
	var releases []release
	target := fmt.Sprintf("%s/repos/%s/%s/releases?per_page=30", c.apiBase, owner, repo)
	if err := c.getJSON(ctx, target, &releases); err != nil {
		return nil, err
	}
	return releases, nil
}

// getJSON makes a conditional API request and reuses the cached body on 304.
func (c *Checker) getJSON(ctx context.Context, target string, out any) error {
	c.mu.Lock()
	cached, hasCached := c.cache[target]
	c.mu.Unlock()
	etag := ""
	if hasCached {
		etag = cached.etag
	}
	body, etag, notModified, err := c.get(ctx, target, etag, maxResponseBytes)
	if err != nil {
		return err
	}
	if notModified {
		if !hasCached {
			return errors.New("GitHub answered 304 without a cached reply")
		}
		body = cached.body
	} else {
		c.mu.Lock()
		if len(c.cache) >= 128 {
			c.cache = map[string]cachedRelease{}
		}
		c.cache[target] = cachedRelease{etag: etag, body: body}
		c.mu.Unlock()
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("GitHub sent an unreadable reply: %w", err)
	}
	return nil
}

// get fetches one URL, following redirects only to trusted hosts.
func (c *Checker) get(ctx context.Context, rawURL, etag string, maxBytes int64) ([]byte, string, bool, error) {
	if err := c.checkURL(rawURL); err != nil {
		return nil, "", false, err
	}
	if err := c.waitBackoff(); err != nil {
		return nil, "", false, err
	}
	current := rawURL
	for redirects := 0; ; redirects++ {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, current, nil)
		if err != nil {
			return nil, "", false, err
		}
		request.Header.Set("User-Agent", c.userAgent)
		request.Header.Set("Accept", "application/vnd.github+json")
		if etag != "" {
			request.Header.Set("If-None-Match", etag)
		}
		response, err := c.client.Do(request)
		if err != nil {
			return nil, "", false, fmt.Errorf("request GitHub: %w", err)
		}
		if isRedirect(response.StatusCode) {
			location := response.Header.Get("Location")
			response.Body.Close()
			if redirects >= maxRedirects {
				return nil, "", false, errors.New("too many redirects")
			}
			next, err := response.Request.URL.Parse(location)
			if err != nil {
				return nil, "", false, fmt.Errorf("bad redirect from GitHub: %w", err)
			}
			if err := c.checkURL(next.String()); err != nil {
				return nil, "", false, err
			}
			current = next.String()
			continue
		}
		return c.readResponse(response, maxBytes)
	}
}

func (c *Checker) readResponse(response *http.Response, maxBytes int64) ([]byte, string, bool, error) {
	defer response.Body.Close()
	c.noteRateLimit(response)
	switch {
	case response.StatusCode == http.StatusNotModified:
		return nil, "", true, nil
	case response.StatusCode == http.StatusForbidden || response.StatusCode == http.StatusTooManyRequests:
		return nil, "", false, c.rateLimitError(response)
	case response.StatusCode != http.StatusOK:
		return nil, "", false, fmt.Errorf("GitHub returned HTTP %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxBytes+1))
	if err != nil {
		return nil, "", false, fmt.Errorf("read GitHub reply: %w", err)
	}
	if int64(len(body)) > maxBytes {
		return nil, "", false, ErrResponseTooLarge
	}
	return body, response.Header.Get("ETag"), false, nil
}

func isRedirect(status int) bool {
	switch status {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		return true
	default:
		return false
	}
}

func (c *Checker) checkURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" {
		return fmt.Errorf("%w: %s", ErrUntrustedHost, raw)
	}
	if !c.hostAllowed(parsed.Hostname()) {
		return fmt.Errorf("%w: %s", ErrUntrustedHost, parsed.Hostname())
	}
	return nil
}

func (c *Checker) hostAllowed(host string) bool {
	host = strings.ToLower(host)
	for _, allowed := range c.assetHosts {
		if host == allowed {
			return true
		}
	}
	return false
}

// waitBackoff refuses a check while a rate limit is active.
func (c *Checker) waitBackoff() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if until := c.limitedUntil; !until.IsZero() && c.now().Before(until) {
		return &RateLimitError{Reset: until}
	}
	return nil
}

// noteRateLimit starts a backoff when GitHub says no requests are left.
func (c *Checker) noteRateLimit(response *http.Response) {
	if response.Header.Get("X-RateLimit-Remaining") != "0" {
		return
	}
	reset := c.resetTime(response)
	c.mu.Lock()
	c.limitedUntil = reset
	c.mu.Unlock()
}

func (c *Checker) rateLimitError(response *http.Response) error {
	reset := c.resetTime(response)
	c.mu.Lock()
	c.limitedUntil = reset
	c.mu.Unlock()
	return &RateLimitError{Reset: reset}
}

func (c *Checker) resetTime(response *http.Response) time.Time {
	now := c.now()
	if raw := response.Header.Get("Retry-After"); raw != "" {
		if seconds, err := strconv.ParseInt(raw, 10, 64); err == nil && seconds > 0 {
			return now.Add(min(time.Duration(seconds)*time.Second, maxBackoff))
		}
	}
	if raw := response.Header.Get("X-RateLimit-Reset"); raw != "" {
		if unix, err := strconv.ParseInt(raw, 10, 64); err == nil {
			reset := time.Unix(unix, 0)
			if reset.After(now) {
				return minTime(reset, now.Add(maxBackoff))
			}
		}
	}
	return now.Add(c.backoff)
}

func minTime(left, right time.Time) time.Time {
	if left.Before(right) {
		return left
	}
	return right
}

// pickAsset finds the plugin package. With several candidates, the one named after the plugin wins.
func pickAsset(id, version string, assets []releaseAsset) (releaseAsset, bool, error) {
	var candidates []releaseAsset
	for _, asset := range assets {
		if strings.HasSuffix(strings.ToLower(asset.Name), ".llplugin") {
			candidates = append(candidates, asset)
		}
	}
	switch len(candidates) {
	case 0:
		return releaseAsset{}, false, nil
	case 1:
		return candidates[0], true, nil
	}
	expected := strings.ToLower(id + "-" + version + ".llplugin")
	for _, candidate := range candidates {
		if strings.ToLower(candidate.Name) == expected {
			return candidate, true, nil
		}
	}
	prefix := strings.ToLower(id) + "-"
	for _, candidate := range candidates {
		if strings.HasPrefix(strings.ToLower(candidate.Name), prefix) {
			return candidate, true, nil
		}
	}
	return releaseAsset{}, false, fmt.Errorf("release has several .llplugin assets and none is named after %s", id)
}

// releaseChecksum reads the optional SHA256SUMS asset. A digest is returned only when the file lists the package.
func (c *Checker) releaseChecksum(ctx context.Context, assets []releaseAsset, assetName string) (string, string, error) {
	for _, asset := range assets {
		if !strings.EqualFold(asset.Name, checksumsAssetName) {
			continue
		}
		if err := c.checkURL(asset.BrowserDownloadURL); err != nil {
			return "", "", err
		}
		body, _, _, err := c.get(ctx, asset.BrowserDownloadURL, "", maxChecksumBytes)
		if err != nil {
			return "", asset.BrowserDownloadURL, err
		}
		digest, err := checksumFor(body, assetName)
		if err != nil {
			return "", asset.BrowserDownloadURL, err
		}
		return digest, asset.BrowserDownloadURL, nil
	}
	return "", "", nil
}

// checksumFor reads a sha256sum-style line for name.
func checksumFor(data []byte, name string) (string, error) {
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		digest, listed, ok := strings.Cut(line, "  ")
		if !ok {
			digest, listed, ok = strings.Cut(line, " *")
		}
		if !ok || listed != name {
			continue
		}
		if len(digest) != 64 {
			return "", fmt.Errorf("%s has a bad digest for %s", checksumsAssetName, name)
		}
		if _, err := hex.DecodeString(digest); err != nil {
			return "", fmt.Errorf("%s has a bad digest for %s", checksumsAssetName, name)
		}
		return strings.ToLower(digest), nil
	}
	return "", fmt.Errorf("%s does not list %s", checksumsAssetName, name)
}
