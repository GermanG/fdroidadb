// Copyright (C) 2026 German Gutierrez
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package adb

import (
	"os"
	"strings"
	"testing"
)

func TestInstallAPK(t *testing.T) {
	mock := NewMockADBDevice()
	dev := &Device{
		Serial: "MOCK_SERIAL_123",
		Adb:    mock,
	}

	// Create a dummy APK file
	tmpFile, err := os.CreateTemp("", "test-*.apk")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("fake apk content"))
	tmpFile.Close()

	err = dev.InstallAPK(tmpFile.Name())
	if err != nil {
		t.Fatalf("InstallAPK failed: %v", err)
	}

	if len(mock.PushedFiles) == 0 {
		t.Errorf("Expected files to be pushed")
	}

	// The last command should have been rm (cleanup)
	if !strings.Contains(mock.LastCommand, "rm") {
		t.Errorf("Expected cleanup command, got: %s", mock.LastCommand)
	}
}

func TestGetPackageVersion(t *testing.T) {
	mock := NewMockADBDevice()
	mock.Responses["dumpsys package com.test | grep versionCode"] = "versionCode=42 targetSdk=30"
	
	dev := &Device{
		Adb: mock,
	}

	version, err := dev.GetPackageVersion("com.test")
	if err != nil {
		t.Fatalf("GetPackageVersion failed: %v", err)
	}

	if version != 42 {
		t.Errorf("Expected version 42, got %d", version)
	}
}
