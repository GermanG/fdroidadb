// Copyright (C) 2026 German Gutierrez
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/GermanG/fdroidadb/internal/xdg"
	_ "modernc.org/sqlite"
)

type App struct {
	ID          int
	PackageName string
	Name        string
	Summary     string
	Description string
	Icon        string
	Signer      string // Certificate fingerprint
	RepoURL     string
}

type Version struct {
	AppID       int
	VersionName string
	VersionCode int
	MinSDK      int
	TargetSDK   int
	Size        int64
	Hash        string
	APKName     string
	Arch        string // Comma separated list of architectures
	RepoURL     string
}

var DB *sql.DB

func BeginTx() (*sql.Tx, error) {
	return DB.Begin()
}

func Init() error {
	dataDir := xdg.DataDir()
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}
	dbPath := filepath.Join(dataDir, "fdroidadb.db")

	var err error
	DB, err = sql.Open("sqlite", dbPath+"?_pragma=journal_mode=WAL")
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	return createTables()
}

func createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS apps (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			package_name TEXT,
			name TEXT,
			summary TEXT,
			description TEXT,
			icon TEXT,
			signer TEXT,
			repo_url TEXT,
			UNIQUE(package_name, repo_url)
		)`,
		`CREATE TABLE IF NOT EXISTS versions (
			app_id INTEGER,
			version_name TEXT,
			version_code INTEGER,
			min_sdk INTEGER,
			target_sdk INTEGER,
			size INTEGER,
			hash TEXT,
			apk_name TEXT,
			arch TEXT,
			repo_url TEXT,
			PRIMARY KEY(app_id, version_code, repo_url),
			FOREIGN KEY(app_id) REFERENCES apps(id)
		)`,
		`CREATE TABLE IF NOT EXISTS repos (
			url TEXT PRIMARY KEY,
			last_index_hash TEXT,
			fingerprint TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_versions_app_id ON versions(app_id)`,
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
	}

	// Migration check for repo_url column
	var count int
	err := DB.QueryRow("SELECT count(*) FROM pragma_table_info('apps') WHERE name='repo_url'").Scan(&count)
	if err == nil && count == 0 {
		fmt.Println("Migrating database schema (adding repo_url support)...")
		// Drop and recreate is cleaner for development when constraints change
		_, _ = DB.Exec("DROP TABLE IF EXISTS versions")
		_, _ = DB.Exec("DROP TABLE IF EXISTS apps")
		_, _ = DB.Exec("DELETE FROM repos") // Force full re-sync
		return createTables()               // Recurse once to recreate
	}

	return nil
}

func SaveApp(app App) (int, error) {
	return SaveAppTx(nil, app)
}

func SaveAppTx(tx *sql.Tx, app App) (int, error) {
	var execer interface {
		Exec(query string, args ...any) (sql.Result, error)
		QueryRow(query string, args ...any) *sql.Row
	} = DB

	if tx != nil {
		execer = tx
	}

	res, err := execer.Exec(`INSERT INTO apps (package_name, name, summary, description, icon, signer, repo_url) 
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(package_name, repo_url) DO UPDATE SET
			name = excluded.name,
			summary = excluded.summary,
			description = excluded.description,
			icon = excluded.icon,
			signer = excluded.signer`, 
		app.PackageName, app.Name, app.Summary, app.Description, app.Icon, app.Signer, app.RepoURL)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %v", err)
	}

	// In SQLite, if the row was updated instead of inserted, LastInsertId() returns
	// the rowid of the updated row. However, to be 100% robust across all environments,
	// if it returns 0, we query the ID.
	if id == 0 {
		err = execer.QueryRow("SELECT id FROM apps WHERE package_name = ? AND repo_url = ?", app.PackageName, app.RepoURL).Scan(&id)
		if err != nil {
			return 0, fmt.Errorf("failed to get app id after upsert: %v", err)
		}
	}

	return int(id), nil
}

func SaveVersion(v Version) error {
	return SaveVersionTx(nil, v)
}

func SaveVersionTx(tx *sql.Tx, v Version) error {
	var execer interface {
		Exec(query string, args ...any) (sql.Result, error)
	} = DB

	if tx != nil {
		execer = tx
	}

	_, err := execer.Exec(`INSERT OR REPLACE INTO versions (app_id, version_name, version_code, min_sdk, target_sdk, size, hash, apk_name, arch, repo_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, v.AppID, v.VersionName, v.VersionCode, v.MinSDK, v.TargetSDK, v.Size, v.Hash, v.APKName, v.Arch, v.RepoURL)
	return err
}

func GetAppByPackage(packageName string) ([]App, error) {
	rows, err := DB.Query("SELECT id, package_name, name, summary, description, icon, signer, repo_url FROM apps WHERE package_name = ?", packageName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []App
	for rows.Next() {
		var a App
		if err := rows.Scan(&a.ID, &a.PackageName, &a.Name, &a.Summary, &a.Description, &a.Icon, &a.Signer, &a.RepoURL); err != nil {
			return nil, err
		}
		apps = append(apps, a)
	}
	return apps, nil
}

func GetVersions(appID int, repoURL string) ([]Version, error) {
	rows, err := DB.Query("SELECT app_id, version_name, version_code, min_sdk, target_sdk, size, hash, apk_name, arch, repo_url FROM versions WHERE app_id = ? AND repo_url = ? ORDER BY version_code DESC", appID, repoURL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []Version
	for rows.Next() {
		var v Version
		if err := rows.Scan(&v.AppID, &v.VersionName, &v.VersionCode, &v.MinSDK, &v.TargetSDK, &v.Size, &v.Hash, &v.APKName, &v.Arch, &v.RepoURL); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, nil
}

func SearchApps(query string) ([]App, error) {
	// Search in name, package_name AND summary
	// We don't GROUP BY here so we can see all repos that have the app,
	// or we handle duplicates in the UI.
	rows, err := DB.Query("SELECT id, package_name, name, summary, repo_url FROM apps WHERE name LIKE ? OR package_name LIKE ? OR summary LIKE ?", "%"+query+"%", "%"+query+"%", "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []App
	for rows.Next() {
		var a App
		if err := rows.Scan(&a.ID, &a.PackageName, &a.Name, &a.Summary, &a.RepoURL); err != nil {
			return nil, err
		}
		apps = append(apps, a)
	}
	return apps, nil
}

func GetRepoHash(url string) string {
	var hash string
	err := DB.QueryRow("SELECT last_index_hash FROM repos WHERE url = ?", url).Scan(&hash)
	if err != nil {
		return ""
	}
	return hash
}

func UpdateRepoHash(url string, hash string) error {
	_, err := DB.Exec("INSERT OR REPLACE INTO repos (url, last_index_hash) VALUES (?, (SELECT last_index_hash FROM repos WHERE url = ?), (SELECT fingerprint FROM repos WHERE url = ?))", url, url, url)
	// That's too complex, let's just use a simple UPSERT if possible or two steps
	_, err = DB.Exec("INSERT INTO repos (url, last_index_hash) VALUES (?, ?) ON CONFLICT(url) DO UPDATE SET last_index_hash=excluded.last_index_hash", url, hash)
	return err
}

func SetRepoFingerprint(url string, fingerprint string) error {
	_, err := DB.Exec("INSERT INTO repos (url, fingerprint) VALUES (?, ?) ON CONFLICT(url) DO UPDATE SET fingerprint=excluded.fingerprint", url, fingerprint)
	return err
}

func GetRepoFingerprint(url string) string {
	var fp string
	err := DB.QueryRow("SELECT fingerprint FROM repos WHERE url = ?", url).Scan(&fp)
	if err != nil {
		return ""
	}
	return fp
}
