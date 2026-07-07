package main

import (
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
