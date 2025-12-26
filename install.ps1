# PowerShell install script
# Created by KevinYouu on 2025-05-21
# Description: Install easyGit
# Github: https://github.com/KevinYouu/easyGit

# Check if Git is installed
if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
    Write-Host "Git is not installed. Please install Git first:"
    Write-Host "https://git-scm.com/downloads"
    exit 1
}

# GitHub repo
$repo = "KevinYouu/easyGit"

# Get the latest release version
try {
    $version = (Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest").tag_name
    Write-Host "Latest version: $version"
} catch {
    Write-Host "Failed to fetch the latest version. Check your internet connection or GitHub API status."
    exit 1
}

# Check system architecture
$arch = $env:PROCESSOR_ARCHITECTURE
switch ($arch) {
    "AMD64" { $package_name = "windows_amd64" }
    "ARM64" { $package_name = "windows_arm64" }
    default {
        Write-Host "Unsupported system architecture: $arch"
        exit 1
    }
}

# Build download URL and file name
$url = "https://github.com/$repo/releases/download/$version/easyGit_${version}_${package_name}.zip"
$file = "easyGit_${version}_${package_name}.zip"

# Download ZIP file
Write-Host "Downloading: $url"
try {
    Invoke-WebRequest -Uri $url -OutFile $file
    Write-Host "Download successful"
} catch {
    Write-Host "Download failed"
    exit 1
}

# Extract ZIP
$destination = "$env:USERPROFILE\easyGit"
if (-not (Test-Path $destination)) {
    New-Item -ItemType Directory -Path $destination | Out-Null
}
try {
    Expand-Archive -Path $file -DestinationPath $destination -Force
    Write-Host "Extraction successful"
} catch {
    Write-Host "Extraction failed"
    exit 1
}

# Add to PATH (current user)
$easyGitPath = "$destination\easyGit.exe"
if (-not (Test-Path $easyGitPath)) {
    Write-Host "Executable not found in extracted files"
    exit 1
}

$envPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (-not ($envPath -split ";" | Where-Object { $_ -eq $destination })) {
    [Environment]::SetEnvironmentVariable("Path", "$envPath;$destination", "User")
    Write-Host "easyGit has been added to PATH. Restart terminal to apply changes."
} else {
    Write-Host "easyGit is already in PATH"
}

# Initialize config
try {
    & "$easyGitPath" init
    Write-Host "Configuration initialized"
} catch {
    Write-Host "Failed to initialize config. You can manually run: easyGit init"
}

# Clean up
Remove-Item $file -Force
Write-Host "Installation completed successfully"
