package main

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	currentSettingsVersion = 3
	githubReleasesAPI      = "https://api.github.com/repos/xiaotianwm/jskern.md/releases"
)

//go:embed internal/i18n/locales/*.json
var localeFiles embed.FS

//go:embed product.manifest.json
var productManifest []byte

var productInfo = mustLoadProductInfo()

// App struct
type App struct {
	ctx               context.Context
	mu                sync.RWMutex
	workspaceScanMu   sync.Mutex
	workspaceRoot     string
	workspaceRootReal string
	workspaceTreeSig  string
	workspaceStates   map[string]workspaceRuntimeState
	appDataRoot       string
	settingsPath      string
	readingMemoryPath string
	settings          Settings
	readingMemory     ReadingMemoryStore
	pendingLaunch     *LaunchRequest
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

type productManifestFile struct {
	AppID       string         `json:"app_id"`
	AppSlug     string         `json:"app_slug"`
	ProductName string         `json:"product_name"`
	Version     string         `json:"version"`
	Repository  string         `json:"repository"`
	Languages   []LocaleOption `json:"languages"`
}

type Bootstrap struct {
	Product        ProductInfo       `json:"product"`
	CurrentLocale  string            `json:"currentLocale"`
	CurrentTheme   string            `json:"currentTheme"`
	ShellLocale    map[string]string `json:"shellLocale"`
	BusinessLocale map[string]string `json:"businessLocale"`
}

type Settings struct {
	StorageVersion       int              `json:"storage_version"`
	LastWorkspace        string           `json:"last_workspace,omitempty"`
	ActiveWorkspaceID    string           `json:"active_workspace_id,omitempty"`
	Workspaces           []WorkspaceEntry `json:"workspaces,omitempty"`
	Locale               string           `json:"locale"`
	Theme                string           `json:"theme"`
	IgnoredUpdateVersion string           `json:"ignored_update_version"`
}

type WorkspaceEntry struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Path         string `json:"path"`
	Order        int    `json:"order"`
	AddedAt      int64  `json:"addedAt"`
	LastOpenedAt int64  `json:"lastOpenedAt"`
}

type workspaceRuntimeState struct {
	Root          string
	RealRoot      string
	TreeSignature string
	Tree          TreeNode
}

type WorkspaceTree struct {
	Workspace WorkspaceEntry `json:"workspace"`
	Root      TreeNode       `json:"root"`
}

type WorkspaceCollection struct {
	Workspaces        []WorkspaceTree `json:"workspaces"`
	ActiveWorkspaceID string          `json:"activeWorkspaceId"`
}

type WorkspaceRefresh struct {
	Changed    bool                 `json:"changed"`
	Tree       *WorkspaceTree       `json:"tree,omitempty"`
	Collection *WorkspaceCollection `json:"collection,omitempty"`
}

type RenameResult struct {
	OldPath  string         `json:"oldPath"`
	NewPath  string         `json:"newPath"`
	NodeType string         `json:"nodeType"`
	Tree     *WorkspaceTree `json:"tree,omitempty"`
}

type TreeNode struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	Type     string     `json:"type"`
	Loaded   bool       `json:"loaded"`
	Children []TreeNode `json:"children,omitempty"`
}

type LaunchRequest struct {
	Collection   *WorkspaceCollection `json:"collection,omitempty"`
	DocumentPath string               `json:"documentPath,omitempty"`
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

type UpdateInfo struct {
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion"`
	UpdateAvailable bool   `json:"updateAvailable"`
	Ignored         bool   `json:"ignored"`
	ReleaseURL      string `json:"releaseUrl"`
	DownloadURL     string `json:"downloadUrl"`
	Sha256          string `json:"sha256"`
	ReleaseNotes    string `json:"releaseNotes"`
	DownloadedPath  string `json:"downloadedPath"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		settings:        defaultSettings(),
		readingMemory:   defaultReadingMemory(),
		workspaceStates: map[string]workspaceRuntimeState{},
	}
}

func mustLoadProductInfo() ProductInfo {
	var manifest productManifestFile
	if err := json.Unmarshal(productManifest, &manifest); err != nil {
		panic(fmt.Sprintf("load product manifest: %v", err))
	}
	product := ProductInfo{
		AppID:       strings.TrimSpace(manifest.AppID),
		AppSlug:     strings.TrimSpace(manifest.AppSlug),
		DisplayName: strings.TrimSpace(manifest.ProductName),
		Version:     strings.TrimSpace(manifest.Version),
		Repository:  strings.TrimSpace(manifest.Repository),
		BrandParts: map[string]string{
			"prefix": "js",
			"core":   "kern",
			"suffix": ".md",
		},
		Languages: append([]LocaleOption(nil), manifest.Languages...),
	}
	if product.AppSlug == "" || product.Version == "" {
		panic("product manifest must include app_slug and version")
	}
	return product
}

func productInfoCopy() ProductInfo {
	product := productInfo
	product.BrandParts = map[string]string{}
	for key, value := range productInfo.BrandParts {
		product.BrandParts[key] = value
	}
	product.Languages = append([]LocaleOption(nil), productInfo.Languages...)
	return product
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	_ = a.initAppData("")
	_ = a.queueLaunchArgs(os.Args[1:], mustGetwd())
}

func (a *App) GetBootstrap(locale string) (Bootstrap, error) {
	a.mu.RLock()
	settingsLocale := a.settings.Locale
	theme := normalizeTheme(a.settings.Theme)
	a.mu.RUnlock()
	if strings.TrimSpace(locale) == "" {
		locale = settingsLocale
	}
	locale = normalizeLocale(locale)
	messages, err := loadLocale(locale)
	if err != nil {
		return Bootstrap{}, err
	}

	return Bootstrap{
		Product:        productInfoCopy(),
		CurrentLocale:  locale,
		CurrentTheme:   theme,
		ShellLocale:    messages["shell"],
		BusinessLocale: messages["business"],
	}, nil
}

func (a *App) SwitchLanguage(locale string) (Bootstrap, error) {
	locale = normalizeLocale(locale)
	a.updateSettings(func(settings *Settings) {
		settings.Locale = locale
	})
	return a.GetBootstrap("")
}

func (a *App) SwitchTheme(theme string) (Bootstrap, error) {
	theme = normalizeTheme(theme)
	a.updateSettings(func(settings *Settings) {
		settings.Theme = theme
	})
	return a.GetBootstrap("")
}

func (a *App) CheckForUpdates() (*UpdateInfo, error) {
	info, err := checkGitHubUpdates(a.ctx, githubReleasesAPI, productInfo.Version)
	if err != nil {
		return nil, err
	}
	a.mu.RLock()
	ignoredVersion := a.settings.IgnoredUpdateVersion
	a.mu.RUnlock()
	if info.UpdateAvailable && info.LatestVersion == ignoredVersion {
		info.UpdateAvailable = false
		info.Ignored = true
	}
	return info, nil
}

func (a *App) DismissUpdate(version string) error {
	version = normalizeVersion(version)
	if version == "" {
		return errors.New("empty update version")
	}
	a.updateSettings(func(settings *Settings) {
		settings.IgnoredUpdateVersion = version
	})
	return nil
}

func (a *App) DownloadUpdate(downloadURL string, sha256Hex string) (*UpdateInfo, error) {
	if !isAllowedUpdateDownloadURL(downloadURL) {
		return nil, errors.New("unsupported update download URL")
	}
	if a.appDataRoot == "" {
		if err := a.initAppData(""); err != nil {
			return nil, err
		}
	}
	parsed, err := url.Parse(downloadURL)
	if err != nil {
		return nil, err
	}
	fileName := filepath.Base(parsed.Path)
	if fileName == "." || fileName == string(filepath.Separator) || !strings.HasSuffix(strings.ToLower(fileName), ".exe") {
		return nil, errors.New("unsupported update installer")
	}
	updateDir := filepath.Join(a.appDataRoot, "temp", "update")
	if err := os.MkdirAll(updateDir, 0o700); err != nil {
		return nil, err
	}
	target := filepath.Join(updateDir, fileName)
	if err := downloadFile(a.ctx, downloadURL, target, sha256Hex); err != nil {
		return nil, err
	}
	return &UpdateInfo{
		CurrentVersion: productInfo.Version,
		DownloadedPath: target,
		DownloadURL:    downloadURL,
		Sha256:         strings.ToLower(strings.TrimSpace(sha256Hex)),
	}, nil
}

func (a *App) OpenDownloadedUpdate(path string) error {
	if path == "" {
		return errors.New("empty update installer path")
	}
	if a.appDataRoot == "" {
		return errors.New("app data root is not initialized")
	}
	cleanPath := filepath.Clean(path)
	updateDir := filepath.Join(a.appDataRoot, "temp", "update")
	rel, err := filepath.Rel(updateDir, cleanPath)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return errors.New("update installer is outside the managed update directory")
	}
	if !strings.HasSuffix(strings.ToLower(cleanPath), ".exe") {
		return errors.New("unsupported update installer")
	}
	if _, err := os.Stat(cleanPath); err != nil {
		return err
	}
	return openFile(cleanPath)
}

func (a *App) RevealPath(path string) error {
	target, err := a.revealableWorkspacePath(path)
	if err != nil {
		return err
	}
	return revealPath(target)
}

func (a *App) RenamePath(path string, newName string) (*RenameResult, error) {
	a.workspaceScanMu.Lock()
	defer a.workspaceScanMu.Unlock()

	source, err := a.renameableWorkspacePath(path)
	if err != nil {
		return nil, err
	}
	name, err := validRenameName(newName)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(source)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() && !isMarkdownFile(name) {
		return nil, errors.New("renamed document must keep a markdown extension")
	}

	target := filepath.Join(filepath.Dir(source), name)
	target = filepath.Clean(target)
	if filepath.Clean(source) == target {
		return &RenameResult{OldPath: source, NewPath: target, NodeType: treeNodeType(info)}, nil
	}
	if !a.isLexicallyWithinWorkspace(target) {
		return nil, errors.New("rename target is outside the current workspace")
	}
	if !a.isRenameTargetWithinRealWorkspace(filepath.Dir(source), target) {
		return nil, errors.New("rename target is outside the current workspace")
	}
	if _, err := os.Stat(target); err == nil {
		return nil, errors.New("rename target already exists")
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if err := os.Rename(source, target); err != nil {
		return nil, err
	}

	node, err := a.refreshWorkspaceTreeAfterMutation(source, target, treeNodeType(info))
	if err != nil {
		return nil, err
	}
	return &RenameResult{
		OldPath:  source,
		NewPath:  target,
		NodeType: treeNodeType(info),
		Tree:     &WorkspaceTree{Root: node},
	}, nil
}

func (a *App) OpenWorkspace() (*WorkspaceCollection, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open Workspace",
	})
	if err != nil || dir == "" {
		return nil, err
	}
	return a.AddWorkspace(dir)
}

func (a *App) AddWorkspace(root string) (*WorkspaceCollection, error) {
	if _, err := a.addWorkspacePath(root, true); err != nil {
		return nil, err
	}
	return a.RestoreWorkspaces()
}

func (a *App) RemoveWorkspace(workspaceID string) (*WorkspaceCollection, error) {
	workspaceID = strings.TrimSpace(workspaceID)
	if workspaceID == "" {
		return nil, errors.New("empty workspace id")
	}
	a.workspaceScanMu.Lock()

	var removedRoot string
	a.mu.Lock()
	next := make([]WorkspaceEntry, 0, len(a.settings.Workspaces))
	for _, workspace := range a.settings.Workspaces {
		if workspace.ID == workspaceID {
			removedRoot = workspace.Path
			delete(a.workspaceStates, workspace.ID)
			continue
		}
		next = append(next, workspace)
	}
	if removedRoot == "" {
		a.mu.Unlock()
		a.workspaceScanMu.Unlock()
		return nil, errors.New("workspace was not found")
	}
	a.settings.Workspaces = next
	if a.settings.ActiveWorkspaceID == workspaceID {
		a.settings.ActiveWorkspaceID = ""
		if len(a.settings.Workspaces) > 0 {
			a.settings.ActiveWorkspaceID = a.settings.Workspaces[0].ID
		}
	}
	a.normalizeSettingsLocked()
	a.applyActiveWorkspaceLocked()
	settingsPath := a.settingsPath
	settings := a.settings
	a.clearReadingSessionForRootLocked(removedRoot)
	memoryPath := a.readingMemoryPath
	memory := a.readingMemory
	a.mu.Unlock()
	a.workspaceScanMu.Unlock()

	if settingsPath != "" {
		_ = saveSettings(settingsPath, settings)
	}
	if memoryPath != "" {
		_ = saveReadingMemory(memoryPath, memory)
	}
	return a.RestoreWorkspaces()
}

func (a *App) ReorderWorkspaces(workspaceIDs []string) (*WorkspaceCollection, error) {
	if len(workspaceIDs) == 0 {
		return a.RestoreWorkspaces()
	}
	a.workspaceScanMu.Lock()
	a.mu.Lock()
	byID := map[string]WorkspaceEntry{}
	for _, workspace := range a.settings.Workspaces {
		byID[workspace.ID] = workspace
	}
	next := make([]WorkspaceEntry, 0, len(a.settings.Workspaces))
	seen := map[string]struct{}{}
	for _, id := range workspaceIDs {
		id = strings.TrimSpace(id)
		workspace, ok := byID[id]
		if !ok {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		next = append(next, workspace)
	}
	for _, workspace := range a.settings.Workspaces {
		if _, exists := seen[workspace.ID]; !exists {
			next = append(next, workspace)
		}
	}
	a.settings.Workspaces = next
	a.normalizeSettingsLocked()
	settingsPath := a.settingsPath
	settings := a.settings
	a.mu.Unlock()
	a.workspaceScanMu.Unlock()
	if settingsPath != "" {
		_ = saveSettings(settingsPath, settings)
	}
	return a.RestoreWorkspaces()
}

func (a *App) RestoreWorkspace() (*WorkspaceTree, error) {
	collection, err := a.RestoreWorkspaces()
	if err != nil || collection == nil || len(collection.Workspaces) == 0 {
		return nil, err
	}
	for _, workspace := range collection.Workspaces {
		if workspace.Workspace.ID == collection.ActiveWorkspaceID {
			return &workspace, nil
		}
	}
	return &collection.Workspaces[0], nil
}

func (a *App) RestoreWorkspaces() (*WorkspaceCollection, error) {
	a.mu.RLock()
	workspaces := append([]WorkspaceEntry(nil), a.settings.Workspaces...)
	activeID := a.settings.ActiveWorkspaceID
	a.mu.RUnlock()
	if len(workspaces) == 0 {
		return nil, nil
	}
	return a.workspaceCollection(workspaces, activeID)
}

func (a *App) LoadDirectory(path string) (*TreeNode, error) {
	a.workspaceScanMu.Lock()
	defer a.workspaceScanMu.Unlock()

	target, err := a.revealableWorkspacePath(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(target)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, errors.New("workspace path is not a directory")
	}
	workspace, _, ok := a.workspaceForPath(target, true)
	if !ok {
		return nil, errors.New("directory is outside the current workspace")
	}

	a.mu.RLock()
	state, ok := a.workspaceStates[workspace.ID]
	a.mu.RUnlock()
	if !ok || state.Root == "" {
		return nil, errors.New("workspace is not open")
	}

	loaded, directory, err := loadTreeDirectory(state.Tree, target)
	if err != nil {
		return nil, err
	}
	state.Tree = loaded
	state.TreeSignature = treeSignature(loaded)
	a.mu.Lock()
	a.workspaceStates[workspace.ID] = state
	a.applyActiveWorkspaceLocked()
	a.mu.Unlock()
	return &directory, nil
}

func (a *App) ScanWorkspace(root string) (*WorkspaceTree, error) {
	tree, err := a.addWorkspacePath(root, true)
	if err != nil {
		return nil, err
	}
	return tree, nil
}

func (a *App) addWorkspacePath(root string, activate bool) (*WorkspaceTree, error) {
	a.workspaceScanMu.Lock()
	defer a.workspaceScanMu.Unlock()

	entry, node, state, err := scanWorkspaceRoot(root)
	if err != nil {
		return nil, err
	}

	now := time.Now().UnixMilli()
	a.mu.Lock()
	if a.workspaceStates == nil {
		a.workspaceStates = map[string]workspaceRuntimeState{}
	}
	a.normalizeSettingsLocked()

	for _, existing := range a.settings.Workspaces {
		if pathWithinRoot(entry.Path, existing.Path) {
			entry = existing
			node, state, err = a.scanExistingWorkspaceLocked(existing)
			if err != nil {
				a.mu.Unlock()
				return nil, err
			}
			if activate {
				a.settings.ActiveWorkspaceID = existing.ID
				a.settings.LastWorkspace = existing.Path
				a.applyActiveWorkspaceLocked()
			}
			settingsPath := a.settingsPath
			settings := a.settings
			a.mu.Unlock()
			if settingsPath != "" {
				_ = saveSettings(settingsPath, settings)
			}
			return &WorkspaceTree{Workspace: entry, Root: node}, nil
		}
	}

	next := make([]WorkspaceEntry, 0, len(a.settings.Workspaces)+1)
	for _, existing := range a.settings.Workspaces {
		if pathWithinRoot(existing.Path, entry.Path) {
			delete(a.workspaceStates, existing.ID)
			continue
		}
		next = append(next, existing)
	}
	entry.AddedAt = now
	entry.LastOpenedAt = now
	next = append(next, entry)
	a.settings.Workspaces = next
	a.workspaceStates[entry.ID] = state
	if activate {
		a.settings.ActiveWorkspaceID = entry.ID
		a.settings.LastWorkspace = entry.Path
	}
	a.normalizeSettingsLocked()
	a.applyActiveWorkspaceLocked()
	settingsPath := a.settingsPath
	settings := a.settings
	a.mu.Unlock()

	if settingsPath != "" {
		_ = saveSettings(settingsPath, settings)
	}
	return &WorkspaceTree{Workspace: entry, Root: node}, nil
}

func (a *App) RefreshWorkspace() (*WorkspaceRefresh, error) {
	refresh, err := a.RefreshWorkspaces()
	if err != nil || refresh == nil || refresh.Collection == nil {
		return refresh, err
	}
	for _, workspace := range refresh.Collection.Workspaces {
		if workspace.Workspace.ID == refresh.Collection.ActiveWorkspaceID {
			refresh.Tree = &workspace
			break
		}
	}
	return refresh, nil
}

func (a *App) RefreshWorkspaces() (*WorkspaceRefresh, error) {
	a.workspaceScanMu.Lock()
	defer a.workspaceScanMu.Unlock()

	a.mu.RLock()
	workspaces := append([]WorkspaceEntry(nil), a.settings.Workspaces...)
	previous := map[string]workspaceRuntimeState{}
	for id, state := range a.workspaceStates {
		previous[id] = state
	}
	activeID := a.settings.ActiveWorkspaceID
	a.mu.RUnlock()
	if len(workspaces) == 0 {
		return &WorkspaceRefresh{Changed: false}, nil
	}

	changed := false
	states := map[string]workspaceRuntimeState{}
	for _, workspace := range workspaces {
		state, existed := previous[workspace.ID]
		if !existed || state.Root == "" {
			_, _, state, _ = scanWorkspaceRoot(workspace.Path)
			changed = changed || state.Root != ""
		}
		if state.Root == "" {
			continue
		}
		node, err := refreshLoadedTree(state.Tree)
		if err != nil {
			continue
		}
		nextSignature := treeSignature(node)
		if state.TreeSignature != nextSignature {
			changed = true
		}
		state.Tree = node
		state.TreeSignature = nextSignature
		states[workspace.ID] = state
	}
	if !changed {
		return &WorkspaceRefresh{Changed: false}, nil
	}

	a.mu.Lock()
	if a.workspaceStates == nil {
		a.workspaceStates = map[string]workspaceRuntimeState{}
	}
	for id, state := range states {
		a.workspaceStates[id] = state
	}
	a.applyActiveWorkspaceLocked()
	allStates := cloneWorkspaceStates(a.workspaceStates)
	a.mu.Unlock()

	return &WorkspaceRefresh{
		Changed:    true,
		Collection: workspaceCollectionFromStates(workspaces, activeID, allStates),
	}, nil
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
	a.setActiveWorkspaceForPath(abs)
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
	_, _, ok := a.workspaceForPath(path, true)
	return ok
}

func (a *App) isLexicallyWithinWorkspace(path string) bool {
	_, _, ok := a.workspaceForPath(path, false)
	return ok
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
	_, state, ok := a.workspaceForPath(path, false)
	if !ok {
		return "", false
	}
	return workspaceRelativePathFromRoot(state.Root, path)
}

func workspaceRelativePathFromRoot(root string, path string) (string, bool) {
	rel, err := filepath.Rel(filepath.Clean(root), filepath.Clean(path))
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", false
	}
	return filepath.ToSlash(rel), true
}

func (a *App) workspaceForPath(path string, checkReal bool) (WorkspaceEntry, workspaceRuntimeState, bool) {
	cleanPath := filepath.Clean(path)
	a.mu.RLock()
	workspaces := append([]WorkspaceEntry(nil), a.settings.Workspaces...)
	states := map[string]workspaceRuntimeState{}
	for id, state := range a.workspaceStates {
		states[id] = state
	}
	a.mu.RUnlock()

	sort.SliceStable(workspaces, func(i, j int) bool {
		return len(workspaces[i].Path) > len(workspaces[j].Path)
	})
	for _, workspace := range workspaces {
		if !pathWithinRoot(cleanPath, workspace.Path) {
			continue
		}
		state := states[workspace.ID]
		if state.Root == "" {
			realRoot, err := filepath.EvalSymlinks(workspace.Path)
			if err != nil {
				continue
			}
			state = workspaceRuntimeState{
				Root:     filepath.Clean(workspace.Path),
				RealRoot: filepath.Clean(realRoot),
			}
		}
		if !checkReal {
			return workspace, state, true
		}
		realPath, err := filepath.EvalSymlinks(cleanPath)
		if err != nil {
			return WorkspaceEntry{}, workspaceRuntimeState{}, false
		}
		realRel, err := filepath.Rel(state.RealRoot, filepath.Clean(realPath))
		if err != nil {
			return WorkspaceEntry{}, workspaceRuntimeState{}, false
		}
		if realRel == "." || (!strings.HasPrefix(realRel, "..") && !filepath.IsAbs(realRel)) {
			return workspace, state, true
		}
	}
	return WorkspaceEntry{}, workspaceRuntimeState{}, false
}

func (a *App) setActiveWorkspaceForPath(path string) {
	workspace, state, ok := a.workspaceForPath(path, false)
	if !ok {
		return
	}
	a.mu.Lock()
	if a.workspaceStates == nil {
		a.workspaceStates = map[string]workspaceRuntimeState{}
	}
	if existing, exists := a.workspaceStates[workspace.ID]; !exists || existing.Root == "" {
		a.workspaceStates[workspace.ID] = state
	}
	a.settings.ActiveWorkspaceID = workspace.ID
	a.settings.LastWorkspace = workspace.Path
	a.applyActiveWorkspaceLocked()
	settingsPath := a.settingsPath
	settings := a.settings
	a.mu.Unlock()
	if settingsPath != "" {
		_ = saveSettings(settingsPath, settings)
	}
}

func (a *App) workspaceCollection(workspaces []WorkspaceEntry, activeID string) (*WorkspaceCollection, error) {
	a.workspaceScanMu.Lock()
	defer a.workspaceScanMu.Unlock()

	a.mu.RLock()
	previous := cloneWorkspaceStates(a.workspaceStates)
	a.mu.RUnlock()
	states := map[string]workspaceRuntimeState{}
	for _, workspace := range workspaces {
		entry, _, state, err := scanWorkspaceRoot(workspace.Path)
		if err != nil {
			continue
		}
		entry.ID = workspace.ID
		entry.Name = firstNonEmpty(workspace.Name, workspaceDisplayName(entry.Path))
		entry.Order = workspace.Order
		entry.AddedAt = workspace.AddedAt
		entry.LastOpenedAt = workspace.LastOpenedAt
		if existing, ok := previous[entry.ID]; ok && samePath(existing.Root, entry.Path) {
			state.Tree = existing.Tree
			state.TreeSignature = existing.TreeSignature
		}
		states[entry.ID] = state
	}
	if len(states) == 0 {
		return &WorkspaceCollection{}, nil
	}
	if activeID == "" {
		for _, workspace := range workspaces {
			if _, ok := states[workspace.ID]; ok {
				activeID = workspace.ID
				break
			}
		}
	}
	a.mu.Lock()
	if a.workspaceStates == nil {
		a.workspaceStates = map[string]workspaceRuntimeState{}
	}
	for id, state := range states {
		a.workspaceStates[id] = state
	}
	a.settings.ActiveWorkspaceID = activeID
	a.applyActiveWorkspaceLocked()
	a.mu.Unlock()
	return workspaceCollectionFromStates(workspaces, activeID, states), nil
}

func (a *App) revealableWorkspacePath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("empty workspace path")
	}
	var (
		target string
		err    error
	)
	if filepath.IsAbs(path) {
		target, err = filepath.Abs(path)
	} else {
		target, err = a.workspacePath(path)
	}
	if err != nil {
		return "", err
	}
	target = filepath.Clean(target)
	if !a.isLexicallyWithinWorkspace(target) {
		return "", errors.New("path is outside the current workspace")
	}
	if _, err := os.Stat(target); err != nil {
		return "", err
	}
	if !a.isWithinWorkspace(target) {
		return "", errors.New("path is outside the current workspace")
	}
	return target, nil
}

func (a *App) renameableWorkspacePath(path string) (string, error) {
	target, err := a.revealableWorkspacePath(path)
	if err != nil {
		return "", err
	}
	_, state, ok := a.workspaceForPath(target, false)
	if !ok || filepath.Clean(target) == filepath.Clean(state.Root) {
		return "", errors.New("workspace root cannot be renamed")
	}
	a.setActiveWorkspaceForPath(target)
	return target, nil
}

func (a *App) isRenameTargetWithinRealWorkspace(sourceParent string, target string) bool {
	_, state, ok := a.workspaceForPath(sourceParent, true)
	if !ok || state.RealRoot == "" {
		return false
	}
	realParent, err := filepath.EvalSymlinks(sourceParent)
	if err != nil {
		return false
	}
	realTarget := filepath.Join(realParent, filepath.Base(target))
	realRel, err := filepath.Rel(state.RealRoot, filepath.Clean(realTarget))
	if err != nil {
		return false
	}
	return realRel == "." || (!strings.HasPrefix(realRel, "..") && !filepath.IsAbs(realRel))
}

func (a *App) refreshWorkspaceTreeAfterMutation(oldPath string, newPath string, nodeType string) (TreeNode, error) {
	workspace, state, ok := a.workspaceForPath(newPath, false)
	if !ok || state.Root == "" {
		return TreeNode{}, errors.New("workspace is not open")
	}
	node := remapTreePaths(state.Tree, oldPath, newPath, nodeType)
	node, err := refreshLoadedTree(node)
	if err != nil {
		return TreeNode{}, err
	}
	state.Tree = node
	state.TreeSignature = treeSignature(node)
	a.mu.Lock()
	a.workspaceStates[workspace.ID] = state
	a.applyActiveWorkspaceLocked()
	a.mu.Unlock()
	return node, nil
}

func scanWorkspaceRoot(root string) (WorkspaceEntry, TreeNode, workspaceRuntimeState, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return WorkspaceEntry{}, TreeNode{}, workspaceRuntimeState{}, err
	}
	abs = filepath.Clean(abs)
	info, err := os.Stat(abs)
	if err != nil {
		return WorkspaceEntry{}, TreeNode{}, workspaceRuntimeState{}, err
	}
	if !info.IsDir() {
		return WorkspaceEntry{}, TreeNode{}, workspaceRuntimeState{}, errors.New("workspace path is not a directory")
	}
	node := TreeNode{Name: info.Name(), Path: abs, Type: "directory"}
	realRoot, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return WorkspaceEntry{}, TreeNode{}, workspaceRuntimeState{}, err
	}
	entry := WorkspaceEntry{
		ID:   workspaceID(abs),
		Name: workspaceDisplayName(abs),
		Path: abs,
	}
	state := workspaceRuntimeState{
		Root:          abs,
		RealRoot:      filepath.Clean(realRoot),
		TreeSignature: treeSignature(node),
		Tree:          node,
	}
	return entry, node, state, nil
}

func (a *App) scanExistingWorkspaceLocked(workspace WorkspaceEntry) (TreeNode, workspaceRuntimeState, error) {
	if state, ok := a.workspaceStates[workspace.ID]; ok && state.Root != "" {
		node, err := refreshLoadedTree(state.Tree)
		if err != nil {
			return TreeNode{}, workspaceRuntimeState{}, err
		}
		state.Tree = node
		state.TreeSignature = treeSignature(node)
		a.workspaceStates[workspace.ID] = state
		return cloneTreeNode(node), state, nil
	}
	_, node, state, err := scanWorkspaceRoot(workspace.Path)
	if err != nil {
		return TreeNode{}, workspaceRuntimeState{}, err
	}
	a.workspaceStates[workspace.ID] = state
	return node, state, nil
}

func (a *App) normalizeSettingsLocked() {
	a.settings.StorageVersion = currentSettingsVersion
	a.settings.Locale = normalizeLocale(a.settings.Locale)
	a.settings.Theme = normalizeTheme(a.settings.Theme)
	if a.settings.Workspaces == nil {
		a.settings.Workspaces = []WorkspaceEntry{}
	}
	if strings.TrimSpace(a.settings.LastWorkspace) != "" && len(a.settings.Workspaces) == 0 {
		path := filepath.Clean(a.settings.LastWorkspace)
		a.settings.Workspaces = append(a.settings.Workspaces, WorkspaceEntry{
			ID:           workspaceID(path),
			Name:         workspaceDisplayName(path),
			Path:         path,
			AddedAt:      time.Now().UnixMilli(),
			LastOpenedAt: time.Now().UnixMilli(),
		})
	}
	seen := map[string]struct{}{}
	next := make([]WorkspaceEntry, 0, len(a.settings.Workspaces))
	for _, workspace := range a.settings.Workspaces {
		path := strings.TrimSpace(workspace.Path)
		if path == "" {
			continue
		}
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
		path = filepath.Clean(path)
		key := strings.ToLower(path)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		workspace.Path = path
		if workspace.ID == "" {
			workspace.ID = workspaceID(path)
		}
		if workspace.Name == "" {
			workspace.Name = workspaceDisplayName(path)
		}
		if workspace.AddedAt == 0 {
			workspace.AddedAt = time.Now().UnixMilli()
		}
		if workspace.LastOpenedAt == 0 {
			workspace.LastOpenedAt = workspace.AddedAt
		}
		workspace.Order = len(next)
		next = append(next, workspace)
	}
	a.settings.Workspaces = next
	if len(next) == 0 {
		a.settings.ActiveWorkspaceID = ""
		a.settings.LastWorkspace = ""
		return
	}
	activeExists := false
	for _, workspace := range next {
		if workspace.ID == a.settings.ActiveWorkspaceID {
			activeExists = true
			a.settings.LastWorkspace = workspace.Path
			break
		}
	}
	if !activeExists {
		a.settings.ActiveWorkspaceID = next[0].ID
		a.settings.LastWorkspace = next[0].Path
	}
}

func (a *App) applyActiveWorkspaceLocked() {
	a.normalizeSettingsLocked()
	a.workspaceRoot = ""
	a.workspaceRootReal = ""
	a.workspaceTreeSig = ""
	for _, workspace := range a.settings.Workspaces {
		if workspace.ID != a.settings.ActiveWorkspaceID {
			continue
		}
		state := a.workspaceStates[workspace.ID]
		a.workspaceRoot = workspace.Path
		a.workspaceRootReal = state.RealRoot
		a.workspaceTreeSig = state.TreeSignature
		if a.workspaceRootReal == "" {
			a.workspaceRootReal = workspace.Path
		}
		return
	}
}

func workspaceID(path string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(filepath.Clean(path))))
	return "ws_" + fmt.Sprintf("%x", sum[:8])
}

func workspaceDisplayName(path string) string {
	name := filepath.Base(filepath.Clean(path))
	if name == "." || name == string(filepath.Separator) || name == "" {
		return filepath.Clean(path)
	}
	return name
}

func pathWithinRoot(path string, root string) bool {
	if strings.TrimSpace(path) == "" || strings.TrimSpace(root) == "" {
		return false
	}
	rel, err := filepath.Rel(filepath.Clean(root), filepath.Clean(path))
	if err != nil {
		return false
	}
	return rel == "." || (!strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel))
}

func (a *App) queueLaunchArgs(args []string, cwd string) error {
	request, err := a.launchRequestFromArgs(args, cwd)
	if err != nil || request == nil {
		return err
	}
	a.mu.Lock()
	a.pendingLaunch = request
	a.mu.Unlock()
	if a.ctx != nil {
		runtime.WindowUnminimise(a.ctx)
		runtime.WindowShow(a.ctx)
		runtime.EventsEmit(a.ctx, "kern:launch-request")
	}
	return nil
}

func (a *App) ConsumeLaunchRequest() (*LaunchRequest, error) {
	a.mu.Lock()
	request := a.pendingLaunch
	a.pendingLaunch = nil
	a.mu.Unlock()
	if request == nil {
		return &LaunchRequest{}, nil
	}
	return request, nil
}

func (a *App) launchRequestFromArgs(args []string, cwd string) (*LaunchRequest, error) {
	mode, target := launchTargetFromArgs(args)
	if target == "" {
		return nil, nil
	}
	target = resolveLaunchPath(target, cwd)
	info, err := os.Stat(target)
	if err != nil {
		return nil, err
	}
	if mode == "workspace" || info.IsDir() {
		collection, err := a.AddWorkspace(target)
		if err != nil {
			return nil, err
		}
		return &LaunchRequest{Collection: collection}, nil
	}
	if !isMarkdownFile(target) {
		return nil, errors.New("not a markdown document")
	}
	if !a.isWithinWorkspace(target) {
		if _, err := a.addWorkspacePath(filepath.Dir(target), true); err != nil {
			return nil, err
		}
	} else {
		a.setActiveWorkspaceForPath(target)
	}
	collection, err := a.RestoreWorkspaces()
	if err != nil {
		return nil, err
	}
	return &LaunchRequest{
		Collection:   collection,
		DocumentPath: filepath.Clean(target),
	}, nil
}

func launchTargetFromArgs(args []string) (string, string) {
	for index := 0; index < len(args); index++ {
		arg := strings.TrimSpace(args[index])
		switch arg {
		case "--open-file":
			if index+1 < len(args) {
				return "file", args[index+1]
			}
		case "--add-workspace":
			if index+1 < len(args) {
				return "workspace", args[index+1]
			}
		default:
			if !strings.HasPrefix(arg, "-") {
				return "auto", arg
			}
		}
	}
	return "", ""
}

func resolveLaunchPath(path string, cwd string) string {
	path = strings.Trim(path, `"`)
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	if cwd == "" {
		cwd = mustGetwd()
	}
	return filepath.Clean(filepath.Join(cwd, path))
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return wd
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

func normalizeTheme(theme string) string {
	switch theme {
	case "light", "dark":
		return theme
	default:
		return "system"
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

func checkGitHubUpdates(ctx context.Context, apiURL string, currentVersion string) (*UpdateInfo, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	requestCtx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", productInfo.AppSlug+"/"+productInfo.Version)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("update check failed: %s", resp.Status)
	}
	var releases []githubRelease
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&releases); err != nil {
		return nil, err
	}
	info := UpdateInfo{CurrentVersion: currentVersion}
	current := parseVersion(currentVersion)
	for _, release := range releases {
		if release.Draft {
			continue
		}
		latest := normalizeVersion(release.TagName)
		if latest == "" || !versionGreater(parseVersion(latest), current) {
			continue
		}
		assetURL, sha := releaseInstallerAsset(release)
		if assetURL == "" {
			continue
		}
		info.LatestVersion = latest
		info.UpdateAvailable = true
		info.ReleaseURL = release.HTMLURL
		info.DownloadURL = assetURL
		info.Sha256 = sha
		info.ReleaseNotes = strings.TrimSpace(release.Body)
		return &info, nil
	}
	return &info, nil
}

type githubRelease struct {
	TagName string        `json:"tag_name"`
	HTMLURL string        `json:"html_url"`
	Body    string        `json:"body"`
	Draft   bool          `json:"draft"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Digest             string `json:"digest"`
}

func releaseInstallerAsset(release githubRelease) (string, string) {
	latest := normalizeVersion(release.TagName)
	expected := "JSKernMD-Setup-" + latest + "-x64.exe"
	var sha string
	for _, asset := range release.Assets {
		if asset.Name == "SHA256SUMS.txt" {
			continue
		}
		if asset.Name == expected {
			sha = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(asset.Digest)), "sha256:")
			return asset.BrowserDownloadURL, sha
		}
	}
	return "", ""
}

func downloadFile(ctx context.Context, sourceURL string, target string, sha256Hex string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	requestCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", productInfo.AppSlug+"/"+productInfo.Version)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("update download failed: %s", resp.Status)
	}
	temp := target + ".part"
	out, err := os.Create(temp)
	if err != nil {
		return err
	}
	hash := sha256.New()
	_, copyErr := io.Copy(io.MultiWriter(out, hash), resp.Body)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(temp)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(temp)
		return closeErr
	}
	actual := fmt.Sprintf("%x", hash.Sum(nil))
	expected := strings.ToLower(strings.TrimSpace(sha256Hex))
	if expected != "" && actual != expected {
		_ = os.Remove(temp)
		return fmt.Errorf("update checksum mismatch")
	}
	_ = os.Remove(target)
	return os.Rename(temp, target)
}

func isAllowedUpdateDownloadURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return parsed.Scheme == "https" &&
		strings.EqualFold(parsed.Host, "github.com") &&
		strings.HasPrefix(parsed.Path, "/xiaotianwm/jskern.md/releases/download/") &&
		strings.HasSuffix(strings.ToLower(parsed.Path), ".exe")
}

func openFile(path string) error {
	switch os.PathSeparator {
	case '\\':
		return exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", path).Start()
	default:
		if _, err := exec.LookPath("open"); err == nil {
			return exec.Command("open", path).Start()
		}
		return exec.Command("xdg-open", path).Start()
	}
}

func revealPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	switch os.PathSeparator {
	case '\\':
		if info.IsDir() {
			return exec.Command("explorer.exe", path).Start()
		}
		return exec.Command("explorer.exe", "/select,"+path).Start()
	default:
		if _, err := exec.LookPath("open"); err == nil {
			return exec.Command("open", "-R", path).Start()
		}
		target := path
		if !info.IsDir() {
			target = filepath.Dir(path)
		}
		return exec.Command("xdg-open", target).Start()
	}
}

func normalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "v")
	version = strings.TrimPrefix(version, "V")
	if parseVersion(version) == nil {
		return ""
	}
	return version
}

func parseVersion(version string) []int {
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	if version == "" {
		return nil
	}
	parts := strings.Split(version, ".")
	if len(parts) == 0 {
		return nil
	}
	result := make([]int, len(parts))
	for index, part := range parts {
		if part == "" {
			return nil
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return nil
			}
		}
		value, err := strconv.Atoi(part)
		if err != nil {
			return nil
		}
		result[index] = value
	}
	return result
}

func versionGreater(left []int, right []int) bool {
	maxParts := len(left)
	if len(right) > maxParts {
		maxParts = len(right)
	}
	for index := 0; index < maxParts; index++ {
		var l, r int
		if index < len(left) {
			l = left[index]
		}
		if index < len(right) {
			r = right[index]
		}
		if l != r {
			return l > r
		}
	}
	return false
}

func (a *App) initAppData(root string) error {
	if root == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return err
		}
		root = filepath.Join(configDir, productInfo.AppSlug)
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
	readingMemoryPath := filepath.Join(root, "data", "reading-memory.json")
	readingMemory, err := loadReadingMemory(readingMemoryPath)
	if err != nil {
		return err
	}
	if err := saveReadingMemory(readingMemoryPath, readingMemory); err != nil {
		return err
	}
	a.mu.Lock()
	a.appDataRoot = root
	a.settingsPath = settingsPath
	a.readingMemoryPath = readingMemoryPath
	a.settings = settings
	a.readingMemory = readingMemory
	if a.workspaceStates == nil {
		a.workspaceStates = map[string]workspaceRuntimeState{}
	}
	a.applyActiveWorkspaceLocked()
	a.mu.Unlock()
	return nil
}

func (a *App) setLastWorkspace(path string) {
	a.updateSettings(func(settings *Settings) {
		settings.LastWorkspace = filepath.Clean(path)
	})
}

func (a *App) updateSettings(mutator func(*Settings)) {
	a.mu.Lock()
	if a.settings.StorageVersion == 0 {
		a.settings = defaultSettings()
	}
	mutator(&a.settings)
	a.normalizeSettingsLocked()
	settingsPath := a.settingsPath
	settings := a.settings
	a.mu.Unlock()
	if settingsPath == "" {
		return
	}
	_ = saveSettings(settingsPath, settings)
}

func defaultSettings() Settings {
	return Settings{
		StorageVersion: currentSettingsVersion,
		Locale:         "zh-CN",
		Theme:          "system",
		Workspaces:     []WorkspaceEntry{},
	}
}

func normalizeSettings(settings *Settings) {
	app := &App{settings: *settings}
	app.normalizeSettingsLocked()
	*settings = app.settings
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
	normalizeSettings(&settings)
	return settings, nil
}

func saveSettings(path string, settings Settings) error {
	normalizeSettings(&settings)
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

func scanDirectoryLevel(path string) (TreeNode, error) {
	info, err := os.Stat(path)
	if err != nil {
		return TreeNode{}, err
	}
	if !info.IsDir() {
		return TreeNode{}, errors.New("workspace path is not a directory")
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
		childType := "file"
		if entry.IsDir() {
			childType = "directory"
		}
		children = append(children, TreeNode{Name: name, Path: childPath, Type: childType})
	}

	sort.Slice(children, func(i, j int) bool {
		if children[i].Type != children[j].Type {
			return children[i].Type == "directory"
		}
		return strings.ToLower(children[i].Name) < strings.ToLower(children[j].Name)
	})
	return TreeNode{
		Name:     info.Name(),
		Path:     filepath.Clean(path),
		Type:     "directory",
		Loaded:   true,
		Children: children,
	}, nil
}

func loadTreeDirectory(root TreeNode, target string) (TreeNode, TreeNode, error) {
	target = filepath.Clean(target)
	if samePath(root.Path, target) {
		loaded, err := scanDirectoryLevel(target)
		if err != nil {
			return TreeNode{}, TreeNode{}, err
		}
		loaded = preserveLoadedChildren(loaded, root)
		return loaded, cloneTreeNode(loaded), nil
	}
	if root.Type != "directory" || !root.Loaded || !pathWithinRoot(target, root.Path) {
		return TreeNode{}, TreeNode{}, errors.New("directory is not present in the loaded workspace tree")
	}

	next := cloneTreeNode(root)
	for index, child := range next.Children {
		if child.Type != "directory" || !pathWithinRoot(target, child.Path) {
			continue
		}
		updated, loaded, err := loadTreeDirectory(child, target)
		if err != nil {
			return TreeNode{}, TreeNode{}, err
		}
		next.Children[index] = updated
		return next, loaded, nil
	}
	return TreeNode{}, TreeNode{}, errors.New("directory is not present in the loaded workspace tree")
}

func refreshLoadedTree(root TreeNode) (TreeNode, error) {
	if root.Type != "directory" || !root.Loaded {
		return cloneTreeNode(root), nil
	}
	fresh, err := scanDirectoryLevel(root.Path)
	if err != nil {
		return TreeNode{}, err
	}
	fresh = preserveLoadedChildren(fresh, root)
	for index, child := range fresh.Children {
		if child.Type != "directory" || !child.Loaded {
			continue
		}
		refreshed, err := refreshLoadedTree(child)
		if err != nil {
			continue
		}
		fresh.Children[index] = refreshed
	}
	return fresh, nil
}

func preserveLoadedChildren(fresh TreeNode, current TreeNode) TreeNode {
	loaded := map[string]TreeNode{}
	for _, child := range current.Children {
		if child.Type == "directory" && child.Loaded {
			loaded[pathKey(child.Path)] = child
		}
	}
	for index, child := range fresh.Children {
		if existing, ok := loaded[pathKey(child.Path)]; ok {
			fresh.Children[index] = cloneTreeNode(existing)
		}
	}
	return fresh
}

func remapTreePaths(root TreeNode, oldPath string, newPath string, nodeType string) TreeNode {
	next := cloneTreeNode(root)
	if mapped, ok := remapTreePath(next.Path, oldPath, newPath, nodeType); ok {
		next.Path = mapped
		next.Name = filepath.Base(mapped)
	}
	for index, child := range next.Children {
		next.Children[index] = remapTreePaths(child, oldPath, newPath, nodeType)
	}
	return next
}

func remapTreePath(path string, oldPath string, newPath string, nodeType string) (string, bool) {
	if samePath(path, oldPath) {
		return filepath.Clean(newPath), true
	}
	if nodeType != "directory" || !pathWithinRoot(path, oldPath) {
		return path, false
	}
	relative, err := filepath.Rel(filepath.Clean(oldPath), filepath.Clean(path))
	if err != nil || relative == "." {
		return path, false
	}
	return filepath.Join(filepath.Clean(newPath), relative), true

}

func workspaceCollectionFromStates(workspaces []WorkspaceEntry, activeID string, states map[string]workspaceRuntimeState) *WorkspaceCollection {
	trees := make([]WorkspaceTree, 0, len(workspaces))
	for _, workspace := range workspaces {
		state, ok := states[workspace.ID]
		if !ok || state.Root == "" {
			continue
		}
		workspace.Path = state.Root
		workspace.Name = firstNonEmpty(workspace.Name, workspaceDisplayName(state.Root))
		trees = append(trees, WorkspaceTree{Workspace: workspace, Root: cloneTreeNode(state.Tree)})
	}
	return &WorkspaceCollection{Workspaces: trees, ActiveWorkspaceID: activeID}
}

func cloneWorkspaceStates(states map[string]workspaceRuntimeState) map[string]workspaceRuntimeState {
	cloned := make(map[string]workspaceRuntimeState, len(states))
	for id, state := range states {
		state.Tree = cloneTreeNode(state.Tree)
		cloned[id] = state
	}
	return cloned
}

func cloneTreeNode(node TreeNode) TreeNode {
	node.Children = append([]TreeNode(nil), node.Children...)
	for index := range node.Children {
		node.Children[index] = cloneTreeNode(node.Children[index])
	}
	return node
}

func samePath(first string, second string) bool {
	return strings.EqualFold(filepath.Clean(first), filepath.Clean(second))
}

func pathKey(path string) string {
	return strings.ToLower(filepath.Clean(path))
}

func treeSignature(root TreeNode) string {
	var builder strings.Builder
	writeTreeSignature(&builder, root)
	return builder.String()
}

func writeTreeSignature(builder *strings.Builder, node TreeNode) {
	builder.WriteString(node.Type)
	builder.WriteByte('\t')
	builder.WriteString(filepath.Clean(node.Path))
	builder.WriteByte('\n')
	for _, child := range node.Children {
		writeTreeSignature(builder, child)
	}
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

func validRenameName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("empty rename target")
	}
	if name == "." || name == ".." {
		return "", errors.New("invalid rename target")
	}
	if strings.ContainsAny(name, `/\`) || strings.ContainsRune(name, 0) || filepath.VolumeName(name) != "" {
		return "", errors.New("rename target must be a file or folder name")
	}
	if strings.ContainsAny(name, `<>:"|?*`) || strings.HasSuffix(name, ".") || strings.HasSuffix(name, " ") {
		return "", errors.New("rename target contains unsupported characters")
	}
	if filepath.Base(name) != name {
		return "", errors.New("rename target must be a file or folder name")
	}
	return name, nil
}

func treeNodeType(info os.FileInfo) string {
	if info.IsDir() {
		return "directory"
	}
	return "file"
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
