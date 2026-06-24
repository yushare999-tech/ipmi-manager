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
├── main.js              # Electron 메인 프로세스 (SSL 우회, IPC 핸들러)
├── preload.js           # 렌더러 ↔ 메인 보안 브릿지
├── renderer/
│   ├── index.html       # 메인 UI
│   ├── style.css        # 다크 테마 스타일
│   └── app.js           # 렌더러 로직
├── vendors/
│   └── java-manager.js  # Java 탐지/설정/실행 모듈
└── api/                 # (추후) API 클라이언트 모듈
```

## 변경 이력

### v1.0.0 (2026-06-25)
- 최초 프로젝트 생성
- Electron 기반 앱 구조 설계
- Dell/HP/SuperMicro/ASUS/ASRock 벤더 지원
- Java 환경 자동 탐지 및 호환성 분석
- Java 보안 예외 사이트 자동 등록
- SSL/TLS 구버전 우회 설정
- API 연동 (장비 정보 가져오기) 기능
- Git 저장소 초기화

## 향후 계획 (Roadmap)

- [ ] 각 벤더별 세부 JNLP URL 자동 생성
- [ ] IPMIView 자동 설치 안내
- [ ] 접속 이력 로그 기능
- [ ] 다중 계정 프로필 관리
- [ ] Windows 인스톨러(.exe) 패키징
- [ ] 자동 업데이트
