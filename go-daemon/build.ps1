# IPMI Manager - Safe Auto-Versioning, Compilation, & Execution Script

$VersionFile = "version.go"
$DaemonProcessName = "ipmi-daemon"

# ==================================================
# 1. Stop running daemon process to release file lock
# ==================================================
$RunningProcess = Get-Process -Name $DaemonProcessName -ErrorAction SilentlyContinue
if ($RunningProcess) {
    Write-Host "[1/5] Stopping active $DaemonProcessName.exe to release file lock..." -ForegroundColor Yellow
    $RunningProcess | Stop-Process -Force
    Start-Sleep -Seconds 1.5  # Give OS time to release the file lock
    Write-Host "Process stopped and file unlocked successfully." -ForegroundColor Green
} else {
    Write-Host "[1/5] No active $DaemonProcessName.exe process detected. Proceeding." -ForegroundColor Cyan
}

if (-not (Test-Path $VersionFile)) {
    Write-Host "[Error] $VersionFile not found." -ForegroundColor Red
    exit 1
}

# ==================================================
# 2. Read and parse current version
# ==================================================
$Content = Get-Content $VersionFile -Raw
$Pattern = 'const Version = "([0-9\.]+)"'

if ($Content -match $Pattern) {
    $CurrentVersion = $Matches[1]
    Write-Host "[2/5] Current version: v$CurrentVersion" -ForegroundColor Cyan
    
    # Increment patch version (e.g. 1.5.3 -> 1.5.4)
    $Parts = $CurrentVersion.Split('.')
    if ($Parts.Count -eq 3) {
        $PatchNum = [int]$Parts[2]
        $NewPatchNum = $PatchNum + 1
        $NewVersion = "$($Parts[0]).$($Parts[1]).$NewPatchNum"
        
        # Update version.go
        $NewContent = $Content -replace $Pattern, ('const Version = "' + $NewVersion + '"')
        Set-Content -Path $VersionFile -Value $NewContent -NoNewline -Encoding UTF8
        Write-Host "Target build version updated to: v$NewVersion" -ForegroundColor Green

        # Synchronize version in viewer/main.go
        $ViewerFile = "viewer/main.go"
        if (Test-Path $ViewerFile) {
            $ViewerContent = Get-Content $ViewerFile -Raw
            $NewViewerContent = $ViewerContent -replace $Pattern, ('const Version = "' + $NewVersion + '"')
            Set-Content -Path $ViewerFile -Value $NewViewerContent -NoNewline -Encoding UTF8
            Write-Host "Viewer version synchronized: v$NewVersion" -ForegroundColor Green
        }
    } else {
        Write-Host "[Warning] Version format is invalid. Resetting to 1.5.0." -ForegroundColor Yellow
        $NewVersion = "1.5.0"
        $NewContent = $Content -replace $Pattern, ('const Version = "1.5.0"')
        Set-Content -Path $VersionFile -Value $NewContent -NoNewline -Encoding UTF8
    }
} else {
    Write-Host "[Error] Could not parse Version constant in version.go" -ForegroundColor Red
    exit 1
}

# ==================================================
# 3. Compile ipmi-viewer.exe (Ultra-lightweight Go viewer)
# ==================================================
Write-Host "`n[3/5] Compiling ipmi-viewer.exe..." -ForegroundColor Yellow
go build -o ipmi-viewer.exe ./viewer
if ($LASTEXITCODE -ne 0) {
    Write-Host "[Error] Failed to build ipmi-viewer.exe. Process aborted." -ForegroundColor Red
    exit 1
}
Write-Host "ipmi-viewer.exe compiled successfully." -ForegroundColor Green

# ==================================================
# 4. Compile ipmi-daemon.exe (Main service daemon)
# ==================================================
Write-Host "[4/5] Compiling ipmi-daemon.exe..." -ForegroundColor Yellow
go build -o ipmi-daemon.exe .
if ($LASTEXITCODE -ne 0) {
    Write-Host "[Error] Failed to build ipmi-daemon.exe. Process aborted." -ForegroundColor Red
    exit 1
}
Write-Host "ipmi-daemon.exe compiled successfully." -ForegroundColor Green

# ==================================================
# 5. Start the newly compiled daemon
# ==================================================
Write-Host "`n[5/5] Starting the newly compiled IPMI Manager Daemon..." -ForegroundColor Yellow
Write-Host "==================================================" -ForegroundColor Green
Write-Host "  IPMI Manager v$NewVersion Build & Run Success!" -ForegroundColor Green
Write-Host "  - Viewer: ipmi-viewer.exe" -ForegroundColor Green
Write-Host "  - Daemon: ipmi-daemon.exe" -ForegroundColor Green
Write-Host "==================================================" -ForegroundColor Green

# Execute the daemon in a new independent background window
Start-Process -FilePath ".\ipmi-daemon.exe" -ArgumentList "-run"
