package fdroid

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"fdroidadb/internal/adb"
	"fdroidadb/internal/db"
	"fdroidadb/internal/logger"
	"fdroidadb/internal/xdg"
)

func InstallApp(query string, device *adb.Device, repoURL string, maxRetries int) error {
	// First, try exact package name
	app, err := db.GetAppByPackage(query)
	if err != nil {
		// If not found, try searching by name
		apps, searchErr := db.SearchApps(query)
		if searchErr != nil || len(apps) == 0 {
			return fmt.Errorf("app '%s' not found in database (tried package name and search)", query)
		}

		if len(apps) > 1 {
			fmt.Printf("Multiple apps found for '%s':\n", query)
			for _, a := range apps {
				fmt.Printf(" - %s (%s)\n", a.Name, a.PackageName)
			}
			return fmt.Errorf("please use the full package name to be specific")
		}

		// If exactly one match, use it
		app = &apps[0]
		fmt.Printf("Resolved '%s' to %s (%s)\n", query, app.Name, app.PackageName)
	}

	versions, err := db.GetVersions(app.ID)
	if err != nil {
		return err
	}

	var bestVersion *db.Version
	for _, v := range versions {
		if isCompatible(v.Arch, device.Arch) {
			bestVersion = &v
			break
		}
	}

	if bestVersion == nil {
		return fmt.Errorf("no compatible version found for architecture %s", device.Arch)
	}

	logger.Info.Printf("Best version: %s (code %d)", bestVersion.VersionName, bestVersion.VersionCode)

	apkPath := filepath.Join(xdg.CacheDir(), bestVersion.APKName)
	url := repoURL + "/" + bestVersion.APKName

	for i := 0; i < maxRetries; i++ {
		err = downloadFileWithResume(url, apkPath)
		if err == nil {
			break
		}
		logger.Warn.Printf("Download attempt %d failed: %v. Retrying...", i+1, err)
	}
	if err != nil {
		return fmt.Errorf("failed to download APK after %d retries: %v", maxRetries, err)
	}

	logger.Info.Printf("Installing %s on %s...", bestVersion.APKName, device.Serial)
	return device.InstallAPK(apkPath)
}

func isCompatible(verArch, devArch string) bool {
	if verArch == "" {
		return true // Universal
	}
	archs := strings.Split(verArch, ",")
	for _, a := range archs {
		if a == devArch {
			return true
		}
		// Basic mapping for arm64-v8a and armeabi-v7a
		if a == "armeabi-v7a" && devArch == "arm64-v8a" {
			return true
		}
	}
	return false
}

func downloadFileWithResume(url string, path string) error {
	fileInfo, err := os.Stat(path)
	var startByte int64 = 0
	if err == nil {
		startByte = fileInfo.Size()
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	if startByte > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", startByte))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
		// Maybe already finished?
		return nil
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	var mode os.FileMode = 0644
	var out *os.File
	if startByte > 0 && resp.StatusCode == http.StatusPartialContent {
		out, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY, mode)
	} else {
		out, err = os.Create(path)
	}
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
