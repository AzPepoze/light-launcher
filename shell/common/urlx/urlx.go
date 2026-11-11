// Package urlx holds the URL rules shared by the shells.
package urlx

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// EnvURL names the environment variable that holds the launcher page.
const EnvURL = "LL_URL"

// ShellURL returns the page URL from the launcher environment value, or from the
// single command argument when the value is empty.
func ShellURL(envURL string, args []string) (string, error) {
	if envURL != "" {
		return envURL, nil
	}
	if len(args) != 1 {
		return "", fmt.Errorf("expected one URL, got %d arguments", len(args))
	}
	return args[0], nil
}

// ParseTarget validates a page URL. url.Parse errors repeat the input, which can
// hold a token, so the error is replaced.
func ParseTarget(raw string) (*url.URL, error) {
	target, err := url.Parse(raw)
	if err != nil {
		return nil, errors.New("URL is not valid")
	}
	if !IsWebURL(target) {
		return nil, errors.New("URL must be http:// or https:// with a host")
	}
	return target, nil
}

// IsWebURL reports whether target is an http or https address with a host.
func IsWebURL(target *url.URL) bool {
	return target != nil && (target.Scheme == "http" || target.Scheme == "https") && target.Host != ""
}

// IsWebAddress applies IsWebURL to a raw address, such as a download or navigation URI.
func IsWebAddress(raw string) bool {
	target, err := url.Parse(raw)
	return err == nil && IsWebURL(target)
}

// Origin returns scheme and authority (host and port) of a parsed URL.
func Origin(target *url.URL) string {
	if target == nil {
		return ""
	}
	return strings.ToLower(target.Scheme + "://" + target.Host)
}

// IsExternal reports whether a URL is outside the launcher origin.
func IsExternal(target *url.URL, launcherOrigin string) bool {
	return Origin(target) != launcherOrigin
}
