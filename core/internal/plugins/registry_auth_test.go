package plugins

import (
	"errors"
	"fmt"
	"testing"

	"light-launcher/core/internal/contract"
)

func TestMapAuthRequiredReturnsSentinel(t *testing.T) {
	wrapped := fmt.Errorf("call failed: %w", &RPCError{Code: CodeAuthRequired, Message: "no key"})
	if err := mapAuthRequired(wrapped); !errors.Is(err, contract.ErrAuthRequired) {
		t.Errorf("error = %v, want contract.ErrAuthRequired", err)
	}
}

func TestMapAuthRequiredKeepsOtherErrors(t *testing.T) {
	other := &RPCError{Code: CodeRemoteError, Message: "site down"}
	if err := mapAuthRequired(other); err != other {
		t.Errorf("error = %v, want the same error", err)
	}
	if err := mapAuthRequired(nil); err != nil {
		t.Errorf("nil error became %v", err)
	}
}
