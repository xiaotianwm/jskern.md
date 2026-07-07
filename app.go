package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	appSlug                = "jskernmd"
	currentSettingsVersion = 1
)

//go:embed internal/i18n/locales/*.json
var localeFiles embed.FS

// App struct
type App struct {
	ctx               context.Context
	mu                sync.RWMutex
	workspaceRoot     string
	workspaceRootReal string
	appDataRoot       string
	settingsPath      string
	settings          Settings
}

type workspaceAssetHandler struct {
	app *App
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

type Settings struct {
	StorageVersion int    `json:"storage_version"`
	LastWorkspace  string `json:"last_workspace"`
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
	Path       string    `json:"path"`
	Name       string    `json:"name"`
	Title      string    `json:"title"`
	HTML       string    `json:"html"`
	Outline    []Heading `json:"outline"`
	ModifiedAt int64     `json:"modifiedAt"`
	Size       int64     `json:"size"`
}

type DocumentStatus struct {
	Path       string `json:"path"`
	Exists     bool   `json:"exists"`
	IsDocument bool   `json:"isDocument"`
	Changed    bool   `json:"changed"`
	ModifiedAt int64  `json:"modifiedAt"`
	Size       int64  `json:"size"`
}

type Heading struct {
	ID    string `json:"id"`
	Level int    `json:"level"`
	Text  string `json:"text"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{settings: defaultSettings()}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	_ = a.initAppData("")
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

func (a *App) RestoreWorkspace() (*WorkspaceTree, error) {
	a.mu.RLock()
	lastWorkspace := a.settings.LastWorkspace
	a.mu.RUnlock()
	if lastWorkspace == "" {
		return nil, nil
	}
	info, err := os.Stat(lastWorkspace)
	if err != nil || !info.IsDir() {
		return nil, nil
	}
	return a.ScanWorkspace(lastWorkspace)
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
	a.setLastWorkspace(abs)
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
	return a.renderMarkdownDocument(abs)
}

func (a *App) OpenWorkspaceDocument(path string) (*Document, error) {
	abs, err := a.workspacePath(path)
	if err != nil {
		return nil, err
	}
	return a.OpenDocument(abs)
}

func (a *App) StatDocument(path string, knownModifiedAt int64, knownSize int64) (*DocumentStatus, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if !isMarkdownFile(abs) {
		return nil, errors.New("not a markdown document")
	}
	if !a.isLexicallyWithinWorkspace(abs) {
		return nil, errors.New("document is outside the current workspace")
	}

	info, err := os.Stat(abs)
	if errors.Is(err, os.ErrNotExist) {
		return &DocumentStatus{
			Path:    filepath.Clean(abs),
			Exists:  false,
			Changed: true,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	if !a.isWithinWorkspace(abs) {
		return nil, errors.New("document is outside the current workspace")
	}

	modifiedAt := info.ModTime().UnixMilli()
	size := info.Size()
	isDocument := !info.IsDir() && isMarkdownFile(abs)
	return &DocumentStatus{
		Path:       filepath.Clean(abs),
		Exists:     true,
		IsDocument: isDocument,
		Changed:    !isDocument || modifiedAt != knownModifiedAt || size != knownSize,
		ModifiedAt: modifiedAt,
		Size:       size,
	}, nil
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

func (a *App) isLexicallyWithinWorkspace(path string) bool {
	a.mu.RLock()
	root := a.workspaceRoot
	a.mu.RUnlock()
	if root == "" {
		return false
	}
	rel, err := filepath.Rel(root, filepath.Clean(path))
	if err != nil {
		return false
	}
	return rel == "." || (!strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel))
}

func (a *App) workspacePath(path string) (string, error) {
	a.mu.RLock()
	root := a.workspaceRoot
	a.mu.RUnlock()
	if root == "" {
		return "", errors.New("workspace is not open")
	}
	if strings.TrimSpace(path) == "" {
		return "", errors.New("empty workspace path")
	}
	localPath := filepath.Clean(filepath.FromSlash(path))
	if filepath.IsAbs(localPath) || localPath == ".." || strings.HasPrefix(localPath, ".."+string(filepath.Separator)) {
		return "", errors.New("workspace path is outside the current workspace")
	}
	return filepath.Join(root, localPath), nil
}

func (a *App) workspaceRelativePath(path string) (string, bool) {
	a.mu.RLock()
	root := a.workspaceRoot
	a.mu.RUnlock()
	if root == "" {
		return "", false
	}
	rel, err := filepath.Rel(root, filepath.Clean(path))
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", false
	}
	return filepath.ToSlash(rel), true
}

func (a *App) assetHandler() http.Handler {
	return workspaceAssetHandler{app: a}
}

func (h workspaceAssetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/kern-asset" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	target, err := h.app.workspacePath(r.URL.Query().Get("path"))
	if err != nil || !h.app.isWithinWorkspace(target) || !isImageFile(target) {
		http.NotFound(w, r)
		return
	}
	info, err := os.Stat(target)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, target)
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

func (a *App) initAppData(root string) error {
	if root == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return err
		}
		root = filepath.Join(configDir, appSlug)
	}
	for _, name := range []string{"config", "data", "logs", "cache", "temp", "runtime", "crash"} {
		if err := os.MkdirAll(filepath.Join(root, name), 0o700); err != nil {
			return err
		}
	}
	settingsPath := filepath.Join(root, "config", "settings.json")
	settings, err := loadSettings(settingsPath)
	if err != nil {
		return err
	}
	if err := saveSettings(settingsPath, settings); err != nil {
		return err
	}
	a.mu.Lock()
	a.appDataRoot = root
	a.settingsPath = settingsPath
	a.settings = settings
	a.mu.Unlock()
	return nil
}

func (a *App) setLastWorkspace(path string) {
	a.mu.Lock()
	if a.settings.StorageVersion == 0 {
		a.settings = defaultSettings()
	}
	a.settings.LastWorkspace = filepath.Clean(path)
	settingsPath := a.settingsPath
	settings := a.settings
	a.mu.Unlock()
	if settingsPath == "" {
		return
	}
	_ = saveSettings(settingsPath, settings)
}

func defaultSettings() Settings {
	return Settings{StorageVersion: currentSettingsVersion}
}

func loadSettings(path string) (Settings, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return defaultSettings(), nil
	}
	if err != nil {
		return Settings{}, err
	}
	settings := defaultSettings()
	if err := json.Unmarshal(data, &settings); err != nil {
		if backupErr := backupBadFile(path); backupErr != nil {
			return Settings{}, backupErr
		}
		return defaultSettings(), nil
	}
	if settings.StorageVersion == 0 {
		settings.StorageVersion = currentSettingsVersion
	}
	return settings, nil
}

func saveSettings(path string, settings Settings) error {
	settings.StorageVersion = currentSettingsVersion
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	dir := filepath.Dir(path)
	temp, err := os.CreateTemp(dir, "settings-*.tmp")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer func() {
		_ = os.Remove(tempPath)
	}()
	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}

func backupBadFile(path string) error {
	stamp := time.Now().UTC().Format("20060102T150405Z")
	backupPath := path + ".bad-" + stamp
	if _, err := os.Stat(backupPath); err == nil {
		backupPath = path + ".bad-" + stamp + "-" + strings.ReplaceAll(time.Now().UTC().Format("150405.000000000"), ".", "")
	}
	return os.Rename(path, backupPath)
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

func isImageFile(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".avif":
		return true
	default:
		return false
	}
}
