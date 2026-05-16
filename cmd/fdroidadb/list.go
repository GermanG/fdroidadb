package main

import (
	"fmt"

	"fdroidadb/internal/adb"
	"fdroidadb/internal/config"
	"fdroidadb/internal/db"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed applications and updates",
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

		fmt.Printf("Installed F-Droid Apps on %s:\n", device.Serial)
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
			
			if len(versions) > 0 {
				latest := versions[0]
				status := ""
				if latest.VersionCode > currentCode {
					status = " [UPDATE AVAILABLE]"
				}
				fmt.Printf("%s (%s)\n  Installed: %d, Latest: %d (%s)%s\n", app.Name, app.PackageName, currentCode, latest.VersionCode, latest.VersionName, status)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
