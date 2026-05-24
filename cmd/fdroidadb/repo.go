// Copyright (C) 2026 German Gutierrez
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package main

import (
	"fmt"

	"github.com/GermanG/fdroidadb/internal/db"
	"github.com/spf13/cobra"
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage repositories",
}

var setFingerprintCmd = &cobra.Command{
	Use:   "set-fingerprint [repo_url] [fingerprint]",
	Short: "Set the expected SHA-256 fingerprint for a repository",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := db.Init(); err != nil {
			return err
		}
		url := args[0]
		fingerprint := args[1]

		if err := db.SetRepoFingerprint(url, fingerprint); err != nil {
			return err
		}

		fmt.Printf("Fingerprint for %s updated successfully.\n", url)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(repoCmd)
	repoCmd.AddCommand(setFingerprintCmd)
}
