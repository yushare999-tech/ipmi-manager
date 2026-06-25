# 🗂️ Git 초기 설정 이력

> **문서 경로**: `docs/setup/git-setup.md`  
> **작성일**: 2026-06-25  
> **작성자**: 삼식이 (AI)

---

## 작업 내용

### 1. GitHub 저장소 신규 생성
- **저장소**: `yushare999-tech/ipmi-manager`
- **URL**: https://github.com/yushare999-tech/ipmi-manager
- **공개 여부**: Public
- **API 방식**: GitHub REST API (`POST /user/repos`) via PowerShell

### 2. 브랜치 통일
- `master` → `main` 으로 변경 (`git branch -M main`)
- 룰 §5 기준: `main` 브랜치를 기본으로 통일

### 3. Remote Origin 연결
```bash
git remote add origin https://yushare999-tech:<PAT>@github.com/yushare999-tech/ipmi-manager.git
git push -u origin main
```

### 4. 협업자 추가
- `koolkuri79` — Write(push) 권한 초대
- API: `PUT /repos/yushare999-tech/ipmi-manager/collaborators/koolkuri79`

### 5. 동기화 유틸리티 추가
| 파일 | 용도 |
|------|------|
| `git_sync.sh` | Git Bash / Linux 환경 |
| `git_sync.ps1` | Windows PowerShell (주 사용) |

**사용법**:
```powershell
# PowerShell
.\git_sync.ps1                          # 자동 타임스탬프
.\git_sync.ps1 -Message "feat: 내용"   # 직접 메시지
```

---
*Managed by [사무실-삼식이] | Last Updated: 2026-06-25*
