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

func readInstallerScript(t *testing.T) string {
	t.Helper()
	content, err := os.ReadFile(`build/windows/installer/project.nsi`)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
