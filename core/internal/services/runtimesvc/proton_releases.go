package runtimesvc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// getProtonReleases returns the GitHub release list for a variant as-is, so the UI sees the same fields.
func getProtonReleases(variantID string) (json.RawMessage, error) {
	variant, ok := findVariant(variantID)
	if !ok {
		return nil, fmt.Errorf("Unknown variant: %s", variantID)
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?per_page=50", variant.RepoOwner, variant.RepoName)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", userAgent)
	request.Header.Set("Accept", "application/vnd.github.v3+json")

	response, err := releaseClient.Do(request)
	if err != nil {
		return nil, requestError(err)
	}
	defer response.Body.Close()

	if err := checkReleaseStatus(response.StatusCode); err != nil {
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, requestError(err)
	}
	if !json.Valid(body) {
		return nil, errors.New("Failed to parse GitHub releases response")
	}
	return json.RawMessage(body), nil
}

func checkReleaseStatus(status int) error {
	if status == http.StatusForbidden || status == http.StatusTooManyRequests {
		return errors.New("GitHub API rate limit exceeded. Please wait a few minutes before trying again.")
	}
	if status != http.StatusOK {
		return fmt.Errorf("GitHub API returned HTTP %d", status)
	}
	return nil
}
