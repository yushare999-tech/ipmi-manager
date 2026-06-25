#!/bin/bash
# ============================================================
# git_sync.sh - [사무실-삼식이] Git 동기화 유틸리티
# Project : ipmi-manager
# Repo    : https://github.com/yushare999-tech/ipmi-manager
# Usage   : ./git_sync.sh [커밋 메시지]
#           커밋 메시지 생략 시 자동 타임스탬프 메시지 사용
# ============================================================

set -e

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BRANCH="main"
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S KST')

# 커밋 메시지 처리
if [ -n "$1" ]; then
  COMMIT_MSG="$1"
else
  COMMIT_MSG="chore: auto-sync @ ${TIMESTAMP}"
fi

echo ""
echo "======================================================"
echo "  [사무실-삼식이] Git Sync 시작"
echo "  Branch : ${BRANCH}"
echo "  Message: ${COMMIT_MSG}"
echo "======================================================"
echo ""

cd "$REPO_ROOT"

# 변경사항 확인
STATUS=$(git status --porcelain)

if [ -z "$STATUS" ]; then
  echo "✅ 변경사항 없음. 동기화할 내용이 없습니다."
  echo ""
  # 최신 상태 pull만 수행
  echo "🔄 원격 최신 상태 Pull 중..."
  git pull origin "${BRANCH}" --rebase
  echo "✅ Pull 완료."
else
  echo "📝 변경된 파일 목록:"
  git status --short
  echo ""

  # Stage → Commit → Push
  echo "📦 Staging all changes..."
  git add -A

  echo "💾 Committing..."
  git commit -m "${COMMIT_MSG}"

  echo "🚀 Pushing to origin/${BRANCH}..."
  git push origin "${BRANCH}"

  echo ""
  echo "======================================================"
  echo "  ✅ Git Sync 완료!"
  echo "  🔗 https://github.com/yushare999-tech/ipmi-manager"
  echo "======================================================"
fi

echo ""
