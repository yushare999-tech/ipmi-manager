# 🔒 비밀번호 보기 기능 및 iframe 내 자동완성 개선

본 문서는 v1.3.0에서 구현된 **비밀번호 보기 토글 기능** 및 **iframe 대응 IPMI 자동 로그인 개선**에 대한 상세 설계 및 변경 사항을 기술합니다.

---

## 1. 비밀번호 일반텍스트 보기 기능

### 배경 및 요구사항
- 사용자가 등록한 IPMI 장비의 계정 정보(특히 비밀번호)를 메인 대시보드 화면에서 바로 확인하고 복사할 수 있는 기능이 필요함.
- 기존 모달 창에서의 토글 기능 외에, 메인 화면의 각 장비 카드 내에서도 비밀번호를 안전하게 확인하고 숨길 수 있어야 함.

### 구현 상세
- **장비 카드 UI 개편 (`renderer/app.js`, `style.css`)**:
  - 장비 카드 내부에 계정이 등록되어 있을 경우(ID 존재 시), `ID: [사용자명] | PW: [••••••••] [👁️]` 형식의 영역을 추가함.
  - 실제 비밀번호 데이터는 `span` 엘리먼트의 `data-password` 속성에 바인딩되어 있으며, 기본 텍스트는 `••••••••`으로 마스킹 처리됨.
- **토글 기능 (`window.toggleCardPw`)**:
  - 눈동자 아이콘(👁️) 클릭 시 `window.toggleCardPw(id)` 함수가 호출되어 `data-password`에 저장된 평문 비밀번호와 마스킹 텍스트를 상호 전환함.
  - 아이콘은 활성화 상태에 따라 👁️(숨김 상태)와 🙈(보임 상태)로 변경됨.

---

## 2. IPMI 페이지 내 iframe 대응 자동 로그인

### 배경 및 요구사항
- 구형 iDRAC, SuperMicro 및 일부 Generic IPMI 웹 페이지는 로그인 화면 전체 또는 로그인 폼 자체가 `iframe`이나 `frame` 내부에 존재함.
- 기존의 단순 `document.querySelector` 방식은 메인 다큐먼트만 탐색하기 때문에 프레임 내부의 `#user`, `#password`, `input[type="password"]` 등의 입력을 수행하지 못해 자동완성이 실패함.

### 구현 상세
- **재귀적 프레임 탐색 헬퍼 (`querySelectorAllAll`)**:
  - `vendors/auto-login-scripts.js` 내의 모든 벤더별 로그인 스크립트에 `querySelectorAllAll(selector, doc)` 공통 헬퍼 함수를 추가함.
  - 이 함수는 현재 `doc`에서 요소를 검색한 후, 문서 내부의 모든 `iframe` 및 `frame`을 찾아 각 프레임의 `contentDocument`를 대상으로 재귀 호출을 수행하여 탐색 결과를 병합함.
  - 이로써 중첩된 프레임 구조 내부의 입력 폼 요소도 완벽히 찾아내어 자동 입력할 수 있게 됨.
- **Electron 이벤트 감지 보완 (`main.js`)**:
  - `BrowserWindow`의 `did-finish-load` 이벤트는 메인 프레임의 로드가 끝났을 때만 호출되므로, 지연 로딩(Lazy loading)되는 `iframe`의 경우에는 스크립트 주입 타이밍을 놓칠 수 있음.
  - 이를 방지하기 위해 메인 프로세스의 `openIpmiWithAutoLogin` 및 `openKvmWithAutoLogin` 핸들러에 `did-frame-finish-load` 이벤트를 추가 등록하여, 하위 프레임이 로드 완료될 때마다 로그인 스크립트가 주입 및 실행되도록 보장함.

---

## 3. 관련 파일 변경 요약

- **`renderer/app.js`**: 장비 카드 템플릿에 계정 정보 출력 추가 및 `window.toggleCardPw` 전역 함수 구현.
- **`renderer/style.css`**: 장비 카드 계정 정보 영역(`.device-account`) 및 비밀번호 토글 버튼의 다크 모드 일관 스타일링 적용.
- **`vendors/auto-login-scripts.js`**: `querySelectorAllAll` 헬퍼 함수를 통한 iframe/frame 내부 요소 재귀 탐색 및 입력 로직 개편.
- **`main.js`**: `did-frame-finish-load` 이벤트를 통한 서브 프레임 로드 시점 스크립트 주입 보완.

---
*Last Updated: 2026-06-30*
*Managed by [사무실-삼식이]*
