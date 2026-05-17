package main

import (
	"fmt"
	"os"

	"github.com/GermanG/fdroidadb/internal/config"
	"github.com/GermanG/fdroidadb/internal/logger"
	"github.com/GermanG/fdroidadb/internal/xdg"
	"github.com/spf13/cobra"
)

var verbose bool
var mockMode bool

var rootCmd = &cobra.Command{
	Use:   "fdroidadb",
	Short: "F-Droid client for Android via ADB",
	Long:  `A command line tool to manage F-Droid applications on your Android device using ADB.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := xdg.EnsureDirs(); err != nil {
			return fmt.Errorf("failed to create directories: %v", err)
		}
		if err := logger.Init(verbose); err != nil {
			return fmt.Errorf("failed to initialize logger: %v", err)
		}
		_, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %v", err)
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&mockMode, "mock", false, "enable mock mode for testing")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
