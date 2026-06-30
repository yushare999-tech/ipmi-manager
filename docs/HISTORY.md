# 📋 IPMI Manager - 작업 이력 (History)

> 본 문서는 모든 작업의 최상위 네비게이션 역할을 합니다.  
> 각 작업의 상세 설계는 `docs/` 하위 문서를 참조하세요.

---

## 📁 문서 구조

```
docs/
├── HISTORY.md              ← 현재 파일 (네비게이션 허브)
├── features/
│   └── auto-login.md       ← IPMI 자동 로그인 기능 설계서
└── setup/
    └── git-setup.md        ← Git 초기 설정 이력
```

---

## 🗂️ 작업 이력

### [v1.0.0] 2026-06-25 — 초기 구조 및 핵심 기능 구현
- **작업자**: 삼식이 (AI) + kuri
- **내용**:
  - Electron 기반 IPMI Manager 초기 구조 설계
  - 장비 등록/수정/삭제 기능
  - HTML5 KVM, JNLP, IPMI 페이지 접속 기능
  - Java 탐지 및 레거시 설정 자동화
  - SSL 우회 처리 (레거시 IPMI 장비 대응)
- **커밋**: `22204e4` feat: IPMI Manager v1.0.0 초기 구조 및 핵심 기능 구현

---

### [v1.0.1] 2026-06-25 — Git 원격 저장소 설정 및 동기화 유틸리티
- **작업자**: 삼식이 (AI) + kuri
- **내용**:
  - GitHub 저장소 신규 생성: `yushare999-tech/ipmi-manager`
  - `master` → `main` 브랜치 통일
  - PAT 기반 Remote origin 연결 및 초기 Push
  - 협업자 `koolkuri79` Write 권한 초대
  - `git_sync.sh` / `git_sync.ps1` 동기화 유틸리티 추가
- **커밋**: `841850b` chore: add git_sync.sh & git_sync.ps1 동기화 유틸리티 추가
- **관련 문서**: [Git 설정 이력](./setup/git-setup.md)

---

### [v1.1.0] 2026-06-25 — IPMI 자동 로그인 기능 구현
- **작업자**: 삼식이 (AI) + kuri
- **내용**:
  - IPMI 페이지 접속 시 저장된 계정으로 자동 로그인 처리
  - HTML5 KVM / JNLP 실행 전 자동 로그인 선행 처리
  - 벤더별(Dell iDRAC / HP iLO / SuperMicro) 로그인 폼 자동화
- **상태**: ✅ 완료
- **관련 문서**: [자동 로그인 기능 설계서](./features/auto-login.md)

---

### [v1.3.1] 2026-06-30 — Dell R630 iDRAC 8 로그인 무한 루프 버그 수정
- **작업자**: 삼식이 (AI) + kuri
- **내용**:
  - Dell R630 (iDRAC 8) 자동 로그인 중 로그인-로그아웃 무한 루프가 발생하는 현상 수정
  - REST API 인증 토큰(`ST1`, `ST2`)을 갖고 대시보드에 진입한 경우, `login-preload.js` 내부에서 새로고침(`window.location.reload()`)을 실행하여 세션이 만료되는 것이 근본 원인임
  - URL에 `st1` 또는 `st2` 매개변수가 포함된 경우 새로고침을 생략하도록 우회 조건문 적용
  - Java 버전과는 무관한 문제임을 분석 및 조치 완료
- **상태**: ✅ 완료

---

### [v1.3.0] 2026-06-30 — 비밀번호 보기 토글 및 IPMI iframe 자동 로그인 개선
- **작업자**: 삼식이 (AI) + kuri
- **내용**:
  - 메인 화면의 장비 카드 내부에 계정(ID/PW) 정보 노출 및 비밀번호 보기/숨기기 토글(👁️/🙈) 기능 추가
  - `iframe` 및 `frame` 구조를 가지는 IPMI 페이지에서 자동완성이 실패하는 문제 해결을 위해 재귀적 프레임 탐색(`querySelectorAllAll`) 도입
  - Electron 메인 프로세스에서 `did-frame-finish-load` 이벤트 감지를 통해 지연 로드되는 프레임에도 로그인 스크립트가 누락 없이 주입되도록 보완
- **상태**: ✅ 완료
- **관련 문서**: [비밀번호 보기 및 iframe 자동 로그인 개선](./features/password-visibility-and-autofill-fix.md)

---

### [v1.2.0] 2026-06-29 — JNLP 실행 차단 해결 및 보안 정책 우회 구현
- **작업자**: 삼식이 (AI) + kuri
- **내용**:
  - JNLP 실행 시 대상 장비의 IP 및 포트 조합(80, 443)을 Java 예외 목록(`exception.sites`)에 자동 추가
  - JRE 내부의 `java.security` 차단 정책(TLS 1.0/1.1 및 MD5/SHA1 차단) 해제를 위한 UAC 권한 파워셸 연동 기능 구현
  - UI (Java 설정 탭)에 "Java 보안 차단 완전 해제" 패치 버튼 적용
- **상태**: ✅ 완료

---

## 🔗 GitHub 저장소
[https://github.com/yushare999-tech/ipmi-manager](https://github.com/yushare999-tech/ipmi-manager)

---
*Managed by [사무실-삼식이] | Last Updated: 2026-06-30*
