package xdg

import (
	"os"
	"path/filepath"
)

func ConfigDir() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "fdroidadb")
	}
	return filepath.Join(os.Getenv("HOME"), ".config", "fdroidadb")
}

func DataDir() string {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return filepath.Join(dir, "fdroidadb")
	}
	return filepath.Join(os.Getenv("HOME"), ".local", "share", "fdroidadb")
}

func CacheDir() string {
	if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
		return filepath.Join(dir, "fdroidadb")
	}
	return filepath.Join(os.Getenv("HOME"), ".cache", "fdroidadb")
}

func EnsureDirs() error {
	dirs := []string{ConfigDir(), DataDir(), CacheDir()}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}
