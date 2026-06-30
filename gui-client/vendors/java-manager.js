/**
 * IPMI Manager - Java 환경 관리 모듈
 * 작성일: 2026-06-25
 * 기능:
 *   - 시스템 Java 버전 탐지 (레지스트리 + 경로 스캔)
 *   - IPMI KVM 호환 Java 버전 판별
 *   - Java 예외 사이트 목록 자동 등록
 *   - Java Web Start(javaws) 실행
 *   - 구버전 Java 다운로드 안내
 */

const { exec, spawn } = require('child_process');
const path = require('path');
const fs = require('fs');
const os = require('os');

// IPMI KVM과 호환되는 Java 버전 범위
// Java 8u161 이후: TLS 1.0/1.1 기본 비활성화 → 일부 구형 IPMI 접속 불가
// Java 8u45 이전: 서명 없는 Applet 허용
const JAVA_COMPAT_INFO = {
  '1.8.0_45':  { level: 'best',    label: '최고 호환 (서명 없는 Applet 허용)' },
  '1.8.0_101': { level: 'good',    label: '양호 (TLS 1.0 허용)' },
  '1.8.0_161': { level: 'warning', label: '주의 (TLS 1.0 기본 비활성화, 수동 설정 필요)' },
  '1.8.0_261': { level: 'warning', label: '주의 (구형 알고리즘 일부 차단)' },
  '1.8.0_441': { level: 'warning', label: '최신 JRE 8 (보안 강화, 설정 조정 필요)' },
  '11':        { level: 'bad',     label: '비호환 (Java Web Start 제거됨)' },
  '17':        { level: 'bad',     label: '비호환 (Applet API 제거됨)' },
  '21':        { level: 'bad',     label: '비호환 (Applet API 제거됨)' },
};

// 구버전 Java 다운로드 안내 링크 (보관용)
const LEGACY_JAVA_LINKS = [
  {
    version: 'Java 8u161',
    url: 'https://www.oracle.com/java/technologies/javase/javase8-archive-downloads.html',
    note: 'TLS 1.0 허용 마지막 기본 버전. Oracle 계정 필요.',
  },
  {
    version: 'Liberica JRE 8 (추천)',
    url: 'https://bell-sw.com/pages/downloads/#jdk-8-lts',
    note: '무료, 계정 불필요. Java Web Start 포함 버전 선택.',
  },
  {
    version: 'Adoptium (Eclipse Temurin) JRE 8',
    url: 'https://adoptium.net/temurin/releases/?version=8',
    note: '무료 오픈소스 JRE 8. Web Start 미포함.',
  },
];

/**
 * 시스템에 설치된 Java 목록을 탐지합니다.
 * 레지스트리 + Program Files 경로 두 가지 방법으로 탐색.
 * @returns {Promise<Array>} Java 설치 목록
 */
async function detectJavaInstallations() {
  const results = [];

  // 방법 1: 레지스트리에서 탐색
  const regKeys = [
    'HKLM\\SOFTWARE\\JavaSoft\\Java Runtime Environment',
    'HKLM\\SOFTWARE\\JavaSoft\\JDK',
    'HKLM\\SOFTWARE\\WOW6432Node\\JavaSoft\\Java Runtime Environment',
  ];

  for (const regKey of regKeys) {
    try {
      const output = await runCmd(`reg query "${regKey}" /s`);
      const lines = output.split('\n');
      let currentVer = null;
      let javaHome = null;

      for (const line of lines) {
        const trimmed = line.trim();
        // 버전 키 탐지
        const verMatch = trimmed.match(/\\(\d[\d._]+)\s*$/);
        if (verMatch) {
          if (currentVer && javaHome) {
            results.push(buildJavaEntry(currentVer, javaHome));
          }
          currentVer = verMatch[1];
          javaHome = null;
        }
        // JavaHome 값 탐지
        if (trimmed.includes('JavaHome') && trimmed.includes('REG_SZ')) {
          const parts = trimmed.split(/\s{2,}/);
          if (parts.length >= 3) {
            javaHome = parts[parts.length - 1].trim();
          }
        }
      }
      if (currentVer && javaHome) {
        results.push(buildJavaEntry(currentVer, javaHome));
      }
    } catch (_) {
      // 레지스트리 키 없으면 스킵
    }
  }

  // 방법 2: Program Files 경로 직접 스캔
  const searchDirs = [
    'C:\\Program Files\\Java',
    'C:\\Program Files (x86)\\Java',
    'C:\\Program Files\\Eclipse Adoptium',
    'C:\\Program Files\\BellSoft',
  ];

  for (const dir of searchDirs) {
    if (!fs.existsSync(dir)) continue;
    const entries = fs.readdirSync(dir, { withFileTypes: true });
    for (const entry of entries) {
      if (!entry.isDirectory()) continue;
      const javaExe = path.join(dir, entry.name, 'bin', 'java.exe');
      const javawsExe = path.join(dir, entry.name, 'bin', 'javaws.exe');
      if (fs.existsSync(javaExe)) {
        const alreadyFound = results.some(r => r.javaHome === path.join(dir, entry.name));
        if (!alreadyFound) {
          try {
            const versionOut = await runCmd(`"${javaExe}" -version`);
            const verMatch = versionOut.match(/(\d+\.\d+\.\d+_\d+|\d+\.\d+\.\d+|\d+)/);
            const ver = verMatch ? verMatch[1] : entry.name;
            results.push(buildJavaEntry(ver, path.join(dir, entry.name)));
          } catch (_) {}
        }
      }
    }
  }

  // 중복 제거
  const unique = [];
  const seen = new Set();
  for (const r of results) {
    const key = r.javaHome?.toLowerCase();
    if (!seen.has(key)) {
      seen.add(key);
      unique.push(r);
    }
  }

  return unique;
}

/**
 * Java 항목 객체 생성
 */
function buildJavaEntry(version, javaHome) {
  const javawsExe = path.join(javaHome || '', 'bin', 'javaws.exe');
  const hasJavaws = fs.existsSync(javawsExe);
  const compat = getCompatLevel(version);

  return {
    version,
    javaHome,
    javaExe: path.join(javaHome || '', 'bin', 'java.exe'),
    javawsExe: hasJavaws ? javawsExe : null,
    hasJavaws,
    compat,
  };
}

/**
 * 버전에 따른 IPMI 호환성 레벨 반환
 */
function getCompatLevel(version) {
  // Java 9+ → 비호환
  if (/^(9|1[0-9]|[2-9]\d)/.test(version)) {
    return { level: 'bad', label: 'Java Web Start 미지원 (Java 9+)' };
  }

  // Java 8 세부 버전 파싱
  const updateMatch = version.match(/1\.8\.0_(\d+)/);
  if (updateMatch) {
    const update = parseInt(updateMatch[1]);
    if (update <= 45)  return { level: 'best',    label: `Java 8u${update} - 최고 호환 (서명 없는 Applet 허용)` };
    if (update <= 101) return { level: 'good',    label: `Java 8u${update} - 양호 (TLS 1.0 허용)` };
    if (update <= 160) return { level: 'warning', label: `Java 8u${update} - 주의 (TLS 설정 조정 필요)` };
    return { level: 'warning', label: `Java 8u${update} - 주의 (보안 설정 조정 필요)` };
  }

  return { level: 'unknown', label: `알 수 없는 버전: ${version}` };
}

/**
 * Java 보안 예외 사이트에 IP/URL 등록
 * 경로: %APPDATA%\Sun\Java\Deployment\security\exception.sites
 */
function addJavaExceptionSite(siteUrl) {
  const exceptionDir = path.join(os.homedir(), 'AppData', 'LocalLow', 'Sun', 'Java', 'Deployment', 'security');
  const exceptionFile = path.join(exceptionDir, 'exception.sites');

  if (!fs.existsSync(exceptionDir)) {
    fs.mkdirSync(exceptionDir, { recursive: true });
  }

  let existing = [];
  if (fs.existsSync(exceptionFile)) {
    existing = fs.readFileSync(exceptionFile, 'utf8').split('\n').map(l => l.trim()).filter(Boolean);
  }

  // IP/도메인만 추출하여 다양한 조합 생성
  let cleanHost = siteUrl.replace(/^(https?:\/\/)/, '').split('/')[0].split(':')[0];

  const urlsToAdd = [
    `http://${cleanHost}`,
    `https://${cleanHost}`,
    `http://${cleanHost}:80`,
    `https://${cleanHost}:80`,
    `http://${cleanHost}:443`,
    `https://${cleanHost}:443`
  ];

  // 원래 보냈던 URL도 그대로 추가
  if (!urlsToAdd.includes(siteUrl)) {
    if (!siteUrl.startsWith('http')) {
      urlsToAdd.push(`http://${siteUrl}`);
      urlsToAdd.push(`https://${siteUrl}`);
    } else {
      urlsToAdd.push(siteUrl);
    }
  }

  let added = 0;
  for (const url of urlsToAdd) {
    if (!existing.includes(url)) {
      existing.push(url);
      added++;
    }
  }

  fs.writeFileSync(exceptionFile, existing.join('\n') + '\n', 'utf8');
  return { added, total: existing.length, file: exceptionFile };
}

/**
 * JRE 내부의 java.security 파일에서 TLS 1.0/1.1 및 MD5/RC4/3DES 제한을 해제합니다.
 * @param {string} javawsPath - javaws.exe 경로
 * @returns {Promise<Object>} 결과 상태
 */
function patchJavaSecurity(javawsPath) {
  return new Promise((resolve, reject) => {
    if (!javawsPath || !fs.existsSync(javawsPath)) {
      return reject(new Error('유효한 javaws.exe 경로가 아닙니다.'));
    }

    const javaHome = path.dirname(path.dirname(javawsPath));
    const candidates = [
      path.join(javaHome, 'lib', 'security', 'java.security'),
      path.join(javaHome, 'jre', 'lib', 'security', 'java.security')
    ];

    const targetFile = candidates.find(f => fs.existsSync(f));
    if (!targetFile) {
      return reject(new Error('java.security 파일을 찾을 수 없습니다. JRE 경로를 확인하세요.'));
    }

    try {
      let content = fs.readFileSync(targetFile, 'utf8');
      let lines = content.split(/\r?\n/);
      let changed = false;

      const toRemove = ['TLSv1', 'TLSv1.1', 'MD5', 'MD5withRSA', 'RC4', '3DES_EDE_CBC', '3DES', 'DES'];

      for (let i = 0; i < lines.length; i++) {
        const line = lines[i];
        if (line.startsWith('jdk.tls.disabledAlgorithms=') ||
            line.startsWith('jdk.certpath.disabledAlgorithms=') ||
            line.startsWith('jdk.jar.disabledAlgorithms=')) {
          
          const eqIdx = line.indexOf('=');
          const key = line.substring(0, eqIdx);
          const val = line.substring(eqIdx + 1);

          let parts = val.split(',').map(p => p.trim());
          const originalLength = parts.length;

          parts = parts.filter(p => {
            const firstWord = p.split(/\s+/)[0];
            return !toRemove.includes(firstWord);
          });

          if (parts.length !== originalLength) {
            lines[i] = `${key}=${parts.join(', ')}`;
            changed = true;
          }
        }
      }

      if (!changed) {
        return resolve({ success: true, message: '이미 제한이 해제되어 있거나 패치할 항목이 없습니다.', file: targetFile });
      }

      // 1. 패치된 java.security 내용을 담을 임시 파일 작성
      const tempFile = path.join(os.tmpdir(), 'java.security.patched');
      fs.writeFileSync(tempFile, lines.join('\r\n'), 'utf8');

      const escapedTarget = targetFile.replace(/'/g, "''");
      const escapedTemp = tempFile.replace(/'/g, "''");

      // 2. 이스케이프 문제를 방지하기 위한 임시 파워쉘 스크립트(.ps1) 작성
      const tempScript = path.join(os.tmpdir(), 'java_patch.ps1');
      const scriptContent = `
$ErrorActionPreference = 'Stop'
Copy-Item -Path '${escapedTarget}' -Destination '${escapedTarget}.bak' -Force
Copy-Item -Path '${escapedTemp}' -Destination '${escapedTarget}' -Force
`.trim();
      
      fs.writeFileSync(tempScript, scriptContent, 'utf8');

      // 3. 임시 스크립트를 UAC 관리자 권한으로 실행
      const psCommand = `Start-Process powershell -ArgumentList "-NoProfile -ExecutionPolicy Bypass -File \\"\${tempScript}\\"" -Verb RunAs -Wait`;

      exec(`powershell -NoProfile -Command "${psCommand.replace(/"/g, '\\"')}"`, (err) => {
        // 임시 파일들 즉시 정리
        try { fs.unlinkSync(tempScript); } catch (e) {}
        try { fs.unlinkSync(tempFile); } catch (e) {}

        if (err) {
          return reject(new Error(`관리자 권한 실행 중 오류 발생: ${err.message}`));
        }

        // 패치 적용 확인을 위해 파일 다시 읽어 비교
        try {
          const postContent = fs.readFileSync(targetFile, 'utf8');
          if (postContent.includes('TLSv1') && targetFile.includes('disabledAlgorithms')) {
            return reject(new Error('보안 설정이 수정되지 않았습니다. 관리자 권한(UAC) 승인이 필요합니다.'));
          }
          resolve({ success: true, message: 'Java 보안 제한 해제 성공! (기존 파일은 .bak로 백업됨)', file: targetFile });
        } catch (e) {
          reject(new Error(`패치 결과 확인 실패: ${e.message}`));
        }
      });

    } catch (e) {
      reject(new Error(`java.security 파일 처리 중 예외 발생: ${e.message}`));
    }
  });
}


/**
 * Java deployment.properties에 보안 레벨을 낮추는 설정 적용
 * (구형 IPMI JNLP 실행을 위해 필요)
 */
function applyLegacyJavaConfig() {
  const deployDir = path.join(
    os.homedir(), 'AppData', 'LocalLow', 'Sun', 'Java', 'Deployment'
  );
  const deployFile = path.join(deployDir, 'deployment.properties');

  if (!fs.existsSync(deployDir)) {
    fs.mkdirSync(deployDir, { recursive: true });
  }

  let content = '';
  if (fs.existsSync(deployFile)) {
    content = fs.readFileSync(deployFile, 'utf8');
  }

  // 적용할 설정들
  const settings = {
    'deployment.security.level': 'HIGH',                   // VERY_HIGH → HIGH로 낮춤
    'deployment.security.askgrantdialog.show': 'true',     // 허용 다이얼로그 표시
    'deployment.security.mixedcode': 'ENABLE',             // 혼합 코드 허용
    'deployment.security.TLSv1': 'true',                   // TLS 1.0 허용
    'deployment.security.TLSv1.1': 'true',                 // TLS 1.1 허용
    'deployment.security.SSLv2Hello': 'true',              // SSL v2 Hello 허용
    'deployment.security.jsse.hostmismatch': 'true',       // 호스트 이름 불일치 경고 무시 (IP 접속 대응)
    'deployment.security.validation.ocsp': 'false',        // OCSP 검증 비활성화 (오프라인/사설망 속도 향상)
    'deployment.security.validation.crl': 'false',         // CRL 검증 비활성화
    'deployment.security.revocation.check': 'NO_CHECK',    // 인증서 만료/서명 유효성 검사 경고창 소멸 핵심 설정
  };

  for (const [key, val] of Object.entries(settings)) {
    const regex = new RegExp(`^${key.replace('.', '\\.')}=.*$`, 'm');
    const newLine = `${key}=${val}`;
    if (regex.test(content)) {
      content = content.replace(regex, newLine);
    } else {
      content += `\n${newLine}`;
    }
  }

  fs.writeFileSync(deployFile, content, 'utf8');
  return { file: deployFile, settings };
}

/**
 * javaws로 JNLP 실행
 * @param {string} javawsPath - javaws.exe 경로
 * @param {string} jnlpUrl - JNLP URL
 */
function launchJnlp(javawsPath, jnlpUrl) {
  if (!javawsPath || !fs.existsSync(javawsPath)) {
    throw new Error(`javaws.exe를 찾을 수 없습니다: ${javawsPath}`);
  }

  const logPath = path.join(__dirname, '..', 'ikvm_error.log');
  const timestamp = new Date().toISOString();
  const fullCmd = `"${javawsPath}" "${jnlpUrl}"`;

  fs.appendFileSync(logPath, `[${timestamp}] [java-manager:launchJnlp] JNLP 실행 시도\n`, 'utf8');
  fs.appendFileSync(logPath, `[${timestamp}] [java-manager:launchJnlp] 커맨드: ${fullCmd}\n`, 'utf8');

  const outFd = fs.openSync(logPath, 'a');
  const errFd = fs.openSync(logPath, 'a');

  const child = spawn(`"${javawsPath}"`, [jnlpUrl], {
    shell: true,
    detached: true,
    stdio: ['ignore', outFd, errFd],
  });

  fs.appendFileSync(logPath, `[${timestamp}] [java-manager:launchJnlp] 프로세스 생성 완료 (PID: ${child.pid})\n`, 'utf8');

  child.on('error', (err) => {
    const ts = new Date().toISOString();
    fs.appendFileSync(logPath, `[${ts}] [java-manager:launchJnlp] [PID: ${child.pid || 'N/A'}] 에러 발생: ${err.message}\n`, 'utf8');
  });

  child.on('exit', (code, signal) => {
    const ts = new Date().toISOString();
    fs.appendFileSync(logPath, `[${ts}] [java-manager:launchJnlp] [PID: ${child.pid}] 프로세스 종료 (Exit Code: ${code}, Signal: ${signal})\n`, 'utf8');
  });

  child.unref();
  fs.closeSync(outFd);
  fs.closeSync(errFd);

  return child;
}

/**
 * 구버전 Java 설치 안내 정보 반환
 */
function getLegacyJavaDownloadInfo() {
  return LEGACY_JAVA_LINKS;
}

// 내부 유틸: 명령 실행 Promise 래퍼
function runCmd(cmd) {
  return new Promise((resolve, reject) => {
    exec(cmd, { encoding: 'utf8' }, (err, stdout, stderr) => {
      resolve((stdout || '') + (stderr || ''));
    });
  });
}

module.exports = {
  detectJavaInstallations,
  addJavaExceptionSite,
  applyLegacyJavaConfig,
  patchJavaSecurity,
  launchJnlp,
  getLegacyJavaDownloadInfo,
  JAVA_COMPAT_INFO,
};
