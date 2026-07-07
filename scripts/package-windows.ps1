Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$manifestPath = Join-Path $repoRoot "product.manifest.json"
$manifest = Get-Content -LiteralPath $manifestPath -Raw -Encoding UTF8 | ConvertFrom-Json

$version = [string]$manifest.version
$binaryName = [string]$manifest.binary_name
$artifactName = "$binaryName-v$version-windows-amd64-setup.exe"
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
[System.IO.File]::WriteAllText($shaPath, $sumLine + [Environment]::NewLine, [System.Text.UTF8Encoding]::new($false))

[PSCustomObject]@{
    Version = $version
    Installer = $target
    Sha256 = $hash
    Sha256File = $shaPath
}
