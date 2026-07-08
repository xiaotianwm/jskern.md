package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
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

func TestRefreshWorkspaceDetectsStructureChangesOnly(t *testing.T) {
	dir := t.TempDir()
	docPath := filepath.Join(dir, "README.md")
	if err := os.WriteFile(docPath, []byte("# Home\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if _, err := app.ScanWorkspace(dir); err != nil {
		t.Fatal(err)
	}

	refresh, err := app.RefreshWorkspace()
	if err != nil {
		t.Fatal(err)
	}
	if refresh.Changed {
		t.Fatalf("expected unchanged workspace structure, got %+v", refresh)
	}

	addedPath := filepath.Join(dir, "guide.md")
	if err := os.WriteFile(addedPath, []byte("# Guide\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	refresh, err = app.RefreshWorkspace()
	if err != nil {
		t.Fatal(err)
	}
	if !refresh.Changed || refresh.Tree == nil || !treeContainsPath(refresh.Tree.Root, addedPath) {
		t.Fatalf("expected added Markdown file to refresh tree, got %+v", refresh)
	}

	if err := os.WriteFile(addedPath, []byte("# Guide\n\ncontent changed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	refresh, err = app.RefreshWorkspace()
	if err != nil {
		t.Fatal(err)
	}
	if refresh.Changed {
		t.Fatalf("expected Markdown content change to stay out of tree refresh, got %+v", refresh)
	}

	if err := os.Remove(addedPath); err != nil {
		t.Fatal(err)
	}
	refresh, err = app.RefreshWorkspace()
	if err != nil {
		t.Fatal(err)
	}
	if !refresh.Changed || refresh.Tree == nil || treeContainsPath(refresh.Tree.Root, addedPath) {
		t.Fatalf("expected deleted Markdown file to refresh tree, got %+v", refresh)
	}
}

func TestRefreshWorkspaceIgnoresSkippedDirectories(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Home\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if _, err := app.ScanWorkspace(dir); err != nil {
		t.Fatal(err)
	}

	skipped := filepath.Join(dir, "node_modules")
	if err := os.MkdirAll(skipped, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skipped, "needle.md"), []byte("# Skipped\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	refresh, err := app.RefreshWorkspace()
	if err != nil {
		t.Fatal(err)
	}
	if refresh.Changed {
		t.Fatalf("expected skipped directory changes to stay out of tree refresh, got %+v", refresh)
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

func TestSearchWorkspaceFindsFileNamesAndContent(t *testing.T) {
	dir := t.TempDir()
	docs := filepath.Join(dir, "docs")
	if err := os.MkdirAll(docs, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Home\n\nNeedle in the haystack.\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(docs, "guide.md"), []byte("# Guide\n\nPlain content.\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	skipped := filepath.Join(dir, "node_modules")
	if err := os.MkdirAll(skipped, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skipped, "needle.md"), []byte("# Should Skip\nNeedle\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if _, err := app.ScanWorkspace(dir); err != nil {
		t.Fatal(err)
	}

	results, err := app.SearchWorkspace("needle")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Kind != "content" || results[0].Name != "README.md" {
		t.Fatalf("expected one content hit in README.md, got %+v", results)
	}
	if strings.Contains(results[0].RelativePath, "node_modules") {
		t.Fatalf("expected skipped directories to stay out of search results, got %+v", results)
	}

	results, err = app.SearchWorkspace("guide")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 || results[0].Kind != "file" || results[0].RelativePath != "docs/guide.md" {
		t.Fatalf("expected file-name hit for docs/guide.md, got %+v", results)
	}
}

func TestSearchWorkspaceRequiresOpenWorkspace(t *testing.T) {
	app := NewApp()
	if _, err := app.SearchWorkspace("readme"); err == nil {
		t.Fatal("expected search without workspace to fail")
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
	if settings.Locale != "zh-CN" || settings.Theme != "system" {
		t.Fatalf("expected default locale/theme, got %+v", settings)
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
	if app.settings.StorageVersion != currentSettingsVersion || app.settings.LastWorkspace != "" || app.settings.Locale != "zh-CN" || app.settings.Theme != "system" {
		t.Fatalf("expected default settings after bad file, got %+v", app.settings)
	}
}

func TestSwitchLanguageAndThemePersistSettings(t *testing.T) {
	appDataRoot := t.TempDir()
	app := NewApp()
	if err := app.initAppData(appDataRoot); err != nil {
		t.Fatal(err)
	}

	bootstrap, err := app.SwitchLanguage("en")
	if err != nil {
		t.Fatal(err)
	}
	if bootstrap.CurrentLocale != "en" {
		t.Fatalf("expected English bootstrap, got %q", bootstrap.CurrentLocale)
	}

	bootstrap, err = app.SwitchTheme("dark")
	if err != nil {
		t.Fatal(err)
	}
	if bootstrap.CurrentLocale != "en" || bootstrap.CurrentTheme != "dark" {
		t.Fatalf("expected persisted English dark bootstrap, got %+v", bootstrap)
	}

	settingsBytes, err := os.ReadFile(filepath.Join(appDataRoot, "config", "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	var settings Settings
	if err := json.Unmarshal(settingsBytes, &settings); err != nil {
		t.Fatal(err)
	}
	if settings.Locale != "en" || settings.Theme != "dark" {
		t.Fatalf("expected persisted locale/theme, got %+v", settings)
	}

	bootstrap, err = app.SwitchLanguage("unknown")
	if err != nil {
		t.Fatal(err)
	}
	if bootstrap.CurrentLocale != "zh-CN" {
		t.Fatalf("expected unsupported language to normalize to zh-CN, got %q", bootstrap.CurrentLocale)
	}

	bootstrap, err = app.SwitchTheme("neon")
	if err != nil {
		t.Fatal(err)
	}
	if bootstrap.CurrentTheme != "system" {
		t.Fatalf("expected unsupported theme to normalize to system, got %q", bootstrap.CurrentTheme)
	}
}

func treeContainsPath(node TreeNode, path string) bool {
	if filepath.Clean(node.Path) == filepath.Clean(path) {
		return true
	}
	for _, child := range node.Children {
		if treeContainsPath(child, path) {
			return true
		}
	}
	return false
}

func TestCheckGitHubUpdatesFindsInstallerRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"tag_name": "v0.1.2",
				"html_url": "https://github.com/xiaotianwm/jskern.md/releases/tag/v0.1.2",
				"body": "Update notes",
				"draft": false,
				"assets": [
					{
						"name": "JSKernMD-Setup-0.1.2-x64.exe",
						"browser_download_url": "https://github.com/xiaotianwm/jskern.md/releases/download/v0.1.2/JSKernMD-Setup-0.1.2-x64.exe",
						"digest": "sha256:abc123"
					}
				]
			}
		]`))
	}))
	defer server.Close()

	info, err := checkGitHubUpdates(nil, server.URL, "0.1.1")
	if err != nil {
		t.Fatal(err)
	}
	if !info.UpdateAvailable || info.LatestVersion != "0.1.2" {
		t.Fatalf("expected 0.1.2 update, got %+v", info)
	}
	if info.Sha256 != "abc123" || info.ReleaseNotes != "Update notes" {
		t.Fatalf("expected sha and notes from release, got %+v", info)
	}
}

func TestCheckGitHubUpdatesIgnoresOlderAndMalformedAssets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"tag_name": "v0.1.2",
				"draft": false,
				"assets": [
					{
						"name": "jskernmd.exe",
						"browser_download_url": "https://github.com/xiaotianwm/jskern.md/releases/download/v0.1.2/jskernmd.exe"
					}
				]
			},
			{
				"tag_name": "v0.1.0",
				"draft": false,
				"assets": [
					{
						"name": "JSKernMD-Setup-0.1.0-x64.exe",
						"browser_download_url": "https://github.com/xiaotianwm/jskern.md/releases/download/v0.1.0/JSKernMD-Setup-0.1.0-x64.exe"
					}
				]
			}
		]`))
	}))
	defer server.Close()

	info, err := checkGitHubUpdates(nil, server.URL, "0.1.1")
	if err != nil {
		t.Fatal(err)
	}
	if info.UpdateAvailable || info.LatestVersion != "" {
		t.Fatalf("expected no usable update, got %+v", info)
	}
}

func TestDownloadFileVerifiesChecksum(t *testing.T) {
	payload := []byte("installer")
	sum := fmt.Sprintf("%x", sha256.Sum256(payload))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	target := filepath.Join(t.TempDir(), "update.exe")
	if err := downloadFile(nil, server.URL, target, sum); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(payload) {
		t.Fatalf("expected downloaded payload, got %q", data)
	}
	if err := downloadFile(nil, server.URL, filepath.Join(t.TempDir(), "bad.exe"), "bad"); err == nil {
		t.Fatal("expected checksum mismatch to fail")
	}
}

func TestDismissUpdatePersistsIgnoredVersion(t *testing.T) {
	appDataRoot := t.TempDir()
	app := NewApp()
	if err := app.initAppData(appDataRoot); err != nil {
		t.Fatal(err)
	}
	if err := app.DismissUpdate("v0.1.2"); err != nil {
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
	if settings.IgnoredUpdateVersion != "0.1.2" {
		t.Fatalf("expected ignored version 0.1.2, got %+v", settings)
	}
}
