package plugins

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

type listGamesRequest struct {
	Source string `json:"source"`
	Scope  string `json:"scope"`
	Cursor string `json:"cursor,omitempty"`
	Limit  int    `json:"limit"`
}

type listGamesAnswer struct {
	Games []struct {
		Source   string `json:"source"`
		SourceID string `json:"sourceId"`
		URL      string `json:"url"`
		Title    string `json:"title"`
	} `json:"games"`
	NextCursor string `json:"nextCursor"`
	Total      int    `json:"total"`
}

func callListGames(t *testing.T, sup *Supervisor, request listGamesRequest) (listGamesAnswer, error) {
	t.Helper()
	raw, err := callFixture(sup, "source.listGames", request)
	if err != nil {
		return listGamesAnswer{}, err
	}
	var answer listGamesAnswer
	if err := json.Unmarshal(raw, &answer); err != nil {
		t.Fatalf("decode listGames answer %s: %v", raw, err)
	}
	return answer, nil
}

func TestListGamesPagesThroughEveryGame(t *testing.T) {
	sup, _, _ := newFixtureSupervisor(t, Permissions{})
	var pages []int
	var ids []string
	cursor := ""
	for {
		answer, err := callListGames(t, sup, listGamesRequest{Source: "fixture", Scope: "mine", Cursor: cursor, Limit: 100})
		if err != nil {
			t.Fatalf("page after cursor %q: %v", cursor, err)
		}
		pages = append(pages, len(answer.Games))
		if answer.Total != 250 {
			t.Fatalf("total = %d, want 250", answer.Total)
		}
		for _, game := range answer.Games {
			ids = append(ids, game.SourceID)
		}
		if answer.NextCursor == "" {
			break
		}
		cursor = answer.NextCursor
	}
	if want := []int{100, 100, 50}; fmt.Sprint(pages) != fmt.Sprint(want) {
		t.Fatalf("page sizes = %v, want %v", pages, want)
	}
	if len(ids) != 250 || ids[0] != "game-001" || ids[249] != "game-250" {
		t.Fatalf("got %d ids, first %q, last %q", len(ids), ids[0], ids[len(ids)-1])
	}
}

func TestListGamesHonoursTheLimit(t *testing.T) {
	cases := []struct {
		limit     int
		wantCount int
		wantNext  string
	}{
		{limit: 1, wantCount: 1, wantNext: "1"},
		{limit: 200, wantCount: 200, wantNext: "200"},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("limit %d", tc.limit), func(t *testing.T) {
			sup, _, _ := newFixtureSupervisor(t, Permissions{})
			answer, err := callListGames(t, sup, listGamesRequest{Source: "fixture", Scope: "mine", Limit: tc.limit})
			if err != nil {
				t.Fatalf("source.listGames: %v", err)
			}
			if len(answer.Games) != tc.wantCount || answer.NextCursor != tc.wantNext {
				t.Fatalf("got %d games, next %q; want %d, next %q", len(answer.Games), answer.NextCursor, tc.wantCount, tc.wantNext)
			}
		})
	}
}

func TestListGamesRejectsBadParams(t *testing.T) {
	cases := []struct {
		name    string
		request listGamesRequest
	}{
		{"limit zero", listGamesRequest{Source: "fixture", Scope: "mine", Limit: 0}},
		{"limit above 200", listGamesRequest{Source: "fixture", Scope: "mine", Limit: 201}},
		{"unknown scope", listGamesRequest{Source: "fixture", Scope: "all", Limit: 10}},
		{"unknown source", listGamesRequest{Source: "other", Scope: "mine", Limit: 10}},
		{"cursor not a number", listGamesRequest{Source: "fixture", Scope: "mine", Cursor: "abc", Limit: 10}},
		{"cursor past the end", listGamesRequest{Source: "fixture", Scope: "mine", Cursor: "250", Limit: 10}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sup, _, _ := newFixtureSupervisor(t, Permissions{})
			_, err := callListGames(t, sup, tc.request)
			var rpcErr *RPCError
			if !errors.As(err, &rpcErr) || rpcErr.Code != CodeInvalidParams {
				t.Fatalf("err = %v, want invalid params (code %d)", err, CodeInvalidParams)
			}
		})
	}
}
