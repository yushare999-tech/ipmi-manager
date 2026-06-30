# 🔐 IPMI 자동 로그인 기능 설계서

> **문서 경로**: `docs/features/auto-login.md`  
> **상태**: 🔄 구현 진행 중  
> **작성일**: 2026-06-25  
> **작성자**: 삼식이 (AI)

---

## 1. 배경 및 목적

현재 IPMI Manager에서 IPMI 페이지, HTML5 KVM, JNLP를 실행하면 매번 브라우저/뷰어에서
수동으로 로그인해야 하는 불편함이 있음.

장비 등록 시 저장된 username/password를 활용하여:
- **IPMI 페이지** 접속 시 자동 로그인 처리
- **HTML5 KVM** 실행 전 자동 로그인 후 KVM 진입
- **JNLP** 실행 전 자동 로그인 선행 처리

---

## 2. 기술 접근 방식

### 핵심 전략: Electron BrowserWindow + webContents 스크립트 주입

Electron의 `BrowserWindow`로 IPMI 웹 페이지를 열고,  
`did-finish-load` / `did-navigate` 이벤트 시점에  
`webContents.executeJavaScript()`로 로그인 폼을 자동 채워넣고 Submit.

```
[사용자 클릭]
    ↓
[main.js: openIpmiWithAutoLogin(device)]
    ↓
[BrowserWindow 생성 → IPMI URL 로드]
    ↓
[did-finish-load 이벤트]
    ↓
[벤더 판별 → 로그인 스크립트 주입]
    ↓
[폼 자동 입력 + Submit]
    ↓
[로그인 완료 → 대시보드/KVM 진입]
```

---

## 3. 벤더별 로그인 폼 분석

| 벤더 | 로그인 URL | username 셀렉터 | password 셀렉터 | submit 방식 |
|------|-----------|----------------|----------------|-------------|
| **Dell iDRAC** | `/login.html` or `/` | `#user` | `#password` | `form.submit()` |
| **HP iLO** | `/ui/` | `#username` | `#password` | 버튼 클릭 |
| **SuperMicro** | `/cgi/login.cgi` | `name="name"` | `name="pwd"` | `form.submit()` |
| **Generic** | `/` | 공통 탐색 시도 | 공통 탐색 시도 | 동적 판별 |

---

## 4. 구현 계획 (단계별)

### Step 1 — IPMI 페이지 자동 로그인 ✅ 1순위
- `main.js`: `openIpmiWithAutoLogin(device)` 함수 신규 추가
- 벤더별 로그인 스크립트 주입 로직 (`vendors/auto-login-scripts.js`)
- `preload.js`: `openIpmiAutoLogin(device)` IPC 노출
- `renderer/app.js`: `openIpmi()` → `openIpmiAutoLogin()` 로 교체

### Step 2 — HTML5 KVM 자동 로그인 선행 처리 ✅ 2순위
- 기존 `openKvmWindow(device)` 함수에 자동 로그인 선행 로직 추가
- 로그인 완료 감지 후 KVM URL로 리다이렉트

### Step 3 — JNLP 실행 전 자동 로그인 ✅ 3순위
- JNLP 실행 전 숨겨진 BrowserWindow로 로그인 완료 후 JNLP 실행
- 세션 쿠키 공유를 통해 Java Viewer가 인증된 세션 활용

---

## 5. 변경 파일 목록

| 파일 | 변경 내용 |
|------|----------|
| `main.js` | `openIpmiWithAutoLogin()`, `injectLoginScript()` 함수 추가, IPC 핸들러 추가 |
| `preload.js` | `openIpmiAutoLogin(device)` 노출 |
| `renderer/app.js` | `openIpmi()` 함수 교체, `connectHtml5()` / `connectJnlp()` 로그인 선행 처리 |
| `vendors/auto-login-scripts.js` | 벤더별 로그인 JS 스크립트 모음 (신규) |

---

## 6. 보안 고려사항

- 비밀번호는 기존처럼 로컬 `ipmi-config.json`에 저장 (Electron 앱이므로 허용 범위)
- 향후 개선: `electron-keytar` 등 OS Keychain 연동 검토

---

## 7. 제약사항

- 일부 벤더(최신 iDRAC9+)는 로그인 후 JS 렌더링 지연이 있어 `setTimeout` 또는 `waitForSelector` 방식 필요
- 2FA(이중 인증)가 활성화된 장비는 자동 로그인 불가 (수동 처리)
- JNLP는 세션 쿠키 공유가 보장되지 않을 수 있어 별도 테스트 필요

---
*Managed by [사무실-삼식이] | Last Updated: 2026-06-25*
