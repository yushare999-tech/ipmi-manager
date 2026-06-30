# 📋 IPMI Manager - 작업 이력 (History)

> 본 문서는 모든 작업의 최상위 네비게이션 역할을 합니다.  
> 각 작업의 상세 설계는 `docs/` 하위 문서를 참조하세요.

---

## 📁 문서 구조

```
docs/
├── HISTORY.md              ← 현재 파일 (네비게이션 허브)
├── features/
│   ├── auto-login.md       ← IPMI 자동 로그인 기능 설계서
│   ├── detailed-logging.md ← 외부 프로세스 실행 상세 로깅 설계서
│   └── ping-check.md       ← 장비 IP 핑 체크 기능 설계서
└── setup/
    └── git-setup.md        ← Git 초기 설정 이력
```

---

## 🗂️ 작업 이력

### [v1.5.0] 2026-06-30 — 장비 목록 실시간 IP 핑 체크 기능 구현 및 iKVM.jar 인자 보완
- **작업자**: 삼식이 (AI) + kuri
- **내용**:
  - 등록된 IPMI 장비의 네트워크 도달 가능 여부(Online/Offline)를 실시간으로 확인하기 위한 비동기 핑 체크 모듈 도입
  - 백엔드(`main.js`): Windows 핑 유틸리티를 활용한 비동기 `device:ping` IPC 핸들러 구현 (타임아웃 800ms 최적화)
  - 프론트엔드(`renderer/app.js`): 장비 목록 렌더링 시 자동으로 핑 체크를 병렬 수행하고, 상태에 따라 뱃지 색상이 바뀌도록 UI 설계
  - **백그라운드 동적 업데이트 고도화**: 장비 수가 많을 때 OS 리소스(ping.exe 프로세스) 과점을 방지하기 위해 **동시 실행 수를 최대 3개로 제한하는 동시성 제어(Limit Concurrency)** 로직 적용
  - **JNLP 실행 전 즉각 선행 검증**: 사용자가 `☕ JNLP` 버튼을 클릭하면 가장 먼저 즉시 핑 체크를 실시하여, **온라인인 경우에만 뷰어를 기동하고 오프라인인 경우 실행을 자동 취소(차단) 및 경고 창 출력** 기능 추가
  - 수동 갱신을 위해 장비 목록 헤더에 "🔄 핑 재체크" 버튼 연동 및 바인딩
  - `iKVM.jar` (Supermicro KVM) 기동 시 0 나누기 예외(`EXCEPTION_INT_DIVIDE_BY_ZERO`) 크래시 우회를 위해 마지막 실행 인자 구성을 `2 0` (암호화 사용, 가상 미디어 비활성화) 표준 8자리 포맷으로 변경 적용하여 앱 내부 기동 안정성 확보
  - 전역 버전을 `1.5.0`으로 공식 변경
- **상태**: ✅ 완료
- **관련 문서**: [장비 IP 핑 체크 기능 설계서](./features/ping-check.md)

---

### [v1.4.3] 2026-06-30 — 외부 프로세스 실행 상세 로깅 및 실제 커맨드 기록 구현
- **작업자**: 삼식이 (AI) + kuri
- **내용**:
  - 외부 창 기동 시 즉시 사라지는 이슈 분석을 보완하기 위해 상세 로깅 모듈 추가
  - `iKVM.jar` (Supermicro), `javaws` (JNLP), `IPMIView.exe` 구동 시 전체 실행 명령 경로 및 인수를 이스케이프된 하나의 문자열로 결합하여 `ikvm_error.log` 파일에 타임스탬프와 함께 로깅
  - `detached: true` 프로세스 기동 시 `stdio`를 파일 디스크립터 (`outFd`, `errFd`)로 리다이렉션하여 자식 프로세스의 에러 출력(Exception 등)이 부모 종료 여부에 무관하게 로그에 저장되도록 개선
  - 프로세스의 `PID` 기록 및 비동기 `exit`, `error` 이벤트를 감지하여 종료 코드(Exit Code) 및 에러 상태 기록
  - `java-manager.js` 내부의 `launchJnlp` 실행 함수에도 동일한 로깅 구조 적용
  - 앱 전역 버전을 `1.4.3`으로 업데이트
- **상태**: ✅ 완료
- **관련 문서**: [외부 프로세스 실행 상세 로깅 설계서](./features/detailed-logging.md)

---

### [v1.4.2] 2026-06-30 — Supermicro iKVM.jar 실행 인자 수정 및 글로벌 버전 관리 시스템 적용
- **작업자**: 삼식이 (AI) + kuri
- **내용**:
  - `iKVM.jar` 기동 시 필수 인자 누락에 따른 `ArrayIndexOutOfBoundsException: 4` 구동 에러 해결 (Aten 표준 6개 인자 규격인 IP, USER, PASS, KVM_PORT, VM_PORT, WEB_PORT에 맞춰 데이터 주입)
  - `iKVM.jar` 내부 바이트코드 디코딩 분석을 바탕으로 파라미터 매핑 순서(IP, USER, PASS, HOSTNAME, KVM_PORT, VM_PORT, WEB_PORT)를 재배치하여 KVM 접속 포트가 VM 포트로 엇갈려 소켓 연결에 실패(`connect failed`)하던 현상 전면 수정
  - Electron 부모 프로세스 및 실행용 쉘 라이프사이클 종료와 연동되어 자바 창이 즉시 닫히는 현상 방지를 위해 `shell: false` 제거 및 `child.unref()`를 통한 완전한 독립 프로세스 분리 적용
  - `shell: false`로 구동 시 OS API(CreateProcess) 규격에 맞춰 실행 파일 및 인자 목록에서 불필요하게 중복 감싸고 있던 수동 큰따옴표(`"`)를 완전히 제거하여 `ENOENT` 경로 예외 완벽 해결
  - Windows 디스플레이 DPI 배율(DPI Scaling) 환경에서 `SharedLibrary64.dll` 내부 graphic 연산 중 `EXCEPTION_INT_DIVIDE_BY_ZERO` (0xc0000094) 오류로 자바 프로세스가 즉시 크래시되던 문제를 해결하기 위해, JVM 실행 아규먼트에 `-Dsun.java2d.uiScale=1.0` 및 DirectDraw/Direct3D 하드웨어 가속 비활성화 옵션 주입
  - JVM 외부의 C++ 네이티브 수준에서 호출되는 Windows DPI 측정 API 우회를 위해, OS 레벨 호환성 환경 변수 `__COMPAT_LAYER=DPIUNAWARE`를 프로세스 생성 시 강제 주입하여 모니터 배율에 따른 화면 리사이징 크래시 원천 차단
  - 자바 KVM 내부적인 예외/크래시 발생 원인 추적을 위해 실행 로그를 `ikvm_error.log` 파일로 리다이렉트하는 디버그 모드 추가
  - 애플리케이션 통합 버전 관리 체계 구축: `package.json` 버전을 `1.4.2`로 공식 갱신
  - 프로그램 구동 시 메인 프로세스 터미널 로그 최상단에 버전 명시 출력
  - 프론트엔드 UI 화면(사이드바 로고 영역)에 dynamic version 뱃지 추가 및 IPC 연동을 통한 동적 바인딩 적용
- **상태**: ✅ 완료

---

### [v1.4.1] 2026-06-30 — Supermicro iKVM.jar 단독 구동 방식 개편
- **작업자**: 삼식이 (AI) + kuri
- **내용**:
  - 불안정한 웹 API 경로 및 SSL 리다이렉트 문제를 우회하기 위해 웹 다운로드 방식 전면 폐기
  - 내장된 IPMIView 2.14.0의 `iKVM.jar`를 로컬 자바를 통해 백그라운드에서 직접 구동하도록 전환
  - 명령어: `java -jar iKVM.jar <IP> <USER> <PASSWORD>`
  - KVM 기동 시 DLL (`iKVM64.dll`) 로딩 경로가 어긋나서 발생하는 `UnsatisfiedLinkError` 방지를 위해 작업 경로(`cwd`)를 JAR 폴더로 강제 매핑하여 호출하도록 보완
- **상태**: ✅ 완료

---

### [v1.4.0] 2026-06-30 — Supermicro JNLP 백그라운드 우회 실행 구현
- **작업자**: 삼식이 (AI) + kuri
- **내용**:
  - Supermicro X9 세대 등 웹 브라우저 SSL 거부 및 리다이렉트가 발생하는 환경을 위한 JNLP 실행 개선
  - 백엔드(Node.js)에서 직접 `cgi/login.cgi`에 로그인하여 세션 쿠키를 획득하는 백그라운드 인증 로직 구축
  - 인증된 세션 쿠키를 기반으로 `cgi/launch_win.cgi`에 접근하여 `launch.jnlp`를 임시 폴더에 자동 다운로드
  - 다운로드 시 HTTP/HTTPS 동적 탐색 지원 및 HTTPS 장애 시 HTTP 자동 폴백(상호 백업) 장치 마련 (보안 무시 옵션 기본 적용)
  - JNLP 다운로드 실패(404 에러 등)에 대응하기 위해 세 가지 유력 다운로드 경로(`/cgi/launch_win.cgi`, `/cgi/viewer.jnlp` 등)를 순차적으로 자동 시도하는 다운로드 체인 구현 추가
  - 다운로드된 로컬 JNLP 파일을 로컬 Java 보안 정책 완화 하에 `javaws.exe`로 즉각 구동
- **상태**: ✅ 완료

---

### [v1.3.2] 2026-06-30 — SSL 접속 실패 시 HTTP 자동 폴백 구현 및 Supermicro X9 검증
- **작업자**: 삼식이 (AI) + kuri
- **내용**:
  - Supermicro X9 세대 (`10.96.19.35`) 장비의 Java KVM 다운로드 및 실행 흐름 성공적 검증 완료
  - 구형 SSL/TLS 규격을 사용하는 레거시 장비 접속 시 `ERR_SSL_VERSION_OR_CIPHER_MISMATCH` 에러로 인한 Electron 블랙아웃 현상 식별
  - `main.js`의 `openIpmiWithAutoLogin` 및 `openKvmWithAutoLogin`에 `did-fail-load` 및 `did-fail-provisional-load` 핸들러 추가
  - SSL 연결 실패(초기 핸드셰이크 실패 포함) 감지 시 자동으로 `https` -> `http`로 치환하여 재접속하는 스마트 폴백(Fallback) 로직 적용 (isMainFrame 및 getURL() 적용)
- **상태**: ✅ 완료

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

### [v1.1.0] 2026-06-25 — IPMI 자동 로그인 기능 구현
- **작업자**: 삼식이 (AI) + kuri
- **내용**:
  - IPMI 페이지 접속 시 저장된 계정으로 자동 로그인 처리
  - HTML5 KVM / JNLP 실행 전 자동 로그인 선행 처리
  - 벤더별(Dell iDRAC / HP iLO / SuperMicro) 로그인 폼 자동화
- **상태**: ✅ 완료
- **관련 문서**: [자동 로그인 기능 설계서](./features/auto-login.md)

---

### [v1.0.1] 2026-06-25 — Git 원격 저장소 설정 및 동기화 유틸리티
- **작업자**: 삼식이 (AI) + kuri
- **내용**:
  - GitHub 저장소 신규 생성: `yushare999-tech/ipmi-manager`
  - `master` → `main` 브랜치 통일
  - PAT 기반 Remote origin 연결 및 초기 Push
  - 협업자 `koolkuri79` Write 권한 초대
  - `git_sync.sh` / `git_sync.ps1` 동기화 유틸리티 추가
- **관련 문서**: [Git 설정 이력](./setup/git-setup.md)

---

### [v1.0.0] 2026-06-25 — 초기 구조 및 핵심 기능 구현
- **작업자**: 삼식이 (AI) + kuri
- **내용**:
  - Electron 기반 IPMI Manager 초기 구조 설계
  - 장비 등록/수정/삭제 기능
  - HTML5 KVM, JNLP, IPMI 페이지 접속 기능
  - Java 탐지 및 레거시 설정 자동화
  - SSL 우회 처리 (레거시 IPMI 장비 대응)
- **상태**: ✅ 완료

---

## 🔗 GitHub 저장소
[https://github.com/yushare999-tech/ipmi-manager](https://github.com/yushare999-tech/ipmi-manager)

---
*Last Updated: 2026-06-30*
*Managed by [사무실-삼식이]*
