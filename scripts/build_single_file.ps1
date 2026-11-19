# Build script for single-file standalone Discord-Portable.exe and Discord.exe
$ErrorActionPreference = "Stop"

Write-Host "=========================================================="
Write-Host " Building Single-File Standalone Discord Portable"
Write-Host "=========================================================="

$rootDir = Split-Path -Parent $PSScriptRoot
Set-Location $rootDir

# 1. Compile launcher
Write-Host "[1/4] Compiling Go launcher..."
go build -v -ldflags "-s -w" -o DiscordPTB-Portable.exe main.go
if ($LASTEXITCODE -ne 0) {
    throw "Go build failed!"
}

# 2. Package payload archive
Write-Host "[2/4] Packaging payload archive (app, res, tools)..."
$scratchDir = Join-Path $rootDir ".agent\scratch"
if (-not (Test-Path $scratchDir)) {
    New-Item -ItemType Directory -Path $scratchDir -Force | Out-Null
}
$payloadZip = Join-Path $scratchDir "payload.zip"
if (Test-Path $payloadZip) {
    Remove-Item -Path $payloadZip -Force
}

$7z = "C:\Program Files\7-Zip\7z.exe"
if (-not (Test-Path $7z)) {
    $7z = (Get-Command 7z.exe -ErrorAction SilentlyContinue).Source
}
if (-not $7z) {
    throw "7-Zip not found!"
}

& $7z a -tzip -mx=5 $payloadZip app/ res/ Discord-ProfileManager.exe portapp.json
if ($LASTEXITCODE -ne 0) {
    throw "7-Zip payload compression failed!"
}

# 3. Combine into single-file executable
Write-Host "[3/4] Creating single-file Discord-Portable.exe..."
$outExe = Join-Path $rootDir "Discord-Portable.exe"
cmd /c "copy /b DiscordPTB-Portable.exe + `"$payloadZip`" `"$outExe`""
if ($LASTEXITCODE -ne 0) {
    throw "Failed to create single-file executable!"
}

# 4. Copy as Discord.exe alias
Write-Host "[4/4] Creating Discord.exe alias..."
$aliasExe = Join-Path $rootDir "Discord.exe"
Copy-Item -Path $outExe -Destination $aliasExe -Force

$fileInfo = Get-Item $outExe
$sizeMB = [math]::Round($fileInfo.Length / 1MB, 2)
Write-Host "=========================================================="
Write-Host " SUCCESS: Single-file Discord built successfully!"
Write-Host " File: $outExe"
Write-Host " Alias: $aliasExe"
Write-Host " Size: $sizeMB MB"
Write-Host "=========================================================="
