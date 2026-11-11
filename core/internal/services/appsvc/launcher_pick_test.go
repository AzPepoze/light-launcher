package appsvc

import (
	"path/filepath"
	"slices"
	"testing"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/types"
)

// exeFile is a candidate executable at the path made of parts.
func exeFile(size int64, parts ...string) scannedFile {
	return scannedFile{Path: filepath.Join(parts...), Size: size}
}

// wantLauncher is an expected launcher at the path made of parts.
func wantLauncher(kind types.LauncherKind, primary bool, parts ...string) types.Launcher {
	path := filepath.Join(parts...)
	return types.Launcher{Name: config.NameWithoutExt(path), Path: path, Kind: kind, Primary: primary}
}

func TestPickPrimaryLauncher(t *testing.T) {
	tests := []struct {
		name       string
		folder     []string
		exes       []scannedFile
		remembered []string
		want       []types.Launcher
	}{
		{
			name:   "unreal shipping beats the root stub",
			folder: []string{"games", "MyGame"},
			exes: []scannedFile{
				exeFile(1000, "games", "MyGame", "MyGame.exe"),
				exeFile(500, "games", "MyGame", "Binaries", "Win64", "MyGame-Win64-Shipping.exe"),
			},
			want: []types.Launcher{
				wantLauncher(types.LauncherMain, true, "games", "MyGame", "Binaries", "Win64", "MyGame-Win64-Shipping.exe"),
				wantLauncher(types.LauncherUnrealStub, false, "games", "MyGame", "MyGame.exe"),
			},
		},
		{
			name:   "shipping outside Binaries/Win64 is not an unreal shipping build",
			folder: []string{"games", "UE"},
			exes: []scannedFile{
				exeFile(900, "games", "UE", "Shipping", "UE-Win64-Shipping.exe"),
				exeFile(100, "games", "UE", "UE.exe"),
			},
			want: []types.Launcher{
				wantLauncher(types.LauncherMain, true, "games", "UE", "UE.exe"),
				wantLauncher(types.LauncherOther, false, "games", "UE", "Shipping", "UE-Win64-Shipping.exe"),
			},
		},
		{
			name:   "name without 32 wins even when the 32-bit file is bigger",
			folder: []string{"games", "Name"},
			exes: []scannedFile{
				exeFile(900, "games", "Name", "Name32.exe"),
				exeFile(100, "games", "Name", "Name.exe"),
			},
			want: []types.Launcher{
				wantLauncher(types.LauncherMain, true, "games", "Name", "Name.exe"),
				wantLauncher(types.LauncherX86, false, "games", "Name", "Name32.exe"),
			},
		},
		{
			name:   "helper programs are removed from the launchers",
			folder: []string{"games", "Game"},
			exes: []scannedFile{
				exeFile(100, "games", "Game", "Game.exe"),
				exeFile(900, "games", "Game", "UnityCrashHandler64.exe"),
				exeFile(900, "games", "Game", "Engine", "CrashReportClient.exe"),
				exeFile(900, "games", "Game", "_Setup", "setup.exe"),
				exeFile(900, "games", "Game", "Redist", "vc_redist.x64.exe"),
				exeFile(900, "games", "Game", "Engine", "UnityCefSubProcess.exe"),
				exeFile(900, "games", "Game", "NotificationHelper.exe"),
				exeFile(900, "games", "Game", "UninstallGame.exe"),
				exeFile(900, "games", "Game", "DXSETUP.exe"),
				exeFile(900, "games", "Game", "installer_x.exe"),
			},
			want: []types.Launcher{
				wantLauncher(types.LauncherMain, true, "games", "Game", "Game.exe"),
			},
		},
		{
			name:   "a folder with only helper programs has no launcher",
			folder: []string{"games", "Junk"},
			exes: []scannedFile{
				exeFile(900, "games", "Junk", "UnityCrashHandler64.exe"),
			},
			want: nil,
		},
		{
			name:   "single exe is the main launcher",
			folder: []string{"games", "Solo"},
			exes: []scannedFile{
				exeFile(10, "games", "Solo", "Play.exe"),
			},
			want: []types.Launcher{
				wantLauncher(types.LauncherMain, true, "games", "Solo", "Play.exe"),
			},
		},
		{
			name:       "remembered launcher wins over the rules",
			folder:     []string{"games", "Rem"},
			remembered: []string{"games", "Rem", "Other.exe"},
			exes: []scannedFile{
				exeFile(900, "games", "Rem", "Rem.exe"),
				exeFile(100, "games", "Rem", "Other.exe"),
			},
			want: []types.Launcher{
				wantLauncher(types.LauncherMain, true, "games", "Rem", "Other.exe"),
				wantLauncher(types.LauncherOther, false, "games", "Rem", "Rem.exe"),
			},
		},
		{
			name:       "remembered launcher that is gone falls back to the rules",
			folder:     []string{"games", "Rem"},
			remembered: []string{"games", "Rem", "Gone.exe"},
			exes: []scannedFile{
				exeFile(100, "games", "Rem", "Other.exe"),
				exeFile(900, "games", "Rem", "Rem.exe"),
			},
			want: []types.Launcher{
				wantLauncher(types.LauncherMain, true, "games", "Rem", "Rem.exe"),
				wantLauncher(types.LauncherOther, false, "games", "Rem", "Other.exe"),
			},
		},
		{
			name:       "remembered helper program is ignored",
			folder:     []string{"games", "Rem"},
			remembered: []string{"games", "Rem", "UnityCrashHandler64.exe"},
			exes: []scannedFile{
				exeFile(900, "games", "Rem", "UnityCrashHandler64.exe"),
				exeFile(100, "games", "Rem", "Rem.exe"),
				exeFile(50, "games", "Rem", "Other.exe"),
			},
			want: []types.Launcher{
				wantLauncher(types.LauncherMain, true, "games", "Rem", "Rem.exe"),
				wantLauncher(types.LauncherOther, false, "games", "Rem", "Other.exe"),
			},
		},
		{
			name:   "name match beats size",
			folder: []string{"games", "Game"},
			exes: []scannedFile{
				exeFile(900, "games", "Game", "Launch.exe"),
				exeFile(10, "games", "Game", "Game.exe"),
			},
			want: []types.Launcher{
				wantLauncher(types.LauncherMain, true, "games", "Game", "Game.exe"),
				wantLauncher(types.LauncherOther, false, "games", "Game", "Launch.exe"),
			},
		},
		{
			name:   "larger file breaks a tie",
			folder: []string{"games", "Q"},
			exes: []scannedFile{
				exeFile(10, "games", "Q", "a.exe"),
				exeFile(20, "games", "Q", "b.exe"),
			},
			want: []types.Launcher{
				wantLauncher(types.LauncherMain, true, "games", "Q", "b.exe"),
				wantLauncher(types.LauncherOther, false, "games", "Q", "a.exe"),
			},
		},
		{
			name:   "shorter path breaks a tie after size",
			folder: []string{"games", "P"},
			exes: []scannedFile{
				exeFile(10, "games", "P", "deep", "x", "a.exe"),
				exeFile(10, "games", "P", "a.exe"),
			},
			want: []types.Launcher{
				wantLauncher(types.LauncherMain, true, "games", "P", "a.exe"),
				wantLauncher(types.LauncherOther, false, "games", "P", "deep", "x", "a.exe"),
			},
		},
		{
			name:   "alphabetical path is the last tie breaker",
			folder: []string{"games", "Other"},
			exes: []scannedFile{
				exeFile(10, "games", "Other", "z.exe"),
				exeFile(10, "games", "Other", "b.exe"),
			},
			want: []types.Launcher{
				wantLauncher(types.LauncherMain, true, "games", "Other", "b.exe"),
				wantLauncher(types.LauncherOther, false, "games", "Other", "z.exe"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			folder := filepath.Join(test.folder...)
			remembered := ""
			if test.remembered != nil {
				remembered = filepath.Join(test.remembered...)
			}
			got := pickPrimaryLauncher(folder, test.exes, remembered)
			if !slices.Equal(got, test.want) {
				t.Fatalf("launchers\n got: %+v\nwant: %+v", got, test.want)
			}
		})
	}
}
