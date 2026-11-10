package plugins

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
)

var (
	// ErrConsent means an install did not list exactly the permissions the plugin asks for.
	ErrConsent = errors.New("acceptedPermissions must list exactly the permissions the plugin asks for")
	// ErrNewPermissions means an update asks for permissions the user has not accepted.
	ErrNewPermissions = errors.New("the update asks for new permissions, accept them to update")
)

// PermissionSummary lists the permissions a manifest asks for, sorted. Entries look like
// "network:example.com", "fs:$GAMES", "secrets", "browser" or "exec". The UI sends them back as acceptedPermissions.
func PermissionSummary(p Permissions) []string {
	summary := []string{}
	for _, host := range p.Network {
		summary = append(summary, "network:"+host)
	}
	for _, folder := range p.FS {
		summary = append(summary, "fs:"+folder)
	}
	if p.Secrets {
		summary = append(summary, "secrets")
	}
	if p.Browser {
		summary = append(summary, "browser")
	}
	if p.Exec {
		summary = append(summary, "exec")
	}
	sort.Strings(summary)
	return summary
}

// checkConsent accepts the request only when accepted holds exactly the summary, in any order.
func checkConsent(summary, accepted []string) error {
	if slices.Equal(sortedCopy(summary), sortedCopy(accepted)) {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrConsent, strings.Join(summary, ", "))
}

// newPermissions returns the entries of next that previous does not have.
func newPermissions(previous, next []string) []string {
	var added []string
	for _, entry := range next {
		if !slices.Contains(previous, entry) {
			added = append(added, entry)
		}
	}
	return added
}

func sortedCopy(values []string) []string {
	out := slices.Clone(values)
	sort.Strings(out)
	return out
}
