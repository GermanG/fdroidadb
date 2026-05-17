// Copyright (C) 2026 German Gutierrez
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/GermanG/fdroidadb/internal/xdg"
)

var (
	Info  *log.Logger
	Error *log.Logger
	Warn  *log.Logger
)

const maxLogSize = 5 * 1024 * 1024 // 5MB

func Init(verbose bool) error {
	logDir := xdg.CacheDir()
	logFile := filepath.Join(logDir, "fdroidadb.log")

	// Basic rotation
	if fi, err := os.Stat(logFile); err == nil {
		if fi.Size() > maxLogSize {
			oldLog := logFile + ".old"
			_ = os.Rename(logFile, oldLog) // Ignore error if rename fails
		}
	}

	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	var mw io.Writer
	if verbose {
		mw = io.MultiWriter(os.Stderr, file)
	} else {
		mw = file
	}

	Info = log.New(mw, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warn = log.New(mw, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(mw, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	return nil
}
