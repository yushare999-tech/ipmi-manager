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
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kardianos/service"
)

var logger service.Logger

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
	flag.Parse()

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
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return config, fmt.Errorf("APPDATA environment variable not found")
	}

	configPath := filepath.Join(appData, "ipmi-manager", "ipmi-config.json")
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
	
	// 규칙 관리 API
	http.HandleFunc("/api/rules", handleGetRules)
	http.HandleFunc("/api/rules/save", handleSaveRules)
	http.HandleFunc("/api/test-proxy", handleTestProxy)
	
	// 연결 처리 API
	http.HandleFunc("/api/connect", handleConnect)
	http.HandleFunc("/api/status", handleStatus)

	port := "8080"
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

// handleGetRules 규칙 설정 불러오기
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

// handleSaveRules 규칙 설정 저장하기
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
	
	// 테스트용으로 간단하게 GET 요청 수행 (타임아웃 3초)
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

func handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte(`{"status":"running", "message":"IPMI Manager Daemon is active"}`))
}

// fetchDeviceFromJsProxy Js-Proxy API로부터 장비 상세 정보 조회
func fetchDeviceFromJsProxy(apiURL, token, target string) (Device, error) {
	var device Device
	
	u, err := url.Parse(apiURL)
	if err != nil {
		return device, err
	}
	
	q := u.Query()
	// 타겟이 IP 형식이면 ip 파라미터로, 아니면 id 파라미터로 전송
	if strings.Contains(target, ".") {
		q.Set("ip", target)
	} else {
		q.Set("id", target)
	}
	u.RawQuery = q.Encode()
	
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return device, err
	}
	
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	
	client := &http.Client{Timeout: 5 * time.Second}
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
	
	err = json.Unmarshal(bodyBytes, &device)
	if err != nil {
		return device, fmt.Errorf("failed to parse JSON: %w (Response: %s)", err, string(bodyBytes))
	}
	
	return device, nil
}

// handleConnect 스마트 라우팅을 반영한 KVM 커넥트 엔드포인트
func handleConnect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// id 또는 ip 파라미터를 유연하게 수용
	target := r.URL.Query().Get("id")
	if target == "" {
		target = r.URL.Query().Get("ip")
	}
	
	// 사용자가 명시적으로 지정을 원할 경우에 대비한 수동 오버라이드 타입 (옵션)
	manualType := r.URL.Query().Get("type") 

	if target == "" {
		http.Error(w, `{"error":"missing query parameter 'id' or 'ip'"}`, http.StatusBadRequest)
		return
	}

	// 1. 규칙 파일 및 연동 정보 로드
	rulesConfig, err := LoadRulesConfig()
	if err != nil {
		logger.Warningf("규칙 파일을 로드하지 못해 기본 규칙을 사용합니다: %v", err)
		rulesConfig.Rules = GetDefaultRules()
	}

	var device Device
	found := false

	// 2. 일차적으로 로컬 ipmi-config.json에서 탐색 시도
	localConfig, err := readConfig()
	if err == nil {
		d, err := findDeviceByIDOrIP(localConfig, target)
		if err == nil {
			device = d
			found = true
			logger.Infof("[Connect] 로컬 설정에서 장비 식별 성공: %s (%s)", device.Name, device.IpmiIP)
		}
	}

	// 3. 로컬에 없거나 Js-Proxy 주소가 설정되어 있다면 외부 API를 통해 동적으로 장비 획득
	if !found && rulesConfig.JsProxyURL != "" {
		logger.Infof("[Connect] Js-Proxy를 통한 장비 조회 시도: %s (URL: %s)", target, rulesConfig.JsProxyURL)
		d, err := fetchDeviceFromJsProxy(rulesConfig.JsProxyURL, rulesConfig.JsProxyToken, target)
		if err == nil {
			device = d
			found = true
			logger.Infof("[Connect] Js-Proxy API 장비 조회 성공: %s (%s)", device.Name, device.IpmiIP)
		} else {
			logger.Errorf("[Connect] Js-Proxy API 조회 실패: %v", err)
		}
	}

	if !found {
		http.Error(w, fmt.Sprintf(`{"error":"device not found locally or via Js-Proxy: %s"}`, target), http.StatusNotFound)
		return
	}

	// 4. KVM 구동 방식 판별 (스마트 라우팅 규칙 적용)
	connectType := manualType
	if connectType == "" {
		connectType = MatchRoute(rulesConfig.Rules, device.Vendor, device.Model)
		logger.Infof("[Router] 스마트 라우팅 판별 완료: [%s / %s] -> 실행 방식: [%s]", device.Vendor, device.Model, connectType)
	} else {
		logger.Infof("[Router] 사용자 수동 지정 적용 -> 실행 방식: [%s]", connectType)
	}

	// 5. 판별 결과에 따른 구동 실행
	if connectType == "ikvm" {
		err = launchSupermicroIKVM(localConfig.JavaPath, device)
	} else if connectType == "jnlp" {
		err = launchJnlp(localConfig, device)
	} else if connectType == "web" {
		err = launchWeb(device)
	} else {
		err = fmt.Errorf("invalid connect type: %s", connectType)
	}

	if err != nil {
		logger.Errorf("[Connect] 실행 실패: %v", err)
		http.Error(w, fmt.Sprintf(`{"error":"failed to launch: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(fmt.Sprintf(`{"success":true, "message":"launch sequence initiated successfully via %s"}`, connectType)))
}

// launchJnlp JNLP (Java Web Start) 실행 처리
func launchJnlp(config AppConfig, device Device) error {
	javaPath := config.JavaPath
	if javaPath == "" {
		javaPath = "C:\\Program Files (x86)\\Java\\jre1.8.0_291\\bin\\javaws.exe"
	}

	if _, err := os.Stat(javaPath); os.IsNotExist(err) {
		return fmt.Errorf("Java Web Start (javaws.exe)가 설정되지 않았거나 경로에 없습니다. 설정에서 자바 경로를 지정해 주세요. (경로: %s)", javaPath)
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
		return launchSupermicroIKVM(javaPath, device)
	} else {
		jnlpURL = fmt.Sprintf("%s://%s/viewer.jnlp", proto, device.IpmiIP)
	}

	logger.Infof("[JNLP] javaws 실행 주소: %s", jnlpURL)
	cmd := exec.Command(javaPath, jnlpURL)
	return cmd.Start()
}

// launchSupermicroIKVM Supermicro iKVM.jar 직접 기동
func launchSupermicroIKVM(javawsPath string, device Device) error {
	javaBin := strings.Replace(javawsPath, "javaws.exe", "java.exe", 1)
	if _, err := os.Stat(javaBin); os.IsNotExist(err) {
		return fmt.Errorf("java.exe를 찾을 수 없습니다. 경로: %s", javaBin)
	}

	executablePath, _ := os.Executable()
	execDir := filepath.Dir(executablePath)
	
	jarPaths := []string{
		filepath.Join(execDir, "IPMIVIEW", "2.14.0", "extracted", "D_", "IPMI20", "FILES FOR IPMI VIEW", "iKVM.jar"),
		filepath.Join("C:\\Users\\kuri\\MyProJ\\ipmi-manager\\gui-client", "IPMIVIEW", "2.14.0", "extracted", "D_", "IPMI20", "FILES FOR IPMI VIEW", "iKVM.jar"),
		filepath.Join("C:\\Users\\kuri\\MyProJ\\ipmi-manager", "IPMIVIEW", "2.14.0", "extracted", "D_", "IPMI20", "FILES FOR IPMI VIEW", "iKVM.jar"),
	}

	var ikvmJar string
	for _, p := range jarPaths {
		if _, err := os.Stat(p); err == nil {
			ikvmJar = p
			break
		}
	}

	if ikvmJar == "" {
		return fmt.Errorf("내장된 Supermicro iKVM.jar 파일을 찾을 수 없습니다.")
	}

	jarDir := filepath.Dir(ikvmJar)
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
		ikvmJar,
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

	cmd := exec.Command(javaBin, cmdArgs...)
	cmd.Dir = jarDir
	return cmd.Start()
}

// launchWeb 웹 브라우저 기동
func launchWeb(device Device) error {
	executablePath, _ := os.Executable()
	execDir := filepath.Dir(executablePath)
	
	electronPaths := []string{
		filepath.Join(execDir, "ipmi-manager.exe"),
		"C:\\Users\\kuri\\MyProJ\\ipmi-manager\\dist\\win-unpacked\\ipmi-manager.exe",
	}

	var electronPath string
	for _, p := range electronPaths {
		if _, err := os.Stat(p); err == nil {
			electronPath = p
			break
		}
	}

	if electronPath != "" {
		logger.Infof("[Web] Electron 뷰어 기동: %s (장비 ID: %s)", electronPath, device.ID)
		cmd := exec.Command(electronPath, "--device-id="+device.ID, "--connect-type=web")
		return cmd.Start()
	}

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

	logger.Infof("[Web] Electron 뷰어를 찾을 수 없어 기본 웹 브라우저로 폴백 오픈: %s", loginUrl)
	cmd := exec.Command("cmd", "/c", "start", "", loginUrl)
	return cmd.Start()
}

