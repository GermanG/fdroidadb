// Copyright (C) 2026 German Gutierrez
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package xdg

import (
	"os"
	"strings"
	"testing"
)

func TestXDGDirs(t *testing.T) {
	home := os.Getenv("HOME")
	
	config := ConfigDir()
	if !strings.Contains(config, ".config/fdroidadb") {
		t.Errorf("expected config dir to contain .config/fdroidadb, got %s", config)
	}

	data := DataDir()
	if !strings.Contains(data, ".local/share/fdroidadb") {
		t.Errorf("expected data dir to contain .local/share/fdroidadb, got %s", data)
	}

	cache := CacheDir()
	if !strings.Contains(cache, ".cache/fdroidadb") {
		t.Errorf("expected cache dir to contain .cache/fdroidadb, got %s", cache)
	}

	if !strings.HasPrefix(config, home) && os.Getenv("XDG_CONFIG_HOME") == "" {
		t.Errorf("expected config dir to be under home, got %s", config)
	}
}

func TestEnsureDirs(t *testing.T) {
	// Use a temp home for testing
	tmpHome, err := os.MkdirTemp("", "fdroidadb-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpHome)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	// Unset XDG envs to ensure we use HOME
	oldConfig := os.Getenv("XDG_CONFIG_HOME")
	oldData := os.Getenv("XDG_DATA_HOME")
	oldCache := os.Getenv("XDG_CACHE_HOME")
	os.Unsetenv("XDG_CONFIG_HOME")
	os.Unsetenv("XDG_DATA_HOME")
	os.Unsetenv("XDG_CACHE_HOME")
	defer func() {
		os.Setenv("XDG_CONFIG_HOME", oldConfig)
		os.Setenv("XDG_DATA_HOME", oldData)
		os.Setenv("XDG_CACHE_HOME", oldCache)
	}()

	err = EnsureDirs()
	if err != nil {
		t.Errorf("EnsureDirs failed: %v", err)
	}

	if _, err := os.Stat(ConfigDir()); os.IsNotExist(err) {
		t.Errorf("ConfigDir was not created")
	}
}
