package library

import (
	"context"
	"runtime"
	"sort"
	"testing"
	"time"
)

// Budgets from the library design. They are checked without the race detector, which slows everything down.
const (
	importBudgetGames = 5000
	importBudget      = 2 * time.Second
	listBudgetRows    = 50000
	listBudget        = 50 * time.Millisecond
	searchBudget      = 80 * time.Millisecond
	// importMemoryGrowth is how much the heap may grow during an import of importBudgetGames games.
	importMemoryGrowth = 8 << 20
	budgetRuns         = 3
	listBudgetRuns     = 5
)

// seedLibrary imports n fixture games through the same path a plugin sync uses.
func seedLibrary(tb testing.TB, n int) *Library {
	tb.Helper()
	lib := openTestLibrary(tb, Options{Sources: feedSources(listGamesOnly(n))})
	result, err := lib.ImportSource(context.Background(), "feed", feedSource, nil)
	if err != nil {
		tb.Fatal(err)
	}
	if result.Imported != n {
		tb.Fatalf("seeded %d games, want %d", result.Imported, n)
	}
	return lib
}

// medianDuration returns the middle value of the durations.
func medianDuration(durations []time.Duration) time.Duration {
	sorted := append([]time.Duration(nil), durations...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	return sorted[len(sorted)/2]
}

// heapInUse returns the live heap after a collection, so the reading is not skewed by garbage.
func heapInUse() uint64 {
	runtime.GC()
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return stats.HeapAlloc
}

func TestImportBudget(t *testing.T) {
	skipUnderRace(t)
	durations := make([]time.Duration, 0, budgetRuns)
	for range budgetRuns {
		lib := openTestLibrary(t, Options{Sources: feedSources(listGamesOnly(importBudgetGames))})
		start := time.Now()
		result, err := lib.ImportSource(context.Background(), "feed", feedSource, nil)
		durations = append(durations, time.Since(start))
		if err != nil || result.Imported != importBudgetGames {
			t.Fatalf("import = %+v, %v", result, err)
		}
	}
	median := medianDuration(durations)
	t.Logf("import of %d games: median %v of %d runs (budget %v)", importBudgetGames, median, budgetRuns, importBudget)
	if median > importBudget {
		t.Errorf("import median %v exceeds the %v budget", median, importBudget)
	}
}

func TestImportMemoryStaysFlat(t *testing.T) {
	skipUnderRace(t)
	lib := openTestLibrary(t, Options{Sources: feedSources(listGamesOnly(importBudgetGames))})
	var before, after uint64
	result, err := lib.ImportSource(context.Background(), "feed", feedSource, func(p ImportProgress) {
		switch p.Imported {
		case 500:
			before = heapInUse()
		case importBudgetGames:
			after = heapInUse()
		}
	})
	if err != nil || result.Imported != importBudgetGames {
		t.Fatalf("import = %+v, %v", result, err)
	}
	growth := int64(after) - int64(before)
	t.Logf("heap after 5 pages %d bytes, after all pages %d bytes, growth %d bytes (limit %d)",
		before, after, growth, importMemoryGrowth)
	if growth > importMemoryGrowth {
		t.Errorf("heap grew %d bytes during the import, limit %d", growth, importMemoryGrowth)
	}
}

func TestListBudgets(t *testing.T) {
	skipUnderRace(t)
	lib := seedLibrary(t, listBudgetRows)
	ctx := context.Background()

	cases := []struct {
		name   string
		filter Filter
		budget time.Duration
	}{
		{"first page of 60", Filter{Limit: 60}, listBudget},
		{"title search", Filter{Search: "harbor", Limit: 60}, searchBudget},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			durations := make([]time.Duration, 0, listBudgetRuns)
			for range listBudgetRuns {
				start := time.Now()
				page, err := lib.List(ctx, tc.filter)
				durations = append(durations, time.Since(start))
				if err != nil || len(page.Items) == 0 {
					t.Fatalf("list = %d items, %v", len(page.Items), err)
				}
			}
			median := medianDuration(durations)
			t.Logf("%s on %d rows: median %v of %d runs (budget %v)", tc.name, listBudgetRows, median, listBudgetRuns, tc.budget)
			if median > tc.budget {
				t.Errorf("%s median %v exceeds the %v budget", tc.name, median, tc.budget)
			}
		})
	}
}

func BenchmarkImport5000(b *testing.B) {
	for range b.N {
		b.StopTimer()
		lib := openTestLibrary(b, Options{Sources: feedSources(listGamesOnly(importBudgetGames))})
		b.StartTimer()
		if _, err := lib.ImportSource(context.Background(), "feed", feedSource, nil); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkListFirstPage50k(b *testing.B) {
	lib := seedLibrary(b, listBudgetRows)
	b.ResetTimer()
	for range b.N {
		if _, err := lib.List(context.Background(), Filter{Limit: 60}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTitleSearch50k(b *testing.B) {
	lib := seedLibrary(b, listBudgetRows)
	b.ResetTimer()
	for range b.N {
		if _, err := lib.List(context.Background(), Filter{Search: "harbor", Limit: 60}); err != nil {
			b.Fatal(err)
		}
	}
}
