package library

import "testing"

// Ported from tests/title-match.test.ts.

func TestCompactTitle(t *testing.T) {
	if got := compactTitle("KelpieRays v1.9"); got != "kelpierays" {
		t.Errorf("compactTitle(KelpieRays v1.9) = %q", got)
	}
	if got := compactTitle("Kel Pie Rays [v2.2.1]"); got != "kelpierays" {
		t.Errorf("compactTitle(Kel Pie Rays [v2.2.1]) = %q", got)
	}
	if got := normalizeTitle("Kel Pie Rays [v2.2.1]"); got != "kel pie rays" {
		t.Errorf("normalizeTitle(Kel Pie Rays [v2.2.1]) = %q", got)
	}
}

func TestScoreMatchFindsSplitTitle(t *testing.T) {
	real := scoreMatch("KelpieRays", "Kel Pie Rays V2.2.1 Studio")
	discussion := scoreMatch("KelpieRays", "Games Similar to Frost Lantern 2 Kelpierays")
	if real <= 0.35 {
		t.Errorf("real title score %.2f, want > 0.35", real)
	}
	if real <= discussion {
		t.Errorf("real title %.2f should beat discussion %.2f", real, discussion)
	}
}

func TestScoreMatchKeepsUnrelatedBelowCutoff(t *testing.T) {
	if got := scoreMatch("DriftRays", "Moonlit Harbor V1.1 Studio"); got >= 0.35 {
		t.Errorf("unrelated score %.2f, want < 0.35", got)
	}
}
