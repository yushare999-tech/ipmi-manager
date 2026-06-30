/**
 * IPMI Manager - Electron 메인 프로세스
 * 작성일: 2026-06-25
 * 변경이력:
 *   - 2026-06-25: 최초 작성 (SSL 우회, KVM 창 열기, Java IPC 통합)
 *   - 2026-06-25: IPMI 자동 로그인 기능 추가 (벤더별 폼 자동 주입)
 *   - 2026-06-26: Dev 모드 DevTools 자동 오픈 + 타이밍 로그 추가 (검수용)
 *   - 2026-06-26: 대시보드 감지 후 자동 reload + iDRAC ST1/ST2 토큰 JNLP 개선
 */

const { app, BrowserWindow, ipcMain, shell, dialog, session } = require('electron');
const path    = require('path');
const https   = require('https');
const { spawn, exec } = require('child_process');
const fs      = require('fs');
const javaManager      = require('./vendors/java-manager');
const autoLoginScripts = require('./vendors/auto-login-scripts');

// ─── iDRAC REST API 직접 로그인 (ST1/ST2 토큰 획득) ────────────────
// POST /data/login → XML 또는 JSON 응답으로 ST1/ST2 토큰 반환
function idracLogin(device) {
  return new Promise((resolve, reject) => {
    const postData = `user=${encodeURIComponent(device.username)}&password=${encodeURIComponent(device.password || '')}`;
    const options = {
      hostname: device.ipmi_ip,
      port: 443,
      path: '/data/login',
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
        'Content-Length': Buffer.byteLength(postData),
      },
      rejectUnauthorized: false,
    };
    const req = https.request(options, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
        try {
          let authResult = -1;
          let tokenString = '';  // "ST1=xxx,ST2=yyy" 형태

          const trimmed = data.trim();

          if (trimmed.startsWith('<?xml') || trimmed.startsWith('<root>')) {
            // ─── XML 응답 (iDRAC 6/7/8) ───────────────────────────
            const authMatch    = trimmed.match(/<authResult>(\d+)<\/authResult>/);
            const forwardMatch = trimmed.match(/ST1=([a-f0-9]+),ST2=([a-f0-9]+)/i);

            authResult  = authMatch    ? parseInt(authMatch[1])    : -1;
            tokenString = forwardMatch ? `ST1=${forwardMatch[1]},ST2=${forwardMatch[2]}` : '';

            console.log(`[iDRAC] XML 파싱 - authResult:${authResult}, token:${tokenString ? '획득' : '없음'}`);
          } else {
            // ─── JSON 응답 (iDRAC 9+) ─────────────────────────────
            const json = JSON.parse(trimmed);
            authResult = json.authResult ?? json.status ?? -1;
            if (json.ST1) tokenString = `ST1=${json.ST1}${json.ST2 ? `,ST2=${json.ST2}` : ''}`;

            console.log(`[iDRAC] JSON 파싱 - authResult:${authResult}, token:${tokenString ? '획득' : '없음'}`);
          }

          if (authResult === 0 && tokenString) {
            const [st1Part, st2Part] = tokenString.split(',');
            const st1 = st1Part?.replace('ST1=', '') || '';
            const st2 = st2Part?.replace('ST2=', '') || '';
            resolve({ success: true, st1, st2, tokenString });
          } else if (authResult === 0) {
            resolve({ success: true, st1: '', st2: '', tokenString: '' });
          } else {
            resolve({ success: false, error: `인증 실패 (authResult:${authResult})` });
          }
        } catch (e) {
          reject(new Error(`iDRAC 응답 파싱 실패: ${e.message}`));
        }
      });
    });
    req.on('error', reject);
    req.setTimeout(10000, () => { req.destroy(); reject(new Error('iDRAC 로그인 타임아웃')); });
    req.write(postData);
    req.end();
  });
}

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

function readConfig() {
  try {
    if (fs.existsSync(CONFIG_PATH)) {
      return JSON.parse(fs.readFileSync(CONFIG_PATH, 'utf8'));
    }
  } catch (e) {
    console.error('[Config] 로드 실패:', e);
  }
  return {};
}

// Dev 모드 여부 (npm run dev 시 --dev 플래그)
const isDev = process.argv.includes('--dev');

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
  kvmWin.$device = device;
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
async function performInteractiveLogin(webContents, device, autoSubmit, log) {
  const querySelectorAllAllHelper = `
    function querySelectorAllAll(selector, doc) {
      doc = doc || document;
      var elements = Array.prototype.slice.call(doc.querySelectorAll(selector));
      var frames = doc.querySelectorAll('iframe, frame');
      for (var i = 0; i < frames.length; i++) {
        try {
          var frameDoc = frames[i].contentDocument || frames[i].contentWindow.document;
          if (frameDoc) {
            elements = elements.concat(querySelectorAllAll(selector, frameDoc));
          }
        } catch (e) {}
      }
      return elements;
    }
  `;

  try {
    webContents.focus();

    // 중복 로그 출력을 방지하기 위한 URL 상태 체크
    const currentUrl = webContents.getURL();
    let shouldLogDiag = false;
    if (!webContents.$lastDiagUrl || webContents.$lastDiagUrl !== currentUrl) {
      webContents.$lastDiagUrl = currentUrl;
      shouldLogDiag = true;
    }

    // 현재 페이지의 Form 및 Input 진단
    const diagScript = `
      (function() {
        ${querySelectorAllAllHelper}
        var inputs = querySelectorAllAll('input');
        var info = inputs.map(function(el) {
          return {
            id: el.id,
            name: el.name,
            type: el.type,
            visible: el.offsetWidth > 0 && el.offsetHeight > 0,
            placeholder: el.placeholder
          };
        });
        var forms = querySelectorAllAll('form').map(function(f) {
          return { id: f.id, name: f.name, action: f.action };
        });
        return { inputs: info, forms: forms, url: window.location.href };
      })()
    `;
    
    const pageDiag = await webContents.executeJavaScript(diagScript).catch(() => null);
    if (pageDiag && shouldLogDiag) {
      log(`[디버그][DOM 진단] URL: ${pageDiag.url}`);
      log(`[디버그][DOM 진단] 발견된 Form: ${JSON.stringify(pageDiag.forms)}`);
      log(`[디버그][DOM 진단] 발견된 Input: ${JSON.stringify(pageDiag.inputs)}`);
    }

    const vendorKey = (device.vendor || 'generic').toLowerCase();
    const uSelectors = userSelectors[vendorKey] || userSelectors.generic;

    const passSelectors = {
      dell: `['#password', 'input[name="password"]', 'input[type="password"]']`,
      hp: `['#password', 'input[name="password"]', 'input[autocomplete="current-password"]', 'input[type="password"]']`,
      supermicro: `['input[name="pwd"]', 'input[name="password"]', 'input[type="password"]']`,
      generic: `['input[type="password"]']`
    };
    const pSelectors = passSelectors[vendorKey] || passSelectors.generic;

    // 단일 동기 자바스크립트 내에서 포커싱, 값 주입, 도메인 변경, 캐릭터별 키 이벤트 디스패치를 완벽히 수행
    const fillScript = `
      (function() {
        ${querySelectorAllAllHelper}
        
        // 0. 도메인 셀렉트 박스가 있을 경우 로컬(This iDRAC, 보통 0번 인덱스)로 강제 복구
        var domainSelect = querySelectorAllAll('#domainDisp').concat(querySelectorAllAll('select[name="domainDisp"]'))[0];
        if (domainSelect && domainSelect.selectedIndex !== 0) {
          domainSelect.selectedIndex = 0;
          domainSelect.dispatchEvent(new Event('change', { bubbles: true }));
        }

        function fillField(selectors, value) {
          var el = null;
          var matched = '';
          for (var i = 0; i < selectors.length; i++) {
            var elements = querySelectorAllAll(selectors[i]);
            for (var j = 0; j < elements.length; j++) {
              var target = elements[j];
              if (selectors[i] === 'input') {
                var t = (target.type || '').toLowerCase();
                var n = (target.name || target.id || '').toLowerCase();
                if ((t === 'text' || t === 'email' || t === 'password') && 
                    (n.includes('user') || n.includes('name') || n.includes('login') || n.includes('id') || n.includes('pwd') || t === 'password')) {
                  el = target;
                  matched = 'input[keyword/type]';
                  break;
                }
              } else {
                el = target;
                matched = selectors[i];
                break;
              }
            }
            if (el) break;
          }
          if (!el && selectors.includes('input')) {
            el = querySelectorAllAll(selectors[0].includes('password') ? 'input[type="password"]' : 'input[type="text"]')[0];
            matched = 'fallback';
          }
          
          if (el) {
            el.focus();
            
            // 1차 값 대입
            var getProto = Object.getPrototypeOf || function(obj) { return obj.__proto__; };
            var proto = getProto(el);
            var desc = Object.getOwnPropertyDescriptor(proto, 'value');
            var setter = desc ? desc.set : null;
            if (setter) {
              setter.call(el, value);
            } else {
              el.value = value;
            }
            
            // 2차 각 글자별 키보드 이벤트 순차 디스패치
            el.dispatchEvent(new Event('focus', { bubbles: true }));
            el.dispatchEvent(new Event('input', { bubbles: true }));
            el.dispatchEvent(new Event('change', { bubbles: true }));
            
            for (var k = 0; k < value.length; k++) {
              var char = value[k];
              var code = char.charCodeAt(0);
              el.dispatchEvent(new KeyboardEvent('keydown', { bubbles: true, key: char, char: char, keyCode: code }));
              el.dispatchEvent(new KeyboardEvent('keypress', { bubbles: true, key: char, char: char, keyCode: code }));
              el.dispatchEvent(new KeyboardEvent('keyup', { bubbles: true, key: char, char: char, keyCode: code }));
            }
            
            el.blur();
            return { success: true, selector: matched, id: el.id, name: el.name };
          }
          return { success: false };
        }

        var userResult = fillField(${JSON.stringify(uSelectors)}, ${JSON.stringify(device.username)});
        if (!userResult.success) return { step: 'user', success: false };
        
        var passResult = fillField(${JSON.stringify(pSelectors)}, ${JSON.stringify(device.password || '')});
        if (!passResult.success) return { step: 'pass', success: false };
        
        return { 
          success: true, 
          userId: userResult.id, userName: userResult.name, 
          passId: passResult.id, passName: passResult.name 
        };
      })()
    `;

    const fillResult = await webContents.executeJavaScript(fillScript);
    if (!fillResult || !fillResult.success) {
      // 요소를 탐색하는 동안 로그가 너무 도배되지 않도록 디버그 로깅 축소
      if (shouldLogDiag) {
        log(`[디버그] 필드 입력 대기 중... 실패 단계: ${fillResult ? fillResult.step : 'unknown'}`);
      }
      return false;
    }

    log(`[디버그] 동기식 주입 성공! ID(ID: "${fillResult.userId}", Name: "${fillResult.userName}"), PW(ID: "${fillResult.passId}", Name: "${fillResult.passName}")`);

    // 5. 자동 제출 처리
    if (autoSubmit) {
      const submitBtnSelectors = {
        dell: `['button[type="submit"]', 'input[type="submit"]', '#btnOK', '.btn-primary']`,
        hp: `['button[type="submit"]', '#btn-login', '.btn-primary', 'input[type="submit"]']`,
        supermicro: `['input[type="submit"]', 'button[type="submit"]', '#login_word']`,
        generic: `['button[type="submit"]', 'input[type="submit"]', '.btn-primary', '.login-btn']`
      };
      const bSelectors = submitBtnSelectors[vendorKey] || submitBtnSelectors.generic;

      const clickSubmitScript = `
        (function() {
          ${querySelectorAllAllHelper}
          var selectors = ${bSelectors};
          var btn = null;
          var matchedSelector = '';
          for (var i = 0; i < selectors.length; i++) {
            var elements = querySelectorAllAll(selectors[i]);
            if (elements.length > 0) {
              btn = elements[0];
              matchedSelector = selectors[i];
              break;
            }
          }
          if (btn) {
            btn.focus();
            btn.click();
            var subMethod = 'click';
            if ('${vendorKey}' === 'dell') {
              var elWin = btn.ownerDocument.defaultView || window;
              if (elWin && typeof elWin.frmSubmit === 'function') {
                elWin.frmSubmit();
                subMethod = 'frmSubmit()';
              }
            }
            return { success: true, selector: matchedSelector, method: subMethod };
          }
          if ('${vendorKey}' === 'dell' && typeof window.frmSubmit === 'function') {
            window.frmSubmit();
            return { success: true, selector: 'global window.frmSubmit', method: 'frmSubmit()' };
          }
          return { success: false };
        })()
      `;

      const submittedResult = await webContents.executeJavaScript(clickSubmitScript);
      if (submittedResult && submittedResult.success) {
        log(`[디버그] 제출 매칭 성공! 셀렉터: "${submittedResult.selector}", 실행 방식: "${submittedResult.method}"`);
      } else {
        log(`[디버그] 제출 버튼 클릭 실패 → Enter 키 직접 전송`);
        webContents.sendInputEvent({ type: 'keyDown', keyCode: 'Enter' });
        webContents.sendInputEvent({ type: 'keyUp', keyCode: 'Enter' });
      }
    }
    return true;
  } catch (e) {
    log(`[디버그] 인터랙티브 로그인 에러: ${e.message}`);
    return false;
  }
}
async function openIpmiWithAutoLogin(device) {
  const winId = `ipmi-${device.id}`;
  if (kvmWindows[winId]) { kvmWindows[winId].focus(); return; }

  const t0  = Date.now();
  const log = (msg) => console.log(`[AutoLogin][IPMI][${Date.now()-t0}ms] ${msg}`);
  const proto = device.https !== false ? 'https' : 'http';

  log(`시작 - 장비: ${device.name} (${device.vendor})`);

  const config = readConfig();
  const autoSubmit = config.autoSubmit === true;
  const enableDevTools = config.enableDevTools === true;

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

  if (enableDevTools) win.webContents.openDevTools({ mode: 'detach' });
  win.webContents.session.setCertificateVerifyProc((_, callback) => callback(0));
  win.once('ready-to-show', () => { log('ready-to-show → 창 표시'); win.show(); });
  win.on('closed', () => { log('창 닫힘'); delete kvmWindows[winId]; if (loginTimer) clearInterval(loginTimer); });
  win.$device = device;
  kvmWindows[winId] = win;

  // Dell iDRAC REST API 직접 로그인 시도
  let restLoginAttempted = false;
  let restLoginUrl = null;

  if (device.username && (device.vendor || '').toLowerCase() === 'dell') {
    log('Dell iDRAC REST API 로그인 시도');
    try {
      const loginResult = await idracLogin(device);
      if (loginResult.success && loginResult.tokenString) {
        // Dell iDRAC 세션 세팅을 위해 프로토콜을 https로 강제 고정합니다.
        restLoginUrl = `https://${device.ipmi_ip}/index.html?${loginResult.tokenString}`;
        log(`REST 로그인 성공! 대시보드로 직접 오픈 예정: ${restLoginUrl}`);
        restLoginAttempted = true;
      } else {
        log(`REST 로그인 실패 (${loginResult.error}), 일반 주입 방식으로 폴백`);
      }
    } catch (e) {
      log(`REST 로그인 예외: ${e.message}, 일반 주입 방식으로 폴백`);
    }
  }

  const loginUrl = buildLoginUrl(device);
  if (restLoginAttempted && restLoginUrl) {
    win.loadURL(restLoginUrl);
    log(`REST 성공 URL 로드 시작: ${restLoginUrl}`);
  } else {
    win.loadURL(loginUrl);
    log(`loginUrl 로드 시작 (인터랙티브 방식): ${loginUrl}`);
  }

  let dashboardLoaded = false;
  let loginDone = false;
  let loginInProgress = false;
  let loginTimer = null;

  const injectLoginScript = async () => {
    if (win.isDestroyed()) return;
    if (dashboardLoaded || loginDone) return;

    const currentUrl = win.webContents.getURL();
    log(`[AutoLogin] 상태 체크 - URL: ${currentUrl}`);

    const dashIndicators = ['index.html?st', 'dashboard', 'sys_summary', 'main.html', 'rfc3986'];
    const isDashboard = dashIndicators.some(kw => currentUrl.toLowerCase().includes(kw));

    if (isDashboard) {
      dashboardLoaded = true;
      log('대시보드 진입 완료! 1.5초 후 새로고침');
      if (loginTimer) { clearInterval(loginTimer); loginTimer = null; }
      setTimeout(() => { if (!win.isDestroyed()) win.webContents.reload(); }, 1500);
      return;
    }

    if (device.username && !loginInProgress) {
      loginInProgress = true;
      log('인터랙티브 로그인 시뮬레이션 시작');
      const success = await performInteractiveLogin(win.webContents, device, autoSubmit, log);
      if (success) {
        log('인터랙티브 로그인 입력 완료.');
        loginDone = true;
        if (loginTimer) { clearInterval(loginTimer); loginTimer = null; }
      } else {
        loginInProgress = false;
      }
    }
  };

  loginTimer = setInterval(injectLoginScript, 800);

  win.webContents.on('did-finish-load',      injectLoginScript);
  win.webContents.on('did-navigate-in-page', injectLoginScript);
  win.webContents.on('did-frame-finish-load', () => injectLoginScript());
  win.webContents.on('did-start-loading',    () => log('did-start-loading'));
  win.webContents.on('dom-ready',            () => log('dom-ready'));
  win.webContents.on('did-navigate',         (_, url) => { log(`did-navigate → ${url}`); injectLoginScript(); });
}

function openKvmWithAutoLogin(device) {
  const winId = device.id;
  if (kvmWindows[winId]) { kvmWindows[winId].focus(); return; }

  const t0 = Date.now();
  const log = (msg) => console.log(`[AutoLogin][KVM][${Date.now()-t0}ms] ${msg}`);

  log(`시작 - 장비: ${device.name} (${device.vendor})`);

  const config = readConfig();
  const autoSubmit = config.autoSubmit === true;
  const enableDevTools = config.enableDevTools === true;

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

  if (enableDevTools) kvmWin.webContents.openDevTools({ mode: 'detach' });
  kvmWin.webContents.session.setCertificateVerifyProc((_, callback) => callback(0));

  if (!device.username) {
    log('계정 정보 없음 → 바로 KVM 오픈');
    kvmWin.loadURL(buildKvmUrl(device));
    kvmWin.once('ready-to-show', () => kvmWin.show());
    kvmWin.on('closed', () => delete kvmWindows[winId]);
    kvmWindows[winId] = kvmWin;
    return;
  }

  const loginUrl = buildLoginUrl(device);
  const kvmUrl   = buildKvmUrl(device);
  let loginDone  = false;
  let loginInProgress = false;
  let loginTimer = null;

  log(`loginUrl 로드: ${loginUrl}`);
  kvmWin.loadURL(loginUrl);

  const injectKvmLoginScript = async () => {
    if (kvmWin.isDestroyed()) return;
    if (loginDone) return;

    const currentUrl = kvmWin.webContents.getURL();
    log(`[AutoLogin KVM] 상태 체크 - URL: ${currentUrl}`);

    const isKvmConsole = currentUrl.toLowerCase().includes('console') 
                         || currentUrl.toLowerCase().includes('kvm')
                         || currentUrl.toLowerCase().includes('viewer');

    if (isKvmConsole) {
      log(`KVM 콘솔 화면 진입 완료`);
      loginDone = true;
      if (loginTimer) { clearInterval(loginTimer); loginTimer = null; }
      return;
    }

    const loginIndicators = ['login', 'signin', 'auth', 'cgi/login'];
    const isLoginPage = loginIndicators.some(kw => currentUrl.toLowerCase().includes(kw))
                        || currentUrl === loginUrl
                        || currentUrl === loginUrl + '/'
                        || (!currentUrl.toLowerCase().includes('console') && !currentUrl.toLowerCase().includes('kvm'));

    if (isLoginPage && device.username && !loginInProgress) {
      loginInProgress = true;
      log('KVM 인터랙티브 로그인 시뮬레이션 시작');
      const success = await performInteractiveLogin(kvmWin.webContents, device, autoSubmit, log);
      if (success) {
        log('KVM 인터랙티브 로그인 완료');
        loginDone = true;
        if (loginTimer) { clearInterval(loginTimer); loginTimer = null; }
      } else {
        loginInProgress = false;
      }
    } else if (!isLoginPage && !isKvmConsole && !loginInProgress) {
      log(`로그인 완료로 판단 (리다이렉션 발생). KVM URL로 이동: ${kvmUrl}`);
      loginDone = true;
      if (loginTimer) { clearInterval(loginTimer); loginTimer = null; }
      setTimeout(() => { if (!kvmWin.isDestroyed()) kvmWin.loadURL(kvmUrl); }, 800);
    }
  };

  loginTimer = setInterval(injectKvmLoginScript, 800);

  kvmWin.webContents.on('did-finish-load',      injectKvmLoginScript);
  kvmWin.webContents.on('did-navigate-in-page', injectKvmLoginScript);
  kvmWin.webContents.on('did-frame-finish-load', () => injectKvmLoginScript());
  kvmWin.webContents.on('did-start-loading',    () => log('did-start-loading'));
  kvmWin.webContents.on('dom-ready',            () => log('dom-ready'));
  kvmWin.webContents.on('did-navigate',         (_, url) => { log(`did-navigate → ${url}`); injectKvmLoginScript(); });

  kvmWin.once('ready-to-show', () => { log('ready-to-show → 창 표시'); kvmWin.show(); });
  kvmWin.on('closed', () => { log('창 닫힘'); delete kvmWindows[winId]; if (loginTimer) clearInterval(loginTimer); });
  kvmWindows[winId] = kvmWin;
}
ipcMain.handle('config:save', async (_, config) => {
  try {
    fs.writeFileSync(CONFIG_PATH, JSON.stringify(config, null, 2), 'utf8');
      // 장비 목록 변경 시 즉각 Java 예외 목록 재동기화
      syncAllDevicesToJavaExceptions();
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
        const proto  = device.https !== false ? 'https' : 'http';
        const javaws = javawsPath || 'C:\\Program Files\\Java\\jre1.8.0_441\\bin\\javaws.exe';
        if (!fs.existsSync(javaws)) {
          return { success: false, error: 'javaws.exe를 찾을 수 없습니다. Java 설정을 확인하세요.' };
        }

        // JNLP 실행 전 해당 장비의 IP를 Java 예외 목록에 자동 등록하고 레거시 설정을 적용합니다.
        try {
          console.log(`[JNLP] 예외 사이트 자동 등록 시도: ${device.ipmi_ip}`);
          javaManager.addJavaExceptionSite(device.ipmi_ip);
          javaManager.applyLegacyJavaConfig();
        } catch (e) {
          console.warn(`[JNLP] 예외 사이트 등록 중 오류 (무시됨): ${e.message}`);
        }

        let jnlpUrl;

        // Dell iDRAC: REST API 직접 로그인 → ST1/ST2 토큰 포함 JNLP URL
        if (device.username && (device.vendor || '').toLowerCase() === 'dell') {
          console.log(`[JNLP] Dell iDRAC REST 로그인 시도: ${device.ipmi_ip}`);
          try {
            const loginResult = await idracLogin(device);
            if (loginResult.success && loginResult.tokenString) {
              // iDRAC JNLP URL: viewer.jnlp?EXTPORT=-1&JNLPSTR=AppletRedirection&ST1=xxx,ST2=yyy
              jnlpUrl = `${proto}://${device.ipmi_ip}/viewer.jnlp?EXTPORT=-1&JNLPSTR=AppletRedirection&${loginResult.tokenString}`;
              console.log(`[JNLP] ✅ 로그인 성공! token: ${loginResult.tokenString}`);
            } else {
              console.warn(`[JNLP] REST 로그인 실패 (${loginResult.error}), 토큰 없이 시도`);
              jnlpUrl = `${proto}://${device.ipmi_ip}/viewer.jnlp?EXTPORT=-1&JNLPSTR=AppletRedirection`;
            }
          } catch (e) {
            console.warn(`[JNLP] REST 예외: ${e.message}, 토큰 없이 시도`);
            jnlpUrl = `${proto}://${device.ipmi_ip}/viewer.jnlp?EXTPORT=-1&JNLPSTR=AppletRedirection`;
          }
        } else {
          jnlpUrl = `${proto}://${device.ipmi_ip}/viewer.jnlp?EXTPORT=-1&JNLPSTR=AppletRedirection`;
        }

        spawn(`"${javaws}"`, [jnlpUrl], { shell: true, detached: true, stdio: 'ignore' });
        console.log(`[JNLP] javaws 실행: ${javaws} ${jnlpUrl}`);
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

// [Java] java.security 패치 (제한 해제)
ipcMain.handle('java:patch-security', async (_, { javawsPath }) => {
  try {
    const result = await javaManager.patchJavaSecurity(javawsPath);
    return { success: true, ...result };
  } catch (e) {
    return { success: false, error: e.message };
  }
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
app.whenReady().then(() => {

  // 다운로드 완료 시 실행 확인 팝업 기능 추가
  session.defaultSession.on('will-download', (event, item, webContents) => {
    // 다운로드 저장 경로가 사전에 정의되지 않은 경우 기본 다운로드 폴더로 강제 지정하여 다운로드 중단을 방지
    const currentPath = item.getSavePath();
    if (!currentPath) {
      const downloadsPath = app.getPath('downloads');
      item.setSavePath(require('path').join(downloadsPath, item.getFilename()));
    }

    item.once('done', async (event, state) => {
      console.log(`[Download] 완료 상태: ${state}, 저장 경로: ${item.getSavePath()}`);
      if (state === 'completed') {
        const filePath = item.getSavePath();
        const fileName = item.getFilename();
        const focusWindow = BrowserWindow.getFocusedWindow() || mainWindow;
        
        // JNLP 실행 전 해당 장비의 IP 및 Hostname을 자바 예외 사이트에 강제 재등록
        try {
          const win = BrowserWindow.fromWebContents(webContents);
          const dev = win ? win.$device : null;
          if (dev) {
            console.log(`[JNLP 실행 전 자바 설정 강제 갱신] 장비: ${dev.name}, IP: ${dev.ipmi_ip}, Hostname: ${dev.hostname || '없음'}`);
            // IP 등록
            javaManager.addJavaExceptionSite(dev.ipmi_ip);
            // 호스트명 등록 (JNLP 내부에 호스트명으로 정의되어 있을 경우 대비)
            if (dev.hostname) {
              javaManager.addJavaExceptionSite(dev.hostname);
            }
            // 자바 보안 우회 및 레거시 TLS 설정 즉각 주입
            javaManager.applyLegacyJavaConfig();
          }
        } catch (e) {
          console.error('[JNLP 실행 전 자바 설정 갱신 실패]:', e);
        }

        
        const { response } = await dialog.showMessageBox(focusWindow, {
          type: 'question',
          buttons: ['예', '아니오'],
          defaultId: 0,
          title: '다운로드 완료',
          message: `파일 다운로드가 완료되었습니다.\n\n파일명: ${fileName}\n\n지금 이 파일을 실행하시겠습니까?`,
          cancelId: 1
        });

        if (response === 0) {
          shell.openPath(filePath).catch(err => {
            dialog.showErrorBox('실행 실패', `파일을 실행하는 중 오류가 발생했습니다.\n경로: ${filePath}\n오류: ${err.message}`);
          });
        }
      }
    });
  });

  createMainWindow();
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit();
});

app.on('activate', () => {
  if (BrowserWindow.getAllWindows().length === 0) createMainWindow();
});


// 저장된 모든 장비의 IP를 Java 예외 목록에 일괄 동기화하는 함수
