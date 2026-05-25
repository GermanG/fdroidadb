// Copyright (C) 2026 German Gutierrez
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package main

import (
	"fmt"

	"github.com/GermanG/fdroidadb/internal/adb"
	"github.com/GermanG/fdroidadb/internal/config"
	"github.com/GermanG/fdroidadb/internal/db"
	"github.com/GermanG/fdroidadb/internal/fdroid"
	"github.com/spf13/cobra"
)

var dryRun bool

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update all installed applications",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if err := db.Init(); err != nil {
			return err
		}

		if err := adb.EnsureServer(cfg.ADBPath); err != nil {
			return err
		}

		// Auto-sync before update
		fmt.Println("\n=== Synchronizing Repositories ===")
		for _, repo := range cfg.Repos {
			fmt.Printf("\nSyncing %s...\n", repo.Name)
			if err := fdroid.SyncRepo(repo.URL); err != nil {
				fmt.Printf("Warning: auto-sync failed for %s: %v\n", repo.Name, err)
			}
		}

		fmt.Println("\n=== Checking for Updates ===")
		device, err := adb.SelectDevice(mockMode)
		if err != nil {
			return err
		}

		packages, err := device.GetInstalledPackages()
		if err != nil {
			return err
		}

		updateCount := 0
		for _, pkg := range packages {
			// Get all instances of this app across all repos
			apps, err := db.GetAppByPackage(pkg)
			if err != nil || len(apps) == 0 {
				continue
			}

			currentCode, err := device.GetPackageVersion(pkg)
			if err != nil {
				currentCode = 0
			}

			var bestVersion *db.Version
			var bestRepoURL string

			for _, app := range apps {
				versions, err := db.GetVersions(app.ID, app.RepoURL)
				if err != nil || len(versions) == 0 {
					continue
				}

				latest := versions[0]
				if latest.VersionCode > currentCode {
					if bestVersion == nil || latest.VersionCode > bestVersion.VersionCode {
						v := latest
						bestVersion = &v
						bestRepoURL = app.RepoURL
					}
				}
			}

			if bestVersion != nil {
				if dryRun {
					updateCount++
					fmt.Printf("[DRY-RUN] Would update %s (%s) from %d to %d (via %s)\n", apps[0].Name, pkg, currentCode, bestVersion.VersionCode, bestRepoURL)
				} else {
					fmt.Printf("Updating %s (%s) from %d to %d (via %s)...\n", apps[0].Name, pkg, currentCode, bestVersion.VersionCode, bestRepoURL)
					if err := fdroid.InstallApp(pkg, device, bestRepoURL, cfg.MaxRetries); err != nil {
						fmt.Printf("Failed to update %s: %v\n", pkg, err)
					}
				}
			}
		}

		if dryRun {
			if updateCount > 0 {
				fmt.Printf("\nDry-run completed. %d updates available.\n", updateCount)
			} else {
				fmt.Println("\nAll applications are up to date.")
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "display updates without installing them")
}

