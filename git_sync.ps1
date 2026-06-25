# ============================================================
# git_sync.ps1 - [사무실-삼식이] Git 동기화 유틸리티 (PowerShell)
# Project : ipmi-manager
# Repo    : https://github.com/yushare999-tech/ipmi-manager
# Usage   : .\git_sync.ps1 [-Message "커밋 메시지"]
#           -Message 생략 시 자동 타임스탬프 메시지 사용
# ============================================================

param(
    [string]$Message = ""
)

$ErrorActionPreference = "Stop"

$RepoRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$Branch = "main"
$Timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss KST"

# 커밋 메시지 처리
if ($Message -eq "") {
    $CommitMsg = "chore: auto-sync @ $Timestamp"
} else {
    $CommitMsg = $Message
}

Write-Host ""
Write-Host "======================================================" -ForegroundColor Cyan
Write-Host "  [사무실-삼식이] Git Sync 시작" -ForegroundColor Cyan
Write-Host "  Branch : $Branch" -ForegroundColor Cyan
Write-Host "  Message: $CommitMsg" -ForegroundColor Cyan
Write-Host "======================================================" -ForegroundColor Cyan
Write-Host ""

Set-Location $RepoRoot

# 변경사항 확인
$Status = git status --porcelain

if ([string]::IsNullOrWhiteSpace($Status)) {
    Write-Host "✅ 변경사항 없음. 동기화할 내용이 없습니다." -ForegroundColor Green
    Write-Host ""
    Write-Host "🔄 원격 최신 상태 Pull 중..." -ForegroundColor Yellow
    git pull origin $Branch --rebase
    Write-Host "✅ Pull 완료." -ForegroundColor Green
} else {
    Write-Host "📝 변경된 파일 목록:" -ForegroundColor Yellow
    git status --short
    Write-Host ""

    Write-Host "📦 Staging all changes..." -ForegroundColor Yellow
    git add -A

    Write-Host "💾 Committing..." -ForegroundColor Yellow
    git commit -m $CommitMsg

    Write-Host "🚀 Pushing to origin/$Branch..." -ForegroundColor Yellow
    git push origin $Branch

    Write-Host ""
    Write-Host "======================================================" -ForegroundColor Green
    Write-Host "  ✅ Git Sync 완료!" -ForegroundColor Green
    Write-Host "  🔗 https://github.com/yushare999-tech/ipmi-manager" -ForegroundColor Green
    Write-Host "======================================================" -ForegroundColor Green
}

Write-Host ""
