/**
 * IPMI Manager - Electron 메인 프로세스
 * 작성일: 2026-06-25
 * 변경이력:
 *   - 2026-06-25: 최초 작성 (SSL 우회, KVM 창 열기, Java IPC 통합)
 *   - 2026-06-25: IPMI 자동 로그인 기능 추가 (벤더별 폼 자동 주입)
 */

const { app, BrowserWindow, ipcMain, shell, dialog } = require('electron');
const path = require('path');
const { spawn, exec } = require('child_process');
const fs = require('fs');
const javaManager    = require('./vendors/java-manager');
const autoLoginScripts = require('./vendors/auto-login-scripts');

// ─── SSL / 보안 우회 설정 (레거시 IPMI 장비 대응) ─────────────────
app.commandLine.appendSwitch('ignore-certificate-errors');
app.commandLine.appendSwitch('ignore-ssl-errors');
app.commandLine.appendSwitch('allow-insecure-localhost');
app.commandLine.appendSwitch('disable-web-security');
// 구형 TLS 프로토콜 허용
app.commandLine.appendSwitch('ssl-version-min', 'tls1');
app.commandLine.appendSwitch('cipher-suite-blacklist', '');

// 설정 파일 경로
const CONFIG_PATH = path.join(app.getPath('userData'), 'ipmi-config.json');

let mainWindow;
let kvmWindows = {};

// ─── 메인 창 생성 ────────────────────────────────────────────────
function createMainWindow() {
  mainWindow = new BrowserWindow({
    width: 1280,
    height: 800,
    minWidth: 960,
    minHeight: 600,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      nodeIntegration: false,
      contextIsolation: true,
      webSecurity: false,
    },
    icon: path.join(__dirname, 'assets', 'icon.ico'),
    show: false,
    backgroundColor: '#0f1117',
    titleBarStyle: 'default',
  });

  mainWindow.loadFile(path.join(__dirname, 'renderer', 'index.html'));
  mainWindow.once('ready-to-show', () => mainWindow.show());
  mainWindow.on('closed', () => { mainWindow = null; });
}

// ─── KVM 창 열기 (HTML5 내장 브라우저) ───────────────────────────
function openKvmWindow(device) {
  const winId = device.id;
  if (kvmWindows[winId]) { kvmWindows[winId].focus(); return; }

  const kvmWin = new BrowserWindow({
    width: 1024,
    height: 768,
    title: `KVM - ${device.name} (${device.ipmi_ip})`,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: false,
      webSecurity: false,
      allowRunningInsecureContent: true,
    },
    show: false,
  });

  // 인증서 오류 무시
  kvmWin.webContents.on('certificate-error', (event) => {
    event.preventDefault();
    // callback(true) 대신 이벤트 기본 동작 막기
  });
  // certificate-error에서 콜백 방식으로 처리
  kvmWin.webContents.session.setCertificateVerifyProc((request, callback) => {
    callback(0); // 0 = 인증서 오류 무시
  });

  kvmWin.loadURL(buildKvmUrl(device));
  kvmWin.once('ready-to-show', () => kvmWin.show());
  kvmWin.on('closed', () => delete kvmWindows[winId]);
  kvmWindows[winId] = kvmWin;
}

// ─── 벤더별 KVM URL 빌더 ─────────────────────────────────────────
function buildKvmUrl(device) {
  const proto = (device.https !== false) ? 'https' : 'http';
  const base  = `${proto}://${device.ipmi_ip}`;
  switch ((device.vendor || '').toLowerCase()) {
    case 'dell':        return `${base}/console`;
    case 'hp':
    case 'hpe':         return `${base}/html5/kvm`;
    case 'supermicro':  return `${base}/cgi/ipmi.cgi`;
    case 'asus':
    case 'asrock':      return `${base}/index.html`;
    default:            return base;
  }
}

// ─── 벤더별 IPMI 로그인 페이지 URL 빌더 ─────────────────────────
function buildLoginUrl(device) {
  const proto = (device.https !== false) ? 'https' : 'http';
  const base  = `${proto}://${device.ipmi_ip}`;
  switch ((device.vendor || '').toLowerCase()) {
    case 'dell':        return `${base}/login.html`;
    case 'hp':
    case 'hpe':         return `${base}/ui/`;
    case 'supermicro':  return `${base}/cgi/login.cgi`;
    default:            return base;
  }
}

// ─── IPMI 페이지 자동 로그인 창 열기 ────────────────────────────
function openIpmiWithAutoLogin(device) {
  const winId = `ipmi-${device.id}`;

  // 이미 열려있으면 포커스
  if (kvmWindows[winId]) { kvmWindows[winId].focus(); return; }

  const win = new BrowserWindow({
    width: 1280,
    height: 800,
    title: `IPMI - ${device.name} (${device.ipmi_ip})`,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: false,
      webSecurity: false,
      allowRunningInsecureContent: true,
    },
    show: false,
  });

  // SSL 인증서 오류 무시
  win.webContents.session.setCertificateVerifyProc((_, callback) => callback(0));

  const loginUrl = buildLoginUrl(device);
  win.loadURL(loginUrl);

  // 페이지 로드 완료 시 로그인 스크립트 주입
  win.webContents.on('did-finish-load', () => {
    const currentUrl = win.webContents.getURL();
    // 이미 로그인된 경우(대시보드 URL) 스크립트 주입 생략
    const loginIndicators = ['login', 'signin', 'auth', 'cgi/login'];
    const isLoginPage = loginIndicators.some(kw => currentUrl.toLowerCase().includes(kw))
                        || currentUrl === loginUrl
                        || currentUrl === loginUrl + '/';

    if (isLoginPage && device.username) {
      const script = autoLoginScripts.getLoginScript(
        device.vendor,
        device.username,
        device.password || ''
      );
      win.webContents.executeJavaScript(script).catch(() => {});
    }
  });

  win.once('ready-to-show', () => win.show());
  win.on('closed', () => delete kvmWindows[winId]);
  kvmWindows[winId] = win;
}

// ─── KVM HTML5 자동 로그인 후 KVM 진입 ──────────────────────────
function openKvmWithAutoLogin(device) {
  const winId = device.id;
  if (kvmWindows[winId]) { kvmWindows[winId].focus(); return; }

  const kvmWin = new BrowserWindow({
    width: 1024,
    height: 768,
    title: `KVM - ${device.name} (${device.ipmi_ip})`,
    webPreferences: {
      nodeIntegration: false,
      contextIsolation: false,
      webSecurity: false,
      allowRunningInsecureContent: true,
    },
    show: false,
  });

  kvmWin.webContents.session.setCertificateVerifyProc((_, callback) => callback(0));

  // 계정 정보 없으면 기존 방식으로 바로 KVM
  if (!device.username) {
    kvmWin.loadURL(buildKvmUrl(device));
    kvmWin.once('ready-to-show', () => kvmWin.show());
    kvmWin.on('closed', () => delete kvmWindows[winId]);
    kvmWindows[winId] = kvmWin;
    return;
  }

  // Step1: 로그인 페이지 먼저 로드 → 자동 로그인
  const loginUrl = buildLoginUrl(device);
  const kvmUrl   = buildKvmUrl(device);
  let loginDone  = false;

  kvmWin.loadURL(loginUrl);

  kvmWin.webContents.on('did-finish-load', () => {
    const currentUrl = kvmWin.webContents.getURL();
    const loginIndicators = ['login', 'signin', 'auth', 'cgi/login'];
    const isLoginPage = loginIndicators.some(kw => currentUrl.toLowerCase().includes(kw))
                        || currentUrl === loginUrl
                        || currentUrl === loginUrl + '/';

    if (!loginDone && isLoginPage) {
      const script = autoLoginScripts.getLoginScript(
        device.vendor,
        device.username,
        device.password || ''
      );
      kvmWin.webContents.executeJavaScript(script).catch(() => {});
    } else if (!loginDone && !isLoginPage) {
      // 로그인 완료 감지 → KVM URL로 이동
      loginDone = true;
      setTimeout(() => kvmWin.loadURL(kvmUrl), 800);
    }
  });

  kvmWin.once('ready-to-show', () => kvmWin.show());
  kvmWin.on('closed', () => delete kvmWindows[winId]);
  kvmWindows[winId] = kvmWin;
}

// ─── IPC 핸들러 ──────────────────────────────────────────────────

// [설정] 저장
ipcMain.handle('config:save', async (_, config) => {
  try {
    fs.writeFileSync(CONFIG_PATH, JSON.stringify(config, null, 2), 'utf8');
    return { success: true };
  } catch (e) { return { success: false, error: e.message }; }
});

// [설정] 로드
ipcMain.handle('config:load', async () => {
  try {
    if (fs.existsSync(CONFIG_PATH)) return JSON.parse(fs.readFileSync(CONFIG_PATH, 'utf8'));
    return {};
  } catch (_) { return {}; }
});

// [KVM] HTML5 창 열기 (기존 - 로그인 없음)
ipcMain.handle('kvm:open-html5', async (_, device) => {
  openKvmWindow(device);
  return { success: true };
});

// [KVM] HTML5 창 열기 (자동 로그인 포함)
ipcMain.handle('kvm:open-html5-autologin', async (_, device) => {
  openKvmWithAutoLogin(device);
  return { success: true };
});

// [IPMI] IPMI 페이지 자동 로그인
ipcMain.handle('ipmi:open-autologin', async (_, device) => {
  openIpmiWithAutoLogin(device);
  return { success: true };
});

// [KVM] 외부 뷰어/JNLP 실행
ipcMain.handle('kvm:launch-external', async (_, { device, method, javawsPath }) => {
  try {
    switch (method) {
      case 'jnlp': {
        const proto = device.https !== false ? 'https' : 'http';
        const jnlpUrl = `${proto}://${device.ipmi_ip}/viewer.jnlp?EXTPORT=-1&JNLPSTR=AppletRedirection`;
        const javaws = javawsPath || 'C:\\Program Files\\Java\\jre1.8.0_441\\bin\\javaws.exe';
        if (!fs.existsSync(javaws)) {
          return { success: false, error: 'javaws.exe를 찾을 수 없습니다. Java 설정을 확인하세요.' };
        }
        spawn(`"${javaws}"`, [jnlpUrl], { shell: true, detached: true, stdio: 'ignore' });
        break;
      }
      case 'ipmiview': {
        const candidates = [
          'C:\\Program Files\\IPMI\\IPMIView\\IPMIView.exe',
          'C:\\Program Files (x86)\\IPMI\\IPMIView\\IPMIView.exe',
        ];
        const found = candidates.find(p => fs.existsSync(p));
        if (!found) return { success: false, error: 'IPMIView가 설치되지 않았습니다.' };
        spawn(`"${found}"`, [device.ipmi_ip], { shell: true, detached: true, stdio: 'ignore' });
        break;
      }
      default:
        shell.openExternal(`https://${device.ipmi_ip}`);
    }
    return { success: true };
  } catch (e) { return { success: false, error: e.message }; }
});

// [Java] 설치된 Java 목록 탐지
ipcMain.handle('java:detect', async () => {
  try {
    const list = await javaManager.detectJavaInstallations();
    return { success: true, list };
  } catch (e) { return { success: false, error: e.message, list: [] }; }
});

// [Java] 예외 사이트 등록
ipcMain.handle('java:add-exception', async (_, { siteUrl }) => {
  try {
    const result = javaManager.addJavaExceptionSite(siteUrl);
    return { success: true, ...result };
  } catch (e) { return { success: false, error: e.message }; }
});

// [Java] 레거시 보안 설정 적용
ipcMain.handle('java:apply-legacy-config', async () => {
  try {
    const result = javaManager.applyLegacyJavaConfig();
    return { success: true, ...result };
  } catch (e) { return { success: false, error: e.message }; }
});

// [Java] 구버전 다운로드 정보
ipcMain.handle('java:get-download-links', async () => {
  return javaManager.getLegacyJavaDownloadInfo();
});

// [시스템] 외부 URL 열기
ipcMain.handle('shell:open-url', async (_, url) => {
  shell.openExternal(url);
  return { success: true };
});

// [다이얼로그] 파일 열기
ipcMain.handle('dialog:open-file', async (_, options) => {
  return dialog.showOpenDialog(mainWindow, options);
});

// ─── 앱 생명주기 ─────────────────────────────────────────────────
app.whenReady().then(createMainWindow);

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit();
});

app.on('activate', () => {
  if (BrowserWindow.getAllWindows().length === 0) createMainWindow();
});
