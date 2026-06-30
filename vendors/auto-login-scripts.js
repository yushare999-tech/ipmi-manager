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

  function tryFill() {
    var users = querySelectorAllAll('#user')
      .concat(querySelectorAllAll('input[name="user"]'))
      .concat(querySelectorAllAll('input[type="text"]'));
    var passes = querySelectorAllAll('#password')
      .concat(querySelectorAllAll('input[name="password"]'))
      .concat(querySelectorAllAll('input[type="password"]'));

    var userEl = users[0];
    var passEl = passes[0];

    if (userEl && passEl) {
      // React/Vue 등 프레임워크 대응: nativeInputValueSetter 방식
      var nativeInputValueSetter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value').set;
      
      if (nativeInputValueSetter) {
        nativeInputValueSetter.call(userEl, ${u});
      } else {
        userEl.value = ${u};
      }
      userEl.dispatchEvent(new Event('input', { bubbles: true }));
      userEl.dispatchEvent(new Event('change', { bubbles: true }));

      if (nativeInputValueSetter) {
        nativeInputValueSetter.call(passEl, ${p});
      } else {
        passEl.value = ${p};
      }
      passEl.dispatchEvent(new Event('input', { bubbles: true }));
      passEl.dispatchEvent(new Event('change', { bubbles: true }));

      if (${autoSubmit}) {
        // Submit 버튼 탐색 후 클릭, 없으면 form submit
        var btn = querySelectorAllAll('button[type="submit"]')
          .concat(querySelectorAllAll('input[type="submit"]'))
          .concat(querySelectorAllAll('#btnOK'))
          .concat(querySelectorAllAll('.btn-primary'))[0];
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
      if (tryFill() || ++attempts > 15) clearInterval(timer);
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

  function tryFill() {
    var users = querySelectorAllAll('#username')
      .concat(querySelectorAllAll('input[name="username"]'))
      .concat(querySelectorAllAll('input[autocomplete="username"]'))
      .concat(querySelectorAllAll('input[type="text"]'));
    var passes = querySelectorAllAll('#password')
      .concat(querySelectorAllAll('input[name="password"]'))
      .concat(querySelectorAllAll('input[autocomplete="current-password"]'))
      .concat(querySelectorAllAll('input[type="password"]'));

    var userEl = users[0];
    var passEl = passes[0];

    if (userEl && passEl) {
      var nativeInputValueSetter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value').set;
      
      if (nativeInputValueSetter) {
        nativeInputValueSetter.call(userEl, ${u});
      } else {
        userEl.value = ${u};
      }
      userEl.dispatchEvent(new Event('input', { bubbles: true }));
      userEl.dispatchEvent(new Event('change', { bubbles: true }));

      if (nativeInputValueSetter) {
        nativeInputValueSetter.call(passEl, ${p});
      } else {
        passEl.value = ${p};
      }
      passEl.dispatchEvent(new Event('input', { bubbles: true }));
      passEl.dispatchEvent(new Event('change', { bubbles: true }));

      if (${autoSubmit}) {
        var btn = querySelectorAllAll('button[type="submit"]')
          .concat(querySelectorAllAll('#btn-login'))
          .concat(querySelectorAllAll('.btn-primary'))
          .concat(querySelectorAllAll('input[type="submit"]'))[0];
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
      if (tryFill() || ++attempts > 15) clearInterval(timer);
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

  function tryFill() {
    var users = querySelectorAllAll('input[name="name"]')
      .concat(querySelectorAllAll('input[name="username"]'))
      .concat(querySelectorAllAll('input[type="text"]'));
    var passes = querySelectorAllAll('input[name="pwd"]')
      .concat(querySelectorAllAll('input[name="password"]'))
      .concat(querySelectorAllAll('input[type="password"]'));

    var userEl = users[0];
    var passEl = passes[0];

    if (userEl && passEl) {
      var nativeInputValueSetter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value').set;
      
      if (nativeInputValueSetter) {
        nativeInputValueSetter.call(userEl, ${u});
      } else {
        userEl.value = ${u};
      }
      userEl.dispatchEvent(new Event('input', { bubbles: true }));
      userEl.dispatchEvent(new Event('change', { bubbles: true }));

      if (nativeInputValueSetter) {
        nativeInputValueSetter.call(passEl, ${p});
      } else {
        passEl.value = ${p};
      }
      passEl.dispatchEvent(new Event('input', { bubbles: true }));
      passEl.dispatchEvent(new Event('change', { bubbles: true }));

      if (${autoSubmit}) {
        var btn = querySelectorAllAll('input[type="submit"]')
          .concat(querySelectorAllAll('button[type="submit"]'))
          .concat(querySelectorAllAll('#login_word'))[0];
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
      if (tryFill() || ++attempts > 15) clearInterval(timer);
    }, 500);
  }
})();
`;
}

// ─── ASUS / ASRock ────────────────────────────────────────────────
function getAsusScript(u, p, autoSubmit) {
  return `
(function() {
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

  function tryFill() {
    var users = querySelectorAllAll('input[type="text"]');
    var passes = querySelectorAllAll('input[type="password"]');

    var userEl = users[0];
    var passEl = passes[0];

    if (userEl && passEl) {
      var nativeInputValueSetter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value').set;
      
      if (nativeInputValueSetter) {
        nativeInputValueSetter.call(userEl, ${u});
      } else {
        userEl.value = ${u};
      }
      userEl.dispatchEvent(new Event('input', { bubbles: true }));
      userEl.dispatchEvent(new Event('change', { bubbles: true }));

      if (nativeInputValueSetter) {
        nativeInputValueSetter.call(passEl, ${p});
      } else {
        passEl.value = ${p};
      }
      passEl.dispatchEvent(new Event('input', { bubbles: true }));
      passEl.dispatchEvent(new Event('change', { bubbles: true }));
      
      if (${autoSubmit}) {
        var btn = querySelectorAllAll('input[type="submit"]')
          .concat(querySelectorAllAll('button[type="submit"]'))[0];
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

// ─── Generic (공통 탐색) ──────────────────────────────────────────
function getGenericScript(u, p, autoSubmit) {
  return `
(function() {
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

  function tryFill() {
    var inputs = querySelectorAllAll('input');
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
    if (!userEl) userEl = querySelectorAllAll('input[type="text"]')[0];

    if (userEl && passEl) {
      var nativeInputValueSetter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value').set;
      
      if (nativeInputValueSetter) {
        nativeInputValueSetter.call(userEl, ${u});
      } else {
        userEl.value = ${u};
      }
      userEl.dispatchEvent(new Event('input', { bubbles: true }));
      userEl.dispatchEvent(new Event('change', { bubbles: true }));

      if (nativeInputValueSetter) {
        nativeInputValueSetter.call(passEl, ${p});
      } else {
        passEl.value = ${p};
      }
      passEl.dispatchEvent(new Event('input', { bubbles: true }));
      passEl.dispatchEvent(new Event('change', { bubbles: true }));

      if (${autoSubmit}) {
        var btn = querySelectorAllAll('button[type="submit"]')
          .concat(querySelectorAllAll('input[type="submit"]'))
          .concat(querySelectorAllAll('.btn-primary'))
          .concat(querySelectorAllAll('.login-btn'))[0];
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
