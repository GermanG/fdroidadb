// Copyright (C) 2026 German Gutierrez
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package config

import (
	"os"
	"testing"
)

func TestConfigLoadSave(t *testing.T) {
	tmpHome, err := os.MkdirTemp("", "fdroidadb-cfg-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpHome)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	os.Unsetenv("XDG_CONFIG_HOME")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if len(cfg.Repos) == 0 {
		t.Errorf("expected default repos, got none")
	}

	cfg.MaxRetries = 10
	err = Save(cfg)
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	cfg2, err := Load()
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	if cfg2.MaxRetries != 10 {
		t.Errorf("expected MaxRetries 10, got %d", cfg2.MaxRetries)
	}
}
