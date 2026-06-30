/**
 * IPMI Manager - 자동 로그인 전용 프리로드 스크립트 (login-preload.js)
 * 작성일: 2026-06-30
 * 기능:
 *   - 모든 프레임/iframe 내부에서 자율적으로 기동하여 로그인 폼을 탐지하고 안전하게 자동 로그인 수행.
 *   - 크로스오리진 및 프레임셋 구조에 따른 보안 차단을 완벽히 우회.
 */

const { ipcRenderer } = require('electron');

(function() {
  // 1. 메인 프로세스로부터 현재 창의 장비 로그인 정보 동기식 획득
  let loginInfo = null;
  try {
    loginInfo = ipcRenderer.sendSync('get-login-device');
  } catch (e) {
    console.error('[Preload] 장비 로그인 정보 획득 실패:', e);
  }

  if (!loginInfo || !loginInfo.device || !loginInfo.device.username) {
    return; // 자동 로그인 대상이 아니거나 정보가 없으면 조기 종료
  }

  const { device, autoSubmit } = loginInfo;
  const vendorKey = (device.vendor || 'generic').toLowerCase();

  let loginDone = false;
  let dashboardDetected = false;
  let checkTimer = null;

  // 2. 로그인 폼 탐지 및 처리 함수
  function attemptLogin() {
    if (loginDone) return;

    // 현재 프레임의 URL 체크
    const currentUrl = window.location.href.toLowerCase();

    // 대시보드 진입 감지 시 자동 새로고침 처리 (1회성)
    const dashIndicators = ['index.html?st', 'dashboard', 'sys_summary', 'main.html', 'rfc3986'];
    const isDashboard = dashIndicators.some(kw => currentUrl.includes(kw));
    if (isDashboard && !dashboardDetected) {
      dashboardDetected = true;
      if (checkTimer) clearInterval(checkTimer);
      
      const isIdrac7 = (device.version || '').toLowerCase().includes('idrac7') || 
                       (device.model || '').toLowerCase().includes('r620') ||
                       (device.version || '').startsWith('1.');

      // REST API 토큰으로 로그인된 경우 새로고침 시 세션이 만료되므로 새로고침을 생략해야 함
      const hasRestToken = currentUrl.includes('st1=') || currentUrl.includes('st2=');

      if (!isIdrac7 && !hasRestToken) {
        console.log('[Preload] 대시보드 진입 감지 → 1.5초 후 새로고침 실행');
        setTimeout(() => {
          window.location.reload();
        }, 1500);
      } else {
        const reason = hasRestToken ? 'REST 토큰 세션 유실 방지' : 'iDRAC 7 세션 유실 방지';
        console.log(`[Preload] 대시보드 진입 감지 → ${reason}를 위해 새로고침 생략`);
      }
      return;
    }

    // 벤더별 입력 필드 셀렉터 설정
    const userSelectors = {
      dell: ['#user', 'input[name="user"]', 'input[type="text"]'],
      hp: ['#username', 'input[name="username"]', 'input[autocomplete="username"]', 'input[type="text"]'],
      supermicro: ['input[name="name"]', 'input[name="username"]', 'input[type="text"]'],
      generic: ['input']
    };
    const passSelectors = {
      dell: ['#password', 'input[name="password"]', 'input[type="password"]'],
      hp: ['#password', 'input[name="password"]', 'input[autocomplete="current-password"]', 'input[type="password"]'],
      supermicro: ['input[name="pwd"]', 'input[name="password"]', 'input[type="password"]'],
      generic: ['input[type="password"]']
    };

    const uSels = userSelectors[vendorKey] || userSelectors.generic;
    const pSels = passSelectors[vendorKey] || passSelectors.generic;

    // 현재 프레임 내에서 엘리먼트 검색
    let userEl = findElement(uSels, false);
    let passEl = findElement(pSels, true);

    if (userEl && passEl) {
      console.log('[Preload] 로그인 필드 감지 성공! 자율 로그인 프로세스 개시.');
      loginDone = true;
      if (checkTimer) clearInterval(checkTimer);

      // A. 도메인 셀렉트 박스 강제 제어 (R620 등 사설 도메인 지정 방지)
      const domainSelect = document.getElementById('domainDisp') || document.querySelector('select[name="domainDisp"]');
      if (domainSelect && domainSelect.selectedIndex !== 0) {
        console.log('[Preload] 도메인 설정 변경 감지 → Local(This iDRAC)로 강제 복구');
        domainSelect.selectedIndex = 0;
        domainSelect.dispatchEvent(new Event('change', { bubbles: true }));
      }

      // B. ID 입력 필드 주입 및 캐릭터별 키 이벤트 순차 발화
      fillFieldWithEvents(userEl, device.username);

      // C. PW 입력 필드 주입 및 캐릭터별 키 이벤트 순차 발화
      fillFieldWithEvents(passEl, device.password || '');

      // D. 자동 제출 처리
      if (autoSubmit) {
        console.log('[Preload] 자동 제출 실행');
        setTimeout(() => {
          submitForm(userEl, vendorKey);
        }, 100);
      }
    }
  }

  // 필드 검색 헬퍼
  function findElement(selectors, isPassword) {
    for (let i = 0; i < selectors.length; i++) {
      const els = document.querySelectorAll(selectors[i]);
      for (let j = 0; j < els.length; j++) {
        const el = els[j];
        if (selectors[i] === 'input') {
          const t = (el.type || '').toLowerCase();
          const n = (el.name || el.id || '').toLowerCase();
          if (isPassword && t === 'password') return el;
          if (!isPassword && (t === 'text' || t === 'email') && 
              (n.includes('user') || n.includes('name') || n.includes('login') || n.includes('id'))) {
            return el;
          }
        } else {
          return el;
        }
      }
    }
    // Fallback
    if (isPassword) {
      return document.querySelector('input[type="password"]');
    } else {
      return document.querySelector('input[type="text"]');
    }
  }

  // 값 주입 및 키보드/인풋 이벤트 에뮬레이션
  function fillFieldWithEvents(el, value) {
    el.focus();

    // 1. 프로토타입 기반의 즉각적 값 강제 주입
    const getProto = Object.getPrototypeOf || function(obj) { return obj.__proto__; };
    const proto = getProto(el);
    const desc = Object.getOwnPropertyDescriptor(proto, 'value');
    const setter = desc ? desc.set : null;
    if (setter) {
      setter.call(el, value);
    } else {
      el.value = value;
    }

    // 2. 인풋 및 변경 이벤트 발생
    el.dispatchEvent(new Event('focus', { bubbles: true }));
    el.dispatchEvent(new Event('input', { bubbles: true }));
    el.dispatchEvent(new Event('change', { bubbles: true }));

    // 3. 캐릭터별 키보드 이벤트 순차 발화 (iDRAC 등의 보안 해시 빌더 작동 보장)
    for (let i = 0; i < value.length; i++) {
      const char = value[i];
      const code = char.charCodeAt(0);
      el.dispatchEvent(new KeyboardEvent('keydown', { bubbles: true, key: char, char: char, keyCode: code }));
      el.dispatchEvent(new KeyboardEvent('keypress', { bubbles: true, key: char, char: char, keyCode: code }));
      el.dispatchEvent(new KeyboardEvent('keyup', { bubbles: true, key: char, char: char, keyCode: code }));
    }

    el.blur();
  }

  // 폼 제출 헬퍼
  function submitForm(activeEl, vendor) {
    const submitBtnSelectors = {
      dell: ['button[type="submit"]', 'input[type="submit"]', '#btnOK', '.btn-primary'],
      hp: ['button[type="submit"]', '#btn-login', '.btn-primary', 'input[type="submit"]'],
      supermicro: ['input[type="submit"]', 'button[type="submit"]', '#login_word'],
      generic: ['button[type="submit"]', 'input[type="submit"]', '.btn-primary', '.login-btn']
    };

    const selectors = submitBtnSelectors[vendor] || submitBtnSelectors.generic;
    let btn = null;
    for (let i = 0; i < selectors.length; i++) {
      const elements = document.querySelectorAll(selectors[i]);
      if (elements.length > 0) {
        btn = elements[0];
        break;
      }
    }

    if (btn) {
      btn.focus();
      btn.click();
      if (vendor === 'dell' && typeof window.frmSubmit === 'function') {
        window.frmSubmit();
      }
    } else if (activeEl && activeEl.form) {
      activeEl.form.submit();
    } else {
      // 최종 폴백: Enter 키 전송
      const enterDown = new KeyboardEvent('keydown', { bubbles: true, cancelable: true, key: 'Enter', keyCode: 13 });
      const enterUp = new KeyboardEvent('keyup', { bubbles: true, cancelable: true, key: 'Enter', keyCode: 13 });
      document.dispatchEvent(enterDown);
      document.dispatchEvent(enterUp);
    }
  }

  // 3. 로딩 상태에 연동하여 주기적 탐색 (DOMContentLoaded 이후 실행되도록 주기 400ms 설정)
  checkTimer = setInterval(attemptLogin, 400);

  window.addEventListener('DOMContentLoaded', attemptLogin);
  window.addEventListener('load', attemptLogin);
})();
