//go:build !windows

package main

import "errors"

func platformMarkdownAssociationStatus() (MarkdownAssociationStatus, error) {
	return MarkdownAssociationStatus{}, nil
}

func platformOpenMarkdownDefaultAppsSettings() error {
	return errors.New("default application settings are only available on Windows")
}
