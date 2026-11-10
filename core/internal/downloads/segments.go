package downloads

const maxConnections = 32

// ClampConnections keeps a connection count between 1 and 32.
func ClampConnections(n int) int {
	return min(max(n, 1), maxConnections)
}

// planSegments splits [0, size) into up to `connections` contiguous, non-empty ranges.
func planSegments(size int64, connections int) []Segment {
	if size <= 0 {
		return []Segment{{Index: 0}}
	}
	count := min(int64(ClampConnections(connections)), size)
	base := size / count
	extra := size % count
	segments := make([]Segment, 0, count)
	var start int64
	for i := int64(0); i < count; i++ {
		length := base
		if i < extra {
			length++
		}
		segments = append(segments, Segment{Index: len(segments), Start: start, End: start + length})
		start += length
	}
	return segments
}

// reuseSegments keeps the saved layout only when it still covers the whole file without gaps.
func reuseSegments(size int64, saved []Segment) ([]Segment, bool) {
	if size <= 0 || len(saved) == 0 {
		return nil, false
	}
	out := make([]Segment, 0, len(saved))
	var cursor int64
	for i, seg := range saved {
		length := seg.End - seg.Start
		if seg.Start != cursor || length <= 0 || seg.End > size || seg.Done < 0 || seg.Done > length {
			return nil, false
		}
		seg.Index = i
		out = append(out, seg)
		cursor = seg.End
	}
	if cursor != size {
		return nil, false
	}
	return out, true
}

// remaining is the number of bytes a ranged segment still needs.
func remaining(seg Segment) int64 {
	return seg.End - seg.Start - seg.Done
}
