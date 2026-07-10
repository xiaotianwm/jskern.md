package main

import (
	"context"
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

func TestProductInfoComesFromManifest(t *testing.T) {
	var manifest productManifestFile
	if err := json.Unmarshal(productManifest, &manifest); err != nil {
		t.Fatal(err)
	}
	bootstrap, err := NewApp().GetBootstrap("")
	if err != nil {
		t.Fatal(err)
	}
	if bootstrap.Product.Version != manifest.Version {
		t.Fatalf("expected bootstrap version %q from manifest, got %q", manifest.Version, bootstrap.Product.Version)
	}
	if bootstrap.Product.AppSlug != manifest.AppSlug {
		t.Fatalf("expected app slug %q from manifest, got %q", manifest.AppSlug, bootstrap.Product.AppSlug)
	}
	if bootstrap.Product.AppID != manifest.AppID {
		t.Fatalf("expected app id %q from manifest, got %q", manifest.AppID, bootstrap.Product.AppID)
	}
}

func TestActionFeedbackLocaleKeys(t *testing.T) {
	keys := []string{
		"feedback.copy_success",
		"feedback.copy_failed",
		"feedback.reveal_success",
		"feedback.reveal_failed",
		"feedback.rename_success",
		"feedback.rename_failed",
		"feedback.remove_workspace_success",
		"feedback.remove_workspace_failed",
		"feedback.dismiss",
	}
	for _, locale := range []string{"zh-CN", "en"} {
		messages, err := loadLocale(locale)
		if err != nil {
			t.Fatal(err)
		}
		for _, key := range keys {
			if strings.TrimSpace(messages["business"][key]) == "" {
				t.Fatalf("expected %s business locale to include %q", locale, key)
			}
		}
	}
}

func TestCurrentManifestVersionIsNotReportedAsUpdate(t *testing.T) {
	assetName := "JSKernMD-Setup-" + productInfo.Version + "-x64.exe"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]githubRelease{{
			TagName: "v" + productInfo.Version,
			HTMLURL: "https://github.com/xiaotianwm/jskern.md/releases/tag/v" + productInfo.Version,
			Assets: []githubAsset{{
				Name:               assetName,
				BrowserDownloadURL: "https://github.com/xiaotianwm/jskern.md/releases/download/v" + productInfo.Version + "/" + assetName,
			}},
		}})
	}))
	defer server.Close()

	info, err := checkGitHubUpdates(context.Background(), server.URL, productInfo.Version)
	if err != nil {
		t.Fatal(err)
	}
	if info.UpdateAvailable {
		t.Fatalf("expected no update for current manifest version, got %+v", info)
	}
}

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

func TestRevealableWorkspacePathValidatesWorkspaceBoundary(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "docs")
	if err := os.MkdirAll(nested, 0o700); err != nil {
		t.Fatal(err)
	}
	docPath := filepath.Join(nested, "guide.md")
	if err := os.WriteFile(docPath, []byte("# Guide\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(outside, []byte("# Outside\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if _, err := app.ScanWorkspace(dir); err != nil {
		t.Fatal(err)
	}
	if got, err := app.revealableWorkspacePath(docPath); err != nil || got != filepath.Clean(docPath) {
		t.Fatalf("expected workspace file to be revealable, got path=%q err=%v", got, err)
	}
	if got, err := app.revealableWorkspacePath("docs"); err != nil || got != filepath.Clean(nested) {
		t.Fatalf("expected workspace-relative directory to be revealable, got path=%q err=%v", got, err)
	}
	if _, err := app.revealableWorkspacePath(outside); err == nil {
		t.Fatal("expected outside workspace path to be rejected")
	}
	if _, err := app.revealableWorkspacePath(filepath.Join(dir, "missing.md")); err == nil {
		t.Fatal("expected missing workspace path to be rejected")
	}
}

func TestRenamePathRenamesWorkspaceDocumentAndRefreshesTree(t *testing.T) {
	dir := t.TempDir()
	docPath := filepath.Join(dir, "README.md")
	if err := os.WriteFile(docPath, []byte("# Home\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if _, err := app.ScanWorkspace(dir); err != nil {
		t.Fatal(err)
	}
	result, err := app.RenamePath(docPath, "Guide.md")
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "Guide.md")
	if result.OldPath != filepath.Clean(docPath) || result.NewPath != filepath.Clean(target) || result.NodeType != "file" {
		t.Fatalf("expected rename result for document, got %+v", result)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatal(err)
	}
	if result.Tree == nil || !treeContainsPath(result.Tree.Root, target) || treeContainsPath(result.Tree.Root, docPath) {
		t.Fatalf("expected refreshed tree with renamed document, got %+v", result.Tree)
	}
}

func TestRenamePathRenamesWorkspaceDirectory(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "docs")
	if err := os.MkdirAll(nested, 0o700); err != nil {
		t.Fatal(err)
	}
	docPath := filepath.Join(nested, "guide.md")
	if err := os.WriteFile(docPath, []byte("# Guide\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if _, err := app.ScanWorkspace(dir); err != nil {
		t.Fatal(err)
	}
	result, err := app.RenamePath(nested, "notes")
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "notes")
	if result.NodeType != "directory" || result.NewPath != filepath.Clean(target) {
		t.Fatalf("expected directory rename result, got %+v", result)
	}
	if result.Tree == nil || !treeContainsPath(result.Tree.Root, filepath.Join(target, "guide.md")) {
		t.Fatalf("expected refreshed tree with renamed directory child, got %+v", result.Tree)
	}
}

func TestRenamePathRejectsUnsafeTargets(t *testing.T) {
	dir := t.TempDir()
	docPath := filepath.Join(dir, "README.md")
	if err := os.WriteFile(docPath, []byte("# Home\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "exists.md"), []byte("# Existing\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if _, err := app.ScanWorkspace(dir); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"", "../escape.md", "notes/guide.md", "guide.txt", "bad:name.md", "exists.md"} {
		if _, err := app.RenamePath(docPath, name); err == nil {
			t.Fatalf("expected rename target %q to be rejected", name)
		}
	}
	if _, err := app.RenamePath(dir, "new-root"); err == nil {
		t.Fatal("expected workspace root rename to be rejected")
	}
	outside := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(outside, []byte("# Outside\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := app.RenamePath(outside, "inside.md"); err == nil {
		t.Fatal("expected outside source to be rejected")
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

func TestReadingMemoryPersistsAndRestoresLastDocument(t *testing.T) {
	appDataRoot := t.TempDir()
	workspace := t.TempDir()
	docPath := filepath.Join(workspace, "README.md")
	if err := os.WriteFile(docPath, []byte("# Saved\n\n## Spot\n\nBody\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if err := app.initAppData(appDataRoot); err != nil {
		t.Fatal(err)
	}
	if _, err := app.ScanWorkspace(workspace); err != nil {
		t.Fatal(err)
	}
	doc, err := app.OpenDocument(docPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.SaveReadingPosition(doc.Path, 240, 0.42, "spot", doc.ModifiedAt, doc.Size); err != nil {
		t.Fatal(err)
	}

	memoryBytes, err := os.ReadFile(filepath.Join(appDataRoot, "data", "reading-memory.json"))
	if err != nil {
		t.Fatal(err)
	}
	var memory ReadingMemoryStore
	if err := json.Unmarshal(memoryBytes, &memory); err != nil {
		t.Fatal(err)
	}
	if memory.StorageVersion != currentReadingMemoryVersion {
		t.Fatalf("expected reading memory version %d, got %d", currentReadingMemoryVersion, memory.StorageVersion)
	}

	nextApp := NewApp()
	if err := nextApp.initAppData(appDataRoot); err != nil {
		t.Fatal(err)
	}
	if _, err := nextApp.RestoreWorkspace(); err != nil {
		t.Fatal(err)
	}
	snapshot, err := nextApp.GetReadingMemory()
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.LastPosition == nil {
		t.Fatal("expected restored reading position")
	}
	if snapshot.LastDocument != filepath.Clean(docPath) || snapshot.LastPosition.Path != filepath.Clean(docPath) {
		t.Fatalf("expected restored document %q, got %+v", docPath, snapshot)
	}
	if snapshot.LastPosition.ScrollTop != 240 || snapshot.LastPosition.ScrollRatio != 0.42 || snapshot.LastPosition.HeadingID != "spot" {
		t.Fatalf("expected restored reading position details, got %+v", snapshot.LastPosition)
	}
	if snapshot.LastPosition.ModifiedAt != doc.ModifiedAt || snapshot.LastPosition.Size != doc.Size {
		t.Fatalf("expected restored document metadata, got %+v for doc %+v", snapshot.LastPosition, doc)
	}
}

func TestReadingSessionPersistsAndRestoresOpenTabs(t *testing.T) {
	appDataRoot := t.TempDir()
	workspace := t.TempDir()
	firstPath := filepath.Join(workspace, "README.md")
	secondPath := filepath.Join(workspace, "guide.md")
	if err := os.WriteFile(firstPath, []byte("# First\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(secondPath, []byte("# Guide\n\n## Spot\n\nBody\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if err := app.initAppData(appDataRoot); err != nil {
		t.Fatal(err)
	}
	if _, err := app.ScanWorkspace(workspace); err != nil {
		t.Fatal(err)
	}
	firstDoc, err := app.OpenDocument(firstPath)
	if err != nil {
		t.Fatal(err)
	}
	secondDoc, err := app.OpenDocument(secondPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.SaveOpenTabs([]string{firstDoc.Path, secondDoc.Path}, secondDoc.Path); err != nil {
		t.Fatal(err)
	}
	if err := app.SaveReadingPosition(secondDoc.Path, 320, 0.5, "spot", secondDoc.ModifiedAt, secondDoc.Size); err != nil {
		t.Fatal(err)
	}

	nextApp := NewApp()
	if err := nextApp.initAppData(appDataRoot); err != nil {
		t.Fatal(err)
	}
	if _, err := nextApp.RestoreWorkspace(); err != nil {
		t.Fatal(err)
	}
	session, err := nextApp.GetReadingSession()
	if err != nil {
		t.Fatal(err)
	}
	if len(session.OpenTabs) != 2 {
		t.Fatalf("expected two restored tabs, got %+v", session)
	}
	if session.OpenTabs[0].Path != filepath.Clean(firstPath) || session.OpenTabs[1].Path != filepath.Clean(secondPath) {
		t.Fatalf("expected restored tab order, got %+v", session.OpenTabs)
	}
	if session.ActiveDocument != filepath.Clean(secondPath) {
		t.Fatalf("expected active tab %q, got %+v", secondPath, session)
	}
	if session.ActivePosition == nil || session.ActivePosition.ScrollTop != 320 || session.ActivePosition.HeadingID != "spot" {
		t.Fatalf("expected active reading position, got %+v", session.ActivePosition)
	}
}

func TestSaveOpenTabsClearsClosedDocumentPositions(t *testing.T) {
	appDataRoot := t.TempDir()
	workspace := t.TempDir()
	firstPath := filepath.Join(workspace, "README.md")
	secondPath := filepath.Join(workspace, "guide.md")
	if err := os.WriteFile(firstPath, []byte("# First\n\nBody\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(secondPath, []byte("# Guide\n\nBody\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if err := app.initAppData(appDataRoot); err != nil {
		t.Fatal(err)
	}
	if _, err := app.ScanWorkspace(workspace); err != nil {
		t.Fatal(err)
	}
	firstDoc, err := app.OpenDocument(firstPath)
	if err != nil {
		t.Fatal(err)
	}
	secondDoc, err := app.OpenDocument(secondPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.SaveOpenTabs([]string{firstDoc.Path, secondDoc.Path}, secondDoc.Path); err != nil {
		t.Fatal(err)
	}
	if err := app.SaveReadingPosition(firstDoc.Path, 120, 0.25, "", firstDoc.ModifiedAt, firstDoc.Size); err != nil {
		t.Fatal(err)
	}
	if err := app.SaveReadingPosition(secondDoc.Path, 360, 0.75, "", secondDoc.ModifiedAt, secondDoc.Size); err != nil {
		t.Fatal(err)
	}

	if err := app.SaveOpenTabs([]string{firstDoc.Path}, firstDoc.Path); err != nil {
		t.Fatal(err)
	}
	firstPosition, err := app.GetReadingPosition(firstDoc.Path)
	if err != nil {
		t.Fatal(err)
	}
	if firstPosition == nil || firstPosition.ScrollTop != 120 {
		t.Fatalf("expected kept first document position, got %+v", firstPosition)
	}
	secondPosition, err := app.GetReadingPosition(secondDoc.Path)
	if err != nil {
		t.Fatal(err)
	}
	if secondPosition != nil {
		t.Fatalf("expected closed document position to be cleared, got %+v", secondPosition)
	}

	if err := app.SaveOpenTabs(nil, ""); err != nil {
		t.Fatal(err)
	}
	firstPosition, err = app.GetReadingPosition(firstDoc.Path)
	if err != nil {
		t.Fatal(err)
	}
	if firstPosition != nil {
		t.Fatalf("expected all document positions to be cleared after closing all tabs, got %+v", firstPosition)
	}
}

func TestGetReadingPositionIgnoresStaleClosedDocumentMemory(t *testing.T) {
	appDataRoot := t.TempDir()
	workspace := t.TempDir()
	firstPath := filepath.Join(workspace, "README.md")
	secondPath := filepath.Join(workspace, "guide.md")
	if err := os.WriteFile(firstPath, []byte("# First\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(secondPath, []byte("# Guide\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if err := app.initAppData(appDataRoot); err != nil {
		t.Fatal(err)
	}
	if _, err := app.ScanWorkspace(workspace); err != nil {
		t.Fatal(err)
	}
	firstDoc, err := app.OpenDocument(firstPath)
	if err != nil {
		t.Fatal(err)
	}
	secondDoc, err := app.OpenDocument(secondPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := app.SaveOpenTabs([]string{firstDoc.Path, secondDoc.Path}, secondDoc.Path); err != nil {
		t.Fatal(err)
	}
	if err := app.SaveReadingPosition(secondDoc.Path, 360, 0.75, "", secondDoc.ModifiedAt, secondDoc.Size); err != nil {
		t.Fatal(err)
	}
	firstRelative, ok := app.workspaceRelativePath(firstDoc.Path)
	if !ok {
		t.Fatal("expected first relative path")
	}
	secondRelative, ok := app.workspaceRelativePath(secondDoc.Path)
	if !ok {
		t.Fatal("expected second relative path")
	}

	app.mu.Lock()
	workspaceLog := app.readingMemory.Workspaces[workspaceMemoryKey(workspace)]
	workspaceLog.OpenTabs = []string{firstRelative}
	workspaceLog.ActiveDocument = firstRelative
	workspaceLog.LastDocument = firstRelative
	workspaceLog.Documents[secondRelative] = DocumentReadingState{
		RelativePath: secondRelative,
		ScrollTop:    360,
		ScrollRatio:  0.75,
		ModifiedAt:   secondDoc.ModifiedAt,
		Size:         secondDoc.Size,
	}
	app.mu.Unlock()

	position, err := app.GetReadingPosition(secondDoc.Path)
	if err != nil {
		t.Fatal(err)
	}
	if position != nil {
		t.Fatalf("expected stale closed document memory to be ignored, got %+v", position)
	}
}

func TestSaveOpenTabsRejectsOutsideWorkspace(t *testing.T) {
	appDataRoot := t.TempDir()
	workspace := t.TempDir()
	inside := filepath.Join(workspace, "README.md")
	outside := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(inside, []byte("# Inside\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outside, []byte("# Outside\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if err := app.initAppData(appDataRoot); err != nil {
		t.Fatal(err)
	}
	if _, err := app.ScanWorkspace(workspace); err != nil {
		t.Fatal(err)
	}
	if err := app.SaveOpenTabs([]string{inside, outside}, inside); err == nil {
		t.Fatal("expected outside workspace tab to be rejected")
	}
}

func TestReadingSessionNormalizesLegacyLastDocument(t *testing.T) {
	workspace := t.TempDir()
	docPath := filepath.Join(workspace, "README.md")
	if err := os.WriteFile(docPath, []byte("# Legacy\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if _, err := app.ScanWorkspace(workspace); err != nil {
		t.Fatal(err)
	}
	app.readingMemory = ReadingMemoryStore{
		StorageVersion: currentReadingMemoryVersion,
		Workspaces: map[string]*WorkspaceReadingLog{
			workspaceMemoryKey(workspace): {
				Root:         filepath.Clean(workspace),
				LastDocument: "README.md",
				Documents: map[string]DocumentReadingState{
					"README.md": {RelativePath: "README.md", ScrollTop: 12},
				},
			},
		},
	}
	normalizeReadingMemory(&app.readingMemory)

	session, err := app.GetReadingSession()
	if err != nil {
		t.Fatal(err)
	}
	if len(session.OpenTabs) != 1 || session.OpenTabs[0].Path != filepath.Clean(docPath) {
		t.Fatalf("expected legacy last document to become a tab, got %+v", session)
	}
	if session.ActiveDocument != filepath.Clean(docPath) {
		t.Fatalf("expected legacy last document to become active, got %+v", session)
	}
}

func TestReadingMemoryRejectsOutsideWorkspace(t *testing.T) {
	appDataRoot := t.TempDir()
	workspace := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(outside, []byte("# Outside\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if err := app.initAppData(appDataRoot); err != nil {
		t.Fatal(err)
	}
	if _, err := app.ScanWorkspace(workspace); err != nil {
		t.Fatal(err)
	}
	if err := app.SaveReadingPosition(outside, 10, 0.1, "", 1, 2); err == nil {
		t.Fatal("expected outside workspace reading position to be rejected")
	}
}

func TestInitAppDataBacksUpBadReadingMemory(t *testing.T) {
	appDataRoot := t.TempDir()
	dataDir := filepath.Join(appDataRoot, "data")
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		t.Fatal(err)
	}
	memoryPath := filepath.Join(dataDir, "reading-memory.json")
	if err := os.WriteFile(memoryPath, []byte("{bad json"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if err := app.initAppData(appDataRoot); err != nil {
		t.Fatal(err)
	}
	backups, err := filepath.Glob(memoryPath + ".bad-*")
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 1 {
		t.Fatalf("expected one bad reading memory backup, got %d", len(backups))
	}
	if app.readingMemory.StorageVersion != currentReadingMemoryVersion || len(app.readingMemory.Workspaces) != 0 {
		t.Fatalf("expected default reading memory after bad file, got %+v", app.readingMemory)
	}
}

func TestReadingMemoryPrunesOldDocuments(t *testing.T) {
	workspace := &WorkspaceReadingLog{
		LastDocument: "keep.md",
		Documents: map[string]DocumentReadingState{
			"keep.md": {RelativePath: "keep.md", UpdatedAt: 0},
		},
	}
	for index := 0; index < maxReadingMemoryDocuments+5; index++ {
		path := fmt.Sprintf("doc-%03d.md", index)
		workspace.Documents[path] = DocumentReadingState{RelativePath: path, UpdatedAt: int64(index + 1)}
	}

	pruneReadingDocuments(workspace)
	if len(workspace.Documents) != maxReadingMemoryDocuments {
		t.Fatalf("expected %d documents after pruning, got %d", maxReadingMemoryDocuments, len(workspace.Documents))
	}
	if _, ok := workspace.Documents["keep.md"]; !ok {
		t.Fatal("expected last document to survive pruning")
	}
	if _, ok := workspace.Documents["doc-000.md"]; ok {
		t.Fatal("expected oldest non-last document to be pruned")
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

func TestSettingsMigratesLegacyLastWorkspaceToWorkspaceList(t *testing.T) {
	appDataRoot := t.TempDir()
	workspace := t.TempDir()
	configDir := filepath.Join(appDataRoot, "config")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(configDir, "settings.json")
	legacy := Settings{
		StorageVersion: 2,
		LastWorkspace:  workspace,
		Locale:         "en",
		Theme:          "dark",
	}
	data, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, data, 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if err := app.initAppData(appDataRoot); err != nil {
		t.Fatal(err)
	}
	collection, err := app.RestoreWorkspaces()
	if err != nil {
		t.Fatal(err)
	}
	if collection == nil || len(collection.Workspaces) != 1 {
		t.Fatalf("expected one migrated workspace, got %+v", collection)
	}
	if collection.Workspaces[0].Root.Path != filepath.Clean(workspace) {
		t.Fatalf("expected migrated workspace path %q, got %+v", workspace, collection.Workspaces[0])
	}
	if app.settings.ActiveWorkspaceID == "" || app.settings.LastWorkspace != filepath.Clean(workspace) {
		t.Fatalf("expected active migrated workspace, got %+v", app.settings)
	}
}

func TestMultipleWorkspacesPersistReorderAndRemove(t *testing.T) {
	appDataRoot := t.TempDir()
	first := t.TempDir()
	second := t.TempDir()
	firstDoc := filepath.Join(first, "README.md")
	if err := os.WriteFile(firstDoc, []byte("# First\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(second, "README.md"), []byte("# Second\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if err := app.initAppData(appDataRoot); err != nil {
		t.Fatal(err)
	}
	if _, err := app.ScanWorkspace(first); err != nil {
		t.Fatal(err)
	}
	collection, err := app.AddWorkspace(second)
	if err != nil {
		t.Fatal(err)
	}
	if len(collection.Workspaces) != 2 {
		t.Fatalf("expected two workspaces, got %+v", collection)
	}
	firstID := collection.Workspaces[0].Workspace.ID
	secondID := collection.Workspaces[1].Workspace.ID
	collection, err = app.ReorderWorkspaces([]string{secondID, firstID})
	if err != nil {
		t.Fatal(err)
	}
	if collection.Workspaces[0].Workspace.ID != secondID {
		t.Fatalf("expected second workspace first after reorder, got %+v", collection.Workspaces)
	}
	if err := app.SaveOpenTabs([]string{firstDoc}, firstDoc); err != nil {
		t.Fatal(err)
	}
	collection, err = app.RemoveWorkspace(firstID)
	if err != nil {
		t.Fatal(err)
	}
	if len(collection.Workspaces) != 1 || collection.Workspaces[0].Workspace.ID != secondID {
		t.Fatalf("expected only second workspace after remove, got %+v", collection)
	}
	if _, err := os.Stat(first); err != nil {
		t.Fatalf("remove workspace must not delete disk folder: %v", err)
	}
	if log := app.readingMemory.Workspaces[workspaceMemoryKey(first)]; log != nil && len(log.OpenTabs) != 0 {
		t.Fatalf("expected removed workspace open session to be cleared, got %+v", log)
	}
}

func TestLaunchArgsAddWorkspaceAndOpenMarkdownFile(t *testing.T) {
	appDataRoot := t.TempDir()
	workspace := t.TempDir()
	docPath := filepath.Join(workspace, "README.md")
	if err := os.WriteFile(docPath, []byte("# Launch\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	if err := app.initAppData(appDataRoot); err != nil {
		t.Fatal(err)
	}
	request, err := app.launchRequestFromArgs([]string{"--open-file", docPath}, "")
	if err != nil {
		t.Fatal(err)
	}
	if request == nil || request.DocumentPath != filepath.Clean(docPath) || request.Collection == nil || len(request.Collection.Workspaces) != 1 {
		t.Fatalf("expected launch request to add parent workspace and open file, got %+v", request)
	}
	if app.settings.Workspaces[0].Path != filepath.Clean(workspace) {
		t.Fatalf("expected parent folder workspace, got %+v", app.settings.Workspaces)
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
