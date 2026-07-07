package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed internal/i18n/locales/*.json
var localeFiles embed.FS

// App struct
type App struct {
	ctx               context.Context
	mu                sync.RWMutex
	workspaceRoot     string
	workspaceRootReal string
}

type ProductInfo struct {
	AppID       string            `json:"appId"`
	AppSlug     string            `json:"appSlug"`
	DisplayName string            `json:"displayName"`
	Version     string            `json:"version"`
	Repository  string            `json:"repository"`
	BrandParts  map[string]string `json:"brandParts"`
	Languages   []LocaleOption    `json:"languages"`
}

type LocaleOption struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

type Bootstrap struct {
	Product        ProductInfo       `json:"product"`
	CurrentLocale  string            `json:"currentLocale"`
	ShellLocale    map[string]string `json:"shellLocale"`
	BusinessLocale map[string]string `json:"businessLocale"`
}

type WorkspaceTree struct {
	Root TreeNode `json:"root"`
}

type TreeNode struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	Type     string     `json:"type"`
	Children []TreeNode `json:"children,omitempty"`
}

type Document struct {
	Path    string    `json:"path"`
	Name    string    `json:"name"`
	Title   string    `json:"title"`
	HTML    string    `json:"html"`
	Outline []Heading `json:"outline"`
}

type Heading struct {
	ID    string `json:"id"`
	Level int    `json:"level"`
	Text  string `json:"text"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GetBootstrap(locale string) (Bootstrap, error) {
	locale = normalizeLocale(locale)
	messages, err := loadLocale(locale)
	if err != nil {
		return Bootstrap{}, err
	}

	return Bootstrap{
		Product: ProductInfo{
			AppID:       "js.kern-md",
			AppSlug:     "jskernmd",
			DisplayName: "JS Kern.md",
			Version:     "0.1.0",
			Repository:  "jskern.md",
			BrandParts: map[string]string{
				"prefix": "js",
				"core":   "kern",
				"suffix": ".md",
			},
			Languages: []LocaleOption{
				{Code: "zh-CN", Label: "中文"},
				{Code: "en", Label: "English"},
			},
		},
		CurrentLocale:  locale,
		ShellLocale:    messages["shell"],
		BusinessLocale: messages["business"],
	}, nil
}

func (a *App) OpenWorkspace() (*WorkspaceTree, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open Workspace",
	})
	if err != nil || dir == "" {
		return nil, err
	}
	return a.ScanWorkspace(dir)
}

func (a *App) ScanWorkspace(root string) (*WorkspaceTree, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	node, err := scanNode(abs, 0)
	if err != nil {
		return nil, err
	}
	realRoot, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return nil, err
	}
	a.mu.Lock()
	a.workspaceRoot = filepath.Clean(abs)
	a.workspaceRootReal = filepath.Clean(realRoot)
	a.mu.Unlock()
	return &WorkspaceTree{Root: node}, nil
}

func (a *App) OpenDocument(path string) (*Document, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if !isMarkdownFile(abs) {
		return nil, errors.New("not a markdown document")
	}
	if !a.isWithinWorkspace(abs) {
		return nil, errors.New("document is outside the current workspace")
	}
	return renderMarkdownDocument(abs)
}

func (a *App) isWithinWorkspace(path string) bool {
	a.mu.RLock()
	root := a.workspaceRoot
	realRoot := a.workspaceRootReal
	a.mu.RUnlock()
	if root == "" || realRoot == "" {
		return false
	}
	cleanPath := filepath.Clean(path)
	rel, err := filepath.Rel(root, cleanPath)
	if err != nil {
		return false
	}
	if rel != "." && (strings.HasPrefix(rel, "..") || filepath.IsAbs(rel)) {
		return false
	}
	realPath, err := filepath.EvalSymlinks(cleanPath)
	if err != nil {
		return false
	}
	realRel, err := filepath.Rel(realRoot, filepath.Clean(realPath))
	if err != nil {
		return false
	}
	return realRel == "." || (!strings.HasPrefix(realRel, "..") && !filepath.IsAbs(realRel))
}

func normalizeLocale(locale string) string {
	switch locale {
	case "en":
		return "en"
	default:
		return "zh-CN"
	}
}

func loadLocale(locale string) (map[string]map[string]string, error) {
	data, err := localeFiles.ReadFile("internal/i18n/locales/" + locale + ".json")
	if err != nil {
		return nil, err
	}
	var messages map[string]map[string]string
	err = json.Unmarshal(data, &messages)
	return messages, err
}

func scanNode(path string, depth int) (TreeNode, error) {
	info, err := os.Stat(path)
	if err != nil {
		return TreeNode{}, err
	}

	nodeType := "file"
	if info.IsDir() {
		nodeType = "directory"
	}
	node := TreeNode{Name: info.Name(), Path: path, Type: nodeType}
	if !info.IsDir() {
		return node, nil
	}

	// ponytail: bounded eager scan; switch to lazy child loading when huge workspaces need it.
	if depth >= 8 {
		return node, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return TreeNode{}, err
	}

	children := make([]TreeNode, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		childPath := filepath.Join(path, name)
		if shouldSkipEntry(name, entry.IsDir()) {
			continue
		}
		if !entry.IsDir() && !isMarkdownFile(name) {
			continue
		}
		child, err := scanNode(childPath, depth+1)
		if err != nil {
			continue
		}
		children = append(children, child)
	}

	sort.Slice(children, func(i, j int) bool {
		if children[i].Type != children[j].Type {
			return children[i].Type == "directory"
		}
		return strings.ToLower(children[i].Name) < strings.ToLower(children[j].Name)
	})
	node.Children = children
	return node, nil
}

func shouldSkipEntry(name string, isDir bool) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	if !isDir {
		return false
	}
	switch strings.ToLower(name) {
	case "node_modules", "dist", "build", "vendor":
		return true
	default:
		return false
	}
}

func isMarkdownFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".md" || ext == ".markdown" || ext == ".mdown"
}
