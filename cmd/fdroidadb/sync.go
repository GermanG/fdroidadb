// Copyright (C) 2026 German Gutierrez
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package main

import (
	"fmt"

	"github.com/GermanG/fdroidadb/internal/config"
	"github.com/GermanG/fdroidadb/internal/db"
	"github.com/GermanG/fdroidadb/internal/fdroid"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync [repo_name]",
	Short: "Synchronize application indices",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := db.Init(); err != nil {
			return err
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		reposToSync := cfg.Repos
		if len(args) > 0 {
			found := false
			for _, repo := range cfg.Repos {
				if repo.Name == args[0] {
					reposToSync = []config.Repo{repo}
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("repository '%s' not found in config", args[0])
			}
		}

		for _, repo := range reposToSync {
			fmt.Printf("\nSyncing %s...\n", repo.Name)
			if err := fdroid.SyncRepo(repo.URL); err != nil {
				return err
			}
			fmt.Printf("Sync for %s completed.\n", repo.Name)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
