package main

// SettingsHTML 내장 설정 GUI 웹페이지 소스코드
const SettingsHTML = `<!DOCTYPE html>
<html lang="ko">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>IPMI Manager - Smart Router Settings</title>
    <link href="https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;600;800&family=Noto+Sans+KR:wght@300;400;700&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg-color: #08080c;
            --card-bg: rgba(18, 18, 28, 0.75);
            --card-border: rgba(255, 255, 255, 0.07);
            --text-main: #f0f0f5;
            --text-muted: #8e93ad;
            --primary: #00f0ff;
            --primary-glow: rgba(0, 240, 255, 0.25);
            --secondary: #b500ff;
            --secondary-glow: rgba(181, 0, 255, 0.25);
            --success: #00ff87;
            --danger: #ff0055;
            --font-family: 'Outfit', 'Noto Sans KR', sans-serif;
        }

        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }

        body {
            background-color: var(--bg-color);
            background-image: 
                radial-gradient(at 0% 0%, rgba(181, 0, 255, 0.08) 0px, transparent 50%),
                radial-gradient(at 100% 0%, rgba(0, 240, 255, 0.06) 0px, transparent 50%);
            color: var(--text-main);
            font-family: var(--font-family);
            min-height: 100vh;
            padding: 2rem;
            display: flex;
            justify-content: center;
            align-items: flex-start;
        }

        .container {
            width: 100%;
            max-width: 1200px;
            display: flex;
            flex-direction: column;
            gap: 2rem;
        }

        /* Header Area */
        header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 1.5rem;
            background: var(--card-bg);
            border: 1px solid var(--card-border);
            border-radius: 16px;
            backdrop-filter: blur(12px);
            box-shadow: 0 8px 32px 0 rgba(0, 0, 0, 0.4);
        }

        .brand {
            display: flex;
            align-items: center;
            gap: 0.8rem;
        }

        .brand-logo {
            font-size: 2rem;
            background: linear-gradient(45deg, var(--primary), var(--secondary));
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            font-weight: 800;
        }

        .brand-title {
            font-size: 1.3rem;
            font-weight: 600;
            letter-spacing: 0.5px;
        }

        .status-badge {
            display: flex;
            align-items: center;
            gap: 0.5rem;
            background: rgba(0, 255, 135, 0.1);
            border: 1px solid rgba(0, 255, 135, 0.3);
            color: var(--success);
            padding: 0.4rem 1rem;
            border-radius: 20px;
            font-size: 0.85rem;
            font-weight: 600;
            text-shadow: 0 0 8px rgba(0, 255, 135, 0.4);
        }

        /* Config Grid */
        .content-grid {
            display: grid;
            grid-template-columns: 400px 1fr;
            gap: 2rem;
        }

        @media (max-width: 950px) {
            .content-grid {
                grid-template-columns: 1fr;
            }
        }

        /* Card Base */
        .card {
            background: var(--card-bg);
            border: 1px solid var(--card-border);
            border-radius: 16px;
            padding: 1.8rem;
            backdrop-filter: blur(12px);
            box-shadow: 0 8px 32px 0 rgba(0, 0, 0, 0.3);
            display: flex;
            flex-direction: column;
            gap: 1.5rem;
        }

        .card-title {
            font-size: 1.15rem;
            font-weight: 600;
            border-bottom: 1px solid rgba(255, 255, 255, 0.08);
            padding-bottom: 0.75rem;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .card-title span {
            background: linear-gradient(90deg, var(--primary), var(--secondary));
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }

        /* Input / Forms */
        .form-group {
            display: flex;
            flex-direction: column;
            gap: 0.5rem;
        }

        label {
            font-size: 0.82rem;
            color: var(--text-muted);
            font-weight: 600;
            letter-spacing: 0.3px;
        }

        input[type="text"], input[type="password"], input[type="number"], select, textarea {
            background: rgba(255, 255, 255, 0.02);
            border: 1px solid rgba(255, 255, 255, 0.08);
            color: var(--text-main);
            padding: 0.75rem 0.9rem;
            border-radius: 8px;
            font-family: var(--font-family);
            font-size: 0.9rem;
            transition: all 0.3s ease;
        }

        input:focus, select:focus, textarea:focus {
            outline: none;
            border-color: var(--primary);
            box-shadow: 0 0 8px var(--primary-glow);
            background: rgba(255, 255, 255, 0.05);
        }

        /* Buttons */
        .btn {
            background: linear-gradient(135deg, var(--primary), #00b0ff);
            color: #0b0b0f;
            border: none;
            padding: 0.75rem 1.3rem;
            border-radius: 8px;
            font-family: var(--font-family);
            font-weight: 700;
            cursor: pointer;
            transition: all 0.3s ease;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            gap: 0.5rem;
            font-size: 0.9rem;
        }

        .btn:hover {
            transform: translateY(-1px);
            box-shadow: 0 4px 12px var(--primary-glow);
            filter: brightness(1.1);
        }

        .btn-secondary {
            background: rgba(255, 255, 255, 0.05);
            color: var(--text-main);
            border: 1px solid rgba(255, 255, 255, 0.08);
        }

        .btn-secondary:hover {
            background: rgba(255, 255, 255, 0.1);
            box-shadow: none;
        }

        .btn-danger {
            background: linear-gradient(135deg, var(--danger), #ff00aa);
            color: #fff;
        }

        .btn-danger:hover {
            box-shadow: 0 4px 12px rgba(255, 0, 85, 0.35);
        }

        .btn-mini {
            padding: 0.25rem 0.5rem;
            font-size: 0.75rem;
            border-radius: 6px;
        }

        /* Lists */
        .list-container {
            display: flex;
            flex-direction: column;
            gap: 0.8rem;
        }

        /* Profile Item Box */
        .profile-item {
            background: rgba(255, 255, 255, 0.015);
            border: 1px solid rgba(255, 255, 255, 0.04);
            border-radius: 10px;
            padding: 1rem;
            display: flex;
            flex-direction: column;
            gap: 0.5rem;
        }

        .profile-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .profile-name {
            font-size: 0.92rem;
            font-weight: 700;
        }

        .profile-diagnose {
            font-size: 0.8rem;
            display: flex;
            gap: 0.5rem;
            color: var(--text-muted);
        }

        .diag-badge {
            display: inline-flex;
            align-items: center;
            gap: 0.2rem;
        }

        .profile-paths {
            font-size: 0.75rem;
            color: var(--text-muted);
            font-family: monospace;
            background: rgba(0, 0, 0, 0.2);
            padding: 0.4rem;
            border-radius: 6px;
            word-break: break-all;
        }

        /* Rule Item Box */
        .rule-item {
            background: rgba(255, 255, 255, 0.015);
            border: 1px solid rgba(255, 255, 255, 0.04);
            border-radius: 12px;
            padding: 1.1rem;
            display: flex;
            align-items: center;
            justify-content: space-between;
            transition: all 0.25s ease;
        }

        .rule-item:hover {
            background: rgba(255, 255, 255, 0.03);
            border-color: rgba(0, 240, 255, 0.15);
        }

        .rule-info {
            display: flex;
            flex-direction: column;
            gap: 0.4rem;
        }

        .rule-badge-row {
            display: flex;
            align-items: center;
            gap: 0.5rem;
            flex-wrap: wrap;
        }

        .badge {
            padding: 0.2rem 0.5rem;
            border-radius: 5px;
            font-size: 0.72rem;
            font-weight: 700;
            text-transform: uppercase;
        }

        .badge-priority {
            background: rgba(255, 255, 255, 0.08);
            color: var(--text-main);
        }

        .badge-vendor {
            background: rgba(181, 0, 255, 0.12);
            border: 1px solid rgba(181, 0, 255, 0.3);
            color: #d880ff;
        }

        .badge-type-ikvm {
            background: rgba(0, 240, 255, 0.12);
            border: 1px solid rgba(0, 240, 255, 0.3);
            color: var(--primary);
        }

        .badge-type-jnlp {
            background: rgba(255, 170, 0, 0.12);
            border: 1px solid rgba(255, 170, 0, 0.3);
            color: #ffa500;
        }

        .badge-type-web {
            background: rgba(0, 255, 135, 0.12);
            border: 1px solid rgba(0, 255, 135, 0.3);
            color: var(--success);
        }

        .badge-profile {
            background: rgba(255, 255, 255, 0.04);
            border: 1px solid rgba(255, 255, 255, 0.15);
            color: var(--text-muted);
        }

        .rule-desc {
            font-size: 0.85rem;
            color: var(--text-muted);
            line-height: 1.4;
        }

        .rule-pattern {
            font-family: monospace;
            background: rgba(0, 0, 0, 0.25);
            padding: 0.05rem 0.3rem;
            border-radius: 4px;
            color: #ffd080;
        }

        .rule-actions {
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .order-btn-group {
            display: flex;
            flex-direction: column;
            gap: 0.1rem;
        }

        .order-btn {
            background: none;
            border: none;
            color: var(--text-muted);
            cursor: pointer;
            font-size: 0.8rem;
            padding: 0.15rem;
            transition: color 0.2s;
        }

        .order-btn:hover {
            color: var(--primary);
        }

        /* Modal styling */
        .modal-overlay {
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0, 0, 0, 0.8);
            backdrop-filter: blur(8px);
            display: flex;
            justify-content: center;
            align-items: center;
            z-index: 1000;
            opacity: 0;
            pointer-events: none;
            transition: all 0.3s ease;
        }

        .modal-overlay.active {
            opacity: 1;
            pointer-events: auto;
        }

        .modal {
            background: #0e0e14;
            border: 1px solid var(--card-border);
            border-radius: 16px;
            width: 100%;
            max-width: 500px;
            padding: 2rem;
            box-shadow: 0 10px 40px rgba(0,0,0,0.6);
            display: flex;
            flex-direction: column;
            gap: 1.2rem;
            transform: scale(0.95);
            transition: all 0.3s ease;
        }

        .modal-overlay.active .modal {
            transform: scale(1);
        }

        .modal-header {
            font-size: 1.15rem;
            font-weight: 700;
            border-bottom: 1px solid rgba(255, 255, 255, 0.08);
            padding-bottom: 0.5rem;
        }

        .modal-actions {
            display: flex;
            justify-content: flex-end;
            gap: 0.8rem;
            margin-top: 0.8rem;
        }

        /* Info Banner & Guide Code Block */
        .info-banner {
            background: rgba(0, 240, 255, 0.04);
            border: 1px solid rgba(0, 240, 255, 0.15);
            padding: 0.9rem 1.1rem;
            border-radius: 8px;
            font-size: 0.82rem;
            color: #b5f5ff;
            line-height: 1.45;
        }

        .guide-section {
            grid-column: 1 / -1;
        }

        .code-block {
            background: #040407;
            border: 1px solid rgba(255, 255, 255, 0.05);
            border-radius: 8px;
            padding: 1.2rem;
            font-family: 'Fira Code', Consolas, monospace;
            font-size: 0.82rem;
            color: #c9d1d9;
            overflow-x: auto;
            line-height: 1.5;
        }

        .code-keyword { color: #ff7b72; }
        .code-string { color: #a5d6ff; }
        .code-comment { color: #8b949e; }
        .code-number { color: #79c0ff; }
    </style>
</head>
<body>
    <div class="container">
        <!-- Header -->
        <header>
            <div class="brand">
                <div class="brand-logo">⚡</div>
                <div>
                    <h1 class="brand-title">IPMI Manager</h1>
                    <p style="font-size: 0.75rem; color: var(--text-muted)">스마트 라우팅 및 다중 프로필 관리 센터 (Port: 4447)</p>
                </div>
            </div>
            <div class="status-badge">
                <span style="display: inline-block; width: 8px; height: 8px; background: var(--success); border-radius: 50%; box-shadow: 0 0 8px var(--success)"></span>
                DAEMON ACTIVE
            </div>
        </header>

        <!-- Main Content Grid -->
        <div class="content-grid">
            <!-- Left Side: Config & Profiles -->
            <div style="display: flex; flex-direction: column; gap: 2rem;">
                <!-- Js-Proxy Card -->
                <div class="card">
                    <h2 class="card-title"><span>Js-Proxy 연동 설정</span></h2>
                    <div class="form-group">
                        <label for="js_proxy_url">API Base URL</label>
                        <input type="text" id="js_proxy_url" placeholder="예: https://js-proxy.jscomz.net/api/devices">
                    </div>
                    <div class="form-group">
                        <label for="js_proxy_token">인증 토큰 (Bearer Token)</label>
                        <input type="password" id="js_proxy_token" placeholder="Bearer 인증 토큰 입력">
                    </div>
                    <div style="display: flex; gap: 0.5rem; margin-top: 0.5rem;">
                        <button class="btn" style="flex: 1" id="btn-save-proxy" onclick="saveProxyConfig()">💾 저장</button>
                        <button class="btn btn-secondary" style="flex: 1" id="btn-test-proxy" onclick="testProxyConnection()">🔄 테스트</button>
                    </div>
                </div>

                <!-- Profiles Card -->
                <div class="card">
                    <div class="card-title">
                        <span>실행 프로필 (Profiles)</span>
                        <button class="btn btn-mini" onclick="openAddProfileModal()">➕ 추가</button>
                    </div>
                    <div class="info-banner" style="background: rgba(181, 0, 255, 0.02); border-color: rgba(181, 0, 255, 0.1)">
                        자바 런타임 및 iKVM.jar 파일 버전을 프로필별로 분리 관리합니다. 각 항목의 파일 정상 존재 여부가 실시간 확인됩니다.
                    </div>
                    <div class="list-container" id="profile-container">
                        <!-- Profiles render dynamically -->
                    </div>
                </div>
            </div>

            <!-- Right Side: Routing Rules -->
            <div class="card">
                <div class="card-title">
                    <span>스마트 라우팅 규칙 매트릭스</span>
                    <button class="btn btn-mini" onclick="openAddRuleModal()">➕ 규칙 추가</button>
                </div>
                <div class="info-banner">
                    호출 API가 접수되면 등록된 규칙 순서대로 판별을 시도합니다. 매칭 규칙의 연동 프로필 환경이 적용되며, 매칭 실패 혹은 연동 실패 시 기본 <strong>WEB</strong> 방식으로 자동 폴백됩니다.
                </div>
                <div class="list-container" id="rule-container">
                    <!-- Rules render dynamically -->
                </div>
            </div>

            <!-- Bottom: Integration Guide -->
            <div class="card guide-section">
                <h2 class="card-title"><span>외부 서비스 연동 가이드 (Integration Guide)</span></h2>
                <div class="rule-desc" style="margin-bottom: 0.5rem;">
                    사내 자산 관리 대시보드나 웹 콘솔 페이지에서 사용자 로컬에 구동 중인 Go 데몬을 호출하여 KVM을 원클릭 실행시킬 수 있습니다. 데몬이 켜져 있지 않을 때 자연스럽게 웹 브라우저 접속으로 연결해주는 안전한 비동기 호출 샘플 코드입니다.
                </div>
                <pre class="code-block"><code><span class="code-keyword">function</span> <span class="code-primary">connectToKvm</span>(ip, fallbackUrl) {
    <span class="code-keyword">const</span> localDaemonUrl = <span class="code-string">'http://127.0.0.1:4447/api/connect?ip='</span> + <span class="code-primary">encodeURIComponent</span>(ip);
    
    <span class="code-comment">// 3초 타임아웃을 설정한 비동기 호출</span>
    <span class="code-keyword">const</span> controller = <span class="code-keyword">new</span> <span class="code-primary">AbortController</span>();
    <span class="code-keyword">const</span> timeoutId = <span class="code-primary">setTimeout</span>(() => controller.<span class="code-primary">abort</span>(), <span class="code-number">3000</span>);

    <span class="code-primary">fetch</span>(localDaemonUrl, { method: <span class="code-string">'GET'</span>, signal: controller.<span class="code-primary">signal</span> })
        .<span class="code-primary">then</span>(res => res.<span class="code-primary">json</span>())
        .<span class="code-primary">then</span>(data => {
            <span class="code-primary">clearTimeout</span>(timeoutId);
            <span class="code-keyword">if</span> (data.success) {
                <span class="code-primary">console</span>.<span class="code-primary">log</span>(<span class="code-string">'KVM 기동 시퀀스 성공'</span>);
            } <span class="code-keyword">else</span> {
                <span class="code-primary">window</span>.<span class="code-primary">open</span>(fallbackUrl, <span class="code-string">'_blank'</span>);
            }
        })
        .<span class="code-primary">catch</span>(err => {
            <span class="code-primary">clearTimeout</span>(timeoutId);
            <span class="code-comment">// 데몬 미기동 시 브라우저 웹페이지 로그인으로 안전한 폴백</span>
            <span class="code-primary">window</span>.<span class="code-primary">open</span>(fallbackUrl, <span class="code-string">'_blank'</span>);
        });
}</code></pre>
            </div>
        </div>
    </div>

    <!-- Rule Modal -->
    <div class="modal-overlay" id="rule-modal">
        <div class="modal">
            <h3 class="modal-header" id="rule-modal-title">라우팅 규칙 추가</h3>
            <input type="hidden" id="rule_id">
            <div class="form-group">
                <label for="rule_vendor">대상 벤더 (Vendor)</label>
                <input type="text" id="rule_vendor" placeholder="예: supermicro, dell, *">
            </div>
            <div class="form-group">
                <label for="rule_model">모델 매칭 패턴 (Model Pattern)</label>
                <input type="text" id="rule_model" placeholder="예: x10, r630, *">
            </div>
            <div class="form-group">
                <label for="rule_type">KVM 실행 방식 (Action)</label>
                <select id="rule_type">
                    <option value="ikvm">ikvm (iKVM.jar 직접실행)</option>
                    <option value="jnlp">jnlp (Java Web Start 기동)</option>
                    <option value="WEB">WEB (웹 자동 로그인)</option>
                </select>
            </div>
            <div class="form-group">
                <label for="rule_profile">연동할 실행 프로필</label>
                <select id="rule_profile">
                    <!-- Dynamic rendering -->
                </select>
            </div>
            <div class="form-group">
                <label for="rule_priority">우선순위 (Priority)</label>
                <input type="number" id="rule_priority" min="1" max="98">
            </div>
            <div class="form-group">
                <label for="rule_desc">규칙 설명</label>
                <textarea id="rule_desc" rows="2" placeholder="규칙 설명 메모"></textarea>
            </div>
            <div class="modal-actions">
                <button class="btn btn-secondary" onclick="closeRuleModal()">취소</button>
                <button class="btn" onclick="saveRule()">저장</button>
            </div>
        </div>
    </div>

    <!-- Profile Modal -->
    <div class="modal-overlay" id="profile-modal">
        <div class="modal">
            <h3 class="modal-header" id="profile-modal-title">실행 프로필 추가</h3>
            <input type="hidden" id="profile_id">
            <div class="form-group">
                <label for="profile_name">프로필 이름</label>
                <input type="text" id="profile_name" placeholder="예: Java 8 최신, JRE 8u161 구형">
            </div>
            <div class="form-group">
                <label for="profile_java_path">javaws.exe 물리 경로</label>
                <input type="text" id="profile_java_path" placeholder="C:\\Program Files (x86)\\Java\\...\\javaws.exe">
            </div>
            <div class="form-group">
                <label for="profile_ikvm_path">iKVM.jar 물리 경로</label>
                <input type="text" id="profile_ikvm_path" placeholder="C:\\Users\\...\\IPMIVIEW\\...\\iKVM.jar">
            </div>
            <div class="form-group">
                <label for="profile_desc">프로필 설명</label>
                <textarea id="profile_desc" rows="2" placeholder="버전별 정보 기재"></textarea>
            </div>
            <div class="modal-actions">
                <button class="btn btn-secondary" onclick="closeProfileModal()">취소</button>
                <button class="btn" onclick="saveProfile()">저장</button>
            </div>
        </div>
    </div>

    <script>
        let configData = { rules: [], profiles: [], js_proxy_url: '', js_proxy_token: '' };
        let profileStatus = {}; // 프로필 ID별 실시간 파일 검증 상태 보관

        window.addEventListener('DOMContentLoaded', () => {
            loadConfig();
        });

        function loadConfig() {
            fetch('/api/rules')
                .then(res => res.json())
                .then(data => {
                    configData = data;
                    document.getElementById('js_proxy_url').value = data.js_proxy_url || '';
                    document.getElementById('js_proxy_token').value = data.js_proxy_token || '';
                    
                    // 각 프로필별로 실시간 경로 진단 수행
                    const diagPromises = configData.profiles.map(p => {
                        return fetch('/api/diagnose?profile_id=' + p.id)
                            .then(r => r.json())
                            .then(diag => {
                                profileStatus[p.id] = diag;
                            })
                            .catch(e => {
                                profileStatus[p.id] = { java_exe: false, javaws_exe: false, ikvm_jar: false };
                            });
                    });

                    Promise.all(diagPromises).then(() => {
                        renderProfiles();
                        renderRules();
                    });
                })
                .catch(err => alert('설정 로드 실패: ' + err));
        }

        // 프로필 목록 렌더링
        function renderProfiles() {
            const container = document.getElementById('profile-container');
            container.innerHTML = '';

            configData.profiles.forEach(p => {
                const status = profileStatus[p.id] || { java_exe: false, javaws_exe: false, ikvm_jar: false };
                
                const javaIndicator = status.javaws_exe ? '<span class="diag-badge" style="color:var(--success)">🟢 javaws</span>' : '<span class="diag-badge" style="color:var(--danger)">🔴 javaws</span>';
                const nativeJavaIndicator = status.java_exe ? '<span class="diag-badge" style="color:var(--success)">🟢 java</span>' : '<span class="diag-badge" style="color:var(--danger)">🔴 java</span>';
                const jarIndicator = status.ikvm_jar ? '<span class="diag-badge" style="color:var(--success)">🟢 iKVM.jar</span>' : '<span class="diag-badge" style="color:var(--danger)">🔴 iKVM.jar</span>';

                const div = document.createElement('div');
                div.className = 'profile-item';
                
                let profileHtml = '\n' +
                '    <div class="profile-header">\n' +
                '        <span class="profile-name">' + p.name + (p.is_default ? ' <span style="color:var(--primary); font-size:0.75rem;">(기본)</span>' : '') + '</span>\n' +
                '        <div>\n' +
                '            <button class="btn btn-secondary btn-mini" onclick="openEditProfileModal(\'' + p.id + '\')">📝</button>\n';
                
                if (!p.is_default) {
                    profileHtml += '            <button class="btn btn-danger btn-mini" onclick="deleteProfile(\'' + p.id + '\')">❌</button>\n';
                }
                
                profileHtml += '        </div>\n' +
                '    </div>\n' +
                '    <div class="profile-diagnose">\n' +
                '        ' + javaIndicator + ' | ' + nativeJavaIndicator + ' | ' + jarIndicator + '\n' +
                '    </div>\n' +
                '    <div class="profile-paths">\n' +
                '        Java: ' + p.java_path + '<br>iKVM: ' + p.ikvm_jar_path + '\n' +
                '    </div>\n';

                div.innerHTML = profileHtml;
                container.appendChild(div);
            });
        }

        // 규칙 목록 렌더링
        function renderRules() {
            const container = document.getElementById('rule-container');
            container.innerHTML = '';

            if (!configData.rules || configData.rules.length === 0) {
                container.innerHTML = '<div style="text-align:center; padding: 2rem; color: var(--text-muted);">등록된 규칙이 없습니다.</div>';
                return;
            }

            configData.rules.forEach((rule, index) => {
                const isDefaultFallback = (rule.vendor === '*' && rule.model_pattern === '*');
                const matchedProfile = configData.profiles.find(p => p.id === rule.profile_id);
                const profileName = matchedProfile ? matchedProfile.name : '기본 프로필';

                const item = document.createElement('div');
                item.className = 'rule-item';
                
                let typeBadgeClass = 'badge-type-web';
                let typeName = 'WEB';
                if (rule.connect_type === 'ikvm') {
                    typeBadgeClass = 'badge-type-ikvm';
                    typeName = 'ikvm';
                } else if (rule.connect_type === 'jnlp') {
                    typeBadgeClass = 'badge-type-jnlp';
                    typeName = 'jnlp';
                }

                let ruleHtml = '\n' +
                '    <div class="rule-info">\n' +
                '        <div class="rule-badge-row">\n' +
                '            <span class="badge badge-priority">우선순위 ' + rule.priority + '</span>\n' +
                '            <span class="badge badge-vendor">' + rule.vendor.toUpperCase() + '</span>\n' +
                '            <span class="badge ' + typeBadgeClass + '">' + typeName + '</span>\n' +
                '            <span class="badge badge-profile">프로필: ' + profileName + '</span>\n' +
                '        </div>\n' +
                '        <p class="rule-desc">\n' +
                '            ' + (rule.description || '설명 없음') + ' \n' +
                '            (매칭 패턴: 벤더 <span class="rule-pattern">' + rule.vendor + '</span> / 모델 <span class="rule-pattern">' + rule.model_pattern + '</span>)\n' +
                '        </p>\n' +
                '    </div>\n' +
                '    <div class="rule-actions">\n';

                if (!isDefaultFallback) {
                    ruleHtml += '\n' +
                    '        <div class="order-btn-group">\n' +
                    '            <button class="order-btn" onclick="moveRule(' + index + ', -1)">▲</button>\n' +
                    '            <button class="order-btn" onclick="moveRule(' + index + ', 1)">▼</button>\n' +
                    '        </div>\n' +
                    '        <button class="btn btn-secondary btn-mini" onclick="openEditRuleModal(' + index + ')">📝 수정</button>\n' +
                    '        <button class="btn btn-danger btn-mini" onclick="deleteRule(\'' + rule.id + '\')">❌ 삭제</button>\n';
                } else {
                    ruleHtml += '        <span style="font-size:0.8rem; color:var(--text-muted);">Fallback 시스템 규칙</span>\n';
                }
                
                ruleHtml += '    </div>\n';

                item.innerHTML = ruleHtml;
                container.appendChild(item);
            });
        }

        // Proxy 설정 및 테스트
        function saveProxyConfig() {
            configData.js_proxy_url = document.getElementById('js_proxy_url').value.trim();
            configData.js_proxy_token = document.getElementById('js_proxy_token').value.trim();
            saveAllConfig('Js-Proxy 설정이 저장되었습니다.');
        }

        function testProxyConnection() {
            const url = document.getElementById('js_proxy_url').value.trim();
            const token = document.getElementById('js_proxy_token').value.trim();
            if (!url) {
                alert('연결 테스트를 진행할 API URL을 입력하십시오.');
                return;
            }
            const btn = document.getElementById('btn-test-proxy');
            btn.innerText = '⏳ 테스트 중...';
            btn.disabled = true;

            fetch('/api/test-proxy', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ url: url, token: token })
            })
            .then(res => res.json())
            .then(data => {
                if (data.success) alert('✅ Js-Proxy API 연동 성공!');
                else alert('❌ 연동 실패: ' + data.error);
            })
            .catch(err => alert('❌ 요청 실패: ' + err))
            .finally(() => {
                btn.innerText = '🔄 테스트';
                btn.disabled = false;
            });
        }

        // 프로필 팝업 제어
        function openAddProfileModal() {
            document.getElementById('profile-modal-title').innerText = '실행 프로필 추가';
            document.getElementById('profile_id').value = '';
            document.getElementById('profile_name').value = '';
            document.getElementById('profile_java_path').value = '';
            document.getElementById('profile_ikvm_path').value = '';
            document.getElementById('profile_desc').value = '';
            document.getElementById('profile-modal').classList.add('active');
        }

        function openEditProfileModal(id) {
            const p = configData.profiles.find(x => x.id === id);
            if (!p) return;
            document.getElementById('profile-modal-title').innerText = '실행 프로필 수정';
            document.getElementById('profile_id').value = p.id;
            document.getElementById('profile_name').value = p.name;
            document.getElementById('profile_java_path').value = p.java_path;
            document.getElementById('profile_ikvm_path').value = p.ikvm_jar_path;
            document.getElementById('profile_desc').value = p.description || '';
            document.getElementById('profile-modal').classList.add('active');
        }

        function closeProfileModal() {
            document.getElementById('profile-modal').classList.remove('active');
        }

        function saveProfile() {
            const id = document.getElementById('profile_id').value;
            const name = document.getElementById('profile_name').value.trim();
            const java = document.getElementById('profile_java_path').value.trim();
            const ikvm = document.getElementById('profile_ikvm_path').value.trim();
            const desc = document.getElementById('profile_desc').value.trim();

            if (!name || !java || !ikvm) {
                alert('모든 입력 칸을 올바르게 채워주세요.');
                return;
            }

            const pData = {
                id: id || ('profile_' + Date.now()),
                name: name,
                java_path: java,
                ikvm_jar_path: ikvm,
                is_default: id ? (configData.profiles.find(x => x.id === id).is_default) : false,
                description: desc
            };

            if (id) {
                const idx = configData.profiles.findIndex(x => x.id === id);
                if (idx !== -1) configData.profiles[idx] = pData;
            } else {
                configData.profiles.push(pData);
            }

            closeProfileModal();
            saveAllConfig('프로필 정보가 저장되었습니다.');
        }

        function deleteProfile(id) {
            if (!confirm('정말 이 프로필을 삭제하시겠습니까? 해당 프로필을 연동 중인 규칙들이 정상 기동하지 않을 수 있습니다.')) return;
            configData.profiles = configData.profiles.filter(x => x.id !== id);
            saveAllConfig('프로필이 삭제되었습니다.');
        }

        // 규칙 팝업 제어
        function openAddRuleModal() {
            document.getElementById('rule-modal-title').innerText = '새 라우팅 규칙 추가';
            document.getElementById('rule_id').value = '';
            document.getElementById('rule_vendor').value = '';
            document.getElementById('rule_model').value = '';
            document.getElementById('rule_type').value = 'ikvm';
            document.getElementById('rule_priority').value = (configData.rules.length > 0) ? configData.rules[configData.rules.length - 1].priority + 1 : 10;
            document.getElementById('rule_desc').value = '';
            
            buildProfileSelectOptions('');
            document.getElementById('rule-modal').classList.add('active');
        }

        function openEditRuleModal(index) {
            const rule = configData.rules[index];
            document.getElementById('rule-modal-title').innerText = '라우팅 규칙 수정';
            document.getElementById('rule_id').value = rule.id;
            document.getElementById('rule_vendor').value = rule.vendor;
            document.getElementById('rule_model').value = rule.model_pattern;
            document.getElementById('rule_type').value = rule.connect_type;
            document.getElementById('rule_priority').value = rule.priority;
            document.getElementById('rule_desc').value = rule.description || '';
            
            buildProfileSelectOptions(rule.profile_id);
            document.getElementById('rule-modal').classList.add('active');
        }

        function buildProfileSelectOptions(selectedId) {
            const select = document.getElementById('rule_profile');
            select.innerHTML = '';
            configData.profiles.forEach(p => {
                const opt = document.createElement('option');
                opt.value = p.id;
                opt.innerText = p.name;
                if (p.id === selectedId) opt.selected = true;
                select.appendChild(opt);
            });
        }

        function closeRuleModal() {
            document.getElementById('rule-modal').classList.remove('active');
        }

        function saveRule() {
            const id = document.getElementById('rule_id').value;
            const vendor = document.getElementById('rule_vendor').value.trim().toLowerCase();
            const model = document.getElementById('rule_model').value.trim().toLowerCase();
            const type = document.getElementById('rule_type').value;
            const profile = document.getElementById('rule_profile').value;
            const priority = parseInt(document.getElementById('rule_priority').value);
            const desc = document.getElementById('rule_desc').value.trim();

            if (!vendor || !model) {
                alert('벤더 및 모델 패턴을 입력하십시오.');
                return;
            }

            const rData = {
                id: id || ('rule_' + Date.now()),
                vendor: vendor,
                model_pattern: model,
                connect_type: type,
                profile_id: profile,
                priority: priority,
                description: desc
            };

            if (id) {
                const idx = configData.rules.findIndex(x => x.id === id);
                if (idx !== -1) configData.rules[idx] = rData;
            } else {
                configData.rules.push(rData);
            }

            closeRuleModal();
            saveAllConfig('규칙이 저장되었습니다.');
        }

        function deleteRule(id) {
            if (!confirm('정말 이 라우팅 규칙을 삭제하시겠습니까?')) return;
            configData.rules = configData.rules.filter(x => x.id !== id);
            saveAllConfig('규칙이 삭제되었습니다.');
        }

        function moveRule(index, direction) {
            const targetIndex = index + direction;
            if (targetIndex < 0 || targetIndex >= configData.rules.length) return;
            if (configData.rules[index].vendor === '*' || configData.rules[targetIndex].vendor === '*') return;

            const temp = configData.rules[index].priority;
            configData.rules[index].priority = configData.rules[targetIndex].priority;
            configData.rules[targetIndex].priority = temp;

            saveAllConfig('우선순위가 변경되었습니다.');
        }

        function saveAllConfig(successMsg) {
            fetch('/api/rules/save', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(configData)
            })
            .then(res => res.json())
            .then(data => {
                if (data.success) {
                    loadConfig();
                } else {
                    alert('저장 오류: ' + data.error);
                }
            })
            .catch(err => alert('요청 에러: ' + err));
        }
    </script>
</body>
</html>`
