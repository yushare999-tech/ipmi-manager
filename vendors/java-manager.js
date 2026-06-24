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
const { Registry } = require('winreg-alt') // fallback: 레지스트리는 직접 파싱
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
  const exceptionDir = path.join(os.homedir(), 'AppData', 'Roaming', 'Sun', 'Java', 'Deployment', 'security');
  const exceptionFile = path.join(exceptionDir, 'exception.sites');

  if (!fs.existsSync(exceptionDir)) {
    fs.mkdirSync(exceptionDir, { recursive: true });
  }

  let existing = [];
  if (fs.existsSync(exceptionFile)) {
    existing = fs.readFileSync(exceptionFile, 'utf8').split('\n').map(l => l.trim()).filter(Boolean);
  }

  // 정규화 (http/https 양쪽 추가)
  const urlsToAdd = [];
  if (!siteUrl.startsWith('http')) {
    urlsToAdd.push(`http://${siteUrl}`);
    urlsToAdd.push(`https://${siteUrl}`);
  } else {
    urlsToAdd.push(siteUrl);
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

  return spawn(`"${javawsPath}"`, [jnlpUrl], {
    shell: true,
    detached: true,
    stdio: 'ignore',
  });
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
  launchJnlp,
  getLegacyJavaDownloadInfo,
  JAVA_COMPAT_INFO,
};
