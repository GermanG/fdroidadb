// Copyright (C) 2026 German Gutierrez
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package db

import (
	"os"
	"testing"
)

func TestDBInitAndOps(t *testing.T) {
	tmpData, err := os.MkdirTemp("", "fdroidadb-db-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpData)

	os.Setenv("XDG_DATA_HOME", tmpData)

	err = Init()
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer DB.Close()

	app := App{
		PackageName: "org.test.app",
		Name:        "Test App",
		Summary:     "A test app",
	}

	id, err := SaveApp(app)
	if err != nil {
		t.Fatalf("failed to save app: %v", err)
	}

	if id == 0 {
		t.Errorf("expected non-zero id")
	}

	apps, err := SearchApps("Test")
	if err != nil {
		t.Fatalf("failed to search apps: %v", err)
	}

	if len(apps) != 1 {
		t.Errorf("expected 1 app, got %d", len(apps))
	}

	if apps[0].Name != "Test App" {
		t.Errorf("expected Test App, got %s", apps[0].Name)
	}

	// Test UPSERT: update the app and verify the ID is preserved
	appUpdated := App{
		PackageName: "org.test.app",
		Name:        "Test App Updated",
		Summary:     "A test app with updated summary",
	}

	id2, err := SaveApp(appUpdated)
	if err != nil {
		t.Fatalf("failed to update app: %v", err)
	}

	if id != id2 {
		t.Errorf("expected ID to be preserved across updates (UPSERT), got first ID: %d, second ID: %d", id, id2)
	}

	apps2, err := SearchApps("Updated")
	if err != nil {
		t.Fatalf("failed to search apps: %v", err)
	}
	if len(apps2) != 1 {
		t.Errorf("expected 1 app matching 'Updated', got %d", len(apps2))
	}
	if apps2[0].Name != "Test App Updated" {
		t.Errorf("expected Test App Updated, got %s", apps2[0].Name)
	}
}
