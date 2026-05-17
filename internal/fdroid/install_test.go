// Copyright (C) 2026 German Gutierrez
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package fdroid

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/GermanG/fdroidadb/internal/adb"
	"github.com/GermanG/fdroidadb/internal/db"
	"github.com/GermanG/fdroidadb/internal/logger"
	"github.com/GermanG/fdroidadb/internal/xdg"
)

func TestInstallAppLogic(t *testing.T) {
	// Setup temp dirs for test
	tmpDir, err := os.MkdirTemp("", "fdroid-install-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))
	os.Setenv("XDG_CACHE_HOME", filepath.Join(tmpDir, "cache"))
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))

	if err := xdg.EnsureDirs(); err != nil {
		t.Fatal(err)
	}

	if err := logger.Init(false); err != nil {
		t.Fatal(err)
	}

	if err := db.Init(); err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.DB.Close()

	// 1. Seed the database
	appID, err := db.SaveApp(db.App{
		PackageName: "org.test.app",
		Name:        "Test App",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = db.SaveVersion(db.Version{
		AppID:       appID,
		VersionName: "1.0",
		VersionCode: 10,
		APKName:     "test-1.0.apk",
		Arch:        "arm64-v8a",
	})
	if err != nil {
		t.Fatal(err)
	}

	// 2. Setup mock device
	mockAdb := adb.NewMockADBDevice()
	device := &adb.Device{
		Serial: "MOCK_SERIAL",
		Arch:   "arm64-v8a",
		Adb:    mockAdb,
	}

	// 3. Setup a dummy server for download
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("fake apk content"))
	}))
	defer ts.Close()

	// 4. Run InstallApp
	err = InstallApp("org.test.app", device, ts.URL, 3)
	if err != nil {
		t.Fatalf("InstallApp failed: %v", err)
	}

	// 5. Verify it called Adb.PushFile
	if len(mockAdb.PushedFiles) == 0 {
		t.Errorf("Expected APK to be pushed to device")
	}
}
