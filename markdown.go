package main

import (
	"bytes"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

const maxMarkdownBytes = 10 * 1024 * 1024

var markdownRenderer = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
)

var markdownPolicy = func() *bluemonday.Policy {
	policy := bluemonday.UGCPolicy()
	policy.AllowRelativeURLs(true)
	policy.AllowAttrs("id").OnElements("h1", "h2", "h3", "h4", "h5", "h6")
	policy.AllowAttrs("data-kern-document", "data-kern-heading").OnElements("a")
	policy.AllowAttrs("class").OnElements("pre", "code")
	return policy
}()

func (a *App) renderMarkdownDocument(path string) (*Document, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, errors.New("document path is a directory")
	}
	if info.Size() > maxMarkdownBytes {
		return nil, errors.New("document is too large")
	}

	source, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	reader := text.NewReader(source)
	doc := markdownRenderer.Parser().Parse(reader)
	outline := collectHeadings(doc, source)
	a.rewriteDocumentLinks(doc, path)

	var html bytes.Buffer
	if err := markdownRenderer.Renderer().Render(&html, source, doc); err != nil {
		return nil, err
	}
	safeHTML := markdownPolicy.SanitizeBytes(html.Bytes())

	title := filepath.Base(path)
	if len(outline) > 0 && outline[0].Text != "" {
		title = outline[0].Text
	}

	return &Document{
		Path:       path,
		Name:       filepath.Base(path),
		Title:      title,
		HTML:       string(safeHTML),
		Outline:    outline,
		ModifiedAt: info.ModTime().UnixMilli(),
		Size:       info.Size(),
	}, nil
}

func (a *App) rewriteDocumentLinks(doc ast.Node, documentPath string) {
	documentDir := filepath.Dir(documentPath)
	_ = ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch typed := node.(type) {
		case *ast.Image:
			raw := string(typed.Destination)
			if assetURL, ok := a.resolveImageURL(documentDir, raw); ok {
				typed.Destination = []byte(assetURL)
			} else if isLocalDestination(raw) {
				typed.Destination = nil
			}
		case *ast.Link:
			target, heading, ok := a.resolveMarkdownLink(documentDir, string(typed.Destination))
			if ok {
				typed.Destination = []byte("#")
				typed.SetAttributeString("data-kern-document", target)
				if heading != "" {
					typed.SetAttributeString("data-kern-heading", heading)
				}
			}
		}
		return ast.WalkContinue, nil
	})
}

func (a *App) resolveImageURL(baseDir string, raw string) (string, bool) {
	target, _, ok := a.resolveRelativeDestination(baseDir, raw)
	if !ok || !isImageFile(target) || !a.isWithinWorkspace(target) {
		return "", false
	}
	rel, ok := a.workspaceRelativePath(target)
	if !ok {
		return "", false
	}
	return "/kern-asset?path=" + url.QueryEscape(rel), true
}

func (a *App) resolveMarkdownLink(baseDir string, raw string) (string, string, bool) {
	target, fragment, ok := a.resolveRelativeDestination(baseDir, raw)
	if !ok || !isMarkdownFile(target) || !a.isWithinWorkspace(target) {
		return "", "", false
	}
	rel, ok := a.workspaceRelativePath(target)
	if !ok {
		return "", "", false
	}
	return rel, fragment, true
}

func (a *App) resolveRelativeDestination(baseDir string, raw string) (string, string, bool) {
	if raw == "" || strings.HasPrefix(raw, "#") || strings.HasPrefix(raw, "//") {
		return "", "", false
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "" || parsed.Host != "" || parsed.Path == "" {
		return "", "", false
	}
	unescapedPath, err := url.PathUnescape(parsed.Path)
	if err != nil {
		return "", "", false
	}
	localPath := filepath.Clean(filepath.FromSlash(unescapedPath))
	if filepath.IsAbs(localPath) {
		return "", "", false
	}
	target := filepath.Clean(filepath.Join(baseDir, localPath))
	return target, parsed.Fragment, true
}

func isLocalDestination(raw string) bool {
	if raw == "" || strings.HasPrefix(raw, "#") || strings.HasPrefix(raw, "//") {
		return false
	}
	parsed, err := url.Parse(raw)
	return err == nil && parsed.Scheme == "" && parsed.Host == ""
}

func collectHeadings(doc ast.Node, source []byte) []Heading {
	headings := []Heading{}
	_ = ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		heading, ok := node.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}
		text := nodeText(heading, source)
		if text == "" {
			return ast.WalkContinue, nil
		}
		id := ""
		if value, ok := heading.AttributeString("id"); ok {
			switch typed := value.(type) {
			case string:
				id = typed
			case []byte:
				id = string(typed)
			}
		}
		headings = append(headings, Heading{
			ID:    id,
			Level: heading.Level,
			Text:  text,
		})
		return ast.WalkContinue, nil
	})
	return headings
}

func nodeText(node ast.Node, source []byte) string {
	var builder strings.Builder
	_ = ast.Walk(node, func(child ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || child == node {
			return ast.WalkContinue, nil
		}
		if textNode, ok := child.(*ast.Text); ok {
			builder.Write(textNode.Segment.Value(source))
			if textNode.SoftLineBreak() || textNode.HardLineBreak() {
				builder.WriteByte(' ')
			}
		}
		return ast.WalkContinue, nil
	})
	return strings.Join(strings.Fields(builder.String()), " ")
}
