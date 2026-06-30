/**
 * IPMI Manager - 렌더러 메인 로직
 * 작성일: 2026-06-25
 * 변경이력:
 *   - 2026-06-25: 최초 작성 (장비 관리, Java 탐지, API 연동, 설정 저장)
 *   - 2026-06-25: IPMI 자동 로그인 기능 적용 (IPMI 페이지, HTML5 KVM, JNLP 선행수행)
 */

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// 상태 (State)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
let state = {
  devices:       [],   // 장비 목록
  javaList:      [],   // 탐지된 Java 목록
  selectedJava:  null, // 선택된 Java
  config: {
    apiBaseUrl:       '',
    apiToken:         '',
    apiEndpoint:      '/ipmi/devices',
    defaultKvmMethod: 'html5',
    javawsPath:       'C:\\Program Files\\Java\\jre1.8.0_441\\bin\\javaws.exe',
    autoSubmit:       false,
    enableDevTools:   false,
  },
  editingDeviceId: null,
};

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// 유틸 함수
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
const $ = (id) => document.getElementById(id);
const uid = () => Math.random().toString(36).slice(2, 10);

function showResult(elId, type, msg) {
  const el = $(elId);
  el.className = `result-box ${type}`;
  el.textContent = msg;
}

function hideResult(elId) {
  const el = $(elId);
  el.className = 'result-box hidden';
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// 탭 전환
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
document.querySelectorAll('.nav-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    const tab = btn.dataset.tab;
    document.querySelectorAll('.nav-btn').forEach(b => b.classList.remove('active'));
    document.querySelectorAll('.tab-content').forEach(t => t.classList.remove('active'));
    btn.classList.add('active');
    $(`tab-${tab}`).classList.add('active');
  });
});

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// 장비 목록 렌더링
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
const VENDOR_LABELS = {
  dell: 'Dell iDRAC', hp: 'HP iLO', hpe: 'HP iLO',
  supermicro: 'SuperMicro', asus: 'ASUS', asrock: 'ASRock', generic: '기타',
};

function renderDevices() {
  const search  = $('device-search').value.toLowerCase();
  const vendor  = $('vendor-filter').value;
  const grid    = $('device-grid');

  const filtered = state.devices.filter(d => {
    const matchSearch = !search || d.name.toLowerCase().includes(search) || d.ipmi_ip.includes(search);
    const matchVendor = !vendor || d.vendor === vendor;
    return matchSearch && matchVendor;
  });

  if (filtered.length === 0) {
    grid.innerHTML = `<div class="empty-state">장비가 없습니다.<br>장비 추가 버튼을 눌러 등록하세요.</div>`;
    return;
  }

  grid.innerHTML = filtered.map(d => `
    <div class="device-card" id="card-${d.id}">
      <div class="device-card-top">
        <div>
          <div class="device-name">${d.name}</div>
          <div class="device-ip">${d.ipmi_ip}</div>
        </div>
        <span class="vendor-badge vendor-${d.vendor}">${VENDOR_LABELS[d.vendor] || d.vendor}</span>
      </div>
      ${d.model ? `<div class="device-note">모델: ${d.model}</div>` : ''}
      ${d.note  ? `<div class="device-note">${d.note}</div>` : ''}
      ${d.username ? `
      <div class="device-account">
        <span>ID: <strong>${d.username}</strong></span>
        <span class="separator">|</span>
        <span class="pw-wrapper">
          PW: 
          <span class="pw-text" id="pw-text-${d.id}" data-password="${d.password || ''}">••••••••</span>
          <button class="btn-toggle-card-pw" onclick="toggleCardPw('${d.id}')">👁️</button>
        </span>
      </div>
      ` : ''}
      <div class="device-actions">
        <button class="btn btn-primary btn-sm" onclick="connectHtml5('${d.id}')">🖥️ HTML5 KVM</button>
        <button class="btn btn-secondary btn-sm" onclick="connectJnlp('${d.id}')">☕ JNLP</button>
        <button class="btn btn-secondary btn-sm" onclick="openIpmi('${d.id}')">🌐 IPMI 페이지</button>
        <button class="btn btn-secondary btn-sm" onclick="editDevice('${d.id}')">✏️</button>
        <button class="btn btn-danger btn-sm" onclick="deleteDevice('${d.id}')">🗑️</button>
      </div>
    </div>
  `).join('');
}

// 검색/필터 이벤트
$('device-search').addEventListener('input', renderDevices);
$('vendor-filter').addEventListener('change', renderDevices);

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// KVM 접속
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// HTML5 KVM: 계정 정보 있으면 자동 로그인 후 KVM 진입
async function connectHtml5(id) {
  const d = state.devices.find(x => x.id === id);
  if (!d) return;
  if (d.username) {
    const res = await window.ipmiAPI.openHtml5KvmAutoLogin(d);
    if (!res.success) alert('KVM 연결 실패: ' + res.error);
  } else {
    const res = await window.ipmiAPI.openHtml5Kvm(d);
    if (!res.success) alert('KVM 연결 실패: ' + res.error);
  }
}

// JNLP: Dell iDRAC은 main process에서 REST API 직접 로그인 후 ST1/ST2 포함 URL로 실행
async function connectJnlp(id) {
  const d = state.devices.find(x => x.id === id);
  if (!d) return;
  const javawsPath = state.config.javawsPath;
  const res = await window.ipmiAPI.launchExternal(d, 'jnlp', javawsPath);
  if (!res.success) alert('JNLP 실행 실패:\n' + res.error);
}

// IPMI 페이지: 계정 정보 있으면 자동 로그인
async function openIpmi(id) {
  const d = state.devices.find(x => x.id === id);
  if (!d) return;
  if (d.username) {
    await window.ipmiAPI.openIpmiAutoLogin(d);
  } else {
    const proto = d.https !== false ? 'https' : 'http';
    await window.ipmiAPI.openUrl(`${proto}://${d.ipmi_ip}`);
  }
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// 장비 추가/수정 모달
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
function openModal(device = null) {
  state.editingDeviceId = device ? device.id : null;
  $('modal-title').textContent = device ? '장비 수정' : '장비 추가';
  $('form-name').value     = device?.name     || '';
  $('form-ip').value       = device?.ipmi_ip  || '';
  $('form-vendor').value   = device?.vendor   || 'dell';
  $('form-model').value    = device?.model    || '';
  $('form-username').value = device?.username || '';
  $('form-password').value = device?.password || '';
  $('form-note').value     = device?.note     || '';
  $('form-https').checked  = device?.https !== false;
  
  // 비밀번호 토글 상태 초기화 (항상 가림)
  $('form-password').type = 'password';
  $('btn-toggle-password').textContent = '👁️';

  $('device-modal').classList.remove('hidden');
}

function closeModal() {
  $('device-modal').classList.add('hidden');
  state.editingDeviceId = null;
  
  // 비밀번호 토글 상태 초기화 (항상 가림)
  $('form-password').type = 'password';
  $('btn-toggle-password').textContent = '👁️';
}

$('btn-add-device').addEventListener('click', () => openModal());
$('modal-close').addEventListener('click', closeModal);
$('btn-modal-cancel').addEventListener('click', closeModal);

$('btn-toggle-password').addEventListener('click', () => {
  const pwdInput = $('form-password');
  const toggleBtn = $('btn-toggle-password');
  if (pwdInput.type === 'password') {
    pwdInput.type = 'text';
    toggleBtn.textContent = '🙈';
  } else {
    pwdInput.type = 'password';
    toggleBtn.textContent = '👁️';
  }
});

$('btn-modal-save').addEventListener('click', async () => {
  const name = $('form-name').value.trim();
  const ip   = $('form-ip').value.trim();
  if (!name || !ip) { alert('장비명과 IP는 필수입니다.'); return; }

  const deviceData = {
    id:       state.editingDeviceId || uid(),
    name,
    ipmi_ip:  ip,
    vendor:   $('form-vendor').value,
    model:    $('form-model').value.trim(),
    username: $('form-username').value.trim(),
    password: $('form-password').value,
    note:     $('form-note').value.trim(),
    https:    $('form-https').checked,
  };

  if (state.editingDeviceId) {
    const idx = state.devices.findIndex(d => d.id === state.editingDeviceId);
    if (idx >= 0) state.devices[idx] = deviceData;
  } else {
    state.devices.push(deviceData);
  }

  await saveDevices();
  renderDevices();
  closeModal();
});

function editDevice(id) {
  const d = state.devices.find(x => x.id === id);
  if (d) openModal(d);
}

async function deleteDevice(id) {
  if (!confirm('이 장비를 삭제하시겠습니까?')) return;
  state.devices = state.devices.filter(d => d.id !== id);
  await saveDevices();
  renderDevices();
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// 설정 저장/로드
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
async function saveDevices() {
  const cfg = await window.ipmiAPI.loadConfig();
  cfg.devices = state.devices;
  await window.ipmiAPI.saveConfig(cfg);
}

async function loadAll() {
  const cfg = await window.ipmiAPI.loadConfig();
  state.devices = cfg.devices || [];
  if (cfg.apiBaseUrl)       { state.config.apiBaseUrl = cfg.apiBaseUrl; $('api-base-url').value = cfg.apiBaseUrl; }
  if (cfg.apiToken)         { state.config.apiToken   = cfg.apiToken;   $('api-token').value    = cfg.apiToken; }
  if (cfg.apiEndpoint)      { state.config.apiEndpoint = cfg.apiEndpoint; $('api-endpoint').value = cfg.apiEndpoint; }
  if (cfg.defaultKvmMethod) { state.config.defaultKvmMethod = cfg.defaultKvmMethod; $('default-kvm-method').value = cfg.defaultKvmMethod; }
  if (cfg.javawsPath)       { state.config.javawsPath = cfg.javawsPath; $('javaws-path').value  = cfg.javawsPath; }
  
  if (cfg.autoSubmit !== undefined) {
    state.config.autoSubmit = cfg.autoSubmit;
    if (cfg.autoSubmit) {
      $('auto-submit-true').checked = true;
    } else {
      $('auto-submit-false').checked = true;
    }
  }
  if (cfg.enableDevTools !== undefined) {
    state.config.enableDevTools = cfg.enableDevTools;
    $('enable-devtools').checked = cfg.enableDevTools;
  }

  renderDevices();
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Java 관련
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
$('btn-detect-java').addEventListener('click', async () => {
  $('java-list').innerHTML = '<div class="empty-state">탐지 중...</div>';
  updateJavaStatus('loading', 'Java 탐지 중...');

  const res = await window.ipmiAPI.detectJava();
  state.javaList = res.list || [];

  if (!res.success || state.javaList.length === 0) {
    $('java-list').innerHTML = `
      <div class="empty-state">
        Java를 찾을 수 없습니다.<br>
        아래 다운로드 안내를 참고해 Java 8을 설치하세요.
      </div>`;
    updateJavaStatus('error', 'Java 없음');
    return;
  }

  renderJavaList();
  // 가장 좋은 호환 버전으로 상태 업데이트
  const best = state.javaList.find(j => j.compat?.level === 'best' || j.compat?.level === 'good');
  if (best) updateJavaStatus('success', `Java ${best.version}`);
  else      updateJavaStatus('warning', `Java ${state.javaList[0].version} (설정 필요)`);
});

function renderJavaList() {
  $('java-list').innerHTML = state.javaList.map(j => `
    <div class="java-item">
      <div class="java-item-left">
        <span class="java-version">☕ Java ${j.version}</span>
        <span class="java-path">${j.javaHome || '경로 불명'}</span>
        <span class="java-path">${j.hasJavaws ? '✅ javaws.exe 있음' : '❌ javaws.exe 없음'}</span>
      </div>
      <div style="display:flex;flex-direction:column;align-items:flex-end;gap:6px;">
        <span class="java-compat compat-${j.compat?.level || 'unknown'}">${j.compat?.label || '알 수 없음'}</span>
        ${j.hasJavaws ? `<button class="btn btn-secondary btn-sm" onclick="useThisJava('${j.javawsExe?.replace(/\\/g, '\\\\') || ''}')">이 버전 사용</button>` : ''}
      </div>
    </div>
  `).join('');
}

function useThisJava(javawsPath) {
  state.config.javawsPath = javawsPath;
  $('javaws-path').value  = javawsPath;
  alert(`javaws 경로가 설정되었습니다:\n${javawsPath}`);
}

$('btn-apply-legacy').addEventListener('click', async () => {
  const res = await window.ipmiAPI.applyLegacyJava();
  if (res.success) {
    showResult('legacy-result', 'success',
      `✅ 레거시 설정 적용 완료\n파일: ${res.file}\nTLS 1.0/1.1 허용, 보안 레벨 조정됨`);
  } else {
    showResult('legacy-result', 'error', `❌ 실패: ${res.error}`);
  }
});

$('btn-patch-java-security').addEventListener('click', async () => {
  if (!confirm('이 작업은 Java JRE 내부의 보안 파일(java.security)을 수정하여 TLS 1.0/1.1 및 MD5/SHA1 차단을 제거합니다.\n\n수정 중 Windows 관리자 권한(UAC) 확인창이 표시됩니다. 계속하시겠습니까?')) {
    return;
  }

  const javawsPath = state.config.javawsPath;
  if (!javawsPath) {
    alert('환경설정 탭에서 먼저 올바른 javaws.exe 경로를 지정해주세요.');
    return;
  }

  showResult('legacy-result', 'info', '⏳ Java 보안 설정을 패치하는 중... (관리자 권한 승인 대기)');

  try {
    const res = await window.ipmiAPI.patchJavaSecurity(javawsPath);
    if (res.success) {
      showResult('legacy-result', 'success', `✅ ${res.message}\n대상 파일: ${res.file}`);
    } else {
      showResult('legacy-result', 'error', `❌ 실패: ${res.error}`);
    }
  } catch (e) {
    showResult('legacy-result', 'error', `❌ 예외 발생: ${e.message || e}`);
  }
});

$('btn-add-exception').addEventListener('click', async () => {
  const site = $('exception-site-input').value.trim();
  if (!site) return;
  const res = await window.ipmiAPI.addJavaException(site);
  if (res.success) {
    showResult('exception-result', 'success',
      `✅ 등록 완료 (신규 ${res.added}개, 전체 ${res.total}개)\n파일: ${res.file}`);
    $('exception-site-input').value = '';
    // IPMI 장비라면 Java 예외 자동 등록 유도
  } else {
    showResult('exception-result', 'error', `❌ 실패: ${res.error}`);
  }
});

// 다운로드 링크 로드
async function loadJavaDownloadLinks() {
  const links = await window.ipmiAPI.getJavaDownloadInfo();
  $('java-download-links').innerHTML = links.map(l => `
    <div class="download-item">
      <div class="download-item-info">
        <span class="download-title">${l.version}</span>
        <span class="download-note">${l.note}</span>
      </div>
      <button class="btn btn-secondary btn-sm" onclick="window.ipmiAPI.openUrl('${l.url}')">🔗 다운로드</button>
    </div>
  `).join('');
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// API 연동
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
$('btn-test-api').addEventListener('click', async () => {
  const baseUrl  = $('api-base-url').value.trim();
  const token    = $('api-token').value.trim();
  const endpoint = $('api-endpoint').value.trim();

  if (!baseUrl) { alert('API Base URL을 입력하세요.'); return; }

  showResult('api-test-result', 'info', '⏳ 연결 테스트 중...');

  try {
    const headers = { 'Content-Type': 'application/json' };
    if (token) headers['Authorization'] = `Bearer ${token}`;

    const response = await fetch(`${baseUrl}${endpoint}`, { headers });
    const data = await response.json();

    if (response.ok) {
      const count = data.devices?.length || 0;
      showResult('api-test-result', 'success',
        `✅ 연결 성공! (HTTP ${response.status})\n장비 수: ${count}개`);
    } else {
      showResult('api-test-result', 'error', `❌ HTTP ${response.status}: ${response.statusText}`);
    }
  } catch (e) {
    showResult('api-test-result', 'error', `❌ 연결 실패: ${e.message}`);
  }
});

$('btn-import-api').addEventListener('click', async () => {
  const baseUrl  = state.config.apiBaseUrl;
  const token    = state.config.apiToken;
  const endpoint = state.config.apiEndpoint;

  if (!baseUrl) {
    alert('먼저 API 연동 탭에서 API 서버를 설정하고 저장해주세요.');
    return;
  }

  try {
    const headers = {};
    if (token) headers['Authorization'] = `Bearer ${token}`;
    const response = await fetch(`${baseUrl}${endpoint}`, { headers });
    const data = await response.json();

    if (data.devices && Array.isArray(data.devices)) {
      const imported = data.devices.map(d => ({ ...d, id: d.id || uid() }));
      // 기존 장비와 병합 (id 기준 중복 제거)
      const existingIds = new Set(state.devices.map(d => d.id));
      const newDevices = imported.filter(d => !existingIds.has(d.id));
      state.devices = [...state.devices, ...newDevices];
      await saveDevices();
      renderDevices();
      alert(`✅ ${newDevices.length}개 장비를 가져왔습니다.`);
    }
  } catch (e) {
    alert(`API 가져오기 실패: ${e.message}`);
  }
});

$('btn-save-api').addEventListener('click', async () => {
  state.config.apiBaseUrl  = $('api-base-url').value.trim();
  state.config.apiToken    = $('api-token').value.trim();
  state.config.apiEndpoint = $('api-endpoint').value.trim();
  const cfg = await window.ipmiAPI.loadConfig();
  Object.assign(cfg, state.config);
  await window.ipmiAPI.saveConfig(cfg);
  alert('✅ API 설정이 저장되었습니다.');
});

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// 환경설정
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
$('btn-browse-javaws').addEventListener('click', async () => {
  const result = await window.ipmiAPI.openFile({
    title: 'javaws.exe 선택',
    filters: [{ name: 'Executable', extensions: ['exe'] }],
  });
  if (!result.canceled && result.filePaths.length > 0) {
    $('javaws-path').value = result.filePaths[0];
  }
});

$('btn-save-settings').addEventListener('click', async () => {
  state.config.defaultKvmMethod = $('default-kvm-method').value;
  state.config.javawsPath       = $('javaws-path').value.trim();
  state.config.autoSubmit       = $('auto-submit-true').checked;
  state.config.enableDevTools   = $('enable-devtools').checked;
  
  const cfg = await window.ipmiAPI.loadConfig();
  Object.assign(cfg, state.config);
  await window.ipmiAPI.saveConfig(cfg);
  alert('✅ 환경설정이 저장되었습니다.');
});

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// 사이드바 Java 상태 표시
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
function updateJavaStatus(level, text) {
  const dot  = $('java-status-dot');
  const span = $('java-status-text');
  dot.className  = `status-dot ${level}`;
  span.textContent = text;
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// 앱 초기화
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
async function init() {
  await loadAll();
  await loadJavaDownloadLinks();

  // 시작 시 Java 빠른 탐지
  updateJavaStatus('loading', 'Java 확인 중...');
  const res = await window.ipmiAPI.detectJava();
  if (res.success && res.list.length > 0) {
    state.javaList = res.list;
    const best = res.list.find(j => j.compat?.level === 'best' || j.compat?.level === 'good');
    if (best) updateJavaStatus('success', `Java ${best.version}`);
    else      updateJavaStatus('warning', `Java ${res.list[0].version}`);
  } else {
    updateJavaStatus('error', 'Java 없음');
  }
}

init();

// 카드 내 비밀번호 토글
window.toggleCardPw = function(id) {
  const pwSpan = document.getElementById(`pw-text-${id}`);
  if (!pwSpan) return;
  const btn = pwSpan.nextElementSibling;
  const rawPw = pwSpan.dataset.password;
  if (pwSpan.textContent === '••••••••') {
    pwSpan.textContent = rawPw;
    btn.textContent = '🙈';
  } else {
    pwSpan.textContent = '••••••••';
    btn.textContent = '👁️';
  }
};
