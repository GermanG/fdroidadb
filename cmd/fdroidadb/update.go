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

			currentCode, _ := device.GetPackageVersion(pkg)
			versions, _ := db.GetVersions(app.ID)

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
