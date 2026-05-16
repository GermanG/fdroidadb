package adb

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/electricbubble/gadb"
)

// ADBDevice interface abstracts the methods we need from gadb.Device
type ADBDevice interface {
	RunShellCommand(cmd string, args ...string) (string, error)
	PushFile(file *os.File, remotePath string, mtime ...time.Time) error
	Serial() string
}

type Device struct {
	Serial string
	Model  string
	Arch   string
	Adb    ADBDevice
}

func EnsureServer(adbPath string) error {
	client, err := gadb.NewClient()
	if err == nil {
		_, err = client.DeviceList()
		if err == nil {
			return nil
		}
	}

	cmd := exec.Command(adbPath, "start-server")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start adb server: %v", err)
	}

	return nil
}

func GetDevices() ([]Device, error) {
	client, err := gadb.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to adb server: %v", err)
	}

	gadbDevices, err := client.DeviceList()
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %v", err)
	}

	var devices []Device
	for _, d := range gadbDevices {
		model, err := d.RunShellCommand("getprop ro.product.model")
		if err != nil {
			model = "Unknown"
		}
		arch, err := d.RunShellCommand("getprop ro.product.cpu.abi")
		if err != nil {
			arch = "Unknown"
		}

		devices = append(devices, Device{
			Serial: d.Serial(),
			Model:  strings.TrimSpace(model),
			Arch:   strings.TrimSpace(arch),
			Adb:    d,
		})
	}

	return devices, nil
}

func GetMockDevices() []Device {
	mock := NewMockADBDevice()
	mock.Responses["pm list packages"] = "package:org.test.app\npackage:de.mm20.launcher2.release"
	mock.Responses["dumpsys package org.test.app | grep versionCode"] = "versionCode=1"
	
	return []Device{
		{
			Serial: "MOCK_SERIAL_1",
			Model:  "Mock Phone 1",
			Arch:   "arm64-v8a",
			Adb:    mock,
		},
	}
}

func (d *Device) InstallAPK(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Use a very specific filename to avoid any collision or accidental deletion
	// Format: /data/local/tmp/fdroidadb_[PID]_[random].apk
	tempName := fmt.Sprintf("fdroidadb_%d_%d.apk", os.Getpid(), time.Now().UnixNano())
	remotePath := "/data/local/tmp/" + tempName
	
	err = d.Adb.PushFile(f, remotePath)
	if err != nil {
		return fmt.Errorf("failed to push APK: %v", err)
	}

	// Always cleanup the specific file we just created
	defer func() {
		if remotePath != "" && strings.HasPrefix(remotePath, "/data/local/tmp/fdroidadb_") {
			_, err := d.Adb.RunShellCommand("rm", remotePath)
			if err != nil {
				// We just log it as a warning, no need to fail the whole install
				fmt.Printf("Warning: failed to remove temporary APK on device: %v\n", err)
			}
		}
	}()

	out, err := d.Adb.RunShellCommand("pm install -r", remotePath)
	if err != nil {
		return err
	}

	if !strings.Contains(out, "Success") {
		return fmt.Errorf("install failed: %s", out)
	}

	return nil
}

func (d *Device) GetInstalledPackages() ([]string, error) {
	out, err := d.Adb.RunShellCommand("pm list packages")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(out, "\n")
	var packages []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "package:") {
			packages = append(packages, strings.TrimPrefix(line, "package:"))
		}
	}
	return packages, nil
}

func (d *Device) GetPackageVersion(packageName string) (int, error) {
	out, err := d.Adb.RunShellCommand(fmt.Sprintf("dumpsys package %s | grep versionCode", packageName))
	if err != nil {
		return 0, err
	}

	out = strings.TrimSpace(out)
	if out == "" {
		return 0, fmt.Errorf("package not found")
	}

	parts := strings.Fields(out)
	for _, part := range parts {
		if strings.HasPrefix(part, "versionCode=") {
			vStr := strings.TrimPrefix(part, "versionCode=")
			return strconv.Atoi(vStr)
		}
	}

	return 0, fmt.Errorf("versionCode not found in output")
}
