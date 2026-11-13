package library

// localSourcesMigration makes the library the single list for local and plugin games.
// Local rows keep their launch options, folder identity and scan state here; the JSON
// config files become import-only data after the backfill.
var localSourcesMigration = migration{
	version: 5,
	statements: []string{
		`alter table games add column origin text not null default 'plugin'`,
		`alter table games add column source_kind text not null default 'plugin'`,
		`alter table games add column folder_path text`,
		`alter table games add column scan_folder text`,
		`alter table games add column root_key text`,
		`alter table games add column exe_key text`,
		`alter table games add column install_dir_key text`,
		`alter table games add column scan_folder_key text`,
		`alter table games add column last_seen_at text`,
		`alter table games add column missing_since text`,
		`alter table games add column miss_count integer not null default 0`,
		`alter table games add column pinned integer not null default 0`,
		`alter table games add column local_config text`,
		`create table if not exists sources (
			id text primary key,
			kind text not null,
			label text not null,
			plugin_id text,
			cover text,
			tile text,
			enabled integer not null default 1,
			first_seen_at text not null,
			last_sync_at text,
			last_error text
		)`,
		`create table if not exists game_launchers (
			id integer primary key autoincrement,
			game_id integer not null references games(id) on delete cascade,
			path text not null,
			path_key text not null,
			name text not null,
			kind text not null,
			primary_launcher integer not null default 0,
			local_config text,
			unique(game_id, path_key)
		)`,
		`create index if not exists game_launchers_game on game_launchers(game_id)`,
		`create table if not exists ignored_paths (
			path_key text primary key,
			path text not null,
			reason text not null,
			created_at text not null
		)`,
		`create unique index if not exists games_local_root on games(root_key)
			where root_key is not null and source_kind = 'local'`,
		`create unique index if not exists games_install_dir_key on games(install_dir_key)
			where install_dir_key is not null`,
		`create index if not exists games_origin on games(origin, hidden)`,
		`create index if not exists games_scan_folder on games(scan_folder_key, hidden)`,
	},
}
