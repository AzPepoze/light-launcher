package plugins

import (
	"errors"
	"testing"
)

func TestIsInvalidCursorCodes(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"sdk code", &RPCError{Code: -32007, Message: "invalid_cursor"}, true},
		{"data code", &RPCError{Code: -32602, Data: map[string]any{"code": "invalid_cursor"}}, true},
		{"other rpc error", &RPCError{Code: -32602, Message: "bad limit"}, false},
		{"plain error", errors.New("boom"), false},
		{"no error", nil, false},
	}
	for _, tc := range cases {
		if got := isInvalidCursor(tc.err); got != tc.want {
			t.Errorf("%s: isInvalidCursor = %v, want %v", tc.name, got, tc.want)
		}
	}
}
