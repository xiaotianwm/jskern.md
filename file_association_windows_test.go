//go:build windows

package main

import (
	"errors"
	"slices"
	"testing"
)

func TestIsJSKernMarkdownProgID(t *testing.T) {
	for _, progID := range []string{
		"JSKernMD.Markdown",
		"jskernmd.markdown",
		`Applications\jskernmd.exe`,
		`applications\JSKERNMD.EXE`,
	} {
		if !isJSKernMarkdownProgID(progID) {
			t.Errorf("expected %q to identify JS Kern.md", progID)
		}
	}
	if isJSKernMarkdownProgID("Other.Markdown") {
		t.Fatal("unexpected match for another Markdown application")
	}
}

func TestOpenMarkdownDefaultAppsSettings(t *testing.T) {
	tests := []struct {
		name     string
		firstErr error
		expected []string
	}{
		{
			name:     "registered app page opens",
			expected: []string{"ms-settings:defaultapps?registeredAppMachine=JS%20Kern.md"},
		},
		{
			name:     "falls back to default apps page",
			firstErr: errors.New("primary URI failed"),
			expected: []string{
				"ms-settings:defaultapps?registeredAppMachine=JS%20Kern.md",
				"ms-settings:defaultapps",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var opened []string
			err := openMarkdownDefaultAppsSettings(func(uri string) error {
				opened = append(opened, uri)
				if len(opened) == 1 {
					return test.firstErr
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(opened, test.expected) {
				t.Fatalf("expected URIs %v, got %v", test.expected, opened)
			}
		})
	}
}
