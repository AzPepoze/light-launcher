package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"
)

const (
	defaultHTTPTimeout = 30 * time.Second
	maxHTTPTimeout     = 2 * time.Minute
	defaultHTTPBytes   = 5 << 20
	maxHTTPBytes       = 20 << 20
	maxHTTPRetries     = 3
	maxHTTPRedirects   = 5
	retryStep          = 500 * time.Millisecond
	pluginUserAgent    = "light-launcher-plugin"
)

var (
	errRedirectNotAllowed = errors.New("redirect to a host the manifest does not allow")
	errTooManyRedirects   = errors.New("too many redirects")
	errResponseTooLarge   = errors.New("response is too large")
)

type httpRequestParams struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body"`
	TimeoutMs   int               `json:"timeoutMs"`
	Retries     int               `json:"retries"`
	CacheKey    string            `json:"cacheKey"`
	MaxBytes    int64             `json:"maxBytes"`
	AllowStatus []int             `json:"allowStatus"`
}

type httpRequestResult struct {
	Status      int               `json:"status"`
	URL         string            `json:"url"`
	Text        string            `json:"text"`
	Headers     map[string]string `json:"headers"`
	NotModified bool              `json:"notModified"`
}

// httpPlan holds the request settings after defaults and limits were applied.
type httpPlan struct {
	method   string
	timeout  time.Duration
	maxBytes int64
	attempts int
	cacheKey string
}

// serviceHTTPRequest fetches a page for the plugin. The host must be in the network grant.
func serviceHTTPRequest(ctx context.Context, env *pluginEnv, params json.RawMessage) (any, error) {
	req, err := decodeParams[httpRequestParams](params)
	if err != nil {
		return nil, err
	}
	target, err := parseHTTPTarget(req.URL)
	if err != nil {
		return nil, err
	}
	if !env.manifest.Permissions.allowsHost(target.Hostname()) {
		return nil, hostDenied(target.Hostname())
	}
	plan, err := planHTTP(req)
	if err != nil {
		return nil, err
	}
	plan.cacheKey = env.cacheID(req.CacheKey)
	return env.fetchWithRetries(ctx, target, req, plan)
}

func parseHTTPTarget(raw string) (*url.URL, error) {
	target, err := url.Parse(raw)
	if err != nil || (target.Scheme != "https" && target.Scheme != "http") || target.Hostname() == "" {
		return nil, invalidParams("url must be an http or https address")
	}
	return target, nil
}

func hostDenied(host string) error {
	return &RPCError{Code: CodePermissionDenied, Message: "host is not in the network permission: " + host}
}

func planHTTP(req httpRequestParams) (httpPlan, error) {
	method := strings.ToUpper(req.Method)
	if method == "" {
		method = http.MethodGet
	}
	if !slices.Contains([]string{http.MethodGet, http.MethodHead, http.MethodPost}, method) {
		return httpPlan{}, invalidParams("method must be GET, HEAD or POST")
	}
	timeout := defaultHTTPTimeout
	if req.TimeoutMs > 0 {
		timeout = min(time.Duration(req.TimeoutMs)*time.Millisecond, maxHTTPTimeout)
	}
	maxBytes := req.MaxBytes
	if maxBytes <= 0 {
		maxBytes = defaultHTTPBytes
	}
	return httpPlan{
		method:   method,
		timeout:  timeout,
		maxBytes: min(maxBytes, maxHTTPBytes),
		attempts: 1 + min(max(req.Retries, 0), maxHTTPRetries),
	}, nil
}

// cacheID scopes a plugin's cache key to that plugin. An empty key means no caching.
func (env *pluginEnv) cacheID(key string) string {
	if key == "" {
		return ""
	}
	return env.id + "\x00" + key
}

func (env *pluginEnv) fetchWithRetries(ctx context.Context, target *url.URL, req httpRequestParams, plan httpPlan) (httpRequestResult, error) {
	var lastErr error
	for attempt := 0; attempt < plan.attempts; attempt++ {
		if attempt > 0 {
			if err := sleepContext(ctx, time.Duration(attempt)*retryStep); err != nil {
				return httpRequestResult{}, contextError(err)
			}
		}
		result, retry, err := env.fetchOnce(ctx, target, req, plan)
		if err == nil {
			return result, nil
		}
		if !retry {
			return httpRequestResult{}, err
		}
		lastErr = err
	}
	return httpRequestResult{}, lastErr
}

// fetchOnce makes one attempt. The bool result says whether a failed attempt may be retried.
func (env *pluginEnv) fetchOnce(ctx context.Context, target *url.URL, req httpRequestParams, plan httpPlan) (httpRequestResult, bool, error) {
	if err := env.deps.limiter.wait(ctx, target.Hostname()); err != nil {
		return httpRequestResult{}, false, contextError(err)
	}
	request, err := newHTTPRequest(ctx, target, req, plan)
	if err != nil {
		return httpRequestResult{}, false, internalError(err.Error())
	}
	etag, hasETag := env.deps.cache.get(plan.cacheKey)
	if hasETag {
		request.Header.Set("If-None-Match", etag)
	}
	response, err := env.httpClient(plan.timeout).Do(request)
	if err != nil {
		retry, classified := classifyDoError(ctx, err)
		return httpRequestResult{}, retry, classified
	}
	defer response.Body.Close()
	return env.readResponse(response, req, plan, hasETag)
}

func newHTTPRequest(ctx context.Context, target *url.URL, req httpRequestParams, plan httpPlan) (*http.Request, error) {
	var body io.Reader
	if req.Body != "" {
		body = strings.NewReader(req.Body)
	}
	request, err := http.NewRequestWithContext(ctx, plan.method, target.String(), body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", pluginUserAgent)
	for name, value := range req.Headers {
		request.Header.Set(name, value)
	}
	return request, nil
}

// httpClient sends redirects only to hosts the manifest allows.
func (env *pluginEnv) httpClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: env.deps.transport,
		Timeout:   timeout,
		CheckRedirect: func(request *http.Request, via []*http.Request) error {
			if len(via) >= maxHTTPRedirects {
				return errTooManyRedirects
			}
			if !env.manifest.Permissions.allowsHost(request.URL.Hostname()) {
				return errRedirectNotAllowed
			}
			return nil
		},
	}
}

func classifyDoError(ctx context.Context, err error) (bool, error) {
	if ctx.Err() != nil {
		return false, contextError(ctx.Err())
	}
	if errors.Is(err, errRedirectNotAllowed) {
		return false, &RPCError{Code: CodePermissionDenied, Message: errRedirectNotAllowed.Error()}
	}
	if errors.Is(err, errTooManyRedirects) {
		return false, &RPCError{Code: CodeRemoteError, Message: errTooManyRedirects.Error()}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true, &RPCError{Code: CodeTimeout, Message: "request timed out"}
	}
	return true, &RPCError{Code: CodeRemoteError, Message: "request failed: " + err.Error()}
}

// readResponse turns an HTTP answer into the protocol result. It keeps the ETag for later requests.
func (env *pluginEnv) readResponse(response *http.Response, req httpRequestParams, plan httpPlan, hasETag bool) (httpRequestResult, bool, error) {
	result := httpRequestResult{
		Status:  response.StatusCode,
		URL:     response.Request.URL.String(),
		Headers: flattenHeaders(response.Header),
	}
	if response.StatusCode == http.StatusNotModified && hasETag {
		result.NotModified = true
		return result, false, nil
	}
	if retry, err := statusError(response.StatusCode, req.AllowStatus); err != nil {
		return httpRequestResult{}, retry, err
	}
	text, err := readLimited(response.Body, plan.maxBytes)
	if errors.Is(err, errResponseTooLarge) {
		return httpRequestResult{}, false, &RPCError{Code: CodeRemoteError, Message: fmt.Sprintf("response is larger than %d bytes", plan.maxBytes)}
	}
	if err != nil {
		return httpRequestResult{}, false, &RPCError{Code: CodeRemoteError, Message: "reading response failed: " + err.Error()}
	}
	result.Text = text
	if response.StatusCode == http.StatusOK && plan.cacheKey != "" {
		env.deps.cache.put(plan.cacheKey, response.Header.Get("ETag"))
	}
	return result, false, nil
}

// statusError maps an HTTP error status to an RPCError. The bool says whether a retry may help.
func statusError(status int, allowed []int) (bool, error) {
	if status < http.StatusBadRequest || slices.Contains(allowed, status) {
		return false, nil
	}
	switch {
	case status == http.StatusTooManyRequests:
		return true, &RPCError{Code: CodeRateLimited, Message: "site rate limited the request (HTTP 429)"}
	case status >= http.StatusInternalServerError:
		return true, &RPCError{Code: CodeRemoteError, Message: fmt.Sprintf("site error (HTTP %d)", status)}
	default:
		return false, &RPCError{Code: CodeRemoteError, Message: fmt.Sprintf("site returned HTTP %d", status)}
	}
}

func readLimited(body io.Reader, limit int64) (string, error) {
	data, err := io.ReadAll(io.LimitReader(body, limit+1))
	if err != nil {
		return "", err
	}
	if int64(len(data)) > limit {
		return "", errResponseTooLarge
	}
	return string(data), nil
}

// flattenHeaders keeps the first value of each header, with lower-case names.
func flattenHeaders(header http.Header) map[string]string {
	flat := make(map[string]string, len(header))
	for name, values := range header {
		if len(values) > 0 {
			flat[strings.ToLower(name)] = values[0]
		}
	}
	return flat
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
