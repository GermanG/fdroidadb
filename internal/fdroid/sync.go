// Copyright (C) 2026 German Gutierrez
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package fdroid

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/GermanG/fdroidadb/internal/db"
	"github.com/GermanG/fdroidadb/internal/logger"
	"github.com/GermanG/fdroidadb/internal/xdg"
	"go.mozilla.org/pkcs7"
	"github.com/schollz/progressbar/v3"
)

func verifyJar(path string, expectedFingerprint string) error {
	r, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer r.Close()

	var sfContent, sigContent []byte
	for _, f := range r.File {
		if strings.HasPrefix(f.Name, "META-INF/") {
			if strings.HasSuffix(f.Name, ".SF") {
				rc, err := f.Open()
				if err != nil {
					return fmt.Errorf("failed to open .SF file in JAR: %v", err)
				}
				sfContent, err = io.ReadAll(rc)
				rc.Close()
				if err != nil {
					return fmt.Errorf("failed to read .SF file in JAR: %v", err)
				}
			}
			if strings.HasSuffix(f.Name, ".RSA") || strings.HasSuffix(f.Name, ".DSA") || strings.HasSuffix(f.Name, ".EC") {
				rc, err := f.Open()
				if err != nil {
					return fmt.Errorf("failed to open signature file in JAR: %v", err)
				}
				sigContent, err = io.ReadAll(rc)
				rc.Close()
				if err != nil {
					return fmt.Errorf("failed to read signature file in JAR: %v", err)
				}
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
	if err := p7.Verify(); err != nil {
		return err
	}

	// Verify against fingerprint if provided
	if expectedFingerprint != "" {
		cert := p7.Certificates[0]
		fingerprint := sha256.Sum256(cert.Raw)
		fpStr := fmt.Sprintf("%x", fingerprint)
		if strings.ToLower(fpStr) != strings.ToLower(expectedFingerprint) {
			return fmt.Errorf("repository fingerprint mismatch!\n  Expected: %s\n  Found:    %s", expectedFingerprint, fpStr)
		}
	}

	return nil
}

func SyncRepo(repoURL string) error {
	logger.Info.Printf("Syncing repo: %s", repoURL)

	err := syncV2(repoURL)
	if err == nil {
		return nil
	}
	logger.Warn.Printf("V2 sync failed, falling back to V1: %v", err)

	return syncV1(repoURL)
}

func syncV2(repoURL string) error {
	entryJarPath := filepath.Join(xdg.CacheDir(), "entry.jar")
	err := downloadFile(repoURL+"/entry.jar", entryJarPath, "Downloading entry.jar")
	if err != nil {
		return err
	}

	// Get expected fingerprint from DB
	expectedFingerprint := db.GetRepoFingerprint(repoURL)

	if err := verifyJar(entryJarPath, expectedFingerprint); err != nil {
		return fmt.Errorf("entry.jar verification failed: %v", err)
	}

	r, err := zip.OpenReader(entryJarPath)
	if err != nil {
		return err
	}
	defer r.Close()

	var entryData []byte
	for _, f := range r.File {
		if f.Name == "entry.json" {
			rc, err := f.Open()
			if err != nil {
				return fmt.Errorf("failed to open entry.json in JAR: %v", err)
			}
			entryData, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return fmt.Errorf("failed to read entry.json in JAR: %v", err)
			}
			break
		}
	}

	var entry EntryV2
	if err := json.Unmarshal(entryData, &entry); err != nil {
		return err
	}

	currentHash := db.GetRepoHash(repoURL)
	if currentHash != "" && currentHash == entry.Index.SHA256 {
		fmt.Printf("Repository up to date (hash: %s...). Skipping full sync.\n", entry.Index.SHA256[:8])
		return nil
	}

	indexV2Path := filepath.Join(xdg.CacheDir(), "index-v2.json")
	err = downloadFile(repoURL+"/index-v2.json", indexV2Path, "Downloading index-v2.json")
	if err != nil {
		return err
	}

	data, err := os.ReadFile(indexV2Path)
	if err != nil {
		return err
	}

	var index IndexV2
	if err := json.Unmarshal(data, &index); err != nil {
		return err
	}

	tx, err := db.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	bar := progressbar.Default(int64(len(index.Packages)), "Updating database (V2)")
	count := 0
	for pkgName, pkg := range index.Packages {
		name := getBestString(pkg.Metadata.Name)
		summary := getBestString(pkg.Metadata.Summary)
		description := getBestString(pkg.Metadata.Description)

		signer := ""
		for _, ver := range pkg.Versions {
			if ver.Manifest.Signer != nil && len(ver.Manifest.Signer.SHA256) > 0 {
				signer = ver.Manifest.Signer.SHA256[0]
				break
			}
		}

		dbApp := db.App{
			PackageName: pkgName,
			Name:        name,
			Summary:     summary,
			Description: description,
			Icon:        getBestIcon(pkg.Metadata.Icon),
			Signer:      signer,
			RepoURL:     repoURL,
		}
		appID, err := db.SaveAppTx(tx, dbApp)
		if err != nil {
			bar.Add(1)
			continue
		}

		for _, ver := range pkg.Versions {
			minSDK := 0
			targetSDK := 0
			if ver.Manifest.UsesSDK != nil {
				minSDK = ver.Manifest.UsesSDK.MinSDK
				targetSDK = ver.Manifest.UsesSDK.TargetSDK
			}

			dbVer := db.Version{
				AppID:       appID,
				VersionName: ver.Manifest.VersionName,
				VersionCode: ver.Manifest.VersionCode,
				MinSDK:      minSDK,
				TargetSDK:   targetSDK,
				Size:        ver.File.Size,
				Hash:        ver.File.SHA256,
				APKName:     ver.File.Name,
				Arch:        strings.Join(ver.Manifest.NativeCode, ","),
				RepoURL:     repoURL,
			}
			if err := db.SaveVersionTx(tx, dbVer); err != nil {
				logger.Warn.Printf("Failed to save version for %s: %v", pkgName, err)
			}
		}
		count++
		bar.Add(1)
	}
	_ = bar.Finish()

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	fmt.Printf("\nSynced %d apps (V2).\n", count)

	db.UpdateRepoHash(repoURL, entry.Index.SHA256)

	return nil
}

func getBestString(m map[string]string) string {
	if s, ok := m["en-US"]; ok { return s }
	if s, ok := m["en"]; ok { return s }
	for _, s := range m { return s }
	return ""
}

func getBestIcon(m map[string]LocalizedFileV2) string {
	if f, ok := m["en-US"]; ok { return f.Name }
	if f, ok := m["en"]; ok { return f.Name }
	for _, f := range m { return f.Name }
	return ""
}

func syncV1(repoURL string) error {
	jarPath := filepath.Join(xdg.CacheDir(), "index-v1.jar")
	err := downloadFile(repoURL+"/index-v1.jar", jarPath, "Downloading index (V1)")
	if err != nil {
		return fmt.Errorf("failed to download index: %v", err)
	}

	expectedFingerprint := db.GetRepoFingerprint(repoURL)
	err = verifyJar(jarPath, expectedFingerprint)
	if err != nil {
		return fmt.Errorf("index signature verification failed: %v", err)
	}

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
				return fmt.Errorf("failed to open index-v1.json in JAR: %v", err)
			}
			indexData, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return fmt.Errorf("failed to read index-v1.json in JAR: %v", err)
			}
			break
		}
	}

	var index IndexV1
	if err := json.Unmarshal(indexData, &index); err != nil {
		return err
	}

	tx, err := db.BeginTx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	bar := progressbar.Default(int64(len(index.Apps)), "Updating database (V1)")
	count := 0
	for _, app := range index.Apps {
		name := app.Name
		summary := app.Summary
		description := app.Description

		if loc, ok := app.Localized["en-US"]; ok {
			if loc.Name != "" { name = loc.Name }
			if loc.Summary != "" { summary = loc.Summary }
			if loc.Description != "" { description = loc.Description }
		} else if loc, ok := app.Localized["en"]; ok {
			if loc.Name != "" { name = loc.Name }
			if loc.Summary != "" { summary = loc.Summary }
			if loc.Description != "" { description = loc.Description }
		}

		signer := ""
		if pkgs, ok := index.Packages[app.PackageName]; ok && len(pkgs) > 0 {
			signer = pkgs[0].Signer
		}

		dbApp := db.App{
			PackageName: app.PackageName,
			Name:        name,
			Summary:     summary,
			Description: description,
			Icon:        app.Icon,
			Signer:      signer,
			RepoURL:     repoURL,
		}
		appID, err := db.SaveAppTx(tx, dbApp)
		if err != nil {
			bar.Add(1)
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
					RepoURL:     repoURL,
				}
				if err := db.SaveVersionTx(tx, dbVer); err != nil {
					logger.Warn.Printf("Failed to save version for %s: %v", app.PackageName, err)
				}
			}
		}
		count++
		bar.Add(1)
	}
	_ = bar.Finish()

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	fmt.Printf("\nSynced %d apps (V1).\n", count)
	return nil
}

func downloadFile(url string, path string, description string) error {
	resp, err := httpClient.Get(url)
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

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		description,
	)

	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	_ = bar.Finish()
	fmt.Println()
	return err
}
