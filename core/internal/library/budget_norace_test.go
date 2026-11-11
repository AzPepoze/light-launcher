//go:build !race

package library

import "testing"

// skipUnderRace runs a timing or memory budget. Without the race detector there is nothing to skip.
func skipUnderRace(testing.TB) {}
