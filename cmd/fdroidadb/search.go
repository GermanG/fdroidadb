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

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for applications",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := db.Init(); err != nil {
			return err
		}
		apps, err := db.SearchApps(args[0])
		if err != nil {
			return err
		}

		for _, app := range apps {
			fmt.Printf("%s (%s)\n  %s\n", app.Name, app.PackageName, app.Summary)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
