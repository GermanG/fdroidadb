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
}

var DB *sql.DB

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
			package_name TEXT UNIQUE,
			name TEXT,
			summary TEXT,
			description TEXT,
			icon TEXT,
			signer TEXT
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
			PRIMARY KEY(app_id, version_code),
			FOREIGN KEY(app_id) REFERENCES apps(id)
		)`,
		`CREATE TABLE IF NOT EXISTS repos (
			url TEXT PRIMARY KEY,
			last_index_hash TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_versions_app_id ON versions(app_id)`,
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
	}

	return nil
}

func SaveApp(app App) (int, error) {
	res, err := DB.Exec(`INSERT OR REPLACE INTO apps (package_name, name, summary, description, icon, signer) 
		VALUES (?, ?, ?, ?, ?, ?)`, app.PackageName, app.Name, app.Summary, app.Description, app.Icon, app.Signer)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %v", err)
	}
	return int(id), nil
}

func SaveVersion(v Version) error {
	_, err := DB.Exec(`INSERT OR REPLACE INTO versions (app_id, version_name, version_code, min_sdk, target_sdk, size, hash, apk_name, arch)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, v.AppID, v.VersionName, v.VersionCode, v.MinSDK, v.TargetSDK, v.Size, v.Hash, v.APKName, v.Arch)
	return err
}

func GetAppByPackage(packageName string) (*App, error) {
	var a App
	err := DB.QueryRow("SELECT id, package_name, name, summary, description, icon, signer FROM apps WHERE package_name = ?", packageName).
		Scan(&a.ID, &a.PackageName, &a.Name, &a.Summary, &a.Description, &a.Icon, &a.Signer)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func GetVersions(appID int) ([]Version, error) {
	rows, err := DB.Query("SELECT app_id, version_name, version_code, min_sdk, target_sdk, size, hash, apk_name, arch FROM versions WHERE app_id = ? ORDER BY version_code DESC", appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []Version
	for rows.Next() {
		var v Version
		if err := rows.Scan(&v.AppID, &v.VersionName, &v.VersionCode, &v.MinSDK, &v.TargetSDK, &v.Size, &v.Hash, &v.APKName, &v.Arch); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, nil
}

func SearchApps(query string) ([]App, error) {
	rows, err := DB.Query("SELECT id, package_name, name, summary FROM apps WHERE name LIKE ? OR package_name LIKE ?", "%"+query+"%", "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []App
	for rows.Next() {
		var a App
		if err := rows.Scan(&a.ID, &a.PackageName, &a.Name, &a.Summary); err != nil {
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
	_, err := DB.Exec("INSERT OR REPLACE INTO repos (url, last_index_hash) VALUES (?, ?)", url, hash)
	return err
}
