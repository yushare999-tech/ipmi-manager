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
            --bg-color: #0b0b0f;
            --card-bg: rgba(22, 22, 33, 0.75);
            --card-border: rgba(255, 255, 255, 0.08);
            --text-main: #f3f3f6;
            --text-muted: #9499b3;
            --primary: #00f0ff;
            --primary-glow: rgba(0, 240, 255, 0.3);
            --secondary: #b500ff;
            --secondary-glow: rgba(181, 0, 255, 0.3);
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
                radial-gradient(at 0% 0%, rgba(181, 0, 255, 0.1) 0px, transparent 50%),
                radial-gradient(at 100% 100%, rgba(0, 240, 255, 0.08) 0px, transparent 50%);
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
            max-width: 1100px;
            display: grid;
            grid-template-columns: 1fr;
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
            box-shadow: 0 8px 32px 0 rgba(0, 0, 0, 0.37);
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
            font-size: 1.25rem;
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
            grid-template-columns: 350px 1fr;
            gap: 2rem;
        }

        @media (max-width: 900px) {
            .content-grid {
                grid-template-columns: 1fr;
            }
        }

        /* Card Base */
        .card {
            background: var(--card-bg);
            border: 1px solid var(--card-border);
            border-radius: 16px;
            padding: 2rem;
            backdrop-filter: blur(12px);
            box-shadow: 0 8px 32px 0 rgba(0, 0, 0, 0.3);
            display: flex;
            flex-direction: column;
            gap: 1.5rem;
        }

        .card-title {
            font-size: 1.2rem;
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
            font-size: 0.85rem;
            color: var(--text-muted);
            font-weight: 600;
        }

        input[type="text"], input[type="password"], input[type="number"], select, textarea {
            background: rgba(255, 255, 255, 0.03);
            border: 1px solid rgba(255, 255, 255, 0.1);
            color: var(--text-main);
            padding: 0.8rem 1rem;
            border-radius: 8px;
            font-family: var(--font-family);
            font-size: 0.95rem;
            transition: all 0.3s ease;
        }

        input:focus, select:focus, textarea:focus {
            outline: none;
            border-color: var(--primary);
            box-shadow: 0 0 10px var(--primary-glow);
            background: rgba(255, 255, 255, 0.06);
        }

        /* Buttons */
        .btn {
            background: linear-gradient(135deg, var(--primary), #00b0ff);
            color: #0b0b0f;
            border: none;
            padding: 0.8rem 1.5rem;
            border-radius: 8px;
            font-family: var(--font-family);
            font-weight: 700;
            cursor: pointer;
            transition: all 0.3s ease;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            gap: 0.5rem;
        }

        .btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 5px 15px var(--primary-glow);
            filter: brightness(1.1);
        }

        .btn-secondary {
            background: rgba(255, 255, 255, 0.06);
            color: var(--text-main);
            border: 1px solid rgba(255, 255, 255, 0.1);
        }

        .btn-secondary:hover {
            background: rgba(255, 255, 255, 0.12);
            box-shadow: none;
        }

        .btn-danger {
            background: linear-gradient(135deg, var(--danger), #ff00aa);
            color: #fff;
        }

        .btn-danger:hover {
            box-shadow: 0 5px 15px rgba(255, 0, 85, 0.4);
        }

        .btn-mini {
            padding: 0.3rem 0.6rem;
            font-size: 0.8rem;
            border-radius: 6px;
        }

        /* Table & Lists */
        .rule-list {
            display: flex;
            flex-direction: column;
            gap: 1rem;
        }

        .rule-item {
            background: rgba(255, 255, 255, 0.02);
            border: 1px solid rgba(255, 255, 255, 0.05);
            border-radius: 12px;
            padding: 1.2rem;
            display: flex;
            align-items: center;
            justify-content: space-between;
            transition: all 0.3s ease;
        }

        .rule-item:hover {
            background: rgba(255, 255, 255, 0.04);
            border-color: rgba(0, 240, 255, 0.2);
            transform: scale(1.01);
        }

        .rule-info {
            display: flex;
            flex-direction: column;
            gap: 0.4rem;
        }

        .rule-badge-row {
            display: flex;
            align-items: center;
            gap: 0.6rem;
        }

        .badge {
            padding: 0.25rem 0.6rem;
            border-radius: 6px;
            font-size: 0.75rem;
            font-weight: 700;
            text-transform: uppercase;
        }

        .badge-priority {
            background: rgba(255, 255, 255, 0.1);
            color: var(--text-main);
        }

        .badge-vendor {
            background: rgba(181, 0, 255, 0.15);
            border: 1px solid rgba(181, 0, 255, 0.4);
            color: #d880ff;
        }

        .badge-type-ikvm {
            background: rgba(0, 240, 255, 0.15);
            border: 1px solid rgba(0, 240, 255, 0.4);
            color: var(--primary);
        }

        .badge-type-jnlp {
            background: rgba(255, 170, 0, 0.15);
            border: 1px solid rgba(255, 170, 0, 0.4);
            color: #ffa500;
        }

        .badge-type-web {
            background: rgba(0, 255, 135, 0.15);
            border: 1px solid rgba(0, 255, 135, 0.4);
            color: var(--success);
        }

        .rule-desc {
            font-size: 0.85rem;
            color: var(--text-muted);
        }

        .rule-pattern {
            font-family: monospace;
            background: rgba(0, 0, 0, 0.3);
            padding: 0.1rem 0.4rem;
            border-radius: 4px;
            color: #ffd080;
        }

        .rule-actions {
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        /* Order buttons */
        .order-btn-group {
            display: flex;
            flex-direction: column;
            gap: 0.2rem;
        }

        .order-btn {
            background: none;
            border: none;
            color: var(--text-muted);
            cursor: pointer;
            font-size: 0.9rem;
            padding: 0.2rem;
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
            background: #11111a;
            border: 1px solid var(--card-border);
            border-radius: 16px;
            width: 100%;
            max-width: 500px;
            padding: 2rem;
            box-shadow: 0 10px 40px rgba(0,0,0,0.5);
            display: flex;
            flex-direction: column;
            gap: 1.5rem;
            transform: scale(0.9);
            transition: all 0.3s ease;
        }

        .modal-overlay.active .modal {
            transform: scale(1);
        }

        .modal-header {
            font-size: 1.25rem;
            font-weight: 700;
            border-bottom: 1px solid rgba(255, 255, 255, 0.08);
            padding-bottom: 0.5rem;
        }

        .modal-actions {
            display: flex;
            justify-content: flex-end;
            gap: 0.8rem;
            margin-top: 1rem;
        }

        /* Info Banner */
        .info-banner {
            background: rgba(0, 240, 255, 0.05);
            border: 1px solid rgba(0, 240, 255, 0.2);
            padding: 1rem;
            border-radius: 8px;
            font-size: 0.85rem;
            color: #b5f5ff;
            line-height: 1.4;
        }
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
                    <p style="font-size: 0.75rem; color: var(--text-muted)">Smart Routing Engine Configurator</p>
                </div>
            </div>
            <div class="status-badge">
                <span style="display: inline-block; width: 8px; height: 8px; background: var(--success); border-radius: 50%; box-shadow: 0 0 8px var(--success)"></span>
                DAEMON ONLINE
            </div>
        </header>

        <!-- Main Content Grid -->
        <div class="content-grid">
            <!-- Left Side: API settings -->
            <div class="card">
                <h2 class="card-title"><span>Js-Proxy 연동 설정</span></h2>
                <div class="info-banner">
                    IP 또는 ID 매칭 요청 발생 시, 지정된 Js-Proxy 서버에 API 요청을 보내서 해당 장비의 세부 정보(벤더, 모델, 계정)를 조회합니다.
                </div>
                
                <div class="form-group">
                    <label for="js_proxy_url">API Base URL</label>
                    <input type="text" id="js_proxy_url" placeholder="http://127.0.0.1:3000/api/device">
                </div>
                
                <div class="form-group">
                    <label for="js_proxy_token">Bearer Token (옵션)</label>
                    <input type="password" id="js_proxy_token" placeholder="API 인증 토큰 입력">
                </div>
                
                <button class="btn" id="btn-save-proxy" onclick="saveProxyConfig()">💾 설정 저장</button>
                <button class="btn btn-secondary" id="btn-test-proxy" onclick="testProxyConnection()">🔄 연결 테스트</button>
            </div>

            <!-- Right Side: Routing Rules -->
            <div class="card" style="flex: 1">
                <div class="card-title">
                    <span>스마트 라우팅 규칙 매트릭스</span>
                    <button class="btn btn-mini" onclick="openAddRuleModal()">➕ 규칙 추가</button>
                </div>
                
                <div class="info-banner" style="background: rgba(181, 0, 255, 0.03); border-color: rgba(181, 0, 255, 0.15)">
                    위에서부터 순차적으로 매칭이 진행됩니다. 매칭 조건 만족 시 지정된 실행 방식이 기동되며 판별 프로세스가 종료됩니다. 화살표 버튼을 통해 우선순위를 바꿀 수 있습니다.
                </div>

                <div class="rule-list" id="rule-container">
                    <!-- Dynamic rendering -->
                </div>
            </div>
        </div>
    </div>

    <!-- Rule Add/Edit Modal -->
    <div class="modal-overlay" id="rule-modal">
        <div class="modal">
            <h3 class="modal-header" id="modal-title">새 라우팅 규칙 추가</h3>
            
            <input type="hidden" id="rule_id">
            
            <div class="form-group">
                <label for="rule_vendor">대상 벤더 (Vendor)</label>
                <input type="text" id="rule_vendor" placeholder="예: supermicro, dell, hp (* = 모든 벤더)">
            </div>
            
            <div class="form-group">
                <label for="rule_model">모델명 매칭 키워드 (Model Keyword)</label>
                <input type="text" id="rule_model" placeholder="예: x10, r630, ilo4 (* = 모든 모델)">
            </div>
            
            <div class="form-group">
                <label for="rule_type">KVM 실행 방식 (Action)</label>
                <select id="rule_type">
                    <option value="ikvm">iKVM (iKVM.jar 직접 구동)</option>
                    <option value="jnlp">JNLP (Java Web Start 구동)</option>
                    <option value="web">Web (브라우저 자동 로그인)</option>
                </select>
            </div>
            
            <div class="form-group">
                <label for="rule_priority">우선순위 (Priority)</label>
                <input type="number" id="rule_priority" value="10" min="1" max="98">
            </div>
            
            <div class="form-group">
                <label for="rule_desc">규칙 설명</label>
                <textarea id="rule_desc" rows="3" placeholder="이 규칙에 대한 메모"></textarea>
            </div>
            
            <div class="modal-actions">
                <button class="btn btn-secondary" onclick="closeRuleModal()">취소</button>
                <button class="btn" onclick="saveRule()">저장하기</button>
            </div>
        </div>
    </div>

    <script>
        let configData = { rules: [], js_proxy_url: '', js_proxy_token: '' };

        window.addEventListener('DOMContentLoaded', () => {
            loadConfig();
        });

        // 설정 정보 일괄 로드
        function loadConfig() {
            fetch('/api/rules')
                .then(res => res.json())
                .then(data => {
                    configData = data;
                    document.getElementById('js_proxy_url').value = data.js_proxy_url || '';
                    document.getElementById('js_proxy_token').value = data.js_proxy_token || '';
                    renderRules();
                })
                .catch(err => alert('설정 로드 실패: ' + err));
        }

        // 규칙 목록 렌더링
        function renderRules() {
            const container = document.getElementById('rule-container');
            container.innerHTML = '';

            if (!configData.rules || configData.rules.length === 0) {
                container.innerHTML = '<div style="text-align:center; padding: 2rem; color: var(--text-muted);">등록된 라우팅 규칙이 없습니다.</div>';
                return;
            }

            configData.rules.forEach((rule, index) => {
                const isDefaultFallback = (rule.vendor === '*' && rule.model_pattern === '*');
                
                const item = document.createElement('div');
                item.className = 'rule-item';
                
                let typeBadgeClass = 'badge-type-web';
                let typeName = 'Web Auto-Login';
                if (rule.connect_type === 'ikvm') {
                    typeBadgeClass = 'badge-type-ikvm';
                    typeName = 'iKVM.jar 직접실행';
                } else if (rule.connect_type === 'jnlp') {
                    typeBadgeClass = 'badge-type-jnlp';
                    typeName = 'JNLP 자바웹스타트';
                }

                let ruleHtml = '\n' +
                '    <div class="rule-info">\n' +
                '        <div class="rule-badge-row">\n' +
                '            <span class="badge badge-priority">우선순위 ' + rule.priority + '</span>\n' +
                '            <span class="badge badge-vendor">' + rule.vendor.toUpperCase() + '</span>\n' +
                '            <span class="badge ' + typeBadgeClass + '">' + typeName + '</span>\n' +
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
                    ruleHtml += '        <span style="font-size:0.8rem; color:var(--text-muted);">기본 시스템 규칙</span>\n';
                }
                
                ruleHtml += '    </div>\n';

                item.innerHTML = ruleHtml;
                container.appendChild(item);
            });
        }

        // Proxy 설정 저장
        function saveProxyConfig() {
            const url = document.getElementById('js_proxy_url').value.trim();
            const token = document.getElementById('js_proxy_token').value.trim();

            configData.js_proxy_url = url;
            configData.js_proxy_token = token;

            saveAllConfig('Js-Proxy 연동 설정이 성공적으로 저장되었습니다.');
        }

        // Proxy 연결 테스트
        function testProxyConnection() {
            const url = document.getElementById('js_proxy_url').value.trim();
            const token = document.getElementById('js_proxy_token').value.trim();

            if (!url) {
                alert('테스트할 API Base URL을 입력해 주세요.');
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
                if (data.success) {
                    alert('✅ 연결 성공! Js-Proxy API에 원활히 접속할 수 있습니다.');
                } else {
                    alert('❌ 연결 실패: ' + data.error);
                }
            })
            .catch(err => alert('❌ 요청 실패: ' + err))
            .finally(() => {
                btn.innerText = '🔄 연결 테스트';
                btn.disabled = false;
            });
        }

        // 규칙 추가 모달 오픈
        function openAddRuleModal() {
            document.getElementById('modal-title').innerText = '새 라우팅 규칙 추가';
            document.getElementById('rule_id').value = '';
            document.getElementById('rule_vendor').value = '';
            document.getElementById('rule_model').value = '';
            document.getElementById('rule_type').value = 'ikvm';
            document.getElementById('rule_priority').value = (configData.rules.length > 0) ? configData.rules[configData.rules.length - 1].priority + 1 : 10;
            document.getElementById('rule_desc').value = '';
            
            document.getElementById('rule-modal').classList.add('active');
        }

        // 규칙 수정 모달 오픈
        function openEditRuleModal(index) {
            const rule = configData.rules[index];
            document.getElementById('modal-title').innerText = '라우팅 규칙 수정';
            document.getElementById('rule_id').value = rule.id;
            document.getElementById('rule_vendor').value = rule.vendor;
            document.getElementById('rule_model').value = rule.model_pattern;
            document.getElementById('rule_type').value = rule.connect_type;
            document.getElementById('rule_priority').value = rule.priority;
            document.getElementById('rule_desc').value = rule.description || '';
            
            document.getElementById('rule-modal').classList.add('active');
        }

        // 모달 닫기
        function closeRuleModal() {
            document.getElementById('rule-modal').classList.remove('active');
        }

        // 규칙 추가/수정 저장
        function saveRule() {
            const id = document.getElementById('rule_id').value;
            const vendor = document.getElementById('rule_vendor').value.trim().toLowerCase();
            const model = document.getElementById('rule_model').value.trim().toLowerCase();
            const type = document.getElementById('rule_type').value;
            const priority = parseInt(document.getElementById('rule_priority').value);
            const desc = document.getElementById('rule_desc').value.trim();

            if (!vendor || !model) {
                alert('대상 벤더와 모델 키워드를 입력해 주세요.');
                return;
            }

            const newRule = {
                id: id || ('rule_' + Date.now()),
                vendor: vendor,
                model_pattern: model,
                connect_type: type,
                priority: priority,
                description: desc
            };

            if (id) {
                // 수정
                const index = configData.rules.findIndex(r => r.id === id);
                if (index !== -1) {
                    configData.rules[index] = newRule;
                }
            } else {
                // 신규 추가
                configData.rules.push(newRule);
            }

            closeRuleModal();
            saveAllConfig('규칙이 반영되었습니다.');
        }

        // 규칙 삭제
        function deleteRule(id) {
            if (!confirm('정말 이 라우팅 규칙을 삭제하시겠습니까?')) return;
            configData.rules = configData.rules.filter(r => r.id !== id);
            saveAllConfig('규칙이 삭제되었습니다.');
        }

        // 규칙 순서 이동 (우선순위 스왑)
        function moveRule(index, direction) {
            const targetIndex = index + direction;
            if (targetIndex < 0 || targetIndex >= configData.rules.length) return;
            
            // 폴백 규칙은 건드리지 않음
            if (configData.rules[index].vendor === '*' || configData.rules[targetIndex].vendor === '*') return;

            // 우선순위(Priority) 스왑
            const temp = configData.rules[index].priority;
            configData.rules[index].priority = configData.rules[targetIndex].priority;
            configData.rules[targetIndex].priority = temp;

            saveAllConfig('우선순위가 업데이트되었습니다.');
        }

        // 최종 저장 API 전송
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
                    if (successMsg) {
                        console.log(successMsg);
                    }
                } else {
                    alert('저장 실패: ' + data.error);
                }
            })
            .catch(err => alert('저장 요청 에러: ' + err));
        }
    </script>
</body>
</html>`
