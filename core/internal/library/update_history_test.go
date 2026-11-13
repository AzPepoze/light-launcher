package library

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"light-launcher/core/internal/contract"
)

func TestUpdateHistoryRecordsEveryOutcome(t *testing.T) {
	cases := []struct {
		name    string
		reply   func(any) (any, error)
		want    string
		wantErr string
	}{
		{
			name: "update",
			reply: func(any) (any, error) {
				return map[string]any{"update": map[string]string{"version": "2.0"}}, nil
			},
			want: ResultUpdate,
		},
		{
			name:  "current",
			reply: func(any) (any, error) { return nil, nil },
			want:  ResultCurrent,
		},
		{
			name:    "failure",
			reply:   func(any) (any, error) { return nil, errors.New("login expired") },
			want:    ResultError,
			wantErr: "login expired",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lib := openTestLibrary(t, Options{Sources: updateSources(tc.reply)})
			game := f95Game(t, lib, "1.0")
			if _, err := lib.CheckGame(context.Background(), game.ID); err != nil {
				t.Fatal(err)
			}
			entries, err := lib.UpdateHistory(context.Background(), 0, 10, 0, "")
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 1 {
				t.Fatalf("history = %+v, want one entry", entries)
			}
			entry := entries[0]
			if entry.Result != tc.want || entry.Title != "Harbor" || entry.GameID != game.ID {
				t.Errorf("entry = %+v, want result %s for Harbor", entry, tc.want)
			}
			if tc.wantErr != "" && (entry.Error == nil || *entry.Error != tc.wantErr) {
				t.Errorf("error = %v, want %q", entry.Error, tc.wantErr)
			}
			if tc.wantErr == "" && entry.Error != nil {
				t.Errorf("error = %v, want none", entry.Error)
			}
		})
	}
}

func TestUpdateHistoryPagesNewestFirst(t *testing.T) {
	sources := updateSources(func(any) (any, error) {
		return map[string]any{"update": map[string]string{"version": "2.0"}}, nil
	})
	lib := openTestLibrary(t, Options{Sources: sources})
	ctx := context.Background()
	var ids []int64
	for i := range 3 {
		id, err := lib.Upsert(ctx, contract.GameDTO{
			Source: "f95", SourceID: fmt.Sprint(i), URL: "https://f95/" + fmt.Sprint(i),
			Title: "Game " + fmt.Sprint(i), Version: "1.0",
		})
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
		if _, err := lib.CheckGame(ctx, id); err != nil {
			t.Fatal(err)
		}
	}

	first, err := lib.UpdateHistory(ctx, 0, 2, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 2 || first[0].GameID != ids[2] || first[1].GameID != ids[1] {
		t.Fatalf("first page = %+v, want the two newest checks", first)
	}
	second, err := lib.UpdateHistory(ctx, 0, 2, first[1].ID, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != 1 || second[0].GameID != ids[0] {
		t.Fatalf("second page = %+v, want the oldest check", second)
	}

	filtered, err := lib.UpdateHistory(ctx, ids[1], 10, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 || filtered[0].GameID != ids[1] {
		t.Errorf("filtered history = %+v, want only game %d", filtered, ids[1])
	}
}
