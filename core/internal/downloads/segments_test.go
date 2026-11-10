package downloads

import "testing"

func TestPlanSegmentsCoversFileWithoutGaps(t *testing.T) {
	segs := planSegments(10, 4)
	if len(segs) != 4 {
		t.Fatalf("segments = %d, want 4", len(segs))
	}
	var cursor, total int64
	for i, seg := range segs {
		if seg.Index != i || seg.Start != cursor || seg.End <= seg.Start {
			t.Fatalf("segment %d out of order: %+v", i, seg)
		}
		cursor = seg.End
		total += seg.End - seg.Start
	}
	if total != 10 || cursor != 10 {
		t.Fatalf("segments cover %d bytes ending at %d, want 10", total, cursor)
	}
}

func TestPlanSegmentsNeverMakesEmptySegments(t *testing.T) {
	if got := len(planSegments(3, 32)); got != 3 {
		t.Fatalf("3 bytes over 32 connections gave %d segments, want 3", got)
	}
}

func TestPlanSegmentsUnknownSizeIsOneSegment(t *testing.T) {
	segs := planSegments(0, 8)
	if len(segs) != 1 || segs[0].End != 0 {
		t.Fatalf("unknown size segments = %+v", segs)
	}
}

func TestClampConnections(t *testing.T) {
	cases := map[int]int{-3: 1, 0: 1, 8: 8, 99: 32}
	for in, want := range cases {
		if got := ClampConnections(in); got != want {
			t.Errorf("ClampConnections(%d) = %d, want %d", in, got, want)
		}
	}
}

func TestReuseSegmentsAcceptsMatchingLayout(t *testing.T) {
	saved := []Segment{{Index: 0, Start: 0, End: 5, Done: 2}, {Index: 1, Start: 5, End: 10}}
	segs, ok := reuseSegments(10, saved)
	if !ok || segs[0].Done != 2 || segs[1].Index != 1 {
		t.Fatalf("reuse failed: ok=%v segs=%+v", ok, segs)
	}
}

func TestReuseSegmentsRejectsBadLayouts(t *testing.T) {
	cases := map[string][]Segment{
		"gap":          {{Start: 0, End: 4}, {Start: 5, End: 10}},
		"short cover":  {{Start: 0, End: 5}},
		"done too big": {{Start: 0, End: 10, Done: 11}},
		"past size":    {{Start: 0, End: 12}},
		"empty":        nil,
	}
	for name, saved := range cases {
		if _, ok := reuseSegments(10, saved); ok {
			t.Errorf("%s: layout should be rejected", name)
		}
	}
}
