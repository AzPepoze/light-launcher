package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/mappu/miqt/qt6/sql"
)

// Chromium stores cookie expiry as microseconds since 1601-01-01 UTC.
const chromiumEpochOffsetMicros = 11644473600000000

type profileCookieRow struct {
	Host       string
	Name       string
	Value      string
	Path       string
	Secure     bool
	ExpiresUTC int64
	HasExpires bool
	Encrypted  bool
}

// readProfileCookies returns the cookies Chromium has already stored for the profile. Qt 6.11
// loadAllCookies emits no cookieAdded for them, but Chromium adds them to its own requests.
func readProfileCookies(profileDir string) ([]cookie, error) {
	path := filepath.Join(profileDir, "Cookies")
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("stat profile cookie database: %w", err)
	}

	connection := "profile-cookies:" + path
	db := sql.QSqlDatabase_AddDatabase2("QSQLITE", connection)
	defer func() {
		db.Close()
		db.Delete()
		sql.QSqlDatabase_RemoveDatabase(connection)
	}()
	db.SetDatabaseName(path)
	db.SetConnectOptionsWithOptions("QSQLITE_OPEN_READONLY;QSQLITE_BUSY_TIMEOUT=1000")
	if !db.Open() {
		return nil, fmt.Errorf("open profile cookie database: %s", db.LastError().Text())
	}

	query := sql.NewQSqlQuery3(db)
	defer query.Delete()
	if !query.Exec("SELECT host_key, name, value, path, is_secure, expires_utc, has_expires, length(encrypted_value) FROM cookies") {
		return nil, fmt.Errorf("query profile cookie database: %s", query.LastError().Text())
	}

	now := time.Now()
	var cookies []cookie
	for query.Next() {
		value := query.Value(2).ToString()
		row := profileCookieRow{
			Host:       query.Value(0).ToString(),
			Name:       query.Value(1).ToString(),
			Value:      value,
			Path:       query.Value(3).ToString(),
			Secure:     query.Value(4).ToBool(),
			ExpiresUTC: query.Value(5).ToLongLong(),
			HasExpires: query.Value(6).ToBool(),
			Encrypted:  value == "" && query.Value(7).ToLongLong() > 0,
		}
		if c, ok := cookieFromProfileRow(row, now); ok {
			cookies = append(cookies, c)
		}
	}
	if err := query.LastError(); err.IsValid() {
		return nil, fmt.Errorf("read profile cookie database: %s", err.Text())
	}
	return cookies, nil
}

func cookieFromProfileRow(row profileCookieRow, now time.Time) (cookie, bool) {
	if row.Name == "" || row.Encrypted {
		return cookie{}, false
	}
	if row.HasExpires && row.ExpiresUTC-chromiumEpochOffsetMicros <= now.UnixMicro() {
		return cookie{}, false
	}
	return cookie{
		Name:   row.Name,
		Value:  row.Value,
		Domain: row.Host,
		Path:   row.Path,
		Secure: row.Secure,
	}, true
}
