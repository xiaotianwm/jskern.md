Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$manifestPath = Join-Path $repoRoot "product.manifest.json"
$manifest = Get-Content -LiteralPath $manifestPath -Raw -Encoding UTF8 | ConvertFrom-Json

$version = [string]$manifest.version
$wailsConfigPath = Join-Path $repoRoot "wails.json"
$wailsConfig = Get-Content -LiteralPath $wailsConfigPath -Raw -Encoding UTF8 | ConvertFrom-Json
function ConvertTo-JsonString {
    param([AllowNull()][string]$Value)

    if ($null -eq $Value) {
        return "null"
    }

    return ($Value | ConvertTo-Json -Compress)
}

$wailsSchema = ConvertTo-JsonString ([string]$wailsConfig.'$schema')
$wailsName = ConvertTo-JsonString ([string]$wailsConfig.name)
$wailsOutputFilename = ConvertTo-JsonString ([string]$wailsConfig.outputfilename)
$wailsFrontendInstall = ConvertTo-JsonString ([string]$wailsConfig.'frontend:install')
$wailsFrontendBuild = ConvertTo-JsonString ([string]$wailsConfig.'frontend:build')
$wailsFrontendWatcher = ConvertTo-JsonString ([string]$wailsConfig.'frontend:dev:watcher')
$wailsFrontendServerUrl = ConvertTo-JsonString ([string]$wailsConfig.'frontend:dev:serverUrl')
$wailsAuthorName = ConvertTo-JsonString ([string]$wailsConfig.author.name)
$wailsAuthorEmail = ConvertTo-JsonString ([string]$wailsConfig.author.email)
$wailsCompanyName = ConvertTo-JsonString ([string]$manifest.company)
$wailsProductName = ConvertTo-JsonString ([string]$manifest.product_name)
$wailsProductVersion = ConvertTo-JsonString $version
$wailsCopyright = ConvertTo-JsonString "Copyright (c) 2026 $($manifest.company)"
$wailsComments = ConvertTo-JsonString "Built using Wails"
$wailsJson = @"
{
  "`$schema": $wailsSchema,
  "name": $wailsName,
  "outputfilename": $wailsOutputFilename,
  "frontend:install": $wailsFrontendInstall,
  "frontend:build": $wailsFrontendBuild,
  "frontend:dev:watcher": $wailsFrontendWatcher,
  "frontend:dev:serverUrl": $wailsFrontendServerUrl,
  "author": {
    "name": $wailsAuthorName,
    "email": $wailsAuthorEmail
  },
  "info": {
    "companyName": $wailsCompanyName,
    "productName": $wailsProductName,
    "productVersion": $wailsProductVersion,
    "copyright": $wailsCopyright,
    "comments": $wailsComments
  }
}
"@
[System.IO.File]::WriteAllText($wailsConfigPath, $wailsJson + "`n", [System.Text.UTF8Encoding]::new($false))

$artifactName = "JSKernMD-Setup-$version-x64.exe"
$releaseDir = Join-Path $repoRoot "dist\releases\v$version"

$makensis = Get-Command makensis -ErrorAction SilentlyContinue
if (-not $makensis) {
    foreach ($candidate in @("C:\Program Files\NSIS", "C:\Program Files (x86)\NSIS")) {
        $candidateExe = Join-Path $candidate "makensis.exe"
        if (Test-Path -LiteralPath $candidateExe) {
            $env:Path = "$candidate;$env:Path"
            break
        }
    }
}

Push-Location $repoRoot
try {
    wails build -clean -nsis -webview2 download
} finally {
    Pop-Location
}

$installer = Get-ChildItem -LiteralPath (Join-Path $repoRoot "build\bin") -File -Filter "*.exe" |
    Where-Object { $_.Name -match "installer|setup" } |
    Sort-Object LastWriteTime -Descending |
    Select-Object -First 1

if (-not $installer) {
    throw "Wails NSIS installer was not found under build\bin."
}

New-Item -ItemType Directory -Force -Path $releaseDir | Out-Null
$target = Join-Path $releaseDir $artifactName
Copy-Item -LiteralPath $installer.FullName -Destination $target -Force

$hash = (Get-FileHash -LiteralPath $target -Algorithm SHA256).Hash.ToLowerInvariant()
$sumLine = "$hash  $artifactName"
$shaPath = Join-Path $releaseDir "SHA256SUMS.txt"
[System.IO.File]::WriteAllText($shaPath, $sumLine + "`n", [System.Text.UTF8Encoding]::new($false))

[PSCustomObject]@{
    Version = $version
    Installer = $target
    Sha256 = $hash
    Sha256File = $shaPath
}
