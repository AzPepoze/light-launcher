package main

import (
	"testing"
	"time"
)

func TestCookieFromProfileRow(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	future := now.Add(time.Hour).UnixMicro() + chromiumEpochOffsetMicros
	past := now.Add(-time.Hour).UnixMicro() + chromiumEpochOffsetMicros

	tests := []struct {
		name string
		row  profileCookieRow
		want cookie
		ok   bool
	}{
		{
			name: "future expiry is kept",
			row:  profileCookieRow{Host: ".example.com", Name: "sid", Value: "s", Path: "/", Secure: true, ExpiresUTC: future, HasExpires: true},
			want: cookie{Name: "sid", Value: "s", Domain: ".example.com", Path: "/", Secure: true},
			ok:   true,
		},
		{
			name: "session cookie is kept",
			row:  profileCookieRow{Host: "example.com", Name: "sid", Value: "s", Path: "/"},
			want: cookie{Name: "sid", Value: "s", Domain: "example.com", Path: "/"},
			ok:   true,
		},
		{
			name: "empty value is kept",
			row:  profileCookieRow{Host: "example.com", Name: "empty", Path: "/"},
			want: cookie{Name: "empty", Domain: "example.com", Path: "/"},
			ok:   true,
		},
		{
			name: "expired cookie is dropped",
			row:  profileCookieRow{Host: "example.com", Name: "old", Value: "s", Path: "/", ExpiresUTC: past, HasExpires: true},
		},
		{
			name: "encrypted value is dropped",
			row:  profileCookieRow{Host: "example.com", Name: "secret", Path: "/", Encrypted: true},
		},
		{
			name: "cookie without name is dropped",
			row:  profileCookieRow{Host: "example.com", Value: "s", Path: "/"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := cookieFromProfileRow(test.row, now)
			if ok != test.ok {
				t.Fatalf("ok = %v, want %v", ok, test.ok)
			}
			if got != test.want {
				t.Fatalf("cookie = %+v, want %+v", got, test.want)
			}
		})
	}
}
