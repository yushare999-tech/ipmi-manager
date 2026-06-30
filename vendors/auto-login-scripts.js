/**
 * vendors/auto-login-scripts.js
 * 벤더별 IPMI 자동 로그인 JS 스크립트 생성기
 * 작성일: 2026-06-25
 * 변경이력:
 *   - 2026-06-25: 최초 작성 (Dell iDRAC, HP iLO, SuperMicro, Generic)
 *   - 2026-06-30: iframe 내 크로스 컨텍스트 대응을 위한 Object.getPrototypeOf 기반 setter 취득 적용 및 focus/blur 이벤트 보완
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
      var getProto = Object.getPrototypeOf || function(obj) { return obj.__proto__; };
      
      // User Input Fill
      var userProto = getProto(userEl);
      var userSetterDesc = Object.getOwnPropertyDescriptor(userProto, 'value');
      var userSetter = userSetterDesc ? userSetterDesc.set : null;
      
      userEl.focus();
      if (userSetter) {
        userSetter.call(userEl, ${u});
      } else {
        userEl.value = ${u};
      }
      userEl.dispatchEvent(new Event('input', { bubbles: true }));
      userEl.dispatchEvent(new Event('change', { bubbles: true }));
      userEl.blur();

      // Password Input Fill
      var passProto = getProto(passEl);
      var passSetterDesc = Object.getOwnPropertyDescriptor(passProto, 'value');
      var passSetter = passSetterDesc ? passSetterDesc.set : null;

      passEl.focus();
      if (passSetter) {
        passSetter.call(passEl, ${p});
      } else {
        passEl.value = ${p};
      }
      passEl.dispatchEvent(new Event('input', { bubbles: true }));
      passEl.dispatchEvent(new Event('change', { bubbles: true }));
      passEl.blur();

      if (${autoSubmit}) {
        var btn = querySelectorAllAll('button[type="submit"]')
          .concat(querySelectorAllAll('input[type="submit"]'))
          .concat(querySelectorAllAll('#btnOK'))
          .concat(querySelectorAllAll('.btn-primary'))[0];
        if (btn) {
          btn.focus();
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

// ─── HP iLO ──────────────────────────────────────────────────────
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
      var getProto = Object.getPrototypeOf || function(obj) { return obj.__proto__; };
      
      // User Input Fill
      var userProto = getProto(userEl);
      var userSetterDesc = Object.getOwnPropertyDescriptor(userProto, 'value');
      var userSetter = userSetterDesc ? userSetterDesc.set : null;

      userEl.focus();
      if (userSetter) {
        userSetter.call(userEl, ${u});
      } else {
        userEl.value = ${u};
      }
      userEl.dispatchEvent(new Event('input', { bubbles: true }));
      userEl.dispatchEvent(new Event('change', { bubbles: true }));
      userEl.blur();

      // Password Input Fill
      var passProto = getProto(passEl);
      var passSetterDesc = Object.getOwnPropertyDescriptor(passProto, 'value');
      var passSetter = passSetterDesc ? passSetterDesc.set : null;

      passEl.focus();
      if (passSetter) {
        passSetter.call(passEl, ${p});
      } else {
        passEl.value = ${p};
      }
      passEl.dispatchEvent(new Event('input', { bubbles: true }));
      passEl.dispatchEvent(new Event('change', { bubbles: true }));
      passEl.blur();

      if (${autoSubmit}) {
        var btn = querySelectorAllAll('button[type="submit"]')
          .concat(querySelectorAllAll('#btn-login'))
          .concat(querySelectorAllAll('.btn-primary'))
          .concat(querySelectorAllAll('input[type="submit"]'))[0];
        if (btn) {
          btn.focus();
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
      var getProto = Object.getPrototypeOf || function(obj) { return obj.__proto__; };
      
      // User Input Fill
      var userProto = getProto(userEl);
      var userSetterDesc = Object.getOwnPropertyDescriptor(userProto, 'value');
      var userSetter = userSetterDesc ? userSetterDesc.set : null;

      userEl.focus();
      if (userSetter) {
        userSetter.call(userEl, ${u});
      } else {
        userEl.value = ${u};
      }
      userEl.dispatchEvent(new Event('input', { bubbles: true }));
      userEl.dispatchEvent(new Event('change', { bubbles: true }));
      userEl.blur();

      // Password Input Fill
      var passProto = getProto(passEl);
      var passSetterDesc = Object.getOwnPropertyDescriptor(passProto, 'value');
      var passSetter = passSetterDesc ? passSetterDesc.set : null;

      passEl.focus();
      if (passSetter) {
        passSetter.call(passEl, ${p});
      } else {
        passEl.value = ${p};
      }
      passEl.dispatchEvent(new Event('input', { bubbles: true }));
      passEl.dispatchEvent(new Event('change', { bubbles: true }));
      passEl.blur();

      if (${autoSubmit}) {
        var btn = querySelectorAllAll('input[type="submit"]')
          .concat(querySelectorAllAll('button[type="submit"]'))
          .concat(querySelectorAllAll('#login_word'))[0];
        if (btn) {
          btn.focus();
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
      var getProto = Object.getPrototypeOf || function(obj) { return obj.__proto__; };
      
      // User Input Fill
      var userProto = getProto(userEl);
      var userSetterDesc = Object.getOwnPropertyDescriptor(userProto, 'value');
      var userSetter = userSetterDesc ? userSetterDesc.set : null;

      userEl.focus();
      if (userSetter) {
        userSetter.call(userEl, ${u});
      } else {
        userEl.value = ${u};
      }
      userEl.dispatchEvent(new Event('input', { bubbles: true }));
      userEl.dispatchEvent(new Event('change', { bubbles: true }));
      userEl.blur();

      // Password Input Fill
      var passProto = getProto(passEl);
      var passSetterDesc = Object.getOwnPropertyDescriptor(passProto, 'value');
      var passSetter = passSetterDesc ? passSetterDesc.set : null;

      passEl.focus();
      if (passSetter) {
        passSetter.call(passEl, ${p});
      } else {
        passEl.value = ${p};
      }
      passEl.dispatchEvent(new Event('input', { bubbles: true }));
      passEl.dispatchEvent(new Event('change', { bubbles: true }));
      passEl.blur();
      
      if (${autoSubmit}) {
        var btn = querySelectorAllAll('input[type="submit"]')
          .concat(querySelectorAllAll('button[type="submit"]'))[0];
        if (btn) {
          btn.focus();
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
    if (!userEl) userEl = querySelectorAllAll('input[type="text"]')[0];

    if (userEl && passEl) {
      var getProto = Object.getPrototypeOf || function(obj) { return obj.__proto__; };
      
      // User Input Fill
      var userProto = getProto(userEl);
      var userSetterDesc = Object.getOwnPropertyDescriptor(userProto, 'value');
      var userSetter = userSetterDesc ? userSetterDesc.set : null;

      userEl.focus();
      if (userSetter) {
        userSetter.call(userEl, ${u});
      } else {
        userEl.value = ${u};
      }
      userEl.dispatchEvent(new Event('input', { bubbles: true }));
      userEl.dispatchEvent(new Event('change', { bubbles: true }));
      userEl.blur();

      // Password Input Fill
      var passProto = getProto(passEl);
      var passSetterDesc = Object.getOwnPropertyDescriptor(passProto, 'value');
      var passSetter = passSetterDesc ? passSetterDesc.set : null;

      passEl.focus();
      if (passSetter) {
        passSetter.call(passEl, ${p});
      } else {
        passEl.value = ${p};
      }
      passEl.dispatchEvent(new Event('input', { bubbles: true }));
      passEl.dispatchEvent(new Event('change', { bubbles: true }));
      passEl.blur();

      if (${autoSubmit}) {
        var btn = querySelectorAllAll('button[type="submit"]')
          .concat(querySelectorAllAll('input[type="submit"]'))
          .concat(querySelectorAllAll('.btn-primary'))
          .concat(querySelectorAllAll('.login-btn'))[0];
        if (btn) {
          btn.focus();
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

module.exports = { getLoginScript };
