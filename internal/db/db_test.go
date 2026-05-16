package db

import (
	"os"
	"testing"
)

func TestDBInitAndOps(t *testing.T) {
	tmpData, err := os.MkdirTemp("", "fdroidadb-db-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpData)

	os.Setenv("XDG_DATA_HOME", tmpData)

	err = Init()
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer DB.Close()

	app := App{
		PackageName: "org.test.app",
		Name:        "Test App",
		Summary:     "A test app",
	}

	id, err := SaveApp(app)
	if err != nil {
		t.Fatalf("failed to save app: %v", err)
	}

	if id == 0 {
		t.Errorf("expected non-zero id")
	}

	apps, err := SearchApps("Test")
	if err != nil {
		t.Fatalf("failed to search apps: %v", err)
	}

	if len(apps) != 1 {
		t.Errorf("expected 1 app, got %d", len(apps))
	}

	if apps[0].Name != "Test App" {
		t.Errorf("expected Test App, got %s", apps[0].Name)
	}
}
