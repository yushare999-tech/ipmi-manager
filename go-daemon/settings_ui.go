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
            grid-template-columns: 380px 1fr;
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

        /* Filter Bar */
        .filter-bar {
            display: flex;
            gap: 0.5rem;
            flex-wrap: wrap;
            padding: 0.3rem 0;
            border-bottom: 1px solid rgba(255, 255, 255, 0.05);
            padding-bottom: 1rem;
        }

        .filter-btn {
            background: rgba(255, 255, 255, 0.03);
            border: 1px solid rgba(255, 255, 255, 0.08);
            color: var(--text-muted);
            padding: 0.4rem 0.8rem;
            border-radius: 6px;
            font-size: 0.8rem;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.2s ease;
        }

        .filter-btn:hover {
            background: rgba(255, 255, 255, 0.08);
            color: var(--text-main);
        }

        .filter-btn.active {
            background: rgba(0, 240, 255, 0.12);
            border-color: var(--primary);
            color: var(--primary);
            box-shadow: 0 0 8px var(--primary-glow);
        }

        /* Vendor Group Box */
        .vendor-group-card {
            background: rgba(255, 255, 255, 0.012);
            border: 1px solid rgba(255, 255, 255, 0.05);
            border-radius: 12px;
            padding: 1.2rem;
            display: flex;
            flex-direction: column;
            gap: 0.8rem;
            transition: all 0.3s ease;
        }

        .vendor-group-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            cursor: pointer;
            user-select: none;
        }

        .vendor-group-title {
            font-size: 0.95rem;
            font-weight: 800;
            color: var(--primary);
            text-transform: uppercase;
            letter-spacing: 0.5px;
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .group-toggle-arrow {
            color: var(--text-muted);
            font-size: 0.8rem;
            transition: transform 0.2s;
        }

        .vendor-group-body {
            display: flex;
            flex-direction: column;
            gap: 0.8rem;
            margin-top: 0.4rem;
        }

        /* Rule Item Box */
        .rule-item {
            background: rgba(255, 255, 255, 0.015);
            border: 1px solid rgba(255, 255, 255, 0.03);
            border-radius: 8px;
            padding: 0.9rem;
            display: flex;
            align-items: center;
            justify-content: space-between;
            transition: all 0.25s ease;
        }

        .rule-item:hover {
            background: rgba(255, 255, 255, 0.025);
            border-color: rgba(0, 240, 255, 0.15);
        }

        .rule-info {
            display: flex;
            flex-direction: column;
            gap: 0.35rem;
        }

        .rule-badge-row {
            display: flex;
            align-items: center;
            gap: 0.4rem;
            flex-wrap: wrap;
        }

        .badge {
            padding: 0.15rem 0.4rem;
            border-radius: 4px;
            font-size: 0.7rem;
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
            font-size: 0.82rem;
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

        /* Inline Proxy Form */
        .inline-proxy-form {
            display: flex;
            gap: 1rem;
            align-items: flex-end;
            background: rgba(255, 255, 255, 0.015);
            border: 1px solid rgba(255, 255, 255, 0.04);
            padding: 1.2rem;
            border-radius: 10px;
            margin-bottom: 1.2rem;
        }

        @media (max-width: 750px) {
            .inline-proxy-form {
                flex-direction: column;
                align-items: stretch;
            }
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
            <!-- Left Side: Profiles ONLY -->
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

            <!-- Right Side: Routing Rules -->
            <div class="card">
                <div class="card-title">
                    <span>스마트 라우팅 규칙 매트릭스</span>
                    <button class="btn btn-mini" onclick="openAddRuleModal()">➕ 규칙 추가</button>
                </div>
                
                <!-- Filter Buttons Bar -->
                <div class="filter-bar" id="filter-bar">
                    <!-- Filters render dynamically -->
                </div>

                <div class="list-container" id="rule-container" style="gap: 1.2rem;">
                    <!-- Rules render dynamically in vendor groups -->
                </div>
            </div>

            <!-- Bottom: Js-Proxy Settings & Integration Guide -->
            <div class="card guide-section">
                <h2 class="card-title"><span>Js-Proxy 연동 설정 (IP 기준 조회 API)</span></h2>
                
                <!-- Js-Proxy Form -->
                <div class="inline-proxy-form">
                    <div class="form-group" style="flex: 2;">
                        <label for="js_proxy_url">Js-Proxy API URL</label>
                        <input type="text" id="js_proxy_url" placeholder="예: https://js-proxy.jscomz.net/api/devices">
                    </div>
                    <div class="form-group" style="flex: 1.2;">
                        <label for="js_proxy_token">인증 토큰 (Bearer Token)</label>
                        <input type="password" id="js_proxy_token" placeholder="Bearer 인증 토큰 입력">
                    </div>
                    <div style="display: flex; gap: 0.4rem;">
                        <button class="btn" id="btn-save-proxy" onclick="saveProxyConfig()">💾 저장</button>
                        <button class="btn btn-secondary" id="btn-test-proxy" onclick="testProxyConnection()">🔄 테스트</button>
                    </div>
                </div>

                <h2 class="card-title" style="margin-top: 1.5rem; border-bottom: none;"><span>외부 서비스 연동 가이드 (Integration Guide)</span></h2>
                <div class="rule-desc" style="margin-bottom: 0.8rem;">
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
                <input type="text" id="rule_vendor" placeholder="예: supermicro, dell, hp">
            </div>
            <div class="form-group">
                <label for="rule_model">모델 매칭 패턴 (Model Pattern)</label>
                <input type="text" id="rule_model" placeholder="예: x10, r630 (벤더 디폴트는 * 입력)">
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
                <label for="rule_priority">동일 그룹 내 우선순위 (Priority)</label>
                <input type="number" id="rule_priority" min="1" max="98" value="1">
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
                <label for="profile_java_path">Java 런타임 경로 (javaws.exe / java.exe)</label>
                <input type="text" id="profile_java_path" placeholder="C:\\Program Files (x86)\\Java\\...\\javaws.exe">
                <span style="font-size:0.75rem; color:var(--primary); line-height: 1.3;">※ 지정하시는 javaws.exe 파일과 동일한 bin 폴더 내에 java.exe가 공존해야 합니다. (실시간 체크 반영)</span>
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
        let profileStatus = {};
        let activeVendorFilter = ''; // 현재 적용된 벤더 필터 값 ('': 필터 없음)
        let collapsedStates = {}; // 벤더 그룹별 접힘 상태 캐시

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

        // 아코디언 토글 제어
        funcToggleGroup = (vendor) => {
            const body = document.getElementById('group-body-' + vendor);
            const arrow = document.getElementById('group-arrow-' + vendor);
            if (!body) return;

            const isCollapsed = body.style.display === 'none';
            if (isCollapsed) {
                body.style.display = 'flex';
                arrow.innerText = '▲';
                collapsedStates[vendor] = false;
            } else {
                body.style.display = 'none';
                arrow.innerText = '▼';
                collapsedStates[vendor] = true;
            }
        };

        // 벤더 필터 토글 클릭
        function toggleFilter(vendor) {
            if (activeVendorFilter === vendor) {
                // 이미 활성화된 필터를 다시 누르면 해제
                activeVendorFilter = '';
            } else {
                // 특정 벤더 필터 활성화
                activeVendorFilter = vendor;
                // 필터링된 벤더 그룹은 자동으로 펼쳐지도록 셋팅
                collapsedStates[vendor] = false;
            }
            renderRules();
        }

        // 벤더별 그룹 박스 렌더링 (글로벌 디폴트 최상단 노출 및 기본 접힘 구현)
        function renderRules() {
            const container = document.getElementById('rule-container');
            const filterBar = document.getElementById('filter-bar');
            container.innerHTML = '';
            filterBar.innerHTML = '';

            if (!configData.rules || configData.rules.length === 0) {
                container.innerHTML = '<div style="text-align:center; padding: 2rem; color: var(--text-muted);">등록된 규칙이 없습니다.</div>';
                return;
            }

            // 1. 벤더별 그룹핑 분리
            const groups = {};
            const fallbackRules = [];
            const vendorsSet = new Set(); // 고유 벤더명 목록

            configData.rules.forEach(rule => {
                if (rule.vendor === '*' && rule.model_pattern === '*') {
                    fallbackRules.push(rule);
                } else {
                    const vendorKey = rule.vendor.toLowerCase();
                    vendorsSet.add(vendorKey);
                    
                    if (!groups[vendorKey]) {
                        groups[vendorKey] = [];
                    }
                    groups[vendorKey].push(rule);
                }
            });

            // 2. 벤더 필터 버튼 생성 및 렌더링
            const sortedVendors = Array.from(vendorsSet).sort();
            
            // 필터 전체 버튼
            const allBtn = document.createElement('button');
            allBtn.className = 'filter-btn' + (activeVendorFilter === '' ? ' active' : '');
            allBtn.innerText = '전체 보기';
            allBtn.onclick = () => { activeVendorFilter = ''; renderRules(); };
            filterBar.appendChild(allBtn);

            sortedVendors.forEach(vendor => {
                const btn = document.createElement('button');
                btn.className = 'filter-btn' + (activeVendorFilter === vendor ? ' active' : '');
                btn.innerText = vendor.toUpperCase();
                btn.onclick = () => toggleFilter(vendor);
                filterBar.appendChild(btn);
            });

            // 3. [최상단] 글로벌 Fallback 렌더링 (항상 열린 상태로 최상단 배치)
            if (fallbackRules.length > 0 && (activeVendorFilter === '')) {
                const fallbackCard = document.createElement('div');
                fallbackCard.className = 'vendor-group-card';
                fallbackCard.style.borderColor = 'rgba(0, 240, 255, 0.15)';
                fallbackCard.style.background = 'rgba(0, 240, 255, 0.01)';

                const header = document.createElement('div');
                header.className = 'vendor-group-header';
                header.style.cursor = 'default';

                const title = document.createElement('h3');
                title.className = 'vendor-group-title';
                title.style.color = 'var(--primary)';
                title.innerHTML = '🌐 글로벌 디폴트 Fallback (최우선 평가)';
                header.appendChild(title);
                fallbackCard.appendChild(header);

                fallbackRules.forEach(rule => {
                    const matchedProfile = configData.profiles.find(p => p.id === rule.profile_id);
                    const profileName = matchedProfile ? matchedProfile.name : '기본 프로필';

                    const item = document.createElement('div');
                    item.className = 'rule-item';
                    item.style.background = 'rgba(0,0,0,0.2)';

                    let ruleHtml = '\n' +
                    '    <div class="rule-info">\n' +
                    '        <div class="rule-badge-row">\n' +
                    '            <span class="badge badge-type-web">WEB (브라우저)</span>\n' +
                    '            <span class="badge badge-profile">프로필: ' + profileName + '</span>\n' +
                    '        </div>\n' +
                    '        <p class="rule-desc">\n' +
                    '            장비 타입이 ipmi가 아니거나, 스마트 라우팅 규칙 매칭에 실패 시 작동하는 전체 시스템 폴백 규칙입니다.\n' +
                    '        </p>\n' +
                    '    </div>\n' +
                    '    <div class="rule-actions">\n' +
                    '        <span style="font-size:0.8rem; color:var(--text-muted); padding-right:0.5rem;">시스템 기본 규칙</span>\n' +
                    '    </div>\n';

                    item.innerHTML = ruleHtml;
                    fallbackCard.appendChild(item);
                });

                container.appendChild(fallbackCard);
            }

            // 4. [그 아래] 벤더 그룹 렌더링
            Object.keys(groups).sort().forEach(vendor => {
                // 필터 조건이 있다면 일치하는 벤더만 렌더링
                if (activeVendorFilter !== '' && activeVendorFilter !== vendor) {
                    return;
                }

                const groupCard = document.createElement('div');
                groupCard.className = 'vendor-group-card';

                // 헤더 클릭 시 아코디언 토글
                const header = document.createElement('div');
                header.className = 'vendor-group-header';
                header.onclick = () => funcToggleGroup(vendor);

                const groupTitle = document.createElement('h3');
                groupTitle.className = 'vendor-group-title';
                groupTitle.innerHTML = '📂 ' + vendor.toUpperCase() + ' 벤더 그룹';
                header.appendChild(groupTitle);

                const arrow = document.createElement('span');
                arrow.className = 'group-toggle-arrow';
                arrow.id = 'group-arrow-' + vendor;
                
                // 접힘 상태 판별 (필터링 중이거나 캐시된 상태에 따름. 기본값은 접힘=true)
                if (collapsedStates[vendor] === undefined) {
                    collapsedStates[vendor] = true; 
                }
                
                const isCollapsed = collapsedStates[vendor];
                arrow.innerText = isCollapsed ? '▼' : '▲';
                header.appendChild(arrow);
                groupCard.appendChild(header);

                const groupBody = document.createElement('div');
                groupBody.className = 'vendor-group-body';
                groupBody.id = 'group-body-' + vendor;
                groupBody.style.display = isCollapsed ? 'none' : 'flex';

                const rulesInGroup = groups[vendor];
                rulesInGroup.forEach((rule, localIdx) => {
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

                    const isDefaultRule = rule.model_pattern === '*';
                    
                    let ruleHtml = '\n' +
                    '    <div class="rule-info">\n' +
                    '        <div class="rule-badge-row">\n' +
                    '            <span class="badge ' + typeBadgeClass + '">' + typeName + '</span>\n' +
                    '            <span class="badge badge-profile">프로필: ' + profileName + '</span>\n' +
                    '            ' + (isDefaultRule ? '<span class="badge" style="background:rgba(255,255,255,0.05); color:#ffd080">벤더 디폴트 규칙</span>' : '') + '\n' +
                    '        </div>\n' +
                    '        <p class="rule-desc">\n' +
                    '            ' + (rule.description || '설명 없음') + ' \n' +
                    '            (매칭 패턴: 벤더 <span class="rule-pattern">' + rule.vendor + '</span> / 모델 <span class="rule-pattern">' + rule.model_pattern + '</span>)\n' +
                    '        </p>\n' +
                    '    </div>\n' +
                    '    <div class="rule-actions">\n' +
                    '        <div class="order-btn-group">\n' +
                    '            <button class="order-btn" onclick="moveRuleInGroup(\'' + vendor + '\', ' + localIdx + ', -1); event.stopPropagation();">▲</button>\n' +
                    '            <button class="order-btn" onclick="moveRuleInGroup(\'' + vendor + '\', ' + localIdx + ', 1); event.stopPropagation();">▼</button>\n' +
                    '        </div>\n' +
                    '        <button class="btn btn-secondary btn-mini" onclick="openEditRuleById(\'' + rule.id + '\'); event.stopPropagation();">📝</button>\n' +
                    '        <button class="btn btn-danger btn-mini" onclick="deleteRule(\'' + rule.id + '\'); event.stopPropagation();">❌</button>\n' +
                    '    </div>\n';

                    item.innerHTML = ruleHtml;
                    groupBody.appendChild(item);
                });

                groupCard.appendChild(groupBody);
                container.appendChild(groupCard);
            });
        }

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

        function openAddRuleModal() {
            document.getElementById('rule-modal-title').innerText = '새 라우팅 규칙 추가';
            document.getElementById('rule_id').value = '';
            document.getElementById('rule_vendor').value = '';
            document.getElementById('rule_model').value = '';
            document.getElementById('rule_type').value = 'ikvm';
            document.getElementById('rule_priority').value = 1;
            document.getElementById('rule_desc').value = '';
            
            buildProfileSelectOptions('');
            document.getElementById('rule-modal').classList.add('active');
        }

        function openEditRuleById(id) {
            const rule = configData.rules.find(r => r.id === id);
            if (!rule) return;
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
            const priority = parseInt(document.getElementById('rule_priority').value) || 1;
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

        // 벤더 그룹 내부에서의 규칙 우선순위 교환
        function moveRuleInGroup(vendor, localIdx, direction) {
            const groupRules = configData.rules.filter(r => r.vendor.toLowerCase() === vendor.toLowerCase());
            
            const targetIdx = localIdx + direction;
            if (targetIdx < 0 || targetIdx >= groupRules.length) return;

            const currentRule = groupRules[localIdx];
            const targetRule = groupRules[targetIdx];

            if (currentRule.model_pattern === '*' || targetRule.model_pattern === '*') return;

            const temp = currentRule.priority;
            currentRule.priority = targetRule.priority;
            targetRule.priority = temp;

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
