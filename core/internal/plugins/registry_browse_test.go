package plugins

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsInvalidCursor(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"plugin says invalid_cursor", &RPCError{Code: CodeInvalidParams, Message: "bad cursor", Data: map[string]any{"code": "invalid_cursor"}}, true},
		{"wrapped invalid_cursor", fmt.Errorf("browse: %w", &RPCError{Code: CodeInvalidParams, Data: map[string]any{"code": "invalid_cursor"}}), true},
		{"other data code", &RPCError{Code: CodeInvalidParams, Data: map[string]any{"code": "bad_limit"}}, false},
		{"no data", &RPCError{Code: CodeInvalidParams, Message: "invalid params"}, false},
		{"data is not an object", &RPCError{Code: CodeInvalidParams, Data: "invalid_cursor"}, false},
		{"timeout", &RPCError{Code: CodeTimeout, Message: "timed out"}, false},
		{"not an rpc error", errors.New("invalid_cursor"), false},
		{"no error", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isInvalidCursor(tc.err); got != tc.want {
				t.Fatalf("isInvalidCursor(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
