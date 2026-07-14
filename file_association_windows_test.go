//go:build windows

package main

import "testing"

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
