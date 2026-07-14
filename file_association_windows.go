//go:build windows

package main

import (
	"net/url"
	"strings"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	markdownProgID            = "JSKernMD.Markdown"
	markdownApplicationProgID = `Applications\jskernmd.exe`
	windowsDefaultAppsURI     = "ms-settings:defaultapps"
)

func platformMarkdownAssociationStatus() (MarkdownAssociationStatus, error) {
	status := MarkdownAssociationStatus{Supported: true}

	registeredApps, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SOFTWARE\RegisteredApplications`,
		registry.QUERY_VALUE|registry.WOW64_64KEY,
	)
	if err == registry.ErrNotExist {
		return status, nil
	}
	if err != nil {
		return status, err
	}
	capabilitiesPath, _, err := registeredApps.GetStringValue(productInfo.DisplayName)
	_ = registeredApps.Close()
	if err == registry.ErrNotExist {
		return status, nil
	}
	if err != nil {
		return status, err
	}

	associations, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		capabilitiesPath+`\FileAssociations`,
		registry.QUERY_VALUE|registry.WOW64_64KEY,
	)
	if err == registry.ErrNotExist {
		return status, nil
	}
	if err != nil {
		return status, err
	}
	status.Registered = true
	for _, extension := range []string{".md", ".markdown", ".mdown"} {
		progID, _, readErr := associations.GetStringValue(extension)
		if readErr != nil || !strings.EqualFold(progID, markdownProgID) {
			status.Registered = false
			break
		}
	}
	_ = associations.Close()

	userChoice, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Explorer\FileExts\.md\UserChoice`,
		registry.QUERY_VALUE,
	)
	if err == nil {
		progID, _, readErr := userChoice.GetStringValue("ProgId")
		status.Default = readErr == nil && isJSKernMarkdownProgID(progID)
		_ = userChoice.Close()
	}

	return status, nil
}

func isJSKernMarkdownProgID(progID string) bool {
	return strings.EqualFold(progID, markdownProgID) || strings.EqualFold(progID, markdownApplicationProgID)
}

func platformOpenMarkdownDefaultAppsSettings() error {
	return openMarkdownDefaultAppsSettings(shellOpenWindowsURI)
}

func openMarkdownDefaultAppsSettings(openURI func(string) error) error {
	settingsURI := windowsDefaultAppsURI + "?registeredAppMachine=" + url.PathEscape(productInfo.DisplayName)
	if err := openURI(settingsURI); err == nil {
		return nil
	}
	return openURI(windowsDefaultAppsURI)
}

func shellOpenWindowsURI(uri string) error {
	uriPointer, err := windows.UTF16PtrFromString(uri)
	if err != nil {
		return err
	}
	return windows.ShellExecute(0, nil, uriPointer, nil, nil, windows.SW_SHOWNORMAL)
}
