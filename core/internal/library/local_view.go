package library

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"

	"light-launcher/core/internal/config"
	"light-launcher/core/internal/types"
)

// LocalGames returns the local rows in the shape the Home page and old callers use.
// A local row may have several launchers; the first primary launcher is its Path.
func (l *Library) LocalGames(ctx context.Context) ([]types.GameInfo, error) {
	rows, err := l.db.QueryContext(ctx,
		`select `+gameColumns+` from games where source_kind = 'local' order by title collate nocase`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	games := []Game{}
	for rows.Next() {
		game, err := scanGame(rows)
		if err != nil {
			return nil, err
		}
		games = append(games, game)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	launchers, err := l.allLaunchers(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]types.GameInfo, 0, len(games))
	for _, game := range games {
		out = append(out, localGameInfo(game, launchers[game.ID]))
	}
	return out, nil
}

// LocalGroups returns the scanned rows grouped by their watched folder.
func (l *Library) LocalGroups(ctx context.Context) ([]types.ScannedFolderGroup, error) {
	rows, err := l.db.QueryContext(ctx,
		`select `+gameColumns+` from games
		 where source_kind = 'local' and origin = 'scan' and scan_folder is not null
		 order by scan_folder, title collate nocase`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	groups := []types.ScannedFolderGroup{}
	index := map[string]int{}
	for rows.Next() {
		game, err := scanGame(rows)
		if err != nil {
			return nil, err
		}
		folder := ""
		if game.ScanFolder != nil {
			folder = *game.ScanFolder
		}
		if folder == "" {
			continue
		}
		at, found := index[folder]
		if !found {
			at = len(groups)
			index[folder] = at
			groups = append(groups, types.ScannedFolderGroup{
				FolderPath: folder,
				FolderName: config.BaseName(folder),
			})
		}
		groups[at].Games = append(groups[at].Games, localGameInfo(game, nil))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	launchers, err := l.allLaunchers(ctx)
	if err != nil {
		return nil, err
	}
	byPath := map[string][]types.Launcher{}
	for _, list := range launchers {
		for _, launcher := range list {
			byPath[launcher.Path] = list
		}
	}
	for groupIndex := range groups {
		for gameIndex := range groups[groupIndex].Games {
			game := &groups[groupIndex].Games[gameIndex]
			if list := byPath[game.Path]; len(list) > 0 {
				game.Launchers = list
				game.LauncherCount = len(list)
			}
		}
	}
	return groups, nil
}

// LocalConfig returns the saved launch settings of one local executable or folder, or nil.
func (l *Library) LocalConfig(ctx context.Context, executable string) *types.LaunchOptions {
	key := config.PathKey(executable)
	if key == "" {
		return nil
	}
	game, err := scanGame(l.db.QueryRowContext(ctx,
		`select `+gameColumns+` from games
		 where source_kind = 'local' and (exe_key = ? or root_key = ?) and local_config is not null limit 1`,
		key, key))
	if err != nil {
		return nil
	}
	return game.LocalConfig
}

// SaveLocalConfig writes a hand-added or changed local game into the library and pins it.
func (l *Library) SaveLocalConfig(ctx context.Context, options types.LaunchOptions) error {
	row, ok := localRowFromConfig(options, nil, nowStamp())
	if !ok {
		return errors.New("a local game needs a game or launcher path")
	}
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var id int64
	err = tx.QueryRowContext(ctx,
		`select id from games where root_key = ? and source_kind = 'local'`, row.RootKey).Scan(&id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		id, err = upsertLocalTx(ctx, tx, row, true)
		if err != nil {
			return err
		}
	case err != nil:
		return err
	default:
		if _, err := tx.ExecContext(ctx, `update games set
			title=?, install_dir=?, exe_path=?, exe_args=?, local_config=?, pinned=1,
			exe_key=?, install_dir_key=?, origin=case when origin='hand' then 'hand' else origin end,
			updated_at=? where id=?`,
			row.Title, row.InstallDir, row.ExePath, nullIfEmpty(row.ExeArgs), string(row.LocalConfig),
			row.ExeKey, row.InstallDirKey, nowStamp(), id); err != nil {
			return err
		}
	}
	if err := replaceLaunchersTx(ctx, tx, id, []types.Launcher{{
		Name: row.Title, Path: row.ExePath, Kind: types.LauncherMain, Primary: true,
	}}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	l.markWritten()
	l.notify("library.changed", map[string]any{})
	return nil
}

// RemoveLocalConfig removes the local row for an executable or folder.
func (l *Library) RemoveLocalConfig(ctx context.Context, executable string) error {
	key := config.PathKey(executable)
	if key == "" {
		return nil
	}
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var id int64
	err = tx.QueryRowContext(ctx,
		`select id from games where source_kind = 'local' and (exe_key = ? or root_key = ?) limit 1`,
		key, key).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := removeLocalTx(ctx, tx, id); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	l.markWritten()
	l.notify("library.changed", map[string]any{})
	return nil
}

// SetLocalLauncher makes executable the primary launcher of the folder's local row.
func (l *Library) SetLocalLauncher(ctx context.Context, folder, executable string) error {
	rootKey := config.PathKey(folder)
	if rootKey == "" || executable == "" {
		return nil
	}
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var id int64
	err = tx.QueryRowContext(ctx, `select id from games where root_key = ? and source_kind = 'local'`, rootKey).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`update game_launchers set primary_launcher = case when path_key = ? then 1 else 0 end where game_id = ?`,
		config.PathKey(executable), id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`update games set exe_path = ?, exe_key = ?, pinned = 1, updated_at = ? where id = ?`,
		executable, config.PathKey(executable), nowStamp(), id); err != nil {
		return err
	}
	return tx.Commit()
}

// SetLocalIgnored adds or removes one ignored path.
func (l *Library) SetLocalIgnored(ctx context.Context, path string, ignored bool) error {
	key := config.PathKey(path)
	if key == "" {
		return nil
	}
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if ignored {
		if _, err := tx.ExecContext(ctx,
			`insert into ignored_paths (path_key, path, reason, created_at) values (?,?,?,?)
			 on conflict(path_key) do nothing`, key, path, "user", nowStamp()); err != nil {
			return err
		}
	} else {
		if _, err := tx.ExecContext(ctx, `delete from ignored_paths where path_key = ?`, key); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (l *Library) allLaunchers(ctx context.Context) (map[int64][]types.Launcher, error) {
	rows, err := l.db.QueryContext(ctx,
		`select game_id, name, path, kind, primary_launcher from game_launchers
		 order by game_id, primary_launcher desc, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64][]types.Launcher{}
	for rows.Next() {
		var gameID int64
		var item types.Launcher
		var primary int
		if err := rows.Scan(&gameID, &item.Name, &item.Path, &item.Kind, &primary); err != nil {
			return nil, err
		}
		item.Primary = primary == 1
		out[gameID] = append(out[gameID], item)
	}
	return out, rows.Err()
}

func localGameInfo(game Game, launchers []types.Launcher) types.GameInfo {
	folder := ""
	if game.FolderPath != nil {
		folder = *game.FolderPath
	}
	path := ""
	if len(launchers) > 0 {
		path = launchers[0].Path
	}
	if path == "" && game.ExePath != nil {
		path = *game.ExePath
	}
	if path == "" {
		path = filepath.Join(folder, game.Title+".exe")
	}
	config := game.LocalConfig
	if config == nil {
		config = &types.LaunchOptions{Name: game.Title, LauncherPath: path, GamePath: folder}
	}
	if len(launchers) == 0 {
		launchers = []types.Launcher{{Name: game.Title, Path: path, Kind: types.LauncherMain, Primary: true}}
	}
	return types.GameInfo{
		Name:          game.Title,
		Path:          path,
		Config:        *config,
		IsAutoScanned: game.Origin == "scan",
		Folder:        folder,
		Launchers:     launchers,
		LauncherCount: len(launchers),
	}
}
