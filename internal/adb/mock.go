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
	"time"
)

type MockADBDevice struct {
	LastCommand string
	PushedFiles map[string]bool
	Responses   map[string]string
}

func NewMockADBDevice() *MockADBDevice {
	return &MockADBDevice{
		PushedFiles: make(map[string]bool),
		Responses:   make(map[string]string),
	}
}

func (m *MockADBDevice) RunShellCommand(cmd string, args ...string) (string, error) {
	m.LastCommand = cmd + " " + strings.Join(args, " ")
	if resp, ok := m.Responses[cmd]; ok {
		return resp, nil
	}
	// Default responses for installation
	if strings.Contains(cmd, "pm install") {
		return "Success", nil
	}
	return "", nil
}

func (m *MockADBDevice) PushFile(file *os.File, remotePath string, mtime ...time.Time) error {
	m.PushedFiles[remotePath] = true
	return nil
}

func (m *MockADBDevice) PullFile(remotePath string, localPath string) error {
	// For testing, we might need to create a dummy file
	return os.WriteFile(localPath, []byte("fake pulled apk"), 0644)
}

func (m *MockADBDevice) Serial() string {
	return "MOCK_SERIAL_123"
}
