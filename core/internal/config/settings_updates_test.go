package config

import (
	"testing"

	"light-launcher/core/internal/types"
)

func TestDefaultSettingsHaveUpdateChecks(t *testing.T) {
	checks := DefaultSettings().UpdateChecks
	if checks == nil || !*checks.Enabled || !*checks.OnStartup {
		t.Fatalf("update checks = %+v, want enabled and on startup", checks)
	}
	if *checks.IntervalHours != types.DefaultUpdateIntervalHours || *checks.RecheckHours != types.DefaultUpdateRecheckHours {
		t.Errorf("hours = %d / %d, want the defaults", *checks.IntervalHours, *checks.RecheckHours)
	}
}

func TestNormalizeFillsMissingUpdateChecks(t *testing.T) {
	settings := &types.AppSettings{}
	if !normalizeUpdateChecks(settings) {
		t.Fatal("a missing update check block did not report a change")
	}
	if settings.UpdateChecks == nil || *settings.UpdateChecks.IntervalHours != types.DefaultUpdateIntervalHours {
		t.Fatalf("update checks = %+v, want the defaults", settings.UpdateChecks)
	}
}

func TestNormalizeClampsUpdateHours(t *testing.T) {
	settings := &types.AppSettings{UpdateChecks: &types.UpdateCheckSettings{
		Enabled:       boolPtr(true),
		OnStartup:     boolPtr(true),
		IntervalHours: intPtr(-1),
		RecheckHours:  intPtr(9999),
		Sources: map[string]types.SourceUpdateCheck{
			"f95":  {Mode: types.UpdateModeCustom, IntervalHours: intPtr(1000), RecheckHours: intPtr(-3)},
			"itch": {Mode: "bogus"},
		},
	}}
	if !normalizeUpdateChecks(settings) {
		t.Fatal("out of range values did not report a change")
	}
	checks := settings.UpdateChecks
	if *checks.IntervalHours != 0 || *checks.RecheckHours != types.MaxUpdateHours {
		t.Errorf("global hours = %d / %d, want 0 / %d", *checks.IntervalHours, *checks.RecheckHours, types.MaxUpdateHours)
	}
	f95 := checks.Sources["f95"]
	if *f95.IntervalHours != types.MaxUpdateHours || *f95.RecheckHours != 0 {
		t.Errorf("f95 hours = %d / %d, want %d / 0", *f95.IntervalHours, *f95.RecheckHours, types.MaxUpdateHours)
	}
	if checks.Sources["itch"].Mode != types.UpdateModeInherit {
		t.Errorf("itch mode = %q, want inherit", checks.Sources["itch"].Mode)
	}
}

func TestNormalizeKeepsValidUpdateChecks(t *testing.T) {
	settings := &types.AppSettings{UpdateChecks: &types.UpdateCheckSettings{
		Enabled:       boolPtr(false),
		OnStartup:     boolPtr(true),
		IntervalHours: intPtr(12),
		RecheckHours:  intPtr(6),
		Sources: map[string]types.SourceUpdateCheck{
			"f95": {Mode: types.UpdateModeOff},
		},
	}}
	if normalizeUpdateChecks(settings) {
		t.Error("valid update checks reported a change")
	}
}
