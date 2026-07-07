package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestOpenDocumentPreservesCodeLanguageClass(t *testing.T) {
	dir := t.TempDir()
	docPath := filepath.Join(dir, "README.md")
	source := "# Code\n\n```go\nfmt.Println(\"hi\")\n```\n"
	if err := os.WriteFile(docPath, []byte(source), 0o600); err != nil {
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
	if !strings.Contains(doc.HTML, `class="language-go"`) {
		t.Fatalf("expected code language class for Shiki, got %q", doc.HTML)
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

func TestOpenDocumentRewritesLocalImagesAndMarkdownLinks(t *testing.T) {
	dir := t.TempDir()
	assetsDir := filepath.Join(dir, "assets")
	if err := os.MkdirAll(assetsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	imagePath := filepath.Join(assetsDir, "pic.png")
	if err := os.WriteFile(imagePath, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}, 0o600); err != nil {
		t.Fatal(err)
	}
	nextPath := filepath.Join(dir, "next.md")
	if err := os.WriteFile(nextPath, []byte("# Next\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	outsideImage := filepath.Join(filepath.Dir(dir), "outside.png")
	if err := os.WriteFile(outsideImage, []byte("outside"), 0o600); err != nil {
		t.Fatal(err)
	}
	docPath := filepath.Join(dir, "README.md")
	source := "# Home\n\n![local](./assets/pic.png)\n\n[Next](next.md#next)\n\n![outside](../outside.png)\n"
	if err := os.WriteFile(docPath, []byte(source), 0o600); err != nil {
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
	if !strings.Contains(doc.HTML, `src="/kern-asset?path=assets%2Fpic.png"`) {
		t.Fatalf("expected local image to use controlled asset URL, got %q", doc.HTML)
	}
	if !strings.Contains(doc.HTML, `data-kern-document="next.md"`) || !strings.Contains(doc.HTML, `data-kern-heading="next"`) {
		t.Fatalf("expected markdown link data attributes, got %q", doc.HTML)
	}
	if strings.Contains(doc.HTML, "outside.png") {
		t.Fatalf("expected outside image not to be rewritten or preserved, got %q", doc.HTML)
	}
}

func TestOpenWorkspaceDocumentUsesWorkspaceRelativePath(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "docs")
	if err := os.MkdirAll(nested, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "next.md"), []byte("# Next\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(filepath.Dir(dir), "outside.md")
	if err := os.WriteFile(outside, []byte("# Outside\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if _, err := app.ScanWorkspace(dir); err != nil {
		t.Fatal(err)
	}
	doc, err := app.OpenWorkspaceDocument("docs/next.md")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Title != "Next" {
		t.Fatalf("expected linked document title, got %q", doc.Title)
	}
	if _, err := app.OpenWorkspaceDocument("../outside.md"); err == nil {
		t.Fatal("expected relative path outside workspace to be rejected")
	}
}

func TestStatDocumentDetectsChangesAndDeletion(t *testing.T) {
	dir := t.TempDir()
	docPath := filepath.Join(dir, "README.md")
	if err := os.WriteFile(docPath, []byte("# One\n"), 0o600); err != nil {
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

	status, err := app.StatDocument(docPath, doc.ModifiedAt, doc.Size)
	if err != nil {
		t.Fatal(err)
	}
	if !status.Exists || !status.IsDocument || status.Changed {
		t.Fatalf("expected unchanged existing document, got %+v", status)
	}

	if err := os.WriteFile(docPath, []byte("# One\n\nchanged\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	status, err = app.StatDocument(docPath, doc.ModifiedAt, doc.Size)
	if err != nil {
		t.Fatal(err)
	}
	if !status.Exists || !status.Changed {
		t.Fatalf("expected changed existing document, got %+v", status)
	}

	if err := os.Remove(docPath); err != nil {
		t.Fatal(err)
	}
	status, err = app.StatDocument(docPath, doc.ModifiedAt, doc.Size)
	if err != nil {
		t.Fatal(err)
	}
	if status.Exists || !status.Changed {
		t.Fatalf("expected deleted document to be reported as changed, got %+v", status)
	}
}

func TestStatDocumentRejectsOutsideWorkspace(t *testing.T) {
	workspace := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(outside, []byte("# Outside\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if _, err := app.ScanWorkspace(workspace); err != nil {
		t.Fatal(err)
	}
	if _, err := app.StatDocument(outside, 0, 0); err == nil {
		t.Fatal("expected outside workspace stat to be rejected")
	}
}

func TestAssetHandlerServesOnlyWorkspaceImages(t *testing.T) {
	dir := t.TempDir()
	imagePath := filepath.Join(dir, "pic.png")
	imageBytes := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	if err := os.WriteFile(imagePath, imageBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(filepath.Dir(dir), "outside.png")
	if err := os.WriteFile(outside, imageBytes, 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if _, err := app.ScanWorkspace(dir); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/kern-asset?path=pic.png", nil)
	rec := httptest.NewRecorder()
	app.assetHandler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected asset status 200, got %d", rec.Code)
	}
	if rec.Body.String() != string(imageBytes) {
		t.Fatalf("expected image bytes, got %q", rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/kern-asset?path=../outside.png", nil)
	rec = httptest.NewRecorder()
	app.assetHandler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected outside asset status 404, got %d", rec.Code)
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
