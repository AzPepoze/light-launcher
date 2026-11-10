package downloads

import "time"

const (
	speedWindow  = 3 * time.Second
	minSpeedSpan = 250 * time.Millisecond
)

// speedMeter reports bytes per second over the last few seconds. It is kept in memory only.
type speedMeter struct {
	samples []meterSample
}

type meterSample struct {
	at    time.Time
	bytes int64
}

func (m *speedMeter) add(now time.Time, n int64) {
	m.samples = append(m.samples, meterSample{at: now, bytes: n})
	m.trim(now)
}

func (m *speedMeter) trim(now time.Time) {
	cutoff := now.Add(-speedWindow)
	drop := 0
	for drop < len(m.samples) && m.samples[drop].at.Before(cutoff) {
		drop++
	}
	m.samples = m.samples[drop:]
}

// rate returns bytes per second over the window.
func (m *speedMeter) rate(now time.Time) float64 {
	m.trim(now)
	if len(m.samples) == 0 {
		return 0
	}
	var total int64
	for _, s := range m.samples {
		total += s.bytes
	}
	span := max(now.Sub(m.samples[0].at), minSpeedSpan)
	return float64(total) / span.Seconds()
}
