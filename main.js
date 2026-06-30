/**
 * IPMI Manager - Electron 硫붿씤 ?꾨줈?몄뒪
 * ?묒꽦?? 2026-06-25
 * 蹂寃쎌씠??
 *   - 2026-06-25: 理쒖큹 ?묒꽦 (SSL ?고쉶, KVM 李??닿린, Java IPC ?듯빀)
 *   - 2026-06-25: IPMI ?먮룞 濡쒓렇??湲곕뒫 異붽? (踰ㅻ뜑蹂????먮룞 二쇱엯)
 *   - 2026-06-26: Dev 紐⑤뱶 DevTools ?먮룞 ?ㅽ뵂 + ??대컢 濡쒓렇 異붽? (寃?섏슜)
 *   - 2026-06-26: ??쒕낫??媛먯? ???먮룞 reload + iDRAC ST1/ST2 ?좏겙 JNLP 媛쒖꽑
 */

const { app, BrowserWindow, ipcMain, shell, dialog, session } = require('electron');
const path    = require('path');
const https   = require('https');
const { spawn, exec } = require('child_process');
const fs      = require('fs');
const javaManager      = require('./vendors/java-manager');
const autoLoginScripts = require('./vendors/auto-login-scripts');

// ??? iDRAC REST API 吏곸젒 濡쒓렇??(ST1/ST2 ?좏겙 ?띾뱷) ????????????????
// POST /data/login ??XML ?먮뒗 JSON ?묐떟?쇰줈 ST1/ST2 ?좏겙 諛섑솚
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
          let tokenString = '';  // "ST1=xxx,ST2=yyy" ?뺥깭

          const trimmed = data.trim();

          if (trimmed.startsWith('<?xml') || trimmed.startsWith('<root>')) {
            // ??? XML ?묐떟 (iDRAC 6/7/8) ???????????????????????????
            const authMatch    = trimmed.match(/<authResult>(\d+)<\/authResult>/);
            const forwardMatch = trimmed.match(/ST1=([a-f0-9]+),ST2=([a-f0-9]+)/i);

            authResult  = authMatch    ? parseInt(authMatch[1])    : -1;
            tokenString = forwardMatch ? `ST1=${forwardMatch[1]},ST2=${forwardMatch[2]}` : '';

            console.log(`[iDRAC] XML ?뚯떛 - authResult:${authResult}, token:${tokenString ? '?띾뱷' : '?놁쓬'}`);
          } else {
            // ??? JSON ?묐떟 (iDRAC 9+) ?????????????????????????????
            const json = JSON.parse(trimmed);
            authResult = json.authResult ?? json.status ?? -1;
            if (json.ST1) tokenString = `ST1=${json.ST1}${json.ST2 ? `,ST2=${json.ST2}` : ''}`;

            console.log(`[iDRAC] JSON ?뚯떛 - authResult:${authResult}, token:${tokenString ? '?띾뱷' : '?놁쓬'}`);
          }

          if (authResult === 0 && tokenString) {
            const [st1Part, st2Part] = tokenString.split(',');
            const st1 = st1Part?.replace('ST1=', '') || '';
            const st2 = st2Part?.replace('ST2=', '') || '';
            resolve({ success: true, st1, st2, tokenString });
          } else if (authResult === 0) {
            resolve({ success: true, st1: '', st2: '', tokenString: '' });
          } else {
            resolve({ success: false, error: `?몄쬆 ?ㅽ뙣 (authResult:${authResult})` });
          }
        } catch (e) {
          reject(new Error(`iDRAC ?묐떟 ?뚯떛 ?ㅽ뙣: ${e.message}`));
        }
      });
    });
    req.on('error', reject);
    req.setTimeout(10000, () => { req.destroy(); reject(new Error('iDRAC 濡쒓렇????꾩븘??)); });
    req.write(postData);
    req.end();
  });
}

// ??? SSL / 蹂댁븞 ?고쉶 ?ㅼ젙 (?덇굅??IPMI ?λ퉬 ??? ?????????????????
app.commandLine.appendSwitch('ignore-certificate-errors');
app.commandLine.appendSwitch('ignore-ssl-errors');
app.commandLine.appendSwitch('allow-insecure-localhost');
app.commandLine.appendSwitch('disable-web-security');
// 援ы삎 TLS ?꾨줈?좎퐳 ?덉슜
app.commandLine.appendSwitch('ssl-version-min', 'tls1');
app.commandLine.appendSwitch('cipher-suite-blacklist', '');

// ?ㅼ젙 ?뚯씪 寃쎈줈
const CONFIG_PATH = path.join(app.getPath('userData'), 'ipmi-config.json');

// Dev 紐⑤뱶 ?щ? (npm run dev ??--dev ?뚮옒洹?
const isDev = process.argv.includes('--dev');

let mainWindow;
let kvmWindows = {};

function readConfig() {
  try {
    if (fs.existsSync(CONFIG_PATH)) return JSON.parse(fs.readFileSync(CONFIG_PATH, 'utf8'));
  } catch (e) {
    console.error('[Config] ?쎄린 ?ㅽ뙣:', e);
  }
  return {};
}

// ??? 硫붿씤 李??앹꽦 ????????????????????????????????????????????????
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
  mainWindow.once('ready-to-show', () => {
    mainWindow.show();
    const config = readConfig();
    if (config.enableDevTools) {
      mainWindow.webContents.openDevTools({ mode: 'detach' });
    }
  });
  mainWindow.on('closed', () => { mainWindow = null; });
}

// ??? KVM 李??닿린 (HTML5 ?댁옣 釉뚮씪?곗?) ???????????????????????????
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

  // ?몄쬆???ㅻ쪟 臾댁떆
  kvmWin.webContents.on('certificate-error', (event) => {
    event.preventDefault();
    // callback(true) ????대깽??湲곕낯 ?숈옉 留됯린
  });
  // certificate-error?먯꽌 肄쒕갚 諛⑹떇?쇰줈 泥섎━
  kvmWin.webContents.session.setCertificateVerifyProc((request, callback) => {
    callback(0); // 0 = ?몄쬆???ㅻ쪟 臾댁떆
  });

  kvmWin.loadURL(buildKvmUrl(device));
  kvmWin.once('ready-to-show', () => kvmWin.show());
  kvmWin.on('closed', () => delete kvmWindows[winId]);
  kvmWindows[winId] = kvmWin;
}

// ??? 踰ㅻ뜑蹂?KVM URL 鍮뚮뜑 ?????????????????????????????????????????
function buildKvmUrl(device) {
  const proto = (device.https !== false) ? 'https' : 'http';
  const base  = `${proto}://${device.ipmasync function openIpmiWithAutoLogin(device) {
  const winId = `ipmi-${device.id}`;
  if (kvmWindows[winId]) { kvmWindows[winId].focus(); return; }

  const t0  = Date.now();
  const log = (msg) => console.log(`[AutoLogin][IPMI][${Date.now()-t0}ms] ${msg}`);
  const proto = device.https !== false ? 'https' : 'http';

  log(`?쒖옉 - ?λ퉬: ${device.name} (${device.vendor})`);

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
  win.once('ready-to-show', () => { log('ready-to-show ??李??쒖떆'); win.show(); });
  win.on('closed', () => { log('李??ロ옒'); delete kvmWindows[winId]; });
  kvmWindows[winId] = win;

  // ?? Dell iDRAC: REST API 吏곸젒 濡쒓렇??????쒕낫??吏곸젒 ?ㅽ뵂 ??????
  if (device.username && (device.vendor || '').toLowerCase() === 'dell') {
    log('??Dell iDRAC REST API 濡쒓렇???쒕룄');
    try {
      const loginResult = await idracLogin(device);
      if (loginResult.success && loginResult.tokenString) {
        const dashUrl = `${proto}://${device.ipmi_ip}/index.html?${loginResult.tokenString}`;
        log(`??REST 濡쒓렇???깃났! ??쒕낫??吏곸젒 ?ㅽ뵂: ${dashUrl}`);
        win.loadURL(dashUrl);
        return;
      }
      log(`??REST 濡쒓렇???ㅽ뙣 (${loginResult.error}), ??二쇱엯 諛⑹떇?쇰줈 ?대갚`);
    } catch (e) {
      log(`??REST 濡쒓렇???덉쇅: ${e.message}, ??二쇱엯 諛⑹떇?쇰줈 ?대갚`);
    }
  }

  // ?? ?대갚: 釉뚮씪?곗? ??二쇱엯 諛⑹떇 (Dell ??踰ㅻ뜑 ?먮뒗 REST ?ㅽ뙣 ?? ??
  const loginUrl = buildLoginUrl(device);
  win.loadURL(loginUrl);
  log(`loginUrl 濡쒕뱶 ?쒖옉 (??二쇱엯): ${loginUrl}`);

  let dashboardLoaded = false;

  const injectLoginScript = () => {
    if (dashboardLoaded) return;
    const currentUrl = win.webContents.getURL();
    log(`[AutoLogin] URL 蹂寃?媛먯?: ${currentUrl}`);

    // ??쒕낫??吏꾩엯 媛먯? ?ㅼ썙??
    const dashIndicators = ['index.html?st', 'dashboard', 'sys_summary', 'main.html', 'rfc3986'];
    const isDashboard = dashIndicators.some(kw => currentUrl.toLowerCase().includes(kw));

    if (isDashboard) {
      dashboardLoaded = true;
      log('????쒕낫??媛먯?! 1.5珥????덈줈怨좎묠');
      setTimeout(() => { if (!win.isDestroyed()) win.webContents.reload(); }, 1500);
      return;
    }

    if (device.username) {
      log('??濡쒓렇???ㅽ겕由쏀듃 二쇱엯 ?쒕룄');
      const script = autoLoginScripts.getLoginScript(device.vendor, device.username, device.password || '', autoSubmit);
      win.webContents.executeJavaScript(script)
        .then(() => log('???ㅽ겕由쏀듃 二쇱엯 ?꾨즺'))
        .catch((e) => log(`???ㅽ겕由쏀듃 ?ㅻ쪟: ${e.message}`));
    }
  };

  win.webContents.on('did-finish-load',      injectLoginScript);
  win.webContents.on('did-navigate-in-page', injectLoginScript); // SPA ?섏씠吏€ ?대? ?쇱슦???€??
  win.webContents.on('did-frame-finish-load', (event, isMainFrame) => { log(`did-frame-finish-load - isMainFrame: ${isMainFrame}`); injectLoginScript(); });
  win.webContents.on('did-start-loading',    () => log('did-start-loading'));
  win.webContents.on('dom-ready',            () => log('dom-ready'));
  win.webContents.on('did-navigate',         (_, url) => { log(`did-navigate ??${url}`); injectLoginScript(); });
}


// ??? KVM HTML5 ?먮룞 濡쒓렇????KVM 吏꾩엯 ??????????????????????????
function openKvmWithAutoLogin(device) {
  const winId = device.id;
  if (kvmWindows[winId]) { kvmWindows[winId].focus(); return; }

  const t0 = Date.now();
  const log = (msg) => console.log(`[AutoLogin][KVM][${Date.now()-t0}ms] ${msg}`);

  log(`?쒖옉 - ?λ퉬: ${device.name} (${device.vendor})`);

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

  // 怨꾩젙 ?뺣낫 ?놁쑝硫?湲곗〈 諛⑹떇?쇰줈 諛붾줈 KVM
  if (!device.username) {
    log('怨꾩젙 ?놁쓬 ??吏곸젒 KVM ?ㅽ뵂');
    kvmWin.loadURL(buildKvmUrl(device));
    kvmWin.once('ready-to-show', () => kvmWin.show());
    kvmWin.on('closed', () => delete kvmWindows[winId]);
    kvmWindows[winId] = kvmWin;
    return;
  }

  // Step1: 濡쒓렇???섏씠吏 癒쇱? 濡쒕뱶 ???먮룞 濡쒓렇??
  const loginUrl = buildLoginUrl(device);
  const kvmUrl   = buildKvmUrl(device);
  let loginDone  = false;

  log(`loginUrl 濡쒕뱶: ${loginUrl}`);
  kvmWin.loadURL(loginUrl);

  const injectKvmLoginScript = () => {
    if (loginDone) return;
    const currentUrl = kvmWin.webContents.getURL();
    log(`[AutoLogin KVM] URL 蹂寃?媛먯?: ${currentUrl}`);

    // KVM ?붾㈃(肄섏넄) 吏꾩엯 議곌굔 媛먯?
    const isKvmConsole = currentUrl.toLowerCase().includes('console') 
                         || currentUrl.toLowerCase().includes('kvm')
                         || currentUrl.toLowerCase().includes('viewer');

    if (isKvmConsole) {
      log(`??KVM 肄섏넄 ?붾㈃ 吏꾩엯 ?꾨즺`);
      loginDone = true;
      return;
    }

    // 濡쒓렇???섏씠吏 ?몃뵒耳?댄꽣
    const loginIndicators = ['login', 'signin', 'auth', 'cgi/login'];
    const isLoginPage = loginIndicators.some(kw => currentUrl.toLowerCase().includes(kw))
                        || currentUrl === loginUrl
                        || currentUrl === loginUrl + '/'
                        || (!currentUrl.toLowerCase().includes('console') && !currentUrl.toLowerCase().includes('kvm'));

    if (isLoginPage && device.username) {
      log('??KVM 濡쒓렇???ㅽ겕由쏀듃 二쇱엯');
      const script = autoLoginScripts.getLoginScript(
        device.vendor,
        device.username,
        device.password || '',
        autoSubmit
      );
      kvmWin.webContents.executeJavaScript(script)
        .then(() => log('???ㅽ겕由쏀듃 二쇱엯 ?꾨즺'))
        .catch((e) => log(`???ㅽ겕由쏀듃 ?ㅻ쪟: ${e.message}`));
    } else {
      log(`??濡쒓렇???꾨즺濡??먮떒 (由щ떎?대젆??諛쒖깮). KVM URL濡??대룞: ${kvmUrl}`);
      loginDone = true;
      setTimeout(() => { if (!kvmWin.isDestroyed()) kvmWin.loadURL(kvmUrl); }, 800);
    }
  };

  kvmWin.webContents.on('did-finish-load',      injectKvmLoginScript);
  kvmWin.webContents.on('did-navigate-in-page', injectKvmLoginScript);
  kvmWin.webContents.on('did-frame-finish-load', (event, isMainFrame) => { log(`did-frame-finish-load - isMainFrame: ${isMainFrame}`); injectKvmLoginScript(); });
  kvmWin.webContents.on('did-start-loading',    () => log('did-start-loading'));
  kvmWin.webContents.on('dom-ready',            () => log('dom-ready'));
  kvmWin.webContents.on('did-navigate',         (_, url) => { log(`did-navigate ??${url}`); injectKvmLoginScript(); });
}넂 ?먮룞 濡쒓렇??
  const loginUrl = buildLoginUrl(device);
  const kvmUrl   = buildKvmUrl(device);
  let loginDone  = false;

  log(`loginUrl 濡쒕뱶: ${loginUrl}`);
  kvmWin.loadURL(loginUrl);

  kvmWin.webContents.on('did-start-loading', () => log('did-start-loading'));
  kvmWin.webContents.on('dom-ready',         () => log('dom-ready'));
  kvmWin.webContents.on('did-finish-load',   () => {
    const currentUrl = kvmWin.webContents.getURL();
    log(`did-finish-load - currentUrl: ${currentUrl}`);

    const loginIndicators = ['login', 'signin', 'auth', 'cgi/login'];
    const isLoginPage = loginIndicators.some(kw => currentUrl.toLowerCase().includes(kw))
                        || currentUrl === loginUrl
                        || currentUrl === loginUrl + '/';

    log(`isLoginPage: ${isLoginPage}, loginDone: ${loginDone}`);

    if (!loginDone && isLoginPage) {
      log('??濡쒓렇???ㅽ겕由쏀듃 二쇱엯');
      const script = autoLoginScripts.getLoginScript(
        device.vendor,
        device.username,
        device.password || '',
        autoSubmit
      );
      kvmWin.webContents.executeJavaScript(script)
        .then(() => log('???ㅽ겕由쏀듃 二쇱엯 ?꾨즺'))
        .catch((e) => log(`???ㅽ겕由쏀듃 ?ㅻ쪟: ${e.message}`));
    } else if (!loginDone && !isLoginPage) {
      log(`??濡쒓렇???꾨즺 媛먯?! KVM URL濡??대룞: ${kvmUrl}`);
      loginDone = true;
      setTimeout(() => kvmWin.loadURL(kvmUrl), 800);
    }
  });
  kvmWin.webContents.on('did-navigate', (_, url) => log(`did-navigate ??${url}`));

  kvmWin.once('ready-to-show', () => { log('ready-to-show ??李??쒖떆'); kvmWin.show(); });
  kvmWin.on('closed', () => { log('李??ロ옒'); delete kvmWindows[winId]; });
  kvmWindows[winId] = kvmWin;
}

// ??? IPC ?몃뱾????????????????????????????????????????????????????

// [?ㅼ젙] ???
ipcMain.handle('config:save', async (_, config) => {
  try {
    fs.writeFileSync(CONFIG_PATH, JSON.stringify(config, null, 2), 'utf8');
    return { success: true };
  } catch (e) { return { success: false, error: e.message }; }
});

// [?ㅼ젙] 濡쒕뱶
ipcMain.handle('config:load', async () => {
  try {
    if (fs.existsSync(CONFIG_PATH)) return JSON.parse(fs.readFileSync(CONFIG_PATH, 'utf8'));
    return {};
  } catch (_) { return {}; }
});

// [KVM] HTML5 李??닿린 (湲곗〈 - 濡쒓렇???놁쓬)
ipcMain.handle('kvm:open-html5', async (_, device) => {
  openKvmWindow(device);
  return { success: true };
});

// [KVM] HTML5 李??닿린 (?먮룞 濡쒓렇???ы븿)
ipcMain.handle('kvm:open-html5-autologin', async (_, device) => {
  openKvmWithAutoLogin(device);
  return { success: true };
});

// [IPMI] IPMI ?섏씠吏 ?먮룞 濡쒓렇??
ipcMain.handle('ipmi:open-autologin', async (_, device) => {
  openIpmiWithAutoLogin(device);
  return { success: true };
});

// [KVM] ?몃? 酉곗뼱/JNLP ?ㅽ뻾
ipcMain.handle('kvm:launch-external', async (_, { device, method, javawsPath }) => {
  try {
    switch (method) {
      case 'jnlp': {
        const proto  = device.https !== false ? 'https' : 'http';
        const javaws = javawsPath || 'C:\\Program Files\\Java\\jre1.8.0_441\\bin\\javaws.exe';
        if (!fs.existsSync(javaws)) {
          return { success: false, error: 'javaws.exe瑜?李얠쓣 ???놁뒿?덈떎. Java ?ㅼ젙???뺤씤?섏꽭??' };
        }

        // JNLP ?ㅽ뻾 ???대떦 ?λ퉬??IP瑜?Java ?덉쇅 紐⑸줉???먮룞 ?깅줉?섍퀬 ?덇굅???ㅼ젙???곸슜?⑸땲??
        try {
          console.log(`[JNLP] ?덉쇅 ?ъ씠???먮룞 ?깅줉 ?쒕룄: ${device.ipmi_ip}`);
          javaManager.addJavaExceptionSite(device.ipmi_ip);
          javaManager.applyLegacyJavaConfig();
        } catch (e) {
          console.warn(`[JNLP] ?덉쇅 ?ъ씠???깅줉 以??ㅻ쪟 (臾댁떆??: ${e.message}`);
        }

        let jnlpUrl;

        // Dell iDRAC: REST API 吏곸젒 濡쒓렇????ST1/ST2 ?좏겙 ?ы븿 JNLP URL
        if (device.username && (device.vendor || '').toLowerCase() === 'dell') {
          console.log(`[JNLP] Dell iDRAC REST 濡쒓렇???쒕룄: ${device.ipmi_ip}`);
          try {
            const loginResult = await idracLogin(device);
            if (loginResult.success && loginResult.tokenString) {
              // iDRAC JNLP URL: viewer.jnlp?EXTPORT=-1&JNLPSTR=AppletRedirection&ST1=xxx,ST2=yyy
              jnlpUrl = `${proto}://${device.ipmi_ip}/viewer.jnlp?EXTPORT=-1&JNLPSTR=AppletRedirection&${loginResult.tokenString}`;
              console.log(`[JNLP] ??濡쒓렇???깃났! token: ${loginResult.tokenString}`);
            } else {
              console.warn(`[JNLP] REST 濡쒓렇???ㅽ뙣 (${loginResult.error}), ?좏겙 ?놁씠 ?쒕룄`);
              jnlpUrl = `${proto}://${device.ipmi_ip}/viewer.jnlp?EXTPORT=-1&JNLPSTR=AppletRedirection`;
            }
          } catch (e) {
            console.warn(`[JNLP] REST ?덉쇅: ${e.message}, ?좏겙 ?놁씠 ?쒕룄`);
            jnlpUrl = `${proto}://${device.ipmi_ip}/viewer.jnlp?EXTPORT=-1&JNLPSTR=AppletRedirection`;
          }
        } else {
          jnlpUrl = `${proto}://${device.ipmi_ip}/viewer.jnlp?EXTPORT=-1&JNLPSTR=AppletRedirection`;
        }

        spawn(`"${javaws}"`, [jnlpUrl], { shell: true, detached: true, stdio: 'ignore' });
        console.log(`[JNLP] javaws ?ㅽ뻾: ${javaws} ${jnlpUrl}`);
        break;
      }
      case 'ipmiview': {
        const candidates = [
          'C:\\Program Files\\IPMI\\IPMIView\\IPMIView.exe',
          'C:\\Program Files (x86)\\IPMI\\IPMIView\\IPMIView.exe',
        ];
        const found = candidates.find(p => fs.existsSync(p));
        if (!found) return { success: false, error: 'IPMIView媛 ?ㅼ튂?섏? ?딆븯?듬땲??' };
        spawn(`"${found}"`, [device.ipmi_ip], { shell: true, detached: true, stdio: 'ignore' });
        break;
      }
      default:
        shell.openExternal(`https://${device.ipmi_ip}`);
    }
    return { success: true };
  } catch (e) { return { success: false, error: e.message }; }
});

// [Java] ?ㅼ튂??Java 紐⑸줉 ?먯?
ipcMain.handle('java:detect', async () => {
  try {
    const list = await javaManager.detectJavaInstallations();
    return { success: true, list };
  } catch (e) { return { success: false, error: e.message, list: [] }; }
});

// [Java] ?덉쇅 ?ъ씠???깅줉
ipcMain.handle('java:add-exception', async (_, { siteUrl }) => {
  try {
    const result = javaManager.addJavaExceptionSite(siteUrl);
    return { success: true, ...result };
  } catch (e) { return { success: false, error: e.message }; }
});

// [Java] ?덇굅??蹂댁븞 ?ㅼ젙 ?곸슜
ipcMain.handle('java:apply-legacy-config', async () => {
  try {
    const result = javaManager.applyLegacyJavaConfig();
    return { success: true, ...result };
  } catch (e) { return { success: false, error: e.message }; }
});

// [Java] java.security ?⑥튂 (?쒗븳 ?댁젣)
ipcMain.handle('java:patch-security', async (_, { javawsPath }) => {
  try {
    const result = await javaManager.patchJavaSecurity(javawsPath);
    return { success: true, ...result };
  } catch (e) {
    return { success: false, error: e.message };
  }
});

// [Java] 援щ쾭???ㅼ슫濡쒕뱶 ?뺣낫
ipcMain.handle('java:get-download-links', async () => {
  return javaManager.getLegacyJavaDownloadInfo();
});

// [?쒖뒪?? ?몃? URL ?닿린
ipcMain.handle('shell:open-url', async (_, url) => {
  shell.openExternal(url);
  return { success: true };
});

// [?ㅼ씠?쇰줈洹? ?뚯씪 ?닿린
ipcMain.handle('dialog:open-file', async (_, options) => {
  return dialog.showOpenDialog(mainWindow, options);
});

// ??? ???앸챸二쇨린 ?????????????????????????????????????????????????
app.whenReady().then(() => {
  // ?ㅼ슫濡쒕뱶 ?꾨즺 ???ㅽ뻾 ?뺤씤 ?앹뾽 湲곕뒫 異붽?
  session.defaultSession.on('will-download', (event, item, webContents) => {
    item.once('done', async (event, state) => {
      if (state === 'completed') {
        const filePath = item.getSavePath();
        const fileName = item.getFilename();
        const focusWindow = BrowserWindow.getFocusedWindow() || mainWindow;
        
        const { response } = await dialog.showMessageBox(focusWindow, {
          type: 'question',
          buttons: ['??, '?꾨땲??],
          defaultId: 0,
          title: '?ㅼ슫濡쒕뱶 ?꾨즺',
          message: `?뚯씪 ?ㅼ슫濡쒕뱶媛 ?꾨즺?섏뿀?듬땲??\n\n?뚯씪紐? ${fileName}\n\n吏湲????뚯씪???ㅽ뻾?섏떆寃좎뒿?덇퉴?`,
          cancelId: 1
        });

        if (response === 0) {
          shell.openPath(filePath).catch(err => {
            dialog.showErrorBox('?ㅽ뻾 ?ㅽ뙣', `?뚯씪???ㅽ뻾?섎뒗 以??ㅻ쪟媛 諛쒖깮?덉뒿?덈떎.\n寃쎈줈: ${filePath}\n?ㅻ쪟: ${err.message}`);
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
