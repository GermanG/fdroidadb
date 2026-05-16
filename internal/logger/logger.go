package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"fdroidadb/internal/xdg"
)

var (
	Info  *log.Logger
	Error *log.Logger
	Warn  *log.Logger
)

func Init(verbose bool) error {
	logDir := xdg.CacheDir()
	logFile := filepath.Join(logDir, "fdroidadb.log")

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
