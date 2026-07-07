package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenDocumentRendersAndSanitizesMarkdown(t *testing.T) {
	dir := t.TempDir()
	docPath := filepath.Join(dir, "README.md")
	err := os.WriteFile(docPath, []byte("# Title\n\n<script>alert(1)</script>\n\n## Next\n\n- item\n"), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if _, err := app.ScanWorkspace(dir); err != nil {
		t.Fatal(err)
	}

	doc, err := app.OpenDocument(docPath)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Title != "Title" {
		t.Fatalf("expected title from first heading, got %q", doc.Title)
	}
	if len(doc.Outline) != 2 {
		t.Fatalf("expected 2 headings, got %d", len(doc.Outline))
	}
	if doc.Outline[0].ID == "" || !strings.Contains(doc.HTML, `id="`+doc.Outline[0].ID+`"`) {
		t.Fatalf("expected sanitized HTML to preserve heading id, outline=%+v html=%q", doc.Outline, doc.HTML)
	}
	if !strings.Contains(doc.HTML, "<h1") {
		t.Fatalf("expected rendered heading HTML, got %q", doc.HTML)
	}
	if strings.Contains(strings.ToLower(doc.HTML), "script") {
		t.Fatalf("expected sanitized HTML, got %q", doc.HTML)
	}
}

func TestOpenDocumentRejectsOutsideWorkspace(t *testing.T) {
	workspace := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(outside, []byte("# Outside\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if _, err := app.ScanWorkspace(workspace); err != nil {
		t.Fatal(err)
	}
	if _, err := app.OpenDocument(outside); err == nil {
		t.Fatal("expected outside workspace document to be rejected")
	}
}

func TestWorkspacePathPersistsAndRestoresFromAppData(t *testing.T) {
	appDataRoot := t.TempDir()
	workspace := t.TempDir()
	if err := os.WriteFile(filepath.Join(workspace, "README.md"), []byte("# Saved\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if err := app.initAppData(appDataRoot); err != nil {
		t.Fatal(err)
	}
	if _, err := app.ScanWorkspace(workspace); err != nil {
		t.Fatal(err)
	}

	settingsBytes, err := os.ReadFile(filepath.Join(appDataRoot, "config", "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	var settings Settings
	if err := json.Unmarshal(settingsBytes, &settings); err != nil {
		t.Fatal(err)
	}
	if settings.StorageVersion != currentSettingsVersion {
		t.Fatalf("expected settings version %d, got %d", currentSettingsVersion, settings.StorageVersion)
	}
	if settings.LastWorkspace != filepath.Clean(workspace) {
		t.Fatalf("expected last workspace %q, got %q", workspace, settings.LastWorkspace)
	}

	nextApp := NewApp()
	if err := nextApp.initAppData(appDataRoot); err != nil {
		t.Fatal(err)
	}
	restored, err := nextApp.RestoreWorkspace()
	if err != nil {
		t.Fatal(err)
	}
	if restored == nil || restored.Root.Path != filepath.Clean(workspace) {
		t.Fatalf("expected restored workspace %q, got %+v", workspace, restored)
	}
}

func TestInitAppDataCreatesStoreLayoutAndBacksUpBadSettings(t *testing.T) {
	appDataRoot := t.TempDir()
	configDir := filepath.Join(appDataRoot, "config")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(configDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte("{bad json"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if err := app.initAppData(appDataRoot); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"config", "data", "logs", "cache", "temp", "runtime", "crash"} {
		info, err := os.Stat(filepath.Join(appDataRoot, name))
		if err != nil {
			t.Fatal(err)
		}
		if !info.IsDir() {
			t.Fatalf("expected %s to be a directory", name)
		}
	}
	backups, err := filepath.Glob(settingsPath + ".bad-*")
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 1 {
		t.Fatalf("expected one bad settings backup, got %d", len(backups))
	}
	if app.settings.StorageVersion != currentSettingsVersion || app.settings.LastWorkspace != "" {
		t.Fatalf("expected default settings after bad file, got %+v", app.settings)
	}
}
