# Clother Installer for Windows
# Usage: irm https://raw.githubusercontent.com/jolehuit/clother/main/scripts/install.ps1 | iex

$ErrorActionPreference = "Stop"

$REPO = "jolehuit/clother"
$INSTALL_DIR = "$env:LOCALAPPDATA\clother\bin"

function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] " -ForegroundColor Blue -NoNewline
    Write-Host $Message
}

function Write-Success {
    param([string]$Message)
    Write-Host "[OK] " -ForegroundColor Green -NoNewline
    Write-Host $Message
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARN] " -ForegroundColor Yellow -NoNewline
    Write-Host $Message
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] " -ForegroundColor Red -NoNewline
    Write-Host $Message
    exit 1
}

# Check if Claude CLI is installed
function Test-ClaudeInstalled {
    $claude = Get-Command claude -ErrorAction SilentlyContinue
    if (-not $claude) {
        Write-Warn "Claude CLI not found. Install it first: https://claude.ai/download"
        Write-Warn "Provider launchers will be created but may not work until Claude CLI is installed."
    }
}

# Get latest release version
function Get-LatestVersion {
    $releases = Invoke-RestMethod "https://api.github.com/repos/$REPO/releases/latest"
    return $releases.tag_name -replace '^v', ''
}

# Download file
function Get-File {
    param([string]$Url, [string]$OutFile)
    Write-Info "Downloading $Url..."
    Invoke-WebRequest -Uri $Url -OutFile $OutFile -UseBasicParsing
}

# Main install function
function Install-Clother {
    Write-Host ""
    Write-Host "  _____ _             _              _ "
    Write-Host " / ____| |           | |            | |"
    Write-Host "| |    | |_ ___   ___| | _____ _ __ | |_"
    Write-Host "| |    | __/ _ \ / __| |/ / _ \ '_ \| __|"
    Write-Host "| |____| || (_) | (__|   <  __/ | | | |_"
    Write-Host " \_____|\__\___/ \___|_|\_\___|_| |_|\__|"
    Write-Host ""
    Write-Host "Installing Clother CLI..."
    Write-Host ""

    Test-ClaudeInstalled

    # Get latest version
    $version = Get-LatestVersion
    Write-Info "Latest version: $version"

    # Determine architecture
    $arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
    $platform = "windows_$arch"

    # Download URL
    $filename = "clother_${version}_${platform}.zip"
    $downloadUrl = "https://github.com/$REPO/releases/download/v${version}/${filename}"

    # Create temp directory
    $tempDirName = "clother-install-" + [System.IO.Path]::GetRandomFileName()
    $tempDir = New-Item -ItemType Directory -Path (Join-Path $env:TEMP $tempDirName) -Force

    try {
        # Download
        $zipPath = Join-Path $tempDir.FullName $filename
        Get-File -Url $downloadUrl -OutFile $zipPath

        # Extract
        Write-Info "Extracting..."
        Expand-Archive -Path $zipPath -DestinationPath $tempDir.FullName -Force

        # Find binary
        $binary = Get-ChildItem -Path $tempDir.FullName -Filter "clother.exe" -Recurse | Select-Object -First 1
        if (-not $binary) {
            Write-Error "Could not find clother.exe in downloaded archive"
        }

        # Create install directory
        if (-not (Test-Path $INSTALL_DIR)) {
            New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
        }

        # Copy binary
        $destBinary = Join-Path $INSTALL_DIR "clother.exe"
        Copy-Item -Path $binary.FullName -Destination $destBinary -Force
        Write-Success "Installed clother.exe to $destBinary"

        # Create batch wrappers for providers
        Write-Info "Creating provider launchers..."
        $providers = @(
            "native", "zai", "zai-cn", "minimax", "minimax-cn", "kimi", "moonshot",
            "ve", "deepseek", "mimo", "alibaba", "alibaba-us", "alibaba-cn",
            "ollama", "lmstudio", "llamacpp", "or", "custom"
        )

        foreach ($provider in $providers) {
            $batPath = Join-Path $INSTALL_DIR "clother-${provider}.bat"
            $batContent = "@echo off" + [Environment]::NewLine + "`"$destBinary`" $provider %*"
            Set-Content -Path $batPath -Value $batContent -Encoding ASCII
        }

        Write-Success "Created provider launchers"

        # Check PATH
        $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if ($currentPath -notlike "*$INSTALL_DIR*") {
            Write-Warn "$INSTALL_DIR is not in your PATH"
            Write-Host ""
            Write-Host "To add it to your PATH:"
            Write-Host "  1. Open System Properties > Environment Variables"
            Write-Host "  2. Edit 'Path' under 'User variables'"
            Write-Host "  3. Add: $INSTALL_DIR"
            Write-Host ""
            Write-Host "Or run this command (restart terminal after):"
            Write-Host "  [Environment]::SetEnvironmentVariable('Path', `"`$env:Path;$INSTALL_DIR`", 'User')"
            Write-Host ""
        }

        Write-Success "Installation complete!"
        Write-Host ""
        Write-Host "Usage:"
        Write-Host "  clother native          # Use your Claude Pro/Max/Team subscription"
        Write-Host "  clother zai             # Z.AI (GLM-5)"
        Write-Host "  clother kimi            # Kimi (kimi-k2.5)"
        Write-Host "  clother config          # Configure providers"
        Write-Host ""
        Write-Host "Legacy launcher style also works:"
        Write-Host "  clother-zai             # Same as: clother zai"
        Write-Host ""

    } finally {
        # Cleanup temp directory
        Remove-Item -Path $tempDir.FullName -Recurse -Force -ErrorAction SilentlyContinue
    }
}

# Run installer
Install-Clother
