package types

import (
	"testing"
	"time"
)

func intOf(value int) *int {
	return &value
}

func boolOf(value bool) *bool {
	return &value
}

func TestResolveUpdateCheckTable(t *testing.T) {
	global := &UpdateCheckSettings{RecheckHours: intOf(6)}
	tests := []struct {
		name        string
		checks      *UpdateCheckSettings
		source      string
		wantEnabled bool
		wantRecheck time.Duration
	}{
		{"nil settings use the defaults", nil, "f95", true, 24 * time.Hour},
		{"inherit uses the global values", &UpdateCheckSettings{RecheckHours: intOf(6)}, "f95", true, 6 * time.Hour},
		{"missing source inherits", global, "itch", true, 6 * time.Hour},
		{"explicit inherit", &UpdateCheckSettings{
			RecheckHours: intOf(6),
			Sources:      map[string]SourceUpdateCheck{"f95": {Mode: UpdateModeInherit, RecheckHours: intOf(2)}},
		}, "f95", true, 6 * time.Hour},
		{"custom with nil fields falls back to global", &UpdateCheckSettings{
			RecheckHours: intOf(6),
			Sources:      map[string]SourceUpdateCheck{"f95": {Mode: UpdateModeCustom}},
		}, "f95", true, 6 * time.Hour},
		{"custom uses its own recheck", &UpdateCheckSettings{
			RecheckHours: intOf(6),
			Sources:      map[string]SourceUpdateCheck{"f95": {Mode: UpdateModeCustom, RecheckHours: intOf(0)}},
		}, "f95", true, 0},
		{"off is never auto checked", &UpdateCheckSettings{
			Sources: map[string]SourceUpdateCheck{"f95": {Mode: UpdateModeOff}},
		}, "f95", false, 24 * time.Hour},
		{"unknown mode inherits", &UpdateCheckSettings{
			RecheckHours: intOf(3),
			Sources:      map[string]SourceUpdateCheck{"f95": {Mode: "weird"}},
		}, "f95", true, 3 * time.Hour},
		{"global off turns every source off", &UpdateCheckSettings{Enabled: boolOf(false)}, "f95", false, 24 * time.Hour},
		{"global off still keeps custom recheck", &UpdateCheckSettings{
			Enabled: boolOf(false),
			Sources: map[string]SourceUpdateCheck{"f95": {Mode: UpdateModeCustom, RecheckHours: intOf(1)}},
		}, "f95", false, time.Hour},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			settings := AppSettings{UpdateChecks: tt.checks}
			enabled, recheck := ResolveUpdateCheck(settings, tt.source)
			if enabled != tt.wantEnabled || recheck != tt.wantRecheck {
				t.Fatalf("got enabled=%v recheck=%v, want enabled=%v recheck=%v", enabled, recheck, tt.wantEnabled, tt.wantRecheck)
			}
		})
	}
}

func TestResolveScheduleDefaultsAndValues(t *testing.T) {
	enabled, onStartup, interval := ResolveSchedule(AppSettings{})
	if !enabled || !onStartup || interval != 24*time.Hour {
		t.Errorf("defaults = %v %v %v, want on, on, 24h", enabled, onStartup, interval)
	}

	settings := AppSettings{UpdateChecks: &UpdateCheckSettings{
		Enabled:       boolOf(false),
		OnStartup:     boolOf(false),
		IntervalHours: intOf(0),
	}}
	enabled, onStartup, interval = ResolveSchedule(settings)
	if enabled || onStartup || interval != 0 {
		t.Errorf("values = %v %v %v, want off, off, 0", enabled, onStartup, interval)
	}
}

func TestResolveClampsHours(t *testing.T) {
	settings := AppSettings{UpdateChecks: &UpdateCheckSettings{
		IntervalHours: intOf(-5),
		RecheckHours:  intOf(10000),
	}}
	_, _, interval := ResolveSchedule(settings)
	if interval != 0 {
		t.Errorf("interval = %v, want 0 for a negative value", interval)
	}
	_, recheck := ResolveUpdateCheck(settings, "f95")
	if recheck != MaxUpdateHours*time.Hour {
		t.Errorf("recheck = %v, want %d hours for a huge value", recheck, MaxUpdateHours)
	}
}

func TestResolveSourceInterval(t *testing.T) {
	tests := []struct {
		name   string
		checks *UpdateCheckSettings
		source string
		want   time.Duration
	}{
		{"defaults to 24 hours", nil, "f95", 24 * time.Hour},
		{"uses the global interval", &UpdateCheckSettings{IntervalHours: intOf(6)}, "f95", 6 * time.Hour},
		{"custom source interval wins", &UpdateCheckSettings{IntervalHours: intOf(6), Sources: map[string]SourceUpdateCheck{"f95": {Mode: UpdateModeCustom, IntervalHours: intOf(72)}}}, "f95", 72 * time.Hour},
		{"custom without an interval inherits", &UpdateCheckSettings{IntervalHours: intOf(6), Sources: map[string]SourceUpdateCheck{"f95": {Mode: UpdateModeCustom}}}, "f95", 6 * time.Hour},
		{"off source has no interval", &UpdateCheckSettings{Sources: map[string]SourceUpdateCheck{"f95": {Mode: UpdateModeOff}}}, "f95", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolveSourceInterval(AppSettings{UpdateChecks: tt.checks}, tt.source); got != tt.want {
				t.Fatalf("interval = %v, want %v", got, tt.want)
			}
		})
	}
}
