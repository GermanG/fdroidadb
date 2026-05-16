package main

import (
	"fmt"

	"fdroidadb/internal/adb"
	"fdroidadb/internal/config"
	"fdroidadb/internal/db"
	"fdroidadb/internal/fdroid"
	"github.com/spf13/cobra"
)

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

		repoURL := cfg.Repos[0].URL

		for _, pkg := range packages {
			app, err := db.GetAppByPackage(pkg)
			if err != nil {
				continue
			}

			currentCode, err := device.GetPackageVersion(pkg)
			if err != nil {
				currentCode = 0
			}
			versions, err := db.GetVersions(app.ID)
			if err != nil {
				continue
			}

			if len(versions) > 0 && versions[0].VersionCode > currentCode {
				fmt.Printf("Updating %s (%s) from %d to %d...\n", app.Name, app.PackageName, currentCode, versions[0].VersionCode)
				if err := fdroid.InstallApp(pkg, device, repoURL, cfg.MaxRetries); err != nil {
					if err.Error() == "signature mismatch" {
						// Already printed explanation in InstallApp
						continue
					}
					fmt.Printf("Failed to update %s: %v\n", pkg, err)
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
