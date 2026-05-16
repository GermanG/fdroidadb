package fdroid

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"fdroidadb/internal/db"
	"fdroidadb/internal/logger"
	"fdroidadb/internal/xdg"
	"go.mozilla.org/pkcs7"
)

func verifyJar(path string) error {
	r, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer r.Close()

	var sfContent, sigContent []byte
	for _, f := range r.File {
		if strings.HasPrefix(f.Name, "META-INF/") {
			if strings.HasSuffix(f.Name, ".SF") {
				rc, _ := f.Open()
				sfContent, _ = io.ReadAll(rc)
				rc.Close()
			}
			if strings.HasSuffix(f.Name, ".RSA") || strings.HasSuffix(f.Name, ".DSA") || strings.HasSuffix(f.Name, ".EC") {
				rc, _ := f.Open()
				sigContent, _ = io.ReadAll(rc)
				rc.Close()
			}
		}
	}

	if sfContent == nil || sigContent == nil {
		return fmt.Errorf("signature files not found in JAR")
	}

	p7, err := pkcs7.Parse(sigContent)
	if err != nil {
		return fmt.Errorf("failed to parse signature: %v", err)
	}

	p7.Content = sfContent
	return p7.Verify()
}

func SyncRepo(repoURL string) error {
	logger.Info.Printf("Syncing repo: %s", repoURL)

	jarPath := filepath.Join(xdg.CacheDir(), "index-v1.jar")
	err := downloadFile(repoURL+"/index-v1.jar", jarPath)
	if err != nil {
		return fmt.Errorf("failed to download index: %v", err)
	}

	// Verify signature
	err = verifyJar(jarPath)
	if err != nil {
		return fmt.Errorf("index signature verification failed: %v", err)
	}

	// Extract index-v1.json
	r, err := zip.OpenReader(jarPath)
	if err != nil {
		return err
	}
	defer r.Close()

	var indexData []byte
	for _, f := range r.File {
		if f.Name == "index-v1.json" {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			indexData, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return err
			}
			break
		}
	}

	if len(indexData) == 0 {
		return fmt.Errorf("index-v1.json not found in JAR")
	}

	var index IndexV1
	if err := json.Unmarshal(indexData, &index); err != nil {
		return fmt.Errorf("failed to parse index JSON: %v", err)
	}

	// Store in DB
	count := 0
	for _, app := range index.Apps {
		name := app.Name
		summary := app.Summary
		description := app.Description

		// Fallback to English if available in localized
		if loc, ok := app.Localized["en-US"]; ok {
			if loc.Name != "" {
				name = loc.Name
			}
			if loc.Summary != "" {
				summary = loc.Summary
			}
			if loc.Description != "" {
				description = loc.Description
			}
		} else if loc, ok := app.Localized["en"]; ok {
			if loc.Name != "" {
				name = loc.Name
			}
			if loc.Summary != "" {
				summary = loc.Summary
			}
			if loc.Description != "" {
				description = loc.Description
			}
		}

		dbApp := db.App{
			PackageName: app.PackageName,
			Name:        name,
			Summary:     summary,
			Description: description,
			Icon:        app.Icon,
		}
		appID, err := db.SaveApp(dbApp)
		if err != nil {
			logger.Error.Printf("Failed to save app %s: %v", app.PackageName, err)
			continue
		}

		if pkgs, ok := index.Packages[app.PackageName]; ok {
			for _, pkg := range pkgs {
				dbVer := db.Version{
					AppID:       appID,
					VersionName: pkg.VersionName,
					VersionCode: pkg.VersionCode,
					MinSDK:      pkg.MinSDK,
					TargetSDK:   pkg.TargetSDK,
					Size:        pkg.Size,
					Hash:        pkg.Hash,
					APKName:     pkg.APKName,
					Arch:        strings.Join(pkg.NativeCode, ","),
				}
				if err := db.SaveVersion(dbVer); err != nil {
					logger.Warn.Printf("Failed to save version for %s: %v", app.PackageName, err)
				}
			}
		}
		count++
	}

	fmt.Printf("Synced %d apps.\n", count)
	return nil
}

func downloadFile(url string, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
