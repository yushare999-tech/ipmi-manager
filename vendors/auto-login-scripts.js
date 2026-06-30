/**
 * vendors/auto-login-scripts.js
 * 벤더별 IPMI 자동 로그인 JS 스크립트 생성기
 * 작성일: 2026-06-25
 * 변경이력:
 *   - 2026-06-25: 최초 작성 (Dell iDRAC, HP iLO, SuperMicro, Generic)
 */

/**
 * 벤더에 맞는 자동 로그인 스크립트 반환
 * @param {string} vendor - 벤더명 (dell/hp/hpe/supermicro/asus/asrock/generic)
 * @param {string} username
 * @param {string} password
 * @returns {string} 브라우저에서 실행할 JS 코드 문자열
 */
function getLoginScript(vendor, username, password, autoSubmit = false) {
  const u = JSON.stringify(username);
  const p = JSON.stringify(password);
  const submitFlag = !!autoSubmit;

  switch ((vendor || '').toLowerCase()) {
    case 'dell':
      return getDellScript(u, p, submitFlag);
    case 'hp':
    case 'hpe':
      return getHpScript(u, p, submitFlag);
    case 'supermicro':
      return getSupermicroScript(u, p, submitFlag);
    case 'asus':
    case 'asrock':
      return getAsusScript(u, p, submitFlag);
    default:
      return getGenericScript(u, p, submitFlag);
  }
}

// ─── Dell iDRAC ──────────────────────────────────────────────────
// iDRAC6/7/8: #user / #password / form submit
// iDRAC9: React 기반 SPA, 셀렉터 다름
// autoSubmit이 true인 경우에만 자동 클릭/제출 수행
function getDellScript(u, p, autoSubmit) {
  return `
(function() {
  function tryFill() {
    // iDRAC 6/7/8
    var userEl = document.getElementById('user') ||
                 document.querySelector('input[name="user"]') ||
                 document.querySelector('input[type="text"]');
    var passEl = document.getElementById('password') ||
                 document.querySelector('input[name="password"]') ||
                 document.querySelector('input[type="password"]');

    if (userEl && passEl) {
      // React/Vue 등 프레임워크 대응: nativeInputValueSetter 방식
      var nativeInputValueSetter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value').set;
      nativeInputValueSetter.call(userEl, ${u});
      userEl.dispatchEvent(new Event('input', { bubbles: true }));
      userEl.dispatchEvent(new Event('change', { bubbles: true }));

      nativeInputValueSetter.call(passEl, ${p});
      passEl.dispatchEvent(new Event('input', { bubbles: true }));
      passEl.dispatchEvent(new Event('change', { bubbles: true }));

      if (${autoSubmit}) {
        // Submit 버튼 탐색 후 클릭, 없으면 form submit
        var btn = document.querySelector('button[type="submit"]') ||
                  document.querySelector('input[type="submit"]') ||
                  document.querySelector('#btnOK') ||
                  document.querySelector('.btn-primary');
        if (btn) {
          btn.click();
        } else if (userEl.form) {
          userEl.form.submit();
        }
      }
      return true;
    }
    return false;
  }

  // 즉시 시도 후 안되면 재시도
  if (!tryFill()) {
    var attempts = 0;
    var timer = setInterval(function() {
      if (tryFill() || ++attempts > 10) clearInterval(timer);
    }, 500);
  }
})();
`;
}

// ─── HP iLO ──────────────────────────────────────────────────────
// iLO4/5: #username / #password
function getHpScript(u, p, autoSubmit) {
  return `
(function() {
  function tryFill() {
    var userEl = document.getElementById('username') ||
                 document.querySelector('input[name="username"]') ||
                 document.querySelector('input[autocomplete="username"]') ||
                 document.querySelector('input[type="text"]');
    var passEl = document.getElementById('password') ||
                 document.querySelector('input[name="password"]') ||
                 document.querySelector('input[autocomplete="current-password"]') ||
                 document.querySelector('input[type="password"]');

    if (userEl && passEl) {
      var nativeInputValueSetter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value').set;
      nativeInputValueSetter.call(userEl, ${u});
      userEl.dispatchEvent(new Event('input', { bubbles: true }));
      userEl.dispatchEvent(new Event('change', { bubbles: true }));

      nativeInputValueSetter.call(passEl, ${p});
      passEl.dispatchEvent(new Event('input', { bubbles: true }));
      passEl.dispatchEvent(new Event('change', { bubbles: true }));

      if (${autoSubmit}) {
        var btn = document.querySelector('button[type="submit"]') ||
                  document.querySelector('#btn-login') ||
                  document.querySelector('.btn-primary') ||
                  document.querySelector('input[type="submit"]');
        if (btn) {
          btn.click();
        } else if (userEl.form) {
          userEl.form.submit();
        }
      }
      return true;
    }
    return false;
  }

  if (!tryFill()) {
    var attempts = 0;
    var timer = setInterval(function() {
      if (tryFill() || ++attempts > 10) clearInterval(timer);
    }, 500);
  }
})();
`;
}

// ─── SuperMicro ───────────────────────────────────────────────────
// name="name" / name="pwd"
function getSupermicroScript(u, p, autoSubmit) {
  return `
(function() {
  function tryFill() {
    var userEl = document.querySelector('input[name="name"]') ||
                 document.querySelector('input[name="username"]') ||
                 document.querySelector('input[type="text"]');
    var passEl = document.querySelector('input[name="pwd"]') ||
                 document.querySelector('input[name="password"]') ||
                 document.querySelector('input[type="password"]');

    if (userEl && passEl) {
      userEl.value = ${u};
      passEl.value = ${p};

      if (${autoSubmit}) {
        var btn = document.querySelector('input[type="submit"]') ||
                  document.querySelector('button[type="submit"]') ||
                  document.querySelector('#login_word');
        if (btn) {
          btn.click();
        } else if (userEl.form) {
          userEl.form.submit();
        }
      }
      return true;
    }
    return false;
  }

  if (!tryFill()) {
    var attempts = 0;
    var timer = setInterval(function() {
      if (tryFill() || ++attempts > 10) clearInterval(timer);
    }, 500);
  }
})();
`;
}

// ─── ASUS / ASRock ────────────────────────────────────────────────
function getAsusScript(u, p, autoSubmit) {
  return `
(function() {
  function tryFill() {
    var userEl = document.querySelector('input[type="text"]');
    var passEl = document.querySelector('input[type="password"]');

    if (userEl && passEl) {
      userEl.value = ${u};
      passEl.value = ${p};
      
      if (${autoSubmit}) {
        var btn = document.querySelector('input[type="submit"]') ||
                  document.querySelector('button[type="submit"]');
        if (btn) btn.click();
        else if (userEl.form) userEl.form.submit();
      }
      return true;
    }
    return false;
  }

  if (!tryFill()) {
    var attempts = 0;
    var timer = setInterval(function() {
      if (tryFill() || ++attempts > 10) clearInterval(timer);
    }, 500);
  }
})();
`;
}

// ─── Generic (공통 탐색) ──────────────────────────────────────────
function getGenericScript(u, p, autoSubmit) {
  return `
(function() {
  function tryFill() {
    var inputs = document.querySelectorAll('input');
    var userEl = null, passEl = null;

    inputs.forEach(function(el) {
      var t = (el.type || '').toLowerCase();
      var n = (el.name || el.id || '').toLowerCase();
      if (!passEl && (t === 'password')) passEl = el;
      if (!userEl && (t === 'text' || t === 'email') &&
          (n.includes('user') || n.includes('name') || n.includes('login') || n.includes('id'))) {
        userEl = el;
      }
    });
    // fallback: 첫번째 text input
    if (!userEl) userEl = document.querySelector('input[type="text"]');

    if (userEl && passEl) {
      var nativeInputValueSetter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value').set;
      nativeInputValueSetter.call(userEl, ${u});
      userEl.dispatchEvent(new Event('input', { bubbles: true }));
      userEl.dispatchEvent(new Event('change', { bubbles: true }));

      nativeInputValueSetter.call(passEl, ${p});
      passEl.dispatchEvent(new Event('input', { bubbles: true }));
      passEl.dispatchEvent(new Event('change', { bubbles: true }));

      if (${autoSubmit}) {
        var btn = document.querySelector('button[type="submit"]') ||
                  document.querySelector('input[type="submit"]') ||
                  document.querySelector('.btn-primary') ||
                  document.querySelector('.login-btn');
        if (btn) btn.click();
        else if (userEl.form) userEl.form.submit();
      }
      return true;
    }
    return false;
  }

  if (!tryFill()) {
    var attempts = 0;
    var timer = setInterval(function() {
      if (tryFill() || ++attempts > 15) clearInterval(timer);
    }, 500);
  }
})();
`;
}

module.exports = { getLoginScript };
