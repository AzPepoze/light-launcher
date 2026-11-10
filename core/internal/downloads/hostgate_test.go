package downloads

import (
	"context"
	"testing"
	"time"
)

func TestHostGatesShareOneGatePerHost(t *testing.T) {
	gates := newHostGates(2)
	a := gates.forURL("https://cdn.example.test/one.zip")
	b := gates.forURL("https://cdn.example.test/two.zip")
	if a != b {
		t.Fatal("two links on one host must share a gate")
	}
	if other := gates.forURL("https://other.example.test/x.zip"); other == a {
		t.Fatal("different hosts must not share a gate")
	}
}

func TestHostGateLimitAppliesToExistingHosts(t *testing.T) {
	gates := newHostGates(2)
	a := gates.forURL("https://cdn.example.test/one.zip")
	gates.setLimit(1)
	if a.Limit() != 1 {
		t.Fatalf("existing host gate limit = %d, want 1", a.Limit())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if err := a.Acquire(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.Acquire(ctx); err == nil {
		t.Fatal("second connection to the host must wait at limit 1")
	}
}
