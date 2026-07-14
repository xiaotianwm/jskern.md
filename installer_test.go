package main

import (
	"os"
	"strings"
	"testing"
)

func TestWindowsInstallerMarkdownRegistrationIsSymmetric(t *testing.T) {
	script := readInstallerScript(t)
	pairs := [][2]string{
		{`WriteRegStr HKLM "Software\Classes\.md\OpenWithProgids" "${MARKDOWN_PROGID}" ""`, `DeleteRegValue HKLM "Software\Classes\.md\OpenWithProgids" "${MARKDOWN_PROGID}"`},
		{`WriteRegStr HKLM "Software\Classes\.markdown\OpenWithProgids" "${MARKDOWN_PROGID}" ""`, `DeleteRegValue HKLM "Software\Classes\.markdown\OpenWithProgids" "${MARKDOWN_PROGID}"`},
		{`WriteRegStr HKLM "Software\Classes\.mdown\OpenWithProgids" "${MARKDOWN_PROGID}" ""`, `DeleteRegValue HKLM "Software\Classes\.mdown\OpenWithProgids" "${MARKDOWN_PROGID}"`},
		{`WriteRegStr HKLM "Software\Classes\${MARKDOWN_PROGID}"`, `DeleteRegKey HKLM "Software\Classes\${MARKDOWN_PROGID}"`},
		{`WriteRegStr HKLM "Software\Classes\Applications\${PRODUCT_EXECUTABLE}"`, `DeleteRegKey HKLM "Software\Classes\Applications\${PRODUCT_EXECUTABLE}"`},
		{`WriteRegStr HKLM "Software\RegisteredApplications" "${INFO_PRODUCTNAME}"`, `DeleteRegValue HKLM "Software\RegisteredApplications" "${INFO_PRODUCTNAME}"`},
		{`WriteRegStr HKLM "${MARKDOWN_CAPABILITIES_KEY}"`, `DeleteRegKey HKLM "Software\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"`},
		{`WriteRegStr HKLM "Software\Classes\SystemFileAssociations\.md\shell\JSKernMD.Open"`, `DeleteRegKey HKLM "Software\Classes\SystemFileAssociations\.md\shell\JSKernMD.Open"`},
		{`WriteRegStr HKLM "Software\Classes\SystemFileAssociations\.markdown\shell\JSKernMD.Open"`, `DeleteRegKey HKLM "Software\Classes\SystemFileAssociations\.markdown\shell\JSKernMD.Open"`},
		{`WriteRegStr HKLM "Software\Classes\SystemFileAssociations\.mdown\shell\JSKernMD.Open"`, `DeleteRegKey HKLM "Software\Classes\SystemFileAssociations\.mdown\shell\JSKernMD.Open"`},
		{`WriteRegStr HKLM "Software\Classes\Directory\shell\JSKernMD.AddWorkspace"`, `DeleteRegKey HKLM "Software\Classes\Directory\shell\JSKernMD.AddWorkspace"`},
		{`WriteRegStr HKLM "Software\Classes\Folder\shell\JSKernMD.AddWorkspace"`, `DeleteRegKey HKLM "Software\Classes\Folder\shell\JSKernMD.AddWorkspace"`},
	}

	for _, pair := range pairs {
		if !strings.Contains(script, pair[0]) {
			t.Errorf("installer is missing registration %q", pair[0])
		}
		if !strings.Contains(script, pair[1]) {
			t.Errorf("uninstaller is missing cleanup %q", pair[1])
		}
	}
	for _, extension := range []string{".md", ".markdown", ".mdown"} {
		if strings.Contains(script, `DeleteRegKey HKLM "Software\Classes\`+extension+`"`) {
			t.Errorf("uninstaller must not delete the shared %s file-class key", extension)
		}
	}
}

func TestWindowsInstallerDoesNotWriteProtectedUserChoice(t *testing.T) {
	for _, line := range strings.Split(readInstallerScript(t), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "WriteReg") && strings.Contains(strings.ToLower(trimmed), "userchoice") {
			t.Fatalf("installer must not write the protected Windows UserChoice key: %s", trimmed)
		}
	}
}

func TestWindowsInstallerUsesDedicatedMarkdownFileIcon(t *testing.T) {
	script := readInstallerScript(t)
	icon, err := os.Stat(`build/windows/markdown-file.ico`)
	if err != nil {
		t.Fatal(err)
	}
	if icon.Size() == 0 {
		t.Fatal("dedicated Markdown icon is empty")
	}

	required := []string{
		`!define MARKDOWN_FILE_ICON "markdown-file.ico"`,
		`File /oname=${MARKDOWN_FILE_ICON} "..\markdown-file.ico"`,
		`WriteRegStr HKLM "Software\Classes\${MARKDOWN_PROGID}\DefaultIcon" "" '"$INSTDIR\${MARKDOWN_FILE_ICON}",0'`,
		`WriteRegStr HKLM "Software\Classes\Applications\${PRODUCT_EXECUTABLE}\DefaultIcon" "" '"$INSTDIR\${MARKDOWN_FILE_ICON}",0'`,
		`WriteRegStr HKLM "${MARKDOWN_CAPABILITIES_KEY}" "ApplicationIcon" '"$INSTDIR\${PRODUCT_EXECUTABLE}",0'`,
	}
	for _, value := range required {
		if !strings.Contains(script, value) {
			t.Errorf("installer is missing dedicated Markdown icon behavior %q", value)
		}
	}

	legacy := []string{
		`WriteRegStr HKLM "Software\Classes\${MARKDOWN_PROGID}\DefaultIcon" "" '"$INSTDIR\${PRODUCT_EXECUTABLE}",0'`,
		`WriteRegStr HKLM "Software\Classes\Applications\${PRODUCT_EXECUTABLE}\DefaultIcon" "" '"$INSTDIR\${PRODUCT_EXECUTABLE}",0'`,
	}
	for _, value := range legacy {
		if strings.Contains(script, value) {
			t.Fatalf("Markdown documents must not reuse the application executable icon: %q", value)
		}
	}
}

func readInstallerScript(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile(`build/windows/installer/project.nsi`)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
