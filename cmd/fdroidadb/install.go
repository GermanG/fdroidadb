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

var installCmd = &cobra.Command{
	Use:   "install [package_name]",
	Short: "Install an application",
	Args:  cobra.ExactArgs(1),
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

		fmt.Printf("Selected device: %s (%s)\n", device.Model, device.Serial)

		var lastErr error
		for _, repo := range cfg.Repos {
			err := fdroid.InstallApp(args[0], device, repo.URL, cfg.MaxRetries)
			if err == nil {
				return nil
			}
			lastErr = err
		}

		return fmt.Errorf("could not install app from any repository. Last error: %v", lastErr)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
