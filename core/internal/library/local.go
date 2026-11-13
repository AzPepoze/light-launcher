package library

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/types"
)

// localBackfillMeta guards the one-time import of the old config files.
const localBackfillMeta = "local-backfill-v1"

// LocalBackfillInput is the data the old stores held before the library became canonical.
// Main reads it from the config, settings and activity stores.
type LocalBackfillInput struct {
	Configs     []types.LaunchOptions
	ScanFolders []string
	Blacklist   []string
	Activity    []types.GameActivity
}

// ScanSyncResult counts what one scan changed.
type ScanSyncResult struct {
	Added   int
	Updated int
	Adopted int
	Missing int
	Removed int
}

// BackfillLocal imports the old hand-added configs, ignored paths and playtime once.
// It is idempotent: after the first successful run the meta key skips it.
func (l *Library) BackfillLocal(ctx context.Context, in LocalBackfillInput) (int, error) {
	if _, found, err := l.GetMeta(ctx, localBackfillMeta); err != nil || found {
		return 0, err
	}
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	now := nowStamp()
	if err := upsertSourceTx(ctx, tx, SourceDescriptor{
		ID: "manual", Kind: "local", Label: "Local games", Cover: "1:1", Tile: "square",
	}); err != nil {
		return 0, err
	}
	if err := insertIgnoredTx(ctx, tx, in.Blacklist, "blacklist", now); err != nil {
		return 0, err
	}

	// One game folder can hold several configs, one per launcher. They become one row
	// with a launcher list, so every old card survives without duplicate roots.
	grouped := map[string][]types.LaunchOptions{}
	var order []string
	for _, config := range in.Configs {
		row, ok := localRowFromConfig(config, in.ScanFolders, now)
		if !ok {
			continue
		}
		if _, seen := grouped[row.RootKey]; !seen {
			order = append(order, row.RootKey)
		}
		grouped[row.RootKey] = append(grouped[row.RootKey], config)
	}

	written := 0
	for _, root := range order {
		configs := grouped[root]
		row, ok := localRowFromConfig(configs[0], in.ScanFolders, now)
		if !ok {
			continue
		}
		if ignored, err := isIgnoredTx(ctx, tx, row.RootKey, row.ExeKey, row.InstallDirKey); err != nil {
			return 0, err
		} else if ignored {
			continue
		}
		gameID, err := upsertLocalTx(ctx, tx, row, row.Origin == "hand")
		if err != nil {
			return 0, err
		}
		launchers := make([]types.Launcher, 0, len(configs))
		for index, entry := range configs {
			exe := entry.LauncherPath
			if exe == "" {
				exe = entry.GamePath
			}
			name := entry.Name
			if name == "" {
				name = config.NameWithoutExt(exe)
			}
			launchers = append(launchers, types.Launcher{
				Name: name, Path: exe, Kind: types.LauncherMain, Primary: index == 0,
			})
		}
		if err := replaceLaunchersTx(ctx, tx, gameID, launchers); err != nil {
			return 0, err
		}
		written++
	}

	for _, activity := range in.Activity {
		if err := applyActivityTx(ctx, tx, activity); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	if err := l.SetMeta(ctx, localBackfillMeta, now); err != nil {
		return 0, err
	}
	if written > 0 {
		l.markWritten()
		l.notify("library.changed", map[string]any{})
	}
	return written, nil
}

// SyncScanned saves the result of one folder scan. Rows that disappear are marked missing
// and removed after the second scan that does not find them; pinned rows are never removed.
func (l *Library) SyncScanned(ctx context.Context, groups []types.ScannedFolderGroup) (ScanSyncResult, error) {
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return ScanSyncResult{}, err
	}
	defer tx.Rollback()

	now := nowStamp()
	result := ScanSyncResult{}
	for _, group := range groups {
		scanFolder := config.CleanPath(group.FolderPath)
		scanKey := config.PathKey(scanFolder)
		seen := map[string]bool{}
		for _, game := range group.Games {
			row, ok := localRowFromGame(game, scanFolder, now)
			if !ok {
				continue
			}
			seen[row.RootKey] = true
			if ignored, err := isIgnoredTx(ctx, tx, row.RootKey, row.ExeKey, row.InstallDirKey); err != nil {
				return ScanSyncResult{}, err
			} else if ignored {
				continue
			}
			state, err := upsertScannedTx(ctx, tx, row)
			if err != nil {
				return ScanSyncResult{}, err
			}
			switch state {
			case "added":
				result.Added++
			case "adopted":
				result.Adopted++
			default:
				result.Updated++
			}
			if err := replaceLaunchersTx(ctx, tx, row.ID, game.Launchers); err != nil {
				return ScanSyncResult{}, err
			}
		}
		missing, removed, err := sweepScanFolderTx(ctx, tx, scanKey, seen, now)
		if err != nil {
			return ScanSyncResult{}, err
		}
		result.Missing += missing
		result.Removed += removed
	}
	if err := tx.Commit(); err != nil {
		return ScanSyncResult{}, err
	}
	if result.Added+result.Updated+result.Adopted+result.Removed > 0 {
		l.markWritten()
		l.notify("library.changed", map[string]any{})
	}
	return result, nil
}

// SweepScanFolder removes the local rows of a folder the user stopped watching.
func (l *Library) SweepScanFolder(ctx context.Context, folder string) (int, error) {
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx,
		`select id from games where scan_folder_key = ? and source_kind = 'local' and pinned = 0`,
		config.PathKey(folder))
	if err != nil {
		return 0, err
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return 0, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}
	for _, id := range ids {
		if err := removeLocalTx(ctx, tx, id); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	if len(ids) > 0 {
		l.markWritten()
		l.notify("library.changed", map[string]any{})
	}
	return len(ids), nil
}

// Launchers returns the stored launchers of one game, primary first.
func (l *Library) Launchers(ctx context.Context, gameID int64) ([]types.Launcher, error) {
	rows, err := l.db.QueryContext(ctx,
		`select name, path, kind, primary_launcher from game_launchers where game_id = ? order by primary_launcher desc, name`,
		gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []types.Launcher{}
	for rows.Next() {
		var item types.Launcher
		var primary int
		if err := rows.Scan(&item.Name, &item.Path, &item.Kind, &primary); err != nil {
			return nil, err
		}
		item.Primary = primary == 1
		out = append(out, item)
	}
	return out, rows.Err()
}

type localRow struct {
	ID            int64
	SourceID      string
	Title         string
	FolderPath    string
	ScanFolder    string
	RootKey       string
	ExeKey        string
	InstallDirKey string
	ScanFolderKey string
	InstallDir    string
	ExePath       string
	ExeArgs       string
	Origin        string
	PlaytimeS     int64
	LastPlayedAt  *string
	LocalConfig   []byte
	Launchers     []types.Launcher
}

// localRowFromConfig turns one old config file into a local row. Hand-added games are pinned.
func localRowFromConfig(c types.LaunchOptions, scanFolders []string, now string) (localRow, bool) {
	exe := strings.TrimSpace(c.LauncherPath)
	if exe == "" {
		exe = strings.TrimSpace(c.GamePath)
	}
	if exe == "" {
		return localRow{}, false
	}
	folder := filepath.Dir(exe)
	rootKey := config.PathKey(folder)
	exeKey := config.PathKey(exe)
	sourceID := "config:" + c.ID
	if c.ID == "" {
		sourceID = "exe:" + exeKey
	}
	origin := "hand"
	scanFolder, scanKey := "", ""
	for _, watched := range scanFolders {
		if config.IsSubPath(watched, folder) {
			origin = "scan"
			scanFolder = config.CleanPath(watched)
			scanKey = config.PathKey(scanFolder)
			break
		}
	}
	title := strings.TrimSpace(c.Name)
	if title == "" {
		title = config.NameWithoutExt(exe)
	}
	raw, _ := json.Marshal(c)
	return localRow{
		SourceID: sourceID, Title: title, FolderPath: folder, ScanFolder: scanFolder,
		RootKey: rootKey, ExeKey: exeKey, InstallDirKey: rootKey,
		ScanFolderKey: scanKey, InstallDir: folder, ExePath: exe, ExeArgs: c.CustomArgs,
		Origin: origin, LocalConfig: raw,
	}, true
}

// localRowFromGame turns one scan card into a local row.
func localRowFromGame(game types.GameInfo, scanFolder, now string) (localRow, bool) {
	folder := config.CleanPath(game.Folder)
	exe := game.Path
	if folder == "" || exe == "" {
		return localRow{}, false
	}
	rootKey := config.PathKey(folder)
	raw, _ := json.Marshal(game.Config)
	return localRow{
		SourceID: "scan:" + rootKey, Title: game.Name, FolderPath: folder, ScanFolder: scanFolder,
		RootKey: rootKey, ExeKey: config.PathKey(exe), InstallDirKey: rootKey,
		ScanFolderKey: config.PathKey(scanFolder), InstallDir: folder, ExePath: exe,
		ExeArgs: game.Config.CustomArgs, Origin: "scan", LocalConfig: raw, Launchers: game.Launchers,
	}, true
}

// upsertLocalTx inserts or refreshes a hand row. It keeps the user's title and playtime.
func upsertLocalTx(ctx context.Context, tx *sql.Tx, row localRow, pinned bool) (int64, error) {
	now := nowStamp()
	var id int64
	err := tx.QueryRowContext(ctx, `insert into games
		(source, source_id, url, title, install_dir, exe_path, exe_args, added_at, updated_at,
		 origin, source_kind, folder_path, scan_folder, root_key, exe_key, install_dir_key,
		 scan_folder_key, last_seen_at, pinned, local_config)
		values ('manual',?,?,?,?,?,?,?,?,?,'local',?,?,?,?,?,?,?,?,?)
		on conflict(source, source_id) do update set
			install_dir=excluded.install_dir, exe_path=excluded.exe_path, exe_args=excluded.exe_args,
			origin=excluded.origin, source_kind='local', folder_path=excluded.folder_path,
			scan_folder=excluded.scan_folder, root_key=excluded.root_key, exe_key=excluded.exe_key,
			install_dir_key=excluded.install_dir_key, scan_folder_key=excluded.scan_folder_key,
			last_seen_at=excluded.last_seen_at, local_config=excluded.local_config,
			pinned=max(games.pinned, excluded.pinned), updated_at=excluded.updated_at
		returning id`,
		row.SourceID, "", row.Title, row.InstallDir, row.ExePath, nullIfEmpty(row.ExeArgs), now, now,
		row.Origin, row.FolderPath, nullIfEmpty(row.ScanFolder), row.RootKey, row.ExeKey,
		row.InstallDirKey, nullIfEmpty(row.ScanFolderKey), now, boolInt(pinned), string(row.LocalConfig)).
		Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("save local game %q: %w", row.Title, err)
	}
	return id, nil
}

// upsertScannedTx finds a row by game root (or adopts a hand row by exe) and refreshes it.
func upsertScannedTx(ctx context.Context, tx *sql.Tx, row localRow) (string, error) {
	var id int64
	var pinned int
	err := tx.QueryRowContext(ctx,
		`select id, pinned from games where root_key = ? and source_kind = 'local'`, row.RootKey).
		Scan(&id, &pinned)
	if err == nil {
		if err := updateScannedTx(ctx, tx, id, row); err != nil {
			return "", err
		}
		row.ID = id
		return "updated", nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}
	// A hand-added row for the same exe becomes the scanned row instead of a duplicate.
	err = tx.QueryRowContext(ctx,
		`select id, pinned from games where exe_key = ? and source_kind = 'local'`, row.ExeKey).
		Scan(&id, &pinned)
	if err == nil {
		if err := updateScannedTx(ctx, tx, id, row); err != nil {
			return "", err
		}
		row.ID = id
		return "adopted", nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}
	now := nowStamp()
	err = tx.QueryRowContext(ctx, `insert into games
		(source, source_id, url, title, install_dir, exe_path, exe_args, added_at, updated_at,
		 origin, source_kind, folder_path, scan_folder, root_key, exe_key, install_dir_key,
		 scan_folder_key, last_seen_at, pinned, local_config)
		values ('manual',?,'',?,?,?,?,?,?,'scan','local',?,?,?,?,?,?,?,0,?)
		returning id`,
		row.SourceID, row.Title, row.InstallDir, row.ExePath, nullIfEmpty(row.ExeArgs), now, now,
		row.FolderPath, nullIfEmpty(row.ScanFolder), row.RootKey, row.ExeKey, row.InstallDirKey,
		nullIfEmpty(row.ScanFolderKey), now, string(row.LocalConfig)).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("save scanned game %q: %w", row.Title, err)
	}
	row.ID = id
	return "added", nil
}

func updateScannedTx(ctx context.Context, tx *sql.Tx, id int64, row localRow) error {
	_, err := tx.ExecContext(ctx, `update games set
		install_dir=?, exe_path=?, exe_args=?, origin=case when origin='hand' then 'hand' else 'scan' end,
		source_kind='local', folder_path=?, scan_folder=?, root_key=?, exe_key=?, install_dir_key=?,
		scan_folder_key=?, last_seen_at=?, missing_since=null, miss_count=0, local_config=?,
		title=case when pinned=1 then title else ? end, updated_at=?
		where id=?`,
		row.InstallDir, row.ExePath, nullIfEmpty(row.ExeArgs), row.FolderPath, nullIfEmpty(row.ScanFolder),
		row.RootKey, row.ExeKey, row.InstallDirKey, nullIfEmpty(row.ScanFolderKey), nowStamp(),
		string(row.LocalConfig), row.Title, nowStamp(), id)
	return err
}

// replaceLaunchersTx stores the launchers of one local game, primary first.
func replaceLaunchersTx(ctx context.Context, tx *sql.Tx, gameID int64, launchers []types.Launcher) error {
	if gameID == 0 {
		return nil
	}
	if _, err := tx.ExecContext(ctx, `delete from game_launchers where game_id = ?`, gameID); err != nil {
		return err
	}
	for _, launcher := range launchers {
		if launcher.Path == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `insert into game_launchers
			(game_id, path, path_key, name, kind, primary_launcher) values (?,?,?,?,?,?)
			on conflict(game_id, path_key) do update set name=excluded.name, kind=excluded.kind,
			primary_launcher=excluded.primary_launcher`,
			gameID, launcher.Path, config.PathKey(launcher.Path), launcher.Name,
			string(launcher.Kind), boolInt(launcher.Primary)); err != nil {
			return err
		}
	}
	return nil
}

// sweepScanFolderTx marks or removes the rows of one scan folder that the scan did not see.
func sweepScanFolderTx(ctx context.Context, tx *sql.Tx, scanKey string, seen map[string]bool, now string) (int, int, error) {
	rows, err := tx.QueryContext(ctx, `select id, root_key, missing_since, miss_count, pinned
		from games where scan_folder_key = ? and source_kind = 'local'`, scanKey)
	if err != nil {
		return 0, 0, err
	}
	type candidate struct {
		id      int64
		root    sql.NullString
		missing sql.NullString
		misses  int
		pinned  int
	}
	var candidates []candidate
	for rows.Next() {
		var c candidate
		if err := rows.Scan(&c.id, &c.root, &c.missing, &c.misses, &c.pinned); err != nil {
			rows.Close()
			return 0, 0, err
		}
		if c.pinned == 1 || (c.root.Valid && seen[c.root.String]) {
			continue
		}
		candidates = append(candidates, c)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, 0, err
	}
	missing, removed := 0, 0
	for _, c := range candidates {
		if c.missing.Valid || c.misses > 0 {
			if err := removeLocalTx(ctx, tx, c.id); err != nil {
				return 0, 0, err
			}
			removed++
			continue
		}
		if _, err := tx.ExecContext(ctx, `update games set missing_since=?, miss_count=miss_count+1 where id=?`, now, c.id); err != nil {
			return 0, 0, err
		}
		missing++
	}
	return missing, removed, nil
}

// removeLocalTx deletes one local row and keeps a compact removal history.
func removeLocalTx(ctx context.Context, tx *sql.Tx, id int64) error {
	if _, err := tx.ExecContext(ctx, `insert into deleted_games
		(game_id, source, source_id, url, title, version, developer, tags, install_dir, deleted_files, deleted_at)
		select id, source, source_id, url, title, version, developer, tags, install_dir, 0, ? from games where id=?`,
		nowStamp(), id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `delete from games where id = ?`, id); err != nil {
		return err
	}
	return nil
}

func insertIgnoredTx(ctx context.Context, tx *sql.Tx, paths []string, reason, now string) error {
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `insert into ignored_paths (path_key, path, reason, created_at)
			values (?,?,?,?) on conflict(path_key) do nothing`,
			config.PathKey(path), path, reason, now); err != nil {
			return err
		}
	}
	return nil
}

func isIgnoredTx(ctx context.Context, tx *sql.Tx, keys ...string) (bool, error) {
	for _, key := range keys {
		if key == "" {
			continue
		}
		var one int
		err := tx.QueryRowContext(ctx, `select 1 from ignored_paths where path_key = ?`, key).Scan(&one)
		if err == nil {
			return true, nil
		}
		if err != sql.ErrNoRows {
			return false, err
		}
	}
	return false, nil
}

// applyActivityTx copies old playtime onto the matching local row.
func applyActivityTx(ctx context.Context, tx *sql.Tx, activity types.GameActivity) error {
	if activity.TotalPlaytimeSeconds <= 0 && activity.LastPlayedAt <= 0 {
		return nil
	}
	var played any
	if activity.LastPlayedAt > 0 {
		played = stamp(time.Unix(activity.LastPlayedAt, 0))
	}
	_, err := tx.ExecContext(ctx, `update games set
		playtime_s = max(playtime_s, ?),
		last_played_at = coalesce(last_played_at, ?)
		where source_kind = 'local' and (exe_key = ? or root_key = ?)`,
		activity.TotalPlaytimeSeconds, played, config.PathKey(activity.GamePath), config.PathKey(activity.GamePath))
	return err
}

// SourceDescriptor is the stored part of a source. The plugin registry supplies the live fields.
type SourceDescriptor struct {
	ID       string
	Kind     string
	Label    string
	PluginID string
	Cover    string
	Tile     string
}

func upsertSourceTx(ctx context.Context, tx *sql.Tx, source SourceDescriptor) error {
	_, err := tx.ExecContext(ctx, `insert into sources
		(id, kind, label, plugin_id, cover, tile, first_seen_at)
		values (?,?,?,?,?,?,?)
		on conflict(id) do update set kind=excluded.kind, label=excluded.label,
		plugin_id=excluded.plugin_id, cover=excluded.cover, tile=excluded.tile`,
		source.ID, source.Kind, source.Label, nullIfEmpty(source.PluginID), nullIfEmpty(source.Cover),
		nullIfEmpty(source.Tile), nowStamp())
	return err
}
