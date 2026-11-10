package logx

import "testing"

func TestCategoryColorIsStable(t *testing.T) {
	first := categoryColor("window")
	second := categoryColor("window")
	if first != second {
		t.Fatalf("category color changed between calls: %q and %q", first, second)
	}
}

func TestCategoryColorsDiffer(t *testing.T) {
	if categoryColor("window") == categoryColor("bridge") {
		t.Fatal("expected different categories to get different colors")
	}
}

func TestHslToRGBIsPastel(t *testing.T) {
	r, g, b := hslToRGB(0, categorySaturation, categoryLightness)
	if r < 200 || g < 100 || b < 100 {
		t.Fatalf("expected a light pastel red, got %d %d %d", r, g, b)
	}
}
