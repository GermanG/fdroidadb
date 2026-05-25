// Copyright (C) 2026 German Gutierrez
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package adb

import (
	"fmt"

	"github.com/GermanG/fdroidadb/internal/cli"
)

func SelectDevice(mock bool) (*Device, error) {
	var devices []Device
	var err error

	if mock {
		devices = GetMockDevices()
	} else {
		devices, err = GetDevices()
		if err != nil {
			return nil, err
		}
	}

	if len(devices) == 0 {
		return nil, fmt.Errorf("no devices connected")
	}

	if len(devices) == 1 {
		return &devices[0], nil
	}

	fmt.Println("Multiple devices detected:")
	for i, d := range devices {
		fmt.Printf("[%d] %s (%s, %s)\n", i, d.Model, d.Serial, d.Arch)
	}

	idx, err := cli.ReadInt("Select device index: ", 0, len(devices)-1)
	if err != nil {
		return nil, fmt.Errorf("invalid selection: %v", err)
	}

	return &devices[idx], nil
}
