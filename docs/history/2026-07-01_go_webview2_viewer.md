# 📄 작업 히스토리: Go 기반 초경량 크로미움 웹 뷰어 전환 및 안전 빌드 파이프라인

- **작성일자**: 2026-07-01
- **작성자**: 삼식이 (Lead Developer)
- **대상 모듈**: `go-daemon` (설정 및 스마트 라우터), `go-daemon/viewer` (초경량 크로미움 웹 뷰어)

---

## 1. 작업 배경
기존 Node.js(Express + Electron) 기반의 뷰어는 배포 시 파일 크기가 수백 MB에 달하고 의존성 관리가 무거웠습니다. 이에 따라 **순수 Go 언어만 사용하여 로컬 웹 GUI 설정 데몬과 초경량 뷰어 프로세스를 이원화하는 경량화 아키텍처**를 수립하였습니다.

---

## 2. 주요 구현 내역

### A. 초경량 Go 크로미움 웹 뷰어 개발 (`go-daemon/viewer/main.go`)
- **기술 스택**: 순수 Go WebView2 바인딩 라이브러리(`github.com/jchv/go-webview2`)를 사용하여 CGO 없이 컴파일 가능한 **5MB 수준의 단독 실행 파일(`ipmi-viewer.exe`)**을 구축했습니다.
- **SSL/TLS 사설 인증서 보안 우회**: 레거시 IPMI 웹 콘솔 접속 시 발생하는 보안 경고를 우회하기 위해 크로미움 아규먼트(`--ignore-certificate-errors --ignore-ssl-errors`)를 환경 변수 수준에서 주입했습니다.
- **자바스크립트 기반 계정 자동완성 (Auto-Fill & Auto-Login)**:
  - WebView2 창이 뜰 때 벤더(Dell iDRAC 7/8/9, Supermicro, HP iLO 3/4/5)의 DOM 양식을 분석하여 ID/PW를 자동으로 입력하고 Submit 이벤트를 강제 실행하는 스크립트를 주입(Injection)했습니다.

### B. 안전 빌드 및 기동 파이프라인 수립 (`go-daemon/build.ps1`)
- **프로세스 파일 잠금(File Lock) 해제**: Windows OS에서 구동 중인 `.exe` 파일은 수정이 제한되므로, 빌드 시작 시 현재 실행 중인 `ipmi-daemon` 프로세스를 감지하여 자동으로 종료(`Stop-Process`)하고 잠금을 안전하게 해제합니다.
- **버전 자동 증감 (Auto-Versioning)**:
  - 빌드 시 `version.go`와 `viewer/main.go` 내부의 `const Version` 문자열을 정규식으로 추적하여 패치 버전(Z)을 **자동으로 1 증가(Count-up)** 시킨 후 저장합니다.
- **원자적 컴파일 및 기동**: 뷰어와 데몬의 컴파일이 **모두 무결하게 성공(Exit Code 0)했을 때만** 최신 버전 데몬(`ipmi-daemon.exe -run`)을 백그라운드 프로세스로 구동시킵니다. 컴파일 실패 시 구동 단계를 전면 차단합니다.

### C. 버전 노출 웹 GUI 연동 (`go-daemon/settings_ui.go`)
- 웹 GUI 초기 기동 시 `/api/status` API를 비동기 호출하여 획득한 데몬의 현재 버전을 상단 헤더 로고 타이틀 옆에 세련된 뱃지 형식으로 실시간 출력합니다. (예: `IPMI Manager v1.5.4`)

---

## 3. 실행 및 관리 방법
1. **신규 빌드 및 재기동**: 
   ```powershell
   cd go-daemon
   .\build.ps1
   ```
   이 명령 한 줄로 `기존 프로세스 종료 ➡️ 버전 1 증가 ➡️ 뷰어 빌드 ➡️ 데몬 빌드 ➡️ 데몬 자동 기동`이 수행됩니다.
2. **웹 GUI 접속**: `http://127.0.0.1:4447`을 브라우저에 입력하여 설정 및 라우팅 매트릭스를 제어합니다.
