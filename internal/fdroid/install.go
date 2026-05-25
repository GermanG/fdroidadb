// Copyright (C) 2026 German Gutierrez
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package fdroid

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/GermanG/fdroidadb/internal/adb"
	"github.com/GermanG/fdroidadb/internal/cli"
	"github.com/GermanG/fdroidadb/internal/db"
	"github.com/GermanG/fdroidadb/internal/logger"
	"github.com/GermanG/fdroidadb/internal/xdg"
	"github.com/schollz/progressbar/v3"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

func InstallApp(query string, device *adb.Device, repoURL string, maxRetries int) error {
	var app *db.App

	// First, try exact package name
	apps, err := db.GetAppByPackage(query)
	if err == nil && len(apps) > 0 {
		if len(apps) > 1 {
			// If no specific repoURL was provided, we must ask or pick the best one
			if repoURL == "" {
				fmt.Printf("\nApp '%s' found in multiple repositories:\n", query)
				for i, a := range apps {
					fmt.Printf("[%d] %s (via %s)\n", i+1, a.Name, a.RepoURL)
				}
				idx, err := cli.ReadInt("Select number (1-"+strconv.Itoa(len(apps))+") or 'q' to cancel: ", 1, len(apps))
				if err != nil {
					return err
				}
				app = &apps[idx-1]
			} else {
				// Try to find the app in the specific repo requested
				for _, a := range apps {
					if a.RepoURL == repoURL {
						app = &a
						break
					}
				}
				if app == nil {
					app = &apps[0]
				}
			}
		} else {
			app = &apps[0]
		}
	}

	if app == nil {
		// If not found, try searching by name/summary
		searchApps, searchErr := db.SearchApps(query)
		if searchErr != nil || len(searchApps) == 0 {
			return fmt.Errorf("app '%s' not found in database", query)
		}

		if len(searchApps) > 1 {
			// Check for exact name match among search results
			var exactMatch *db.App
			for _, a := range searchApps {
				if a.PackageName == query || a.Name == query {
					exactMatch = &a
					break
				}
			}

			if exactMatch != nil {
				app = exactMatch
				fmt.Printf("Resolved '%s' to %s (%s) [exact match]\n", query, app.Name, app.PackageName)
			} else {
				fmt.Printf("\nMultiple apps found for '%s':\n", query)
				for i, a := range searchApps {
					fmt.Printf("[%d] %s (%s) [via %s]\n    %s\n", i+1, a.Name, a.PackageName, a.RepoURL, a.Summary)
				}
				idx, err := cli.ReadInt("\nSelect number to install (1-"+strconv.Itoa(len(searchApps))+") or 'q' to cancel: ", 1, len(searchApps))
				if err != nil {
					return err
				}
				app = &searchApps[idx-1]
			}
		} else {
			app = &searchApps[0]
			fmt.Printf("Resolved '%s' to %s (%s)\n", query, app.Name, app.PackageName)
		}
	}

	// Important: Use the repoURL from the selected app entry
	actualRepoURL := app.RepoURL

	versions, err := db.GetVersions(app.ID, actualRepoURL)
	if err != nil {
		return err
	}

	var bestVersion *db.Version
	for _, v := range versions {
		if isCompatible(v.Arch, device.Arch) {
			bestVersion = &v
			break
		}
	}

	if bestVersion == nil {
		return fmt.Errorf("no compatible version found for architecture %s in repository %s", device.Arch, actualRepoURL)
	}

	logger.Info.Printf("Best version: %s (code %d) from %s", bestVersion.VersionName, bestVersion.VersionCode, actualRepoURL)

	apkPath := filepath.Join(xdg.CacheDir(), bestVersion.APKName)
	url := actualRepoURL + "/" + bestVersion.APKName

	for i := 0; i < maxRetries; i++ {
		err = downloadFileWithResume(url, apkPath, "Downloading APK")
		if err == nil {
			break
		}
		logger.Warn.Printf("Download attempt %d failed: %v. Retrying...", i+1, err)
	}
	if err != nil {
		return fmt.Errorf("failed to download APK after %d retries: %v", maxRetries, err)
	}

	fmt.Printf("Installing %s on %s... (this may take a moment)\n", bestVersion.APKName, device.Serial)
	err = device.InstallAPK(apkPath)
	if err != nil {
		return err
	}
	fmt.Println("Installation successful!")
	return nil
}

func isCompatible(verArch, devArch string) bool {
	if verArch == "" || devArch == "" || devArch == "Unknown" {
		return true // Universal or unknown device
	}
	archs := strings.Split(verArch, ",")
	for _, a := range archs {
		a = strings.TrimSpace(a)
		if a == devArch {
			return true
		}
		// Basic mapping for arm64-v8a and armeabi-v7a
		if a == "armeabi-v7a" && devArch == "arm64-v8a" {
			return true
		}
	}
	return false
}

func downloadFileWithResume(url string, path string, description string) error {
	fileInfo, err := os.Stat(path)
	var startByte int64 = 0
	if err == nil {
		startByte = fileInfo.Size()
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	if startByte > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", startByte))
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
		return nil
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	var mode os.FileMode = 0644
	var out *os.File
	if startByte > 0 && resp.StatusCode == http.StatusPartialContent {
		out, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY, mode)
	} else {
		out, err = os.Create(path)
		startByte = 0
	}
	if err != nil {
		return err
	}
	defer out.Close()

	contentLength := resp.ContentLength
	if contentLength != -1 {
		contentLength += startByte
	}

	bar := progressbar.DefaultBytes(
		contentLength,
		description,
	)
	_ = bar.Add64(startByte)

	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	_ = bar.Finish()
	fmt.Println()
	return err
}
