# IPMI Manager

> IPMI/BMC 장비 통합 관리 도구 (Windows Electron 앱)

## 개요

다양한 벤더의 IPMI(Intelligent Platform Management Interface) 장비에 대한
KVM 원격 콘솔 접속을 통합 관리하는 Windows 데스크탑 앱입니다.

## 지원 벤더

| 벤더 | 모델 | 접속 방식 |
|------|------|-----------|
| Dell | iDRAC 6/7/8/9 | HTML5, JNLP |
| HP/HPE | iLO 3/4/5 | HTML5, JNLP |
| SuperMicro | IPMI 2.0 | HTML5, IPMIView |
| ASUS | ASMB 시리즈 | HTML5, JNLP |
| ASRock | IPMI | HTML5, JNLP |

## 주요 기능

- **SSL/TLS 우회**: 구형 IPMI 장비의 자체 서명 인증서 자동 허용
- **Java 환경 관리**: 설치된 Java 자동 탐지, 버전 호환성 분석
- **Java 보안 설정**: TLS 1.0/1.1 허용, 예외 사이트 자동 등록
- **API 연동**: 웹 관리 페이지에서 장비 정보 자동 가져오기
- **다중 접속 방식**: HTML5 내장 브라우저, Java JNLP, IPMIView

## 개발 환경

- Node.js 24.x
- Electron 32.x
- Windows 10/11

## 실행 방법

```bash
npm install
npm start
```

## 프로젝트 구조

```
ipmi-manager/
├── main.js              # Electron 메인 프로세스 (SSL 우회, IPC 핸들러, 자동로그인)
├── preload.js           # 렌더러 ↔ 메인 보안 브릿지
├── login-preload.js     # 자동 로그인 전용 프리로드 스크립트
├── renderer/
│   ├── index.html       # 메인 UI
│   ├── style.css        # 다크 테마 스타일
│   └── app.js           # 렌더러 로직
├── vendors/
│   ├── java-manager.js  # Java 탐지/설정/실행 모듈
│   └── auto-login-scripts.js # 벤더별 자동 로그인 스크립트 빌더
├── go-daemon/           # Go 기반 로컬 서비스 데몬 (자바 뷰어 기동 등 백그라운드 대행)
│   ├── main.go          # 데몬 엔트리 포인트 및 윈도우 서비스 관리
│   ├── java.go          # 로컬 자바 환경 탐지 및 실행 모듈
│   ├── session.go       # 세션 관리 및 백그라운드 웹 서버 핸들러
│   ├── go.mod           # Go 모듈 의존성 정의
│   └── ipmi-daemon.exe  # 빌드된 윈도우 데몬 실행 파일
├── gui-client/          # Electron 기반 GUI 클라이언트
├── api/                 # (추후) API 클라이언트 모듈
└── docs/                # 상세 설계 및 작업 이력 문서
    └── history.md       # [최상위 네비게이션] 작업 히스토리 및 개발 문서 이력
```

## 변경 이력

### v1.5.0 (2026-06-30)
- **장비 IP 실시간 핑 체크 기능 추가**: 장비 목록 UI에서 각 장비의 온라인/오프라인 네트워크 도달 상태를 비동기로 실시간 검사하고 표시하는 기능 구현. 상단에 "🔄 핑 재체크" 수동 갱신 버튼 연동.
- **Supermicro iKVM.jar 실행 인자 보완 (크래시 해결)**: 가상 미디어 포트 및 연결 옵션을 `2 0` 규격에 맞춰 8자리로 교정함으로써, 네이티브 DLL(`SharedLibrary64.dll`)의 0 나누기 예외 크래시 문제를 근본적으로 우회 해결.

### v1.4.3 (2026-06-30)
- **외부 프로세스 실행 상세 로깅 및 실제 커맨드 기록**: 외부 창(`iKVM.jar`, `javaws`, `IPMIView.exe`) 기동 시 실제 실행되는 전체 커맨드 라인을 기록하고, `stdio` 파일 리다이렉션을 통해 프로세스 비정상 종료 예외 및 표준 출력을 `ikvm_error.log` 파일에 누적 기록하여 디버깅 편의성을 극대화함.

### v1.4.2 (2026-06-30)
- **Supermicro iKVM.jar 실행 인자 보완**: 포트 인자 누락으로 인한 기동 오류 해결.
- **글로벌 버전 관리 시스템 적용**: `package.json` 기반으로 Electron 구동 시 터미널 로그 및 프론트엔드 UI(사이드바 로고 영역)에 dynamic version 뱃지 추가 및 자동화 적용.

### v1.4.1 (2026-06-30)
- **Supermicro iKVM.jar 직접 구동 구현**: 구형 웹 서버의 SSL 한계 및 JNLP 파편화 문제를 완전히 배제하고, 내장된 IPMIView 2.14.0의 `iKVM.jar`를 직접 로컬 자바로 호출하여 무인 기동시키는 고성능 전용 우회 기법 도입 (네이티브 DLL 연동 지원).

### v1.4.0 (2026-06-30)
- **Supermicro JNLP 백그라운드 우회 실행 구현**: 웹 브라우저가 SSL 제약이나 리다이렉션 루프로 인해 먹통이 되는 상황에 대비하여, 백엔드(Node.js)에서 직접 로그인 세션을 얻어 JNLP 파일을 로컬에 받아 `javaws.exe`로 즉각 원클릭 실행시키는 백그라운드 자동화 지원.

### v1.3.2 (2026-06-30)
- **구형 장비의 SSL 접속 실패 시 HTTP 자동 폴백(Fallback) 기능 추가**: `ERR_SSL_VERSION_OR_CIPHER_MISMATCH` 등 최신 크롬 엔진이 거부하는 구형 SSL/TLS 규격 탐지 시 자동으로 `http://`로 재접속 처리.
- **Supermicro X9 세대 KVM 실행 검증**: `10.96.19.35` 장비의 로그인 및 가상 콘솔(JNLP) 실행 프로세스 완료.

### v1.3.1 (2026-06-30)
- **Dell R630 (iDRAC 8) 자동 로그인 시 무한 로그인-로그아웃 루프가 발생하는 버그 수정**: REST API 토큰(`ST1`, `ST2`)으로 대시보드 진입 시 자동 새로고침을 생략하도록 `login-preload.js` 개선.

### v1.3.0 (2026-06-30)
- 장비 카드 내부에 계정(ID/PW) 정보 표시 및 비밀번호 토글(👁️/🙈) 기능 추가
- iframe 구조의 IPMI 로그인 페이지 대응을 위한 재귀적 프레임 탐색(`querySelectorAllAll`) 로직 도입
- Electron 메인 프로세스에서 `did-frame-finish-load` 이벤트를 통한 스크립트 주입 보완

### v1.2.0 (2026-06-29)
- JNLP 실행 시 대상 장비 IP를 Java 예외 목록에 자동 등록 및 보안 해제 적용
- `java.security` 차단 해제 UAC 연동 기능 구현

### v1.1.0 (2026-06-25)
- IPMI 자동 로그인 기능 구현 (Dell iDRAC, HP iLO, SuperMicro 대응)

### v1.0.0 (2026-06-25)
- 최초 프로젝트 생성 및 데스크탑 앱 기본 구조 구현
- Java 환경 자동 탐지 및 SSL/TLS 구버전 우회 설정
- API 연동 (장비 목록 연동) 및 기본 CRUD 구현

## 향후 계획 (Roadmap)

- [ ] 각 벤더별 세부 JNLP URL 자동 생성
- [ ] IPMIView 자동 설치 안내
- [ ] 접속 이력 로그 기능
- [ ] 다중 계정 프로필 관리
- [ ] Windows 인스톨러(.exe) 패키징
- [ ] 자동 업데이트
