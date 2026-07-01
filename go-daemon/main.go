package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/kardianos/service"
)

var logger service.Logger
var debugMode bool

// javaProcessMap tracks running iKVM/JNLP Java process PIDs per device IP for duplicate prevention.
var (
	javaProcessMu  sync.Mutex
	javaProcessMap = make(map[string]uint32) // key: deviceIP → child PID
)

type program struct {
	exit chan struct{}
}

func (p *program) Start(s service.Service) error {
	p.exit = make(chan struct{})
	go p.run()
	return nil
}

func (p *program) run() {
	logger.Info("IPMI Manager 서비스 데몬 구동 시작")
	
	// 백그라운드 웹 서버 기동 (기본 포트 8080)
	go runWebServer()

	<-p.exit
}

func (p *program) Stop(s service.Service) error {
	logger.Info("IPMI Manager 서비스 데몬 중지")
	close(p.exit)
	return nil
}

func main() {
	svcFlag := flag.String("service", "", "Control the system service: install, uninstall, start, stop")
	runFlag := flag.Bool("run", false, "Run the daemon in foreground")
	debugFlag := flag.Bool("debug", false, "Enable debug mode with verbose logging")
	flag.Parse()
	debugMode = *debugFlag

	svcConfig := &service.Config{
		Name:        "IPMIManagerDaemon",
		DisplayName: "IPMI Manager Service Daemon",
		Description: "IPMI/KVM 장비 자동 로그인 및 자바 뷰어 기동을 대행하는 백그라운드 로컬 데몬 서비스입니다.",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	if *svcFlag != "" {
		err := service.Control(s, *svcFlag)
		if err != nil {
			log.Fatalf("서비스 제어 실패 (%s): %v", *svcFlag, err)
		}
		fmt.Printf("서비스 제어 성공: %s\n", *svcFlag)
		return
	}

	// 서비스로 실행되거나 -run 플래그가 설정된 경우 실행
	if *runFlag || !service.Interactive() {
		err = s.Run()
		if err != nil {
			logger.Error(err)
		}
	} else {
		fmt.Println("사용법:")
		fmt.Println("  ipmi-daemon.exe -run                  : 포그라운드에서 즉시 실행")
		fmt.Println("  ipmi-daemon.exe -service install      : 윈도우 서비스에 등록")
		fmt.Println("  ipmi-daemon.exe -service uninstall    : 윈도우 서비스에서 삭제")
		fmt.Println("  ipmi-daemon.exe -service start        : 서비스 시작")
		fmt.Println("  ipmi-daemon.exe -service stop         : 서비스 중지")
	}
}

// readConfig Electron AppData 경로에 있는 ipmi-config.json 로드
func readConfig() (AppConfig, error) {
	var config AppConfig
	configPath := FindConfigFile("ipmi-config.json")
	
	file, err := os.Open(configPath)
	if err != nil {
		return config, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	return config, err
}

// findDeviceByIDOrIP 장비 식별자 또는 IP로 장비 탐색
func findDeviceByIDOrIP(config AppConfig, target string) (Device, error) {
	for _, d := range config.Devices {
		if d.ID == target || d.IpmiIP == target {
			return d, nil
		}
	}
	return Device{}, fmt.Errorf("device not found: %s", target)
}

// runWebServer 로컬 HTTP API 및 설정 GUI 웹서버 구동
func runWebServer() {
	// GUI 설정 페이지 서빙
	http.HandleFunc("/", handleHome)
	
	// 규칙 및 프로필 관리 API
	http.HandleFunc("/api/rules", handleGetRules)
	http.HandleFunc("/api/rules/save", handleSaveRules)
	http.HandleFunc("/api/test-proxy", handleTestProxy)
	http.HandleFunc("/api/diagnose", handleDiagnose)
	
	// 연결 처리 API
	http.HandleFunc("/api/connect", handleConnect)
	http.HandleFunc("/api/status", handleStatus)

	port := "4447"
	logger.Infof("로컬 웹 서버 및 설정 GUI 기동 완료... http://127.0.0.1:%s", port)
	
	// 로컬 요청만 수용하기 위해 127.0.0.1로 바인딩
	err := http.ListenAndServe("127.0.0.1:"+port, nil)
	if err != nil {
		logger.Errorf("웹 서버 기동 실패: %v", err)
	}
}

// handleHome 설정 GUI HTML 서빙
func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(SettingsHTML))
}

// handleGetRules 규칙 및 프로필 설정 불러오기
func handleGetRules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	config, err := LoadRulesConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to load rules: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(config)
}

// handleSaveRules 규칙 및 프로필 설정 저장하기
func handleSaveRules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	if r.Method != "POST" {
		http.Error(w, `{"error":"only POST method is allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	
	var config RulesConfig
	err := json.NewDecoder(r.Body).Decode(&config)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to decode JSON: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	
	err = SaveRulesConfig(config)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to save rules: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	
	w.Write([]byte(`{"success":true}`))
}

// handleTestProxy Js-Proxy 연결성 테스트
func handleTestProxy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	if r.Method != "POST" {
		http.Error(w, `{"error":"only POST method is allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	
	var reqBody struct {
		URL   string `json:"url"`
		Token string `json:"token"`
	}
	
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"invalid request: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	
	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest("GET", reqBody.URL, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"success":false, "error":"invalid url: %s"}`, err.Error()), http.StatusOK)
		return
	}
	
	if reqBody.Token != "" {
		req.Header.Set("Authorization", "Bearer "+reqBody.Token)
	}
	
	resp, err := client.Do(req)
	if err != nil {
		w.Write([]byte(fmt.Sprintf(`{"success":false, "error":"connection failed: %s"}`, err.Error())))
		return
	}
	defer resp.Body.Close()
	
	w.Write([]byte(`{"success":true}`))
}

// handleDiagnose 특정 프로필 경로의 자바 및 JAR 파일 실시간 검증 API
func handleDiagnose(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	profileID := r.URL.Query().Get("profile_id")
	if profileID == "" {
		http.Error(w, `{"error":"missing profile_id"}`, http.StatusBadRequest)
		return
	}

	config, err := LoadRulesConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	var matchedProfile Profile
	found := false
	for _, p := range config.Profiles {
		if p.ID == profileID {
			matchedProfile = p
			found = true
			break
		}
	}

	if !found {
		http.Error(w, `{"error":"profile not found"}`, http.StatusNotFound)
		return
	}

	// 1. javaws.exe 파일 검증
	javawsFound := false
	if matchedProfile.JavaPath != "" {
		if _, err := os.Stat(matchedProfile.JavaPath); err == nil {
			javawsFound = true
		}
	}

	// 2. java.exe 파일 검증 (javaws.exe 경로 기반 유추)
	javaFound := false
	if matchedProfile.JavaPath != "" {
		javaBin := strings.Replace(matchedProfile.JavaPath, "javaws.exe", "java.exe", 1)
		if _, err := os.Stat(javaBin); err == nil {
			javaFound = true
		}
	}

	// 3. iKVM.jar 파일 검증
	jarFound := false
	if matchedProfile.IkvmJarPath != "" {
		if _, err := os.Stat(matchedProfile.IkvmJarPath); err == nil {
			jarFound = true
		}
	}

	response := map[string]interface{}{
		"javaws_exe": javawsFound,
		"java_exe":   javaFound,
		"ikvm_jar":   jarFound,
	}

	json.NewEncoder(w).Encode(response)
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	response := map[string]string{
		"status":  "running",
		"message": "IPMI Manager Daemon is active",
		"version": Version,
	}
	json.NewEncoder(w).Encode(response)
}

// fetchDeviceFromJsProxy Js-Proxy API로부터 단일 IP 기반 장비 정보 조회
func fetchDeviceFromJsProxy(apiURL, token, ip string) (Device, error) {
	var device Device
	
	u, err := url.Parse(apiURL)
	if err != nil {
		return device, err
	}
	
	q := u.Query()
	q.Set("ip", ip)
	u.RawQuery = q.Encode()
	
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return device, err
	}
	
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	
	client := &http.Client{Timeout: 4 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return device, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return device, fmt.Errorf("API returned HTTP %d", resp.StatusCode)
	}
	
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return device, err
	}
	
	// API 응답 데이터 파싱
	// 만약 배열 형태로 들어올 경우 첫번째 데이터 추출 대응
	if strings.HasPrefix(strings.TrimSpace(string(bodyBytes)), "[") {
		var devices []Device
		err = json.Unmarshal(bodyBytes, &devices)
		if err != nil {
			return device, fmt.Errorf("failed to parse JSON array: %w", err)
		}
		if len(devices) == 0 {
			return device, fmt.Errorf("device array is empty")
		}
		device = devices[0]
	} else {
		err = json.Unmarshal(bodyBytes, &device)
		if err != nil {
			return device, fmt.Errorf("failed to parse JSON object: %w", err)
		}
	}

	// used_protocol 필드 값에 HTTPS가 포함되어 있는지 확인하여 HTTPS 설정 자동화
	device.HTTPS = strings.Contains(strings.ToUpper(device.UsedProtocol), "HTTPS")

	logger.Infof("[Connect] Js-Proxy 조회 완료 - ID: %s, Name: %s, IP: %s, Vendor: %s, Model: %s, HTTPS: %t, UsedProtocol: %s",
		device.ID, device.Name, device.IpmiIP, device.Vendor, device.Model, device.HTTPS, device.UsedProtocol)
	
	return device, nil
}

// handleConnect IP 기반 스마트 라우팅 연결 엔드포인트
func handleConnect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	ip := r.URL.Query().Get("ip")
	if ip == "" {
		// 예비 호환용 ID 파라미터 폴백 지원
		ip = r.URL.Query().Get("id")
	}

	if ip == "" {
		http.Error(w, `{"error":"missing query parameter 'ip'"}`, http.StatusBadRequest)
		return
	}

	// 규칙 설정 로드
	rulesConfig, err := LoadRulesConfig()
	if err != nil {
		logger.Warningf("[Connect] 규칙 로드 실패: %v. 기본 규칙을 적용합니다.", err)
		rulesConfig.Rules = GetDefaultRules()
		rulesConfig.Profiles = GetDefaultProfiles()
	}

	var device Device
	found := false

	// 1. Js-Proxy가 설정되어 있다면 실시간 연동 조회 우선 적용
	if rulesConfig.JsProxyURL != "" {
		logger.Infof("[Connect] Js-Proxy 조회 시도 (IP: %s, URL: %s)", ip, rulesConfig.JsProxyURL)
		d, err := fetchDeviceFromJsProxy(rulesConfig.JsProxyURL, rulesConfig.JsProxyToken, ip)
		if err == nil {
			device = d
			found = true
			logger.Infof("[Connect] Js-Proxy 조회 완료: %s (%s, Type: %s)", device.Name, device.IpmiIP, device.Type)
		} else {
			logger.Errorf("[Connect] Js-Proxy 조회 실패: %v. 로컬 설정 조회를 시도합니다.", err)
		}
	}

	// 2. Js-Proxy 실패 시 로컬 ipmi-config.json에서 검색 시도 (폴백)
	if !found {
		localConfig, err := readConfig()
		if err == nil {
			d, err := findDeviceByIDOrIP(localConfig, ip)
			if err == nil {
				device = d
				found = true
				// 로컬 조회 장비는 기본적으로 ipmi 타입으로 보정
				if device.Type == "" {
					device.Type = "ipmi"
				}
				logger.Infof("[Connect] 로컬 설정 장비 조회 성공: %s (%s)", device.Name, device.IpmiIP)
			}
		}
	}

	// 3. 만약 어떤 경로로도 장비 정보를 획득하지 못했을 경우
	if !found {
		logger.Warningf("[Connect] 장비 조회 실패 (%s). 기본 정보로 WEB 방식 강제 기동합니다.", ip)
		fallbackDevice := Device{IpmiIP: ip, HTTPS: true, Type: "ipmi"}
		viewerFound, err := launchWeb(fallbackDevice)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"success":false, "error":"failed to launch WEB fallback: %s"}`, err.Error())))
		} else {
			w.Write([]byte(fmt.Sprintf(`{"success":true, "message":"device not found. launched WEB fallback.", "viewer_found":%t, "fallback":true}`, viewerFound)))
		}
		return
	}

	// 4. 장비의 type이 "ipmi"가 아닐 경우 즉각 WEB 방식으로 연결 처리
	if strings.ToLower(device.Type) != "ipmi" {
		logger.Infof("[Router] 장비 type이 'ipmi'가 아님 (%s). WEB 방식으로 직접 연결 처리합니다.", device.Type)
		viewerFound, err := launchWeb(device)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"success":false, "error":"failed to launch WEB: %s"}`, err.Error())))
		} else {
			w.Write([]byte(fmt.Sprintf(`{"success":true, "message":"non-ipmi type device. launched WEB console.", "viewer_found":%t, "fallback":false}`, viewerFound)))
		}
		return
	}

	// 5. 스마트 라우팅 규칙 매칭
	connectType, profileID := MatchRoute(rulesConfig.Rules, device.Vendor, device.Model)
	logger.Infof("[Router] 스마트 라우팅 결과: [%s / %s] -> 방식: [%s], 프로필 ID: [%s]", device.Vendor, device.Model, connectType, profileID)

	// 6. 매칭된 프로필의 실제 경로 자원 획득
	var activeProfile Profile
	profileFound := false
	for _, p := range rulesConfig.Profiles {
		if p.ID == profileID {
			activeProfile = p
			profileFound = true
			break
		}
	}
	
	// 프로필이 없는 경우 기본 프로필 적용
	if !profileFound {
		for _, p := range rulesConfig.Profiles {
			if p.IsDefault {
				activeProfile = p
				profileFound = true
				break
			}
		}
	}
	
	// 최후의 안전장치
	if !profileFound && len(rulesConfig.Profiles) > 0 {
		activeProfile = rulesConfig.Profiles[0]
	}

	// 7. 지정된 프로필 환경 경로로 구동 처리
	runErr := error(nil)
	viewerFound := false
	if connectType == "ikvm" {
		runErr = launchSupermicroIKVM(activeProfile.JavaPath, activeProfile.IkvmJarPath, device)
	} else if connectType == "jnlp" {
		runErr = launchJnlp(activeProfile.JavaPath, device)
	} else if connectType == "WEB" {
		viewerFound, runErr = launchWeb(device)
	} else if connectType == "WEB-HTTP" {
		// WEB-HTTP: 규칙에 의해 HTTP 강제 사용 (device.HTTPS 무시)
		httpDevice := device
		httpDevice.HTTPS = false
		viewerFound, runErr = launchWeb(httpDevice)
	} else {
		runErr = fmt.Errorf("unknown connect type: %s", connectType)
	}

	// 8. 구동 에러 발생 시 최후의 보루로 WEB 방식 자동 폴백 기동
	if runErr != nil {
		logger.Errorf("[Connect] 방식 기동 실패 (%v). WEB 방식으로 최종 폴백을 구동합니다.", runErr)
		fallbackViewerFound, fallbackErr := launchWeb(device)
		w.Header().Set("Content-Type", "application/json")
		if fallbackErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"success":false, "error":"failed to launch connect and fallback: %s"}`, fallbackErr.Error())))
			return
		}
		
		// 에러 스트링 이스케이프 처리
		errStr := strings.ReplaceAll(runErr.Error(), `"`, `\"`)
		errStr = strings.ReplaceAll(errStr, `\`, `\\`)
		w.Write([]byte(fmt.Sprintf(`{"success":true, "message":"connect failed. fallback to WEB initiated.", "viewer_found":%t, "fallback":true, "connect_error":"%s"}`, fallbackViewerFound, errStr)))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"success":true, "message":"launch sequence initiated successfully via %s", "viewer_found":%t, "fallback":false}`, connectType, viewerFound)))
}

// launchJnlp JNLP (Java Web Start) 실행 처리
func launchJnlp(javaPath string, device Device) error {
	if javaPath == "" {
		javaPath = "C:\\Program Files (x86)\\Java\\jre1.8.0_291\\bin\\javaws.exe"
	}

	if _, err := os.Stat(javaPath); os.IsNotExist(err) {
		return fmt.Errorf("Java Web Start (javaws.exe)가 경로에 존재하지 않습니다. 경로: %s", javaPath)
	}

	if err := AddJavaExceptionSite(device.IpmiIP); err != nil {
		logger.Errorf("[Java] 예외 사이트 등록 실패: %v", err)
	}

	if err := ApplyLegacyJavaConfig(); err != nil {
		logger.Errorf("[Java] 레거시 자바 설정 적용 실패: %v", err)
	}

	proto := "https"
	if !device.HTTPS {
		proto = "http"
	}

	var jnlpURL string
	vendor := strings.ToLower(device.Vendor)

	if vendor == "dell" {
		logger.Infof("[JNLP] Dell iDRAC REST 로그인 시도: %s", device.IpmiIP)
		tokenString, err := IdracLogin(device)
		if err != nil {
			logger.Errorf("[JNLP] REST 로그인 실패 (%v). 토큰 없이 JNLP 호출을 수행합니다.", err)
			jnlpURL = fmt.Sprintf("%s://%s/viewer.jnlp?EXTPORT=-1&JNLPSTR=AppletRedirection", proto, device.IpmiIP)
		} else {
			jnlpURL = fmt.Sprintf("%s://%s/viewer.jnlp?EXTPORT=-1&JNLPSTR=AppletRedirection&%s", proto, device.IpmiIP, tokenString)
			logger.Infof("[JNLP] REST 로그인 성공 및 토큰 연동 완료: %s", device.IpmiIP)
		}
	} else if vendor == "supermicro" || strings.Contains(strings.ToLower(device.Model), "x10drl") || strings.Contains(strings.ToLower(device.Model), "x9drl") {
		// Supermicro의 경우 예외적으로 iKVM 직접 실행으로 리다이렉션 (안전장치)
		executablePath, _ := os.Executable()
		execDir := filepath.Dir(executablePath)
		defaultJar := filepath.Join(execDir, "IPMIVIEW", "2.14.0", "extracted", "D_", "IPMI20", "FILES FOR IPMI VIEW", "iKVM.jar")
		return launchSupermicroIKVM(javaPath, defaultJar, device)
	} else {
		jnlpURL = fmt.Sprintf("%s://%s/viewer.jnlp", proto, device.IpmiIP)
	}

	logger.Infof("[JNLP] javaws 실행 주소: %s", jnlpURL)

	// ── 중복 c29지 ────────────────────────────────────────────────────────────────────
	javaProcessMu.Lock()
	existingPID, hasPID := javaProcessMap[device.IpmiIP]
	javaProcessMu.Unlock()

	if hasPID && isProcessRunning(existingPID) {
		logger.Infof("[JNLP] 동일 IP에 대한 실행 중인 Java 프로세스 감지 (PID: %d). 전면 활성화.", existingPID)
		return activateJavaWindow(existingPID)
	}
	// ─────────────────────────────────────────────────────────────────────────

	pid, err := launchAsUserWithPID(javaPath, []string{jnlpURL}, "")
	if err == nil && pid > 0 {
		javaProcessMu.Lock()
		javaProcessMap[device.IpmiIP] = pid
		javaProcessMu.Unlock()
		// 실행 직후 잘보이는 위치로 이동
		go activateJavaWindowDelayed(pid, 3*time.Second)
	}
	return err
}

// launchSupermicroIKVM Supermicro iKVM.jar 직접 기동
func launchSupermicroIKVM(javaPath, jarPath string, device Device) error {
	javaBin := strings.Replace(javaPath, "javaws.exe", "java.exe", 1)
	if _, err := os.Stat(javaBin); os.IsNotExist(err) {
		return fmt.Errorf("java.exe를 찾을 수 없습니다. 경로: %s", javaBin)
	}

	if _, err := os.Stat(jarPath); os.IsNotExist(err) {
		return fmt.Errorf("iKVM.jar 파일을 찾을 수 없습니다. 경로: %s", jarPath)
	}

	jarDir := filepath.Dir(jarPath)
	hostName := device.Name
	if hostName == "" {
		hostName = device.IpmiIP
	}

	webPort := "443"
	if !device.HTTPS {
		webPort = "80"
	}

	cmdArgs := []string{
		"-Dsun.java2d.noddraw=true",
		"-Dsun.java2d.d3d=false",
		"-Dsun.java2d.uiScale=1.0",
		"-jar",
		jarPath,
		device.IpmiIP,
		device.Username,
		device.Password,
		hostName,
		"5900",
		"623",
		"2",
		"0",
	}

	if !device.HTTPS {
		cmdArgs[11] = "1"
	}
	_ = webPort

	// ── 중복 방지 ────────────────────────────────────────────────────────────────────
	javaProcessMu.Lock()
	existingPID, hasPID := javaProcessMap[device.IpmiIP]
	javaProcessMu.Unlock()

	if hasPID && isProcessRunning(existingPID) {
		logger.Infof("[iKVM] 동일 IP에 대한 실행 중인 iKVM 프로세스 감지 (PID: %d). 전면 활성화.", existingPID)
		return activateJavaWindow(existingPID)
	}
	// ─────────────────────────────────────────────────────────────────────────

	pid, err := launchAsUserWithPID(javaBin, cmdArgs, jarDir)
	if err == nil && pid > 0 {
		javaProcessMu.Lock()
		javaProcessMap[device.IpmiIP] = pid
		javaProcessMu.Unlock()
		go activateJavaWindowDelayed(pid, 4*time.Second)
	}
	return err
}

// launchWeb 웹 브라우저 기동 (WEB 방식)
// 반환: (viewerFound bool, err error)
func launchWeb(device Device) (bool, error) {
	executablePath, _ := os.Executable()
	execDir := filepath.Dir(executablePath)

	proto := "https"
	if !device.HTTPS {
		proto = "http"
	}
	loginUrl := fmt.Sprintf("%s://%s", proto, device.IpmiIP)

	vendor := strings.ToLower(device.Vendor)
	model := strings.ToLower(device.Model)
	if vendor == "dell" && (strings.Contains(model, "r640") || strings.Contains(model, "r740")) {
		loginUrl = fmt.Sprintf("%s://%s/restgui/start.html", proto, device.IpmiIP)
	}

	// 순수 Go 기반 초경량 크로미움 웹 뷰어 탐색 경로
	viewerPaths := []string{
		filepath.Join(execDir, "ipmi-viewer.exe"),
		filepath.Join(filepath.Dir(execDir), "ipmi-viewer.exe"),
		"C:\\Users\\kuri\\MyProJ\\ipmi-manager\\go-daemon\\ipmi-viewer.exe",
	}

	var viewerPath string
	for _, p := range viewerPaths {
		if _, err := os.Stat(p); err == nil {
			viewerPath = p
			break
		}
	}

	// 초경량 웹 뷰어가 존재하면 자식 프로세스로 기동하여 계정 자동 완성 처리
	if viewerPath != "" {
		logger.Infof("[WEB] 초경량 Go 웹 뷰어 기동: %s (IP: %s, URL: %s)", viewerPath, device.IpmiIP, loginUrl)
		args := []string{
			"--url=" + loginUrl,
			"--user=" + device.Username,
			"--pass=" + device.Password,
			"--vendor=" + device.Vendor,
			"--ip=" + device.IpmiIP, // 중복 방지용 기기 식별자
		}
		if debugMode {
			args = append(args, "--debug")
			logger.Infof("[WEB] [DEBUG] 디버그 옵션을 활성화하여 뷰어를 구동합니다. (인자: %v)", args)
		}
		return true, launchAsUser(viewerPath, args, "")
	}

	// 뷰어를 찾을 수 없는 경우 최후의 폴백으로 기본 웹 브라우저 직접 기동
	logger.Warningf("[WEB] 초경량 웹 뷰어(ipmi-viewer.exe)를 찾을 수 없어 기본 웹 브라우저로 직접 접속: %s", loginUrl)
	return false, launchAsUser("cmd", []string{"/c", "start", "", loginUrl}, "")
}

// resolveViewerPath finds ipmi-viewer.exe relative to the daemon executable.
func resolveViewerPath() string {
	executablePath, _ := os.Executable()
	execDir := filepath.Dir(executablePath)
	candidates := []string{
		filepath.Join(execDir, "ipmi-viewer.exe"),
		filepath.Join(filepath.Dir(execDir), "ipmi-viewer.exe"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// activateJavaWindow launches ipmi-viewer.exe in user session with --activate-pid flag.
// ipmi-viewer.exe will enumerate Session 1 windows for the given PID and bring it to front.
func activateJavaWindow(pid uint32) error {
	viewerPath := resolveViewerPath()
	if viewerPath == "" {
		logger.Warningf("[Activate] ipmi-viewer.exe를 찾을 수 없어 창 활성화를 건너뜁니다.")
		return nil
	}
	logger.Infof("[Activate] ipmi-viewer.exe --activate-pid=%d 실행", pid)
	return launchAsUser(viewerPath, []string{fmt.Sprintf("--activate-pid=%d", pid)}, "")
}

// activateJavaWindowDelayed waits for the Java process to create its window, then activates it.
func activateJavaWindowDelayed(pid uint32, delay time.Duration) {
	time.Sleep(delay)
	if isProcessRunning(pid) {
		if err := activateJavaWindow(pid); err != nil {
			logger.Warningf("[Activate] 창 활성화 실패 (PID: %d): %v", pid, err)
		}
	}
}
