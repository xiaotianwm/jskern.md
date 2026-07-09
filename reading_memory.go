package main

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	currentReadingMemoryVersion = 1
	maxReadingMemoryDocuments   = 300
	maxOpenReadingTabs          = 24
)

type ReadingMemoryStore struct {
	StorageVersion int                             `json:"storage_version"`
	Workspaces     map[string]*WorkspaceReadingLog `json:"workspaces"`
}

type WorkspaceReadingLog struct {
	Root           string                          `json:"root"`
	LastDocument   string                          `json:"last_document"`
	ActiveDocument string                          `json:"active_document"`
	OpenTabs       []string                        `json:"open_tabs"`
	Documents      map[string]DocumentReadingState `json:"documents"`
}

type DocumentReadingState struct {
	RelativePath string  `json:"relative_path"`
	ScrollTop    int     `json:"scroll_top"`
	ScrollRatio  float64 `json:"scroll_ratio"`
	HeadingID    string  `json:"heading_id"`
	ModifiedAt   int64   `json:"modified_at"`
	Size         int64   `json:"size"`
	UpdatedAt    int64   `json:"updated_at"`
}

type ReadingMemorySnapshot struct {
	LastDocument string           `json:"lastDocument"`
	LastPosition *ReadingPosition `json:"lastPosition,omitempty"`
}

type WorkspaceReadingSession struct {
	OpenTabs       []ReadingTab     `json:"openTabs"`
	ActiveDocument string           `json:"activeDocument"`
	ActivePosition *ReadingPosition `json:"activePosition,omitempty"`
}

type ReadingTab struct {
	Path         string `json:"path"`
	RelativePath string `json:"relativePath"`
	Name         string `json:"name"`
}

type ReadingPosition struct {
	Path         string  `json:"path"`
	RelativePath string  `json:"relativePath"`
	ScrollTop    int     `json:"scrollTop"`
	ScrollRatio  float64 `json:"scrollRatio"`
	HeadingID    string  `json:"headingId"`
	ModifiedAt   int64   `json:"modifiedAt"`
	Size         int64   `json:"size"`
	UpdatedAt    int64   `json:"updatedAt"`
}

func defaultReadingMemory() ReadingMemoryStore {
	return ReadingMemoryStore{
		StorageVersion: currentReadingMemoryVersion,
		Workspaces:     map[string]*WorkspaceReadingLog{},
	}
}

func (a *App) GetReadingMemory() (*ReadingMemorySnapshot, error) {
	a.mu.RLock()
	root := a.workspaceRoot
	var relativePath string
	var state DocumentReadingState
	var found bool
	if root != "" {
		if workspace := a.readingMemory.Workspaces[workspaceMemoryKey(root)]; workspace != nil {
			relativePath = firstNonEmpty(workspace.ActiveDocument, workspace.LastDocument)
			if relativePath != "" {
				state, found = workspace.Documents[relativePath]
			}
		}
	}
	a.mu.RUnlock()
	if root == "" {
		return &ReadingMemorySnapshot{}, nil
	}
	if relativePath == "" || !found {
		return &ReadingMemorySnapshot{}, nil
	}
	position, err := a.readingPositionFromState(root, relativePath, state)
	if err != nil {
		return &ReadingMemorySnapshot{}, nil
	}
	return &ReadingMemorySnapshot{
		LastDocument: position.Path,
		LastPosition: position,
	}, nil
}

func (a *App) GetReadingSession() (*WorkspaceReadingSession, error) {
	a.mu.RLock()
	workspaces := append([]WorkspaceEntry(nil), a.settings.Workspaces...)
	activeWorkspaceID := a.settings.ActiveWorkspaceID
	memory := a.readingMemory
	a.mu.RUnlock()
	if len(workspaces) == 0 {
		return &WorkspaceReadingSession{}, nil
	}

	tabs := []ReadingTab{}
	activePath := ""
	for _, workspaceEntry := range workspaces {
		root := workspaceEntry.Path
		workspace := memory.Workspaces[workspaceMemoryKey(root)]
		if workspace == nil {
			continue
		}
		openTabs := append([]string(nil), workspace.OpenTabs...)
		activeDocument := firstNonEmpty(workspace.ActiveDocument, workspace.LastDocument)
		if len(openTabs) == 0 && activeDocument != "" {
			openTabs = []string{activeDocument}
		}
		for _, relativePath := range openTabs {
			tab, err := a.readingTabFromRelativePath(root, relativePath)
			if err != nil {
				continue
			}
			tabs = append(tabs, tab)
			if workspaceEntry.ID == activeWorkspaceID && relativePath == activeDocument {
				activePath = tab.Path
			}
		}
	}
	if len(tabs) == 0 {
		return &WorkspaceReadingSession{}, nil
	}
	if activePath == "" {
		activePath = tabs[0].Path
	}
	position, _ := a.GetReadingPosition(activePath)
	return &WorkspaceReadingSession{
		OpenTabs:       tabs,
		ActiveDocument: activePath,
		ActivePosition: position,
	}, nil
}

func (a *App) GetReadingPosition(path string) (*ReadingPosition, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if !isMarkdownFile(abs) || !a.isWithinWorkspace(abs) {
		return nil, nil
	}
	_, workspaceState, ok := a.workspaceForPath(abs, true)
	if !ok {
		return nil, nil
	}
	root := workspaceState.Root
	relativePath, ok := workspaceRelativePathFromRoot(root, abs)
	if !ok {
		return nil, nil
	}

	a.mu.RLock()
	var state DocumentReadingState
	var found bool
	if workspace := a.readingMemory.Workspaces[workspaceMemoryKey(root)]; workspace != nil {
		if len(workspace.OpenTabs) > 0 && !containsString(workspace.OpenTabs, relativePath) {
			a.mu.RUnlock()
			return nil, nil
		}
		state, found = workspace.Documents[relativePath]
	}
	a.mu.RUnlock()
	if !found {
		return nil, nil
	}
	return a.readingPositionFromState(root, relativePath, state)
}

func (a *App) SaveOpenTabs(paths []string, activePath string) error {
	type groupedTabs struct {
		root string
		tabs []string
	}
	a.mu.RLock()
	configuredRoots := make([]string, 0, len(a.settings.Workspaces))
	for _, workspace := range a.settings.Workspaces {
		configuredRoots = append(configuredRoots, workspace.Path)
	}
	activeRootSnapshot := a.workspaceRoot
	a.mu.RUnlock()

	groups := map[string]*groupedTabs{}
	order := []string{}
	for _, path := range paths {
		root, relativePath, err := a.readingTabRelativePath(path)
		if err != nil {
			return err
		}
		group := groups[root]
		if group == nil {
			group = &groupedTabs{root: root}
			groups[root] = group
			order = append(order, root)
		}
		if containsString(group.tabs, relativePath) {
			continue
		}
		if len(group.tabs) < maxOpenReadingTabs {
			group.tabs = append(group.tabs, relativePath)
		}
	}

	activeRoot := ""
	activeDocument := ""
	if strings.TrimSpace(activePath) != "" {
		root, relativePath, err := a.readingTabRelativePath(activePath)
		if err != nil {
			return err
		}
		activeRoot = root
		activeDocument = relativePath
		group := groups[root]
		if group == nil {
			group = &groupedTabs{root: root}
			groups[root] = group
			order = append(order, root)
		}
		if !containsString(group.tabs, relativePath) {
			if len(group.tabs) >= maxOpenReadingTabs {
				group.tabs[len(group.tabs)-1] = relativePath
			} else {
				group.tabs = append(group.tabs, relativePath)
			}
		}
	}
	if len(order) == 0 {
		if activeRootSnapshot == "" {
			return nil
		}
		groups[activeRootSnapshot] = &groupedTabs{root: activeRootSnapshot}
		order = append(order, activeRootSnapshot)
	}
	for _, root := range configuredRoots {
		if groups[root] != nil {
			continue
		}
		groups[root] = &groupedTabs{root: root}
		order = append(order, root)
	}

	a.mu.Lock()
	if a.readingMemory.StorageVersion == 0 {
		a.readingMemory = defaultReadingMemory()
	}
	if a.readingMemory.Workspaces == nil {
		a.readingMemory.Workspaces = map[string]*WorkspaceReadingLog{}
	}
	for _, root := range order {
		group := groups[root]
		key := workspaceMemoryKey(root)
		workspace := a.readingMemory.Workspaces[key]
		if workspace == nil {
			workspace = &WorkspaceReadingLog{
				Root:      filepath.Clean(root),
				Documents: map[string]DocumentReadingState{},
			}
			a.readingMemory.Workspaces[key] = workspace
		}
		if workspace.Documents == nil {
			workspace.Documents = map[string]DocumentReadingState{}
		}
		workspace.Root = filepath.Clean(root)
		workspace.OpenTabs = group.tabs
		if root == activeRoot {
			workspace.ActiveDocument = activeDocument
			workspace.LastDocument = activeDocument
		} else {
			workspace.ActiveDocument = ""
			if len(group.tabs) > 0 {
				workspace.ActiveDocument = group.tabs[0]
				workspace.LastDocument = workspace.ActiveDocument
			}
		}
		pruneClosedReadingDocuments(workspace)
		pruneReadingDocuments(workspace)
	}
	memoryPath := a.readingMemoryPath
	if memoryPath == "" {
		a.mu.Unlock()
		return nil
	}
	err := saveReadingMemory(memoryPath, a.readingMemory)
	a.mu.Unlock()
	return err
}

func pruneClosedReadingDocuments(workspace *WorkspaceReadingLog) {
	keep := map[string]struct{}{}
	for _, path := range workspace.OpenTabs {
		keep[path] = struct{}{}
	}
	for path := range workspace.Documents {
		if _, ok := keep[path]; !ok {
			delete(workspace.Documents, path)
		}
	}
}

func (a *App) clearReadingSessionForRootLocked(root string) {
	if a.readingMemory.Workspaces == nil {
		return
	}
	workspace := a.readingMemory.Workspaces[workspaceMemoryKey(root)]
	if workspace == nil {
		return
	}
	workspace.OpenTabs = nil
	workspace.ActiveDocument = ""
	workspace.LastDocument = ""
}

func (a *App) SaveReadingPosition(path string, scrollTop int, scrollRatio float64, headingID string, modifiedAt int64, size int64) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if !isMarkdownFile(abs) {
		return errors.New("not a markdown document")
	}
	if !a.isWithinWorkspace(abs) {
		return errors.New("document is outside the current workspace")
	}
	_, state, ok := a.workspaceForPath(abs, true)
	if !ok {
		return errors.New("document is outside the current workspace")
	}
	root := state.Root
	relativePath, ok := workspaceRelativePathFromRoot(root, abs)
	if !ok {
		return errors.New("document is outside the current workspace")
	}
	if scrollTop < 0 {
		scrollTop = 0
	}
	if math.IsNaN(scrollRatio) || math.IsInf(scrollRatio, 0) || scrollRatio < 0 {
		scrollRatio = 0
	}
	if scrollRatio > 1 {
		scrollRatio = 1
	}
	headingID = strings.TrimSpace(headingID)
	if len(headingID) > 512 {
		headingID = headingID[:512]
	}

	a.mu.Lock()
	if a.readingMemory.StorageVersion == 0 {
		a.readingMemory = defaultReadingMemory()
	}
	if a.readingMemory.Workspaces == nil {
		a.readingMemory.Workspaces = map[string]*WorkspaceReadingLog{}
	}
	key := workspaceMemoryKey(root)
	workspace := a.readingMemory.Workspaces[key]
	if workspace == nil {
		workspace = &WorkspaceReadingLog{
			Root:      filepath.Clean(root),
			Documents: map[string]DocumentReadingState{},
		}
		a.readingMemory.Workspaces[key] = workspace
	}
	if workspace.Documents == nil {
		workspace.Documents = map[string]DocumentReadingState{}
	}
	workspace.Root = filepath.Clean(root)
	workspace.LastDocument = relativePath
	workspace.ActiveDocument = relativePath
	if len(workspace.OpenTabs) == 0 {
		workspace.OpenTabs = []string{relativePath}
	}
	workspace.Documents[relativePath] = DocumentReadingState{
		RelativePath: relativePath,
		ScrollTop:    scrollTop,
		ScrollRatio:  scrollRatio,
		HeadingID:    headingID,
		ModifiedAt:   modifiedAt,
		Size:         size,
		UpdatedAt:    time.Now().UnixMilli(),
	}
	pruneReadingDocuments(workspace)
	memoryPath := a.readingMemoryPath
	if memoryPath == "" {
		a.mu.Unlock()
		return nil
	}
	err = saveReadingMemory(memoryPath, a.readingMemory)
	a.mu.Unlock()
	return err
}

func (a *App) readingTabRelativePath(path string) (string, string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", "", err
	}
	if !isMarkdownFile(abs) {
		return "", "", errors.New("not a markdown document")
	}
	_, state, ok := a.workspaceForPath(abs, true)
	if !ok {
		return "", "", errors.New("document is outside the current workspace")
	}
	relativePath, ok := workspaceRelativePathFromRoot(state.Root, abs)
	if !ok {
		return "", "", errors.New("document is outside the current workspace")
	}
	return state.Root, relativePath, nil
}

func (a *App) readingTabFromRelativePath(root string, relativePath string) (ReadingTab, error) {
	relativePath = normalizeMemoryRelativePath(relativePath)
	if relativePath == "" {
		return ReadingTab{}, errors.New("empty reading tab path")
	}
	abs, err := workspacePathFromRoot(root, relativePath)
	if err != nil {
		return ReadingTab{}, err
	}
	info, err := os.Stat(abs)
	if err != nil || info.IsDir() || !isMarkdownFile(abs) || !a.isWithinWorkspace(abs) {
		return ReadingTab{}, errors.New("reading tab is unavailable")
	}
	return ReadingTab{
		Path:         filepath.Clean(abs),
		RelativePath: relativePath,
		Name:         filepath.Base(abs),
	}, nil
}

func (a *App) readingPositionFromState(root string, relativePath string, state DocumentReadingState) (*ReadingPosition, error) {
	relativePath = normalizeMemoryRelativePath(firstNonEmpty(state.RelativePath, relativePath))
	if relativePath == "" {
		return nil, errors.New("empty reading path")
	}
	abs, err := workspacePathFromRoot(root, relativePath)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(abs)
	if err != nil || info.IsDir() || !isMarkdownFile(abs) || !a.isWithinWorkspace(abs) {
		return nil, errors.New("reading document is unavailable")
	}
	return &ReadingPosition{
		Path:         filepath.Clean(abs),
		RelativePath: relativePath,
		ScrollTop:    state.ScrollTop,
		ScrollRatio:  state.ScrollRatio,
		HeadingID:    state.HeadingID,
		ModifiedAt:   state.ModifiedAt,
		Size:         state.Size,
		UpdatedAt:    state.UpdatedAt,
	}, nil
}

func loadReadingMemory(path string) (ReadingMemoryStore, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return defaultReadingMemory(), nil
	}
	if err != nil {
		return ReadingMemoryStore{}, err
	}
	memory := defaultReadingMemory()
	if err := json.Unmarshal(data, &memory); err != nil {
		if backupErr := backupBadFile(path); backupErr != nil {
			return ReadingMemoryStore{}, backupErr
		}
		return defaultReadingMemory(), nil
	}
	normalizeReadingMemory(&memory)
	return memory, nil
}

func saveReadingMemory(path string, memory ReadingMemoryStore) error {
	normalizeReadingMemory(&memory)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(memory, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	dir := filepath.Dir(path)
	temp, err := os.CreateTemp(dir, "reading-memory-*.tmp")
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

func normalizeReadingMemory(memory *ReadingMemoryStore) {
	memory.StorageVersion = currentReadingMemoryVersion
	if memory.Workspaces == nil {
		memory.Workspaces = map[string]*WorkspaceReadingLog{}
	}
	for key, workspace := range memory.Workspaces {
		if workspace == nil {
			delete(memory.Workspaces, key)
			continue
		}
		workspace.Root = filepath.Clean(workspace.Root)
		if workspace.Root == "." {
			workspace.Root = ""
		}
		if workspace.Documents == nil {
			workspace.Documents = map[string]DocumentReadingState{}
		}
		workspace.OpenTabs = normalizeOpenTabs(workspace.OpenTabs)
		for documentKey, state := range workspace.Documents {
			relativePath := normalizeMemoryRelativePath(firstNonEmpty(state.RelativePath, documentKey))
			if relativePath == "" {
				delete(workspace.Documents, documentKey)
				continue
			}
			if state.ScrollTop < 0 {
				state.ScrollTop = 0
			}
			if math.IsNaN(state.ScrollRatio) || math.IsInf(state.ScrollRatio, 0) || state.ScrollRatio < 0 {
				state.ScrollRatio = 0
			}
			if state.ScrollRatio > 1 {
				state.ScrollRatio = 1
			}
			state.RelativePath = relativePath
			if documentKey != relativePath {
				delete(workspace.Documents, documentKey)
			}
			workspace.Documents[relativePath] = state
		}
		workspace.LastDocument = normalizeMemoryRelativePath(workspace.LastDocument)
		if workspace.LastDocument != "" {
			if _, ok := workspace.Documents[workspace.LastDocument]; !ok {
				workspace.LastDocument = ""
			}
		}
		workspace.ActiveDocument = normalizeMemoryRelativePath(firstNonEmpty(workspace.ActiveDocument, workspace.LastDocument))
		if len(workspace.OpenTabs) == 0 && workspace.ActiveDocument != "" {
			workspace.OpenTabs = []string{workspace.ActiveDocument}
		}
		if workspace.ActiveDocument != "" && !containsString(workspace.OpenTabs, workspace.ActiveDocument) {
			workspace.OpenTabs = append(workspace.OpenTabs, workspace.ActiveDocument)
			workspace.OpenTabs = normalizeOpenTabs(workspace.OpenTabs)
		}
		if workspace.ActiveDocument == "" && len(workspace.OpenTabs) > 0 {
			workspace.ActiveDocument = workspace.OpenTabs[0]
		}
		workspace.LastDocument = workspace.ActiveDocument
		pruneReadingDocuments(workspace)
	}
}

func pruneReadingDocuments(workspace *WorkspaceReadingLog) {
	if len(workspace.Documents) <= maxReadingMemoryDocuments {
		return
	}
	type entry struct {
		path      string
		updatedAt int64
	}
	entries := make([]entry, 0, len(workspace.Documents))
	for path, state := range workspace.Documents {
		entries = append(entries, entry{path: path, updatedAt: state.UpdatedAt})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].updatedAt > entries[j].updatedAt
	})
	keep := map[string]struct{}{}
	if workspace.LastDocument != "" {
		keep[workspace.LastDocument] = struct{}{}
	}
	if workspace.ActiveDocument != "" {
		keep[workspace.ActiveDocument] = struct{}{}
	}
	for _, path := range workspace.OpenTabs {
		keep[path] = struct{}{}
	}
	for _, entry := range entries {
		if len(keep) >= maxReadingMemoryDocuments {
			break
		}
		keep[entry.path] = struct{}{}
	}
	for path := range workspace.Documents {
		if _, ok := keep[path]; !ok {
			delete(workspace.Documents, path)
		}
	}
}

func workspaceMemoryKey(root string) string {
	return filepath.Clean(root)
}

func workspacePathFromRoot(root string, path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("empty workspace path")
	}
	localPath := filepath.Clean(filepath.FromSlash(path))
	if filepath.IsAbs(localPath) || localPath == ".." || strings.HasPrefix(localPath, ".."+string(filepath.Separator)) {
		return "", errors.New("workspace path is outside the current workspace")
	}
	return filepath.Join(filepath.Clean(root), localPath), nil
}

func normalizeMemoryRelativePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	localPath := filepath.Clean(filepath.FromSlash(path))
	if filepath.IsAbs(localPath) || localPath == "." || localPath == ".." || strings.HasPrefix(localPath, ".."+string(filepath.Separator)) {
		return ""
	}
	return filepath.ToSlash(localPath)
}

func normalizeOpenTabs(paths []string) []string {
	tabs := make([]string, 0, minInt(len(paths), maxOpenReadingTabs))
	seen := map[string]struct{}{}
	for _, path := range paths {
		relativePath := normalizeMemoryRelativePath(path)
		if relativePath == "" {
			continue
		}
		if _, ok := seen[relativePath]; ok {
			continue
		}
		seen[relativePath] = struct{}{}
		tabs = append(tabs, relativePath)
		if len(tabs) >= maxOpenReadingTabs {
			break
		}
	}
	return tabs
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
