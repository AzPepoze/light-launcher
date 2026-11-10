package downloads

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// SQLiteStore keeps downloads in the `downloads` and `segments` tables. The caller opens the *sql.DB
// and registers the driver.
type SQLiteStore struct {
	db *sql.DB
}

const sqliteDownloadsTable = `
create table if not exists downloads (
  id integer primary key autoincrement,
  game_id integer not null default 0,
  url text not null,
  filename text not null,
  dest text not null,
  size integer not null default 0,
  bytes_done integer not null default 0,
  connections integer not null default 1,
  status text not null default 'queued',
  error text,
  etag text,
  headers text,
  install integer not null default 0,
  created_at text not null,
  started_at text,
  finished_at text
)`

const sqliteSegmentsTable = `
create table if not exists segments (
  download_id integer not null references downloads(id) on delete cascade,
  idx integer not null,
  start integer not null,
  "end" integer not null,
  done integer not null default 0,
  primary key (download_id, idx)
)`

const downloadColumns = `id, game_id, url, filename, dest, size, bytes_done, connections, status, error, etag, headers, install, created_at, started_at, finished_at`

const updateDownloadSQL = `
update downloads set
  status = coalesce(?, status),
  error = coalesce(?, error),
  size = coalesce(?, size),
  bytes_done = coalesce(?, bytes_done),
  connections = coalesce(?, connections),
  etag = coalesce(?, etag),
  started_at = coalesce(?, started_at),
  finished_at = coalesce(?, finished_at)
where id = ?`

// NewSQLiteStore creates the tables if they are missing.
func NewSQLiteStore(ctx context.Context, db *sql.DB) (*SQLiteStore, error) {
	for _, stmt := range []string{sqliteDownloadsTable, sqliteSegmentsTable} {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return nil, fmt.Errorf("create download tables: %w", err)
		}
	}
	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) Insert(ctx context.Context, d Download) (int64, error) {
	headers, err := encodeHeaders(d.Headers)
	if err != nil {
		return 0, err
	}
	created := d.CreatedAt
	if created.IsZero() {
		created = time.Now()
	}
	res, err := s.db.ExecContext(ctx, `
insert into downloads (game_id, url, filename, dest, size, bytes_done, connections, status, error, etag, headers, install, created_at, started_at, finished_at)
values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		d.GameID, d.URL, d.Filename, d.Dest, d.Size, d.BytesDone, d.Connections, string(d.State),
		nullString(d.Error), nullString(d.ETag), headers, boolInt(d.Install),
		formatTime(created), optionalTime(d.StartedAt), optionalTime(d.FinishedAt),
	)
	if err != nil {
		return 0, fmt.Errorf("insert download: %w", err)
	}
	return res.LastInsertId()
}

func (s *SQLiteStore) Get(ctx context.Context, id int64) (Download, error) {
	row := s.db.QueryRowContext(ctx, `select `+downloadColumns+` from downloads where id = ?`, id)
	d, err := scanDownload(row)
	if err == sql.ErrNoRows {
		return Download{}, ErrNotFound
	}
	return d, err
}

func (s *SQLiteStore) List(ctx context.Context) ([]Download, error) {
	rows, err := s.db.QueryContext(ctx, `select `+downloadColumns+` from downloads order by id desc`)
	if err != nil {
		return nil, fmt.Errorf("list downloads: %w", err)
	}
	defer rows.Close()
	var out []Download
	for rows.Next() {
		d, err := scanDownload(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *SQLiteStore) Update(ctx context.Context, id int64, p Patch) error {
	res, err := s.db.ExecContext(ctx, updateDownloadSQL,
		stateArg(p.State), ptrArg(p.Error), ptrArg(p.Size), ptrArg(p.BytesDone),
		ptrArg(p.Connections), ptrArg(p.ETag), timeArg(p.StartedAt), timeArg(p.FinishedAt), id,
	)
	if err != nil {
		return fmt.Errorf("update download %d: %w", id, err)
	}
	return requireRow(res)
}

func (s *SQLiteStore) Delete(ctx context.Context, id int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `delete from segments where download_id = ?`, id); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `delete from downloads where id = ?`, id)
	if err != nil {
		return err
	}
	if err := requireRow(res); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *SQLiteStore) SaveSegments(ctx context.Context, id int64, segments []Segment) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `delete from segments where download_id = ?`, id); err != nil {
		return err
	}
	for _, seg := range segments {
		_, err := tx.ExecContext(ctx,
			`insert into segments (download_id, idx, start, "end", done) values (?,?,?,?,?)`,
			id, seg.Index, seg.Start, seg.End, seg.Done)
		if err != nil {
			return fmt.Errorf("save segment %d of download %d: %w", seg.Index, id, err)
		}
	}
	return tx.Commit()
}

func (s *SQLiteStore) Segments(ctx context.Context, id int64) ([]Segment, error) {
	rows, err := s.db.QueryContext(ctx,
		`select idx, start, "end", done from segments where download_id = ? order by idx`, id)
	if err != nil {
		return nil, fmt.Errorf("load segments: %w", err)
	}
	defer rows.Close()
	var out []Segment
	for rows.Next() {
		var seg Segment
		if err := rows.Scan(&seg.Index, &seg.Start, &seg.End, &seg.Done); err != nil {
			return nil, err
		}
		out = append(out, seg)
	}
	return out, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanDownload(row rowScanner) (Download, error) {
	var (
		d                      Download
		state, created         string
		errText, etag, headers sql.NullString
		started, finished      sql.NullString
		install                int
	)
	err := row.Scan(&d.ID, &d.GameID, &d.URL, &d.Filename, &d.Dest, &d.Size, &d.BytesDone,
		&d.Connections, &state, &errText, &etag, &headers, &install, &created, &started, &finished)
	if err != nil {
		return Download{}, err
	}
	d.State = State(state)
	d.Error = errText.String
	d.ETag = etag.String
	d.Install = install != 0
	if d.CreatedAt, err = parseTime(created); err != nil {
		return Download{}, err
	}
	if d.StartedAt, err = parseOptionalTime(started); err != nil {
		return Download{}, err
	}
	if d.FinishedAt, err = parseOptionalTime(finished); err != nil {
		return Download{}, err
	}
	d.Headers = decodeHeaders(headers)
	return d, nil
}

func requireRow(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func encodeHeaders(h map[string]string) (any, error) {
	if len(h) == 0 {
		return nil, nil
	}
	data, err := json.Marshal(h)
	if err != nil {
		return nil, fmt.Errorf("encode headers: %w", err)
	}
	return string(data), nil
}

// decodeHeaders ignores a damaged value rather than failing the whole row.
func decodeHeaders(v sql.NullString) map[string]string {
	if !v.Valid || v.String == "" {
		return nil
	}
	var h map[string]string
	if err := json.Unmarshal([]byte(v.String), &h); err != nil {
		return nil
	}
	return h
}

func nullString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func optionalTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return formatTime(t)
}

func parseTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, s)
}

func parseOptionalTime(v sql.NullString) (time.Time, error) {
	if !v.Valid || v.String == "" {
		return time.Time{}, nil
	}
	return parseTime(v.String)
}

func stateArg(s *State) any {
	if s == nil {
		return nil
	}
	return string(*s)
}

func timeArg(t *time.Time) any {
	if t == nil {
		return nil
	}
	return formatTime(*t)
}

// ptrArg turns an optional value into a SQL argument; nil keeps the column as it is.
func ptrArg[T any](p *T) any {
	if p == nil {
		return nil
	}
	return *p
}
