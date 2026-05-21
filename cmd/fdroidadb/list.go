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
			apps, err := db.GetAppByPackage(pkg)
			if err != nil || len(apps) == 0 {
				continue
			}

			currentCode, err := device.GetPackageVersion(pkg)
			if err != nil {
				currentCode = 0
			}

			var bestVersion *db.Version
			for _, app := range apps {
				versions, err := db.GetVersions(app.ID, app.RepoURL)
				if err != nil || len(versions) == 0 {
					continue
				}
				if bestVersion == nil || versions[0].VersionCode > bestVersion.VersionCode {
					v := versions[0]
					bestVersion = &v
				}
			}
			
			if bestVersion != nil {
				status := ""
				if bestVersion.VersionCode > currentCode {
					status = " [UPDATE AVAILABLE]"
				}
				fmt.Printf("%s (%s)\n  Installed: %d, Latest: %d (%s)%s\n", apps[0].Name, pkg, currentCode, bestVersion.VersionCode, bestVersion.VersionName, status)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
