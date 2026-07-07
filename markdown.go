package main

import (
	"bytes"
	"errors"
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
	policy.AllowAttrs("id").OnElements("h1", "h2", "h3", "h4", "h5", "h6")
	return policy
}()

func renderMarkdownDocument(path string) (*Document, error) {
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
		Path:    path,
		Name:    filepath.Base(path),
		Title:   title,
		HTML:    string(safeHTML),
		Outline: outline,
	}, nil
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
