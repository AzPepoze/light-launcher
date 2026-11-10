package library

import "testing"

// Ported from tests/watch-match.test.ts. Titles are passed as slices, and the
// index of the matched title stands in for the game object.

const harborTitle = "Moonlit Harbor [v1.1] [Studio]"

func matchIndex(t *testing.T, fileName string, titles []string) (int, bool) {
	t.Helper()
	result, ok := bestConfidentMatch(fileName, titles)
	return result.Index, ok
}

func TestNormalizeName(t *testing.T) {
	if got := normalizeName("Moonlit Harbor [v1.1] [Studio]"); got != "moonlitharbor" {
		t.Errorf("normalizeName(harbor) = %q", got)
	}
	if got := normalizeName("Game (Final) Edition!"); got != "gameedition" {
		t.Errorf("normalizeName(edition) = %q", got)
	}
}

func TestMatchesArchiveNamedAfterGame(t *testing.T) {
	if idx, ok := matchIndex(t, "Moonlit-Harbor-v1.1-PC.zip", []string{harborTitle}); !ok || idx != 0 {
		t.Errorf("want harbor match, got %d %v", idx, ok)
	}
}

func TestUnrelatedFileMatchesNothing(t *testing.T) {
	if _, ok := matchIndex(t, "readme.txt", []string{harborTitle}); ok {
		t.Error("readme.txt should not match")
	}
}

func TestPrefersLongestMatchingTitle(t *testing.T) {
	titles := []string{"Moonlit Harbor", "Moonlit Harbor Deluxe"}
	if idx, ok := matchIndex(t, "Moonlit Harbor Deluxe v2.zip", titles); !ok || idx != 1 {
		t.Errorf("want Deluxe (1), got %d %v", idx, ok)
	}
}

func TestIgnoresTitlesShorterThanFourCharacters(t *testing.T) {
	if _, ok := matchIndex(t, "Go-Game-v1.zip", []string{"Go!"}); ok {
		t.Error("tiny title should not match")
	}
}

func TestMatchesSmallTypo(t *testing.T) {
	titles := []string{"Cobalt Tide [v0.16] [Studio]"}
	if idx, ok := matchIndex(t, "Cobalt_Tyde_0.16.zip", titles); !ok || idx != 0 {
		t.Errorf("want typo match, got %d %v", idx, ok)
	}
}

func TestMatchesMainTitleOnly(t *testing.T) {
	titles := []string{"Ember Vow: The Long Road of Ash and Silver [v0.129] [Studio]"}
	if idx, ok := matchIndex(t, "Ember_Vow_v0129.zip", titles); !ok || idx != 0 {
		t.Errorf("want main title match, got %d %v", idx, ok)
	}
}

func TestIgnoresWinExtensionAndVersions(t *testing.T) {
	titles := []string{"Glass Orchard [v0.4.2] [Studio]"}
	if idx, ok := matchIndex(t, "Glass Orchard 0.3.2 WIN.zip", titles); !ok || idx != 0 {
		t.Errorf("want orchard match, got %d %v", idx, ok)
	}
}

func TestKeepsSequelsApart(t *testing.T) {
	titles := []string{"Game 2", "Game 3"}
	if idx, ok := matchIndex(t, "Game 3 v1.0.zip", titles); !ok || idx != 1 {
		t.Errorf("want Game 3 (1), got %d %v", idx, ok)
	}
	if _, ok := matchIndex(t, "Game 3 v1.0.zip", []string{"Game 2"}); ok {
		t.Error("Game 3 file must not match Game 2 alone")
	}
}

func TestRanksSuggestionsWhenNotObvious(t *testing.T) {
	titles := []string{"Moonlit Harbor", "Moonlit Harbor Deluxe"}
	hits := rankTitles("Moonlit Harbor Deluxe v2.zip", titles, 2)
	if len(hits) < 2 {
		t.Fatalf("want 2 suggestions, got %d", len(hits))
	}
	if hits[0].Index != 1 {
		t.Errorf("best suggestion index = %d, want 1", hits[0].Index)
	}
	if hits[0].Score <= hits[1].Score {
		t.Errorf("best score %.2f should exceed runner-up %.2f", hits[0].Score, hits[1].Score)
	}
}

func TestNeverAutoImportsPatchDemoOrMod(t *testing.T) {
	titles := []string{"Moonlit Harbor"}
	if _, ok := matchIndex(t, "Moonlit Harbor patch v1.zip", titles); ok {
		t.Error("patch must not auto-import")
	}
	if ranked := rankTitles("Moonlit Harbor patch v1.zip", titles, 5); len(ranked) == 0 || ranked[0].Index != 0 {
		t.Error("patch should still be suggested for the game")
	}
	if _, ok := matchIndex(t, "Moonlit Harbor demo.zip", titles); ok {
		t.Error("demo must not auto-import")
	}
}

func TestNeverAutoImportsUpdateDlcOrSoundtrack(t *testing.T) {
	titles := []string{"Paper Lantern"}
	for _, file := range []string{"Paper Lantern Update.zip", "Paper Lantern DLC.zip", "Paper Lantern Soundtrack.zip"} {
		if _, ok := matchIndex(t, file, titles); ok {
			t.Errorf("%s must not auto-import", file)
		}
	}
	if ranked := rankTitles("Paper Lantern Update.zip", titles, 5); len(ranked) == 0 || ranked[0].Index != 0 {
		t.Error("update should still be suggested for the game")
	}
}
