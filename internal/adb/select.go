package adb

import (
	"fmt"
	"strconv"
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
	fmt.Print("Select device index: ")

	var input string
	fmt.Scanln(&input)
	idx, err := strconv.Atoi(input)
	if err != nil || idx < 0 || idx >= len(devices) {
		return nil, fmt.Errorf("invalid selection")
	}

	return &devices[idx], nil
}
