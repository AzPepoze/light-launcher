//go:build race

package library

import "testing"

// skipUnderRace skips a timing or memory budget, which the race detector would distort.
func skipUnderRace(t testing.TB) {
	t.Skip("budgets are measured without the race detector")
}
