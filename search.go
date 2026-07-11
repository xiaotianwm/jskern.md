package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const maxSearchResults = 50

type SearchResult struct {
	Path         string `json:"path"`
	Name         string `json:"name"`
	RelativePath string `json:"relativePath"`
	Kind         string `json:"kind"`
	Snippet      string `json:"snippet"`
	MatchLine    int    `json:"matchLine"`
}

func (a *App) SearchWorkspace(query string) ([]SearchResult, error) {
	query = strings.TrimSpace(query)
	if len([]rune(query)) < 2 {
		return []SearchResult{}, nil
	}

	a.mu.RLock()
	workspaces := append([]WorkspaceEntry(nil), a.settings.Workspaces...)
	a.mu.RUnlock()
	if len(workspaces) == 0 {
		return nil, errors.New("workspace is not open")
	}

	lowerQuery := strings.ToLower(query)
	results := []SearchResult{}
	for _, workspace := range workspaces {
		root := workspace.Path
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				if entry != nil && entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if path != root && shouldSkipEntry(entry.Name(), entry.IsDir()) {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if entry.IsDir() || !isMarkdownFile(entry.Name()) {
				return nil
			}

			abs := filepath.Clean(path)
			if !a.isWithinWorkspace(abs) {
				return nil
			}
			rel, ok := workspaceRelativePathFromRoot(root, abs)
			if !ok {
				return nil
			}
			displayRel := rel
			if len(workspaces) > 1 {
				displayRel = workspace.Name + "/" + rel
			}

			name := entry.Name()
			if strings.Contains(strings.ToLower(name), lowerQuery) || strings.Contains(strings.ToLower(displayRel), lowerQuery) {
				results = append(results, SearchResult{
					Path:         abs,
					Name:         name,
					RelativePath: displayRel,
					Kind:         "file",
					Snippet:      displayRel,
					MatchLine:    0,
				})
				if len(results) >= maxSearchResults {
					return filepath.SkipAll
				}
			}

			info, err := entry.Info()
			if err != nil || info.Size() > maxMarkdownBytes {
				return nil
			}
			data, err := os.ReadFile(abs)
			if err != nil {
				return nil
			}
			if snippet, matchLine, ok := searchSnippet(string(data), query); ok {
				results = append(results, SearchResult{
					Path:         abs,
					Name:         name,
					RelativePath: displayRel,
					Kind:         "content",
					Snippet:      snippet,
					MatchLine:    matchLine,
				})
				if len(results) >= maxSearchResults {
					return filepath.SkipAll
				}
			}
			return nil
		})
		if err != nil && err != filepath.SkipAll {
			return nil, err
		}
		if len(results) >= maxSearchResults {
			break
		}
	}
	return results, nil
}

func searchSnippet(content string, query string) (string, int, bool) {
	lowerQuery := strings.ToLower(query)
	for index, line := range strings.Split(content, "\n") {
		compact := strings.Join(strings.Fields(line), " ")
		if compact == "" || !strings.Contains(strings.ToLower(compact), lowerQuery) {
			continue
		}
		runes := []rune(compact)
		if len(runes) <= 180 {
			return compact, index + 1, true
		}
		return string(runes[:177]) + "...", index + 1, true
	}
	return "", 0, false
}
