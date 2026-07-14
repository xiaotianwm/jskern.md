Unicode true

####
## Please note: Template replacements don't work in this file. They are provided with default defines like
## mentioned underneath.
## If the keyword is not defined, "wails_tools.nsh" will populate them with the values from ProjectInfo.
## If they are defined here, "wails_tools.nsh" will not touch them. This allows to use this project.nsi manually
## from outside of Wails for debugging and development of the installer.
##
## For development first make a wails nsis build to populate the "wails_tools.nsh":
## > wails build --target windows/amd64 --nsis
## Then you can call makensis on this file with specifying the path to your binary:
## For a AMD64 only installer:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app.exe
## For a ARM64 only installer:
## > makensis -DARG_WAILS_ARM64_BINARY=..\..\bin\app.exe
## For a installer with both architectures:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app-amd64.exe -DARG_WAILS_ARM64_BINARY=..\..\bin\app-arm64.exe
####
## The following information is taken from the ProjectInfo file, but they can be overwritten here.
####
## !define INFO_PROJECTNAME    "MyProject" # Default "{{.Name}}"
## !define INFO_COMPANYNAME    "MyCompany" # Default "{{.Info.CompanyName}}"
## !define INFO_PRODUCTNAME    "MyProduct" # Default "{{.Info.ProductName}}"
## !define INFO_PRODUCTVERSION "1.0.0"     # Default "{{.Info.ProductVersion}}"
## !define INFO_COPYRIGHT      "Copyright" # Default "{{.Info.Copyright}}"
###
## !define PRODUCT_EXECUTABLE  "Application.exe"      # Default "${INFO_PROJECTNAME}.exe"
## !define UNINST_KEY_NAME     "UninstKeyInRegistry"  # Default "${INFO_COMPANYNAME}${INFO_PRODUCTNAME}"
####
## !define REQUEST_EXECUTION_LEVEL "admin"            # Default "admin"  see also https://nsis.sourceforge.io/Docs/Chapter4.html
####
## Include the wails tools
####
!include "wails_tools.nsh"
!include "LogicLib.nsh"

!define MARKDOWN_PROGID "JSKernMD.Markdown"
!define MARKDOWN_CAPABILITIES_KEY "Software\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}\Capabilities"

# The version information for this two must consist of 4 parts
VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"

VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

# Enable HiDPI support. https://nsis.sourceforge.io/Reference/ManifestDPIAware
ManifestDPIAware true

!include "MUI.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
# !define MUI_WELCOMEFINISHPAGE_BITMAP "resources\leftimage.bmp" #Include this to add a bitmap on the left side of the Welcome Page. Must be a size of 164x314
!define MUI_FINISHPAGE_NOAUTOCLOSE # Wait on the INSTFILES page so the user can take a look into the details of the installation steps
!define MUI_ABORTWARNING # This will warn the user if they exit from the installer.

!insertmacro MUI_PAGE_WELCOME # Welcome to the installer page.
# !insertmacro MUI_PAGE_LICENSE "resources\eula.txt" # Adds a EULA page to the installer
!insertmacro MUI_PAGE_DIRECTORY # In which folder install page.
!insertmacro MUI_PAGE_INSTFILES # Installing page.
!insertmacro MUI_PAGE_FINISH # Finished installation page.

!insertmacro MUI_UNPAGE_INSTFILES # Uinstalling page

!insertmacro MUI_LANGUAGE "English"
!insertmacro MUI_LANGUAGE "SimpChinese"

## The following two statements can be used to sign the installer and the uninstaller. The path to the binaries are provided in %1
#!uninstfinalize 'signtool --file "%1"'
#!finalize 'signtool --file "%1"'

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe" # Name of the installer's file.
InstallDir "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}" # Default installing folder ($PROGRAMFILES is Program Files folder).
ShowInstDetails show # This will always show the installation details.

Function .onInit
   Call SelectInstallerLanguage
   !insertmacro wails.checkArchitecture
   Call ResolvePreviousInstallDir
FunctionEnd

Function SelectInstallerLanguage
    System::Call 'kernel32::GetUserDefaultUILanguage() i .r0'
    IntOp $0 $0 & 0x3ff
    ${If} $0 == 4
        StrCpy $LANGUAGE ${LANG_SIMPCHINESE}
    ${Else}
        StrCpy $LANGUAGE ${LANG_ENGLISH}
    ${EndIf}
FunctionEnd

Function un.onInit
    System::Call 'kernel32::GetUserDefaultUILanguage() i .r0'
    IntOp $0 $0 & 0x3ff
    ${If} $0 == 4
        StrCpy $LANGUAGE ${LANG_SIMPCHINESE}
    ${Else}
        StrCpy $LANGUAGE ${LANG_ENGLISH}
    ${EndIf}
FunctionEnd

Function ResolvePreviousInstallDir
    SetRegView 64
    ReadRegStr $0 HKLM "${UNINST_KEY}" "InstallLocation"
    ${If} $0 != ""
        IfFileExists "$0\*.*" 0 noInstallLocation
        StrCpy $INSTDIR "$0"
noInstallLocation:
        Return
    ${EndIf}

    ReadRegStr $0 HKLM "${UNINST_KEY}" "UninstallString"
    ${If} $0 == ""
        Return
    ${EndIf}

    StrCpy $1 $0 1
    ${If} $1 == '"'
        StrCpy $0 $0 "" 1
        StrLen $1 $0
        IntOp $1 $1 - 1
        StrCpy $0 $0 $1
    ${EndIf}

    IfFileExists "$0" 0 donePreviousInstallDir
    ${GetParent} "$0" $1
    IfFileExists "$1\*.*" 0 donePreviousInstallDir
    StrCpy $INSTDIR "$1"
donePreviousInstallDir:
FunctionEnd

Function RegisterWindowsIntegration
    SetRegView 64

    # Remove current-user keys created by releases before machine-wide registration.
    DeleteRegKey HKCU "Software\Classes\SystemFileAssociations\.md\shell\JSKernMD.Open"
    DeleteRegKey HKCU "Software\Classes\SystemFileAssociations\.markdown\shell\JSKernMD.Open"
    DeleteRegKey HKCU "Software\Classes\SystemFileAssociations\.mdown\shell\JSKernMD.Open"
    DeleteRegKey HKCU "Software\Classes\Directory\shell\JSKernMD.AddWorkspace"
    DeleteRegKey HKCU "Software\Classes\Folder\shell\JSKernMD.AddWorkspace"

    WriteRegStr HKLM "Software\Classes\${MARKDOWN_PROGID}" "" "JS Kern.md Markdown Document"
    WriteRegStr HKLM "Software\Classes\${MARKDOWN_PROGID}\DefaultIcon" "" '"$INSTDIR\${PRODUCT_EXECUTABLE}",0'
    WriteRegStr HKLM "Software\Classes\${MARKDOWN_PROGID}\shell\open\command" "" '"$INSTDIR\${PRODUCT_EXECUTABLE}" --open-file "%1"'

    WriteRegStr HKLM "Software\Classes\.md\OpenWithProgids" "${MARKDOWN_PROGID}" ""
    WriteRegStr HKLM "Software\Classes\.markdown\OpenWithProgids" "${MARKDOWN_PROGID}" ""
    WriteRegStr HKLM "Software\Classes\.mdown\OpenWithProgids" "${MARKDOWN_PROGID}" ""

    WriteRegStr HKLM "Software\Classes\Applications\${PRODUCT_EXECUTABLE}" "FriendlyAppName" "${INFO_PRODUCTNAME}"
    WriteRegStr HKLM "Software\Classes\Applications\${PRODUCT_EXECUTABLE}\DefaultIcon" "" '"$INSTDIR\${PRODUCT_EXECUTABLE}",0'
    WriteRegStr HKLM "Software\Classes\Applications\${PRODUCT_EXECUTABLE}\shell\open\command" "" '"$INSTDIR\${PRODUCT_EXECUTABLE}" --open-file "%1"'
    WriteRegStr HKLM "Software\Classes\Applications\${PRODUCT_EXECUTABLE}\SupportedTypes" ".md" ""
    WriteRegStr HKLM "Software\Classes\Applications\${PRODUCT_EXECUTABLE}\SupportedTypes" ".markdown" ""
    WriteRegStr HKLM "Software\Classes\Applications\${PRODUCT_EXECUTABLE}\SupportedTypes" ".mdown" ""

    WriteRegStr HKLM "${MARKDOWN_CAPABILITIES_KEY}" "ApplicationName" "${INFO_PRODUCTNAME}"
    WriteRegStr HKLM "${MARKDOWN_CAPABILITIES_KEY}" "ApplicationDescription" "Desktop Markdown reader"
    WriteRegStr HKLM "${MARKDOWN_CAPABILITIES_KEY}" "ApplicationIcon" '"$INSTDIR\${PRODUCT_EXECUTABLE}",0'
    WriteRegStr HKLM "${MARKDOWN_CAPABILITIES_KEY}\FileAssociations" ".md" "${MARKDOWN_PROGID}"
    WriteRegStr HKLM "${MARKDOWN_CAPABILITIES_KEY}\FileAssociations" ".markdown" "${MARKDOWN_PROGID}"
    WriteRegStr HKLM "${MARKDOWN_CAPABILITIES_KEY}\FileAssociations" ".mdown" "${MARKDOWN_PROGID}"
    WriteRegStr HKLM "Software\RegisteredApplications" "${INFO_PRODUCTNAME}" "${MARKDOWN_CAPABILITIES_KEY}"

    WriteRegStr HKLM "Software\Classes\SystemFileAssociations\.md\shell\JSKernMD.Open" "" "Open with JS Kern.md"
    WriteRegStr HKLM "Software\Classes\SystemFileAssociations\.md\shell\JSKernMD.Open" "Icon" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    WriteRegStr HKLM "Software\Classes\SystemFileAssociations\.md\shell\JSKernMD.Open\command" "" '"$INSTDIR\${PRODUCT_EXECUTABLE}" --open-file "%1"'
    WriteRegStr HKLM "Software\Classes\SystemFileAssociations\.markdown\shell\JSKernMD.Open" "" "Open with JS Kern.md"
    WriteRegStr HKLM "Software\Classes\SystemFileAssociations\.markdown\shell\JSKernMD.Open" "Icon" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    WriteRegStr HKLM "Software\Classes\SystemFileAssociations\.markdown\shell\JSKernMD.Open\command" "" '"$INSTDIR\${PRODUCT_EXECUTABLE}" --open-file "%1"'
    WriteRegStr HKLM "Software\Classes\SystemFileAssociations\.mdown\shell\JSKernMD.Open" "" "Open with JS Kern.md"
    WriteRegStr HKLM "Software\Classes\SystemFileAssociations\.mdown\shell\JSKernMD.Open" "Icon" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    WriteRegStr HKLM "Software\Classes\SystemFileAssociations\.mdown\shell\JSKernMD.Open\command" "" '"$INSTDIR\${PRODUCT_EXECUTABLE}" --open-file "%1"'
    WriteRegStr HKLM "Software\Classes\Directory\shell\JSKernMD.AddWorkspace" "" "Add to JS Kern.md workspace"
    WriteRegStr HKLM "Software\Classes\Directory\shell\JSKernMD.AddWorkspace" "Icon" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    WriteRegStr HKLM "Software\Classes\Directory\shell\JSKernMD.AddWorkspace\command" "" '"$INSTDIR\${PRODUCT_EXECUTABLE}" --add-workspace "%1"'
    WriteRegStr HKLM "Software\Classes\Folder\shell\JSKernMD.AddWorkspace" "" "Add to JS Kern.md workspace"
    WriteRegStr HKLM "Software\Classes\Folder\shell\JSKernMD.AddWorkspace" "Icon" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    WriteRegStr HKLM "Software\Classes\Folder\shell\JSKernMD.AddWorkspace\command" "" '"$INSTDIR\${PRODUCT_EXECUTABLE}" --add-workspace "%1"'

    System::Call 'shell32::SHChangeNotify(i 0x08000000, i 0, p 0, p 0)'
FunctionEnd

Function un.UnregisterWindowsIntegration
    SetRegView 64
    DeleteRegValue HKLM "Software\Classes\.md\OpenWithProgids" "${MARKDOWN_PROGID}"
    DeleteRegValue HKLM "Software\Classes\.markdown\OpenWithProgids" "${MARKDOWN_PROGID}"
    DeleteRegValue HKLM "Software\Classes\.mdown\OpenWithProgids" "${MARKDOWN_PROGID}"
    DeleteRegKey HKLM "Software\Classes\${MARKDOWN_PROGID}"
    DeleteRegKey HKLM "Software\Classes\Applications\${PRODUCT_EXECUTABLE}"
    DeleteRegValue HKLM "Software\RegisteredApplications" "${INFO_PRODUCTNAME}"
    DeleteRegKey HKLM "Software\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
    DeleteRegKey HKLM "Software\Classes\SystemFileAssociations\.md\shell\JSKernMD.Open"
    DeleteRegKey HKLM "Software\Classes\SystemFileAssociations\.markdown\shell\JSKernMD.Open"
    DeleteRegKey HKLM "Software\Classes\SystemFileAssociations\.mdown\shell\JSKernMD.Open"
    DeleteRegKey HKLM "Software\Classes\Directory\shell\JSKernMD.AddWorkspace"
    DeleteRegKey HKLM "Software\Classes\Folder\shell\JSKernMD.AddWorkspace"
    DeleteRegKey HKCU "Software\Classes\SystemFileAssociations\.md\shell\JSKernMD.Open"
    DeleteRegKey HKCU "Software\Classes\SystemFileAssociations\.markdown\shell\JSKernMD.Open"
    DeleteRegKey HKCU "Software\Classes\SystemFileAssociations\.mdown\shell\JSKernMD.Open"
    DeleteRegKey HKCU "Software\Classes\Directory\shell\JSKernMD.AddWorkspace"
    DeleteRegKey HKCU "Software\Classes\Folder\shell\JSKernMD.AddWorkspace"
    System::Call 'shell32::SHChangeNotify(i 0x08000000, i 0, p 0, p 0)'
FunctionEnd

Section
    !insertmacro wails.setShellContext

    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR

    !insertmacro wails.files

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols
    Call RegisterWindowsIntegration

    !insertmacro wails.writeUninstaller
    SetRegView 64
    WriteRegStr HKLM "${UNINST_KEY}" "InstallLocation" "$INSTDIR"
    WriteRegStr HKLM "${UNINST_KEY}" "InstallerLanguage" "$LANGUAGE"
SectionEnd

Section "uninstall"
    !insertmacro wails.setShellContext

    RMDir /r "$AppData\${PRODUCT_EXECUTABLE}" # Remove the WebView2 DataPath

    RMDir /r $INSTDIR

    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols
    Call un.UnregisterWindowsIntegration

    !insertmacro wails.deleteUninstaller
SectionEnd
