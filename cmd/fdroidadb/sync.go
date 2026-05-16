package main

import (
	"fmt"

	"fdroidadb/internal/config"
	"fdroidadb/internal/db"
	"fdroidadb/internal/fdroid"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize application indices",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := db.Init(); err != nil {
			return err
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		for _, repo := range cfg.Repos {
			fmt.Printf("Syncing %s...\n", repo.Name)
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
