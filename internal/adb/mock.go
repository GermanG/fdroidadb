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

func (m *MockADBDevice) Serial() string {
	return "MOCK_SERIAL_123"
}
