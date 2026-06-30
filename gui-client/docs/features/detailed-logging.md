# 📝 외부 프로세스 실행 상세 로깅 (Detailed Process Logging)

이 문서는 외부 자바 KVM 및 JNLP 뷰어, IPMIView 등의 외부 프로세스가 구동된 후 즉시 창이 사라지거나 비정상 종료되는 현상을 추적하기 위해 구현된 상세 로깅 기능에 대한 설계 사양서입니다.

---

## 🔍 배경 및 목적

기존 버전에서는 외부 프로세스(`javaws`, `IPMIView`)를 실행할 때 `stdio: 'ignore'` 설정을 사용하였기 때문에, 해당 프로세스가 실행된 이후에 출력하는 에러(예: JVM 크래시, SSL 접속 거부, 클래스 로딩 실패 등)를 전혀 수집할 수 없었습니다.
이를 보완하기 위해 다음과 같은 기능을 제공하도록 로깅 시스템을 개편하였습니다.

1. **실제 실행 명령어 누락 없는 기록**: OS로 전달되는 전체 실행 인자와 커맨드 라인을 사람이 읽기 편하게 조합하여 로그에 남깁니다.
2. **에러 및 표준 출력 수집**: `stdio`를 파일 디스크립터로 리다이렉션하여 자식 프로세스의 출력을 부모 프로세스 라이프사이클에 무관하게 디스크 파일에 기록되도록 합니다.
3. **프로세스 종료 및 장애 감지**: 프로세스가 생성된 직후의 `PID`와 정상/비정상 종료 시점의 `Exit Code`, `Signal`, 실행 실패 시의 `Error` 이벤트를 실시간 로깅합니다.

---

## 📂 로그 파일 위치

- **파일명**: `ikvm_error.log` (프로젝트 루트 디렉토리에 생성)
- **작성 방식**: 기존 로그 파일 끝에 누적 기록 (Append Mode)

---

## 🛠️ 기능 사양

### 1. 로그 기록 헬퍼
`main.js` 및 `vendors/java-manager.js`에서 공통적으로 로그 파일에 타임스탬프와 함께 라인을 추가하는 방식을 사용합니다.

```javascript
function logToErrorFile(message) {
  const timestamp = new Date().toISOString();
  fs.appendFileSync(logPath, `[${timestamp}] ${message}\n`, 'utf8');
}
```

### 2. 표준 입출력 파일 디스크립터(FD) 리다이렉션
Windows 환경에서 `detached: true` 상태로 실행되는 자식 프로세스의 출력 파이프를 유지하면 부모 프로세스 종료 시 자식 프로세스가 동반 종료되거나, 파이프 버퍼 누적으로 인한 데드락이 발생할 수 있습니다.
이를 방지하기 위해 **OS 레벨의 파일 디스크립터를 생성하여 직접 전달**합니다.

```javascript
const outFd = fs.openSync(logPath, 'a');
const errFd = fs.openSync(logPath, 'a');

const child = spawn(command, args, {
  detached: true,
  stdio: ['ignore', outFd, errFd] // stdout, stderr를 파일 디스크립터로 리다이렉션
});

// spawn 호출 완료 후 부모 프로세스 측의 FD는 즉시 해제하여 리소스 누수 방지
fs.closeSync(outFd);
fs.closeSync(errFd);
```

### 3. 프로세스 이벤트 모니터링
부모 프로세스가 실행 중인 동안, 실행된 자식 프로세스의 상태 변화를 감지하기 위해 아래 이벤트를 리스닝하여 로그에 기록합니다.

- `error`: 프로세스를 시작할 수 없거나, 종료할 수 없거나, 프로세스로 메시지를 보낼 수 없는 경우 발생.
- `exit`: 자식 프로세스가 종료되었을 때 발생 (종료 코드 및 시그널 수집).

---

## 📝 로그 포맷 예시

실행 시 `ikvm_error.log`에 작성되는 로그의 전형적인 패턴은 다음과 같습니다.

### JNLP 실행 예시
```text
[2026-06-30T14:10:00.123Z] [JNLP Launch] javaws 실행 (원격 URL) 시도: 10.96.19.35
[2026-06-30T14:10:00.124Z] [JNLP Launch] 커맨드: "C:\Program Files\Java\jre1.8.0_441\bin\javaws.exe" "https://10.96.19.35/viewer.jnlp?EXTPORT=-1&JNLPSTR=AppletRedirection"
[2026-06-30T14:10:00.150Z] [JNLP Launch] 프로세스 생성 완료 (PID: 12432)
[2026-06-30T14:10:05.567Z] [JNLP Launch] [PID: 12432] 프로세스 종료 (Exit Code: 0, Signal: null)
```

### Supermicro `iKVM.jar` 실행 및 내부 크래시 예시
```text
[2026-06-30T14:12:00.456Z] [Supermicro KVM] iKVM.jar 직접 구동 시도: 10.96.19.40 (KVM Port: 5900, Web Port: 443)
[2026-06-30T14:12:00.457Z] [Supermicro KVM] 커맨드: "C:\Program Files\Java\jre1.8.0_441\bin\java.exe" -Dsun.java2d.noddraw=true -Dsun.java2d.d3d=false -Dsun.java2d.uiScale=1.0 -jar "c:\Users\kuri\MyProJ\ipmi-manager\IPMIVIEW\2.14.0\extracted\D_\IPMI20\FILES FOR IPMI VIEW\iKVM.jar" 10.96.19.40 admin admin123 10.96.19.40 5900 623 443
[2026-06-30T14:12:00.458Z] [Supermicro KVM] 작업 디렉토리: c:\Users\kuri\MyProJ\ipmi-manager\IPMIVIEW\2.14.0\extracted\D_\IPMI20\FILES FOR IPMI VIEW
[2026-06-30T14:12:00.500Z] [Supermicro KVM] 프로세스 생성 완료 (PID: 8540)
# (여기에 java.exe가 파일 디스크립터로 직접 출력한 표준 에러 내용이 누적됨)
# java.lang.NullPointerException ...
[2026-06-30T14:12:02.120Z] [Supermicro KVM] [PID: 8540] 프로세스 종료 (Exit Code: 1, Signal: null)
```

---

## ⚠️ 예외 처리 및 주의 사항

- **경로 내 공백 처리**: 명령어 출력 로깅 시 가독성을 위해 각 실행 인자 중 공백이 포함된 경우 큰따옴표(`"`)로 이스케이프하여 출력합니다. 단, `spawn` 함수 호출 시에는 내부적으로 배열 형태로 원본 값을 그대로 넘겨주므로 이중 이스케이프 오류가 발생하지 않습니다.
- **부모 프로세스 생존 여부**: `detached: true` 상태에서 부모 프로세스(Electron 메인 프로세스)가 먼저 종료되는 경우, 자식 프로세스의 `exit` 이벤트는 로그 파일에 추가되지 않을 수 있습니다. 그러나 자식 프로세스가 스스로 파일에 쓰는 표준 출력/에러(stdout/stderr)는 자식이 살아있는 한 끝까지 로그 파일에 누적됩니다.
