package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

// runWebServer 로컬 HTTP API 웹서버 구동
func runWebServer() {
	http.HandleFunc("/api/connect", handleConnect)
	http.HandleFunc("/api/status", handleStatus)

	port := "8080"
	logger.Infof("로컬 웹 서버 기동 중... http://127.0.0.1:%s", port)
	
	// 로컬 요청만 수용하기 위해 127.0.0.1로 바인딩
	err := http.ListenAndServe("127.0.0.1:"+port, nil)
	if err != nil {
		logger.Errorf("웹 서버 기동 실패: %v", err)
	}
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte(`{"status":"running", "message":"IPMI Manager Daemon is active"}`))
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	// CORS 허용 (외부 웹페이지 호출 연동 대비)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	id := r.URL.Query().Get("id")
	connectType := r.URL.Query().Get("type") // jnlp 또는 web

	if id == "" || connectType == "" {
		http.Error(w, `{"error":"missing query parameters 'id' or 'type'"}`, http.StatusBadRequest)
		return
	}

	config, err := readConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to read config: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	device, err := findDeviceByIDOrIP(config, id)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusNotFound)
		return
	}

	if connectType == "jnlp" {
		err = launchJnlp(config, device)
	} else if connectType == "web" {
		err = launchWeb(device)
	} else {
		err = fmt.Errorf("invalid connect type: %s", connectType)
	}

	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to launch: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(`{"success":true, "message":"launch sequence initiated successfully"}`))
}

// launchJnlp JNLP (Java Web Start) 실행 처리
func launchJnlp(config AppConfig, device Device) error {
	javaPath := config.JavaPath
	if javaPath == "" {
		// 기본 경로 자동 추정
		javaPath = "C:\\Program Files (x86)\\Java\\jre1.8.0_291\\bin\\javaws.exe" // 예시
		// 레지스트리나 환경 변수 등에서 자바 경로를 찾아보도록 폴백 가능
	}

	if _, err := os.Stat(javaPath); os.IsNotExist(err) {
		return fmt.Errorf("Java Web Start (javaws.exe)가 설정되지 않았거나 경로에 없습니다. 설정에서 자바 경로를 지정해 주세요. (경로: %s)", javaPath)
	}

	// 1. 자바 보안 예외 목록에 장비 IP 등록
	if err := AddJavaExceptionSite(device.IpmiIP); err != nil {
		logger.Errorf("[Java] 예외 사이트 등록 실패: %v", err)
	}

	// 2. 자바 레거시 TLS 규격 완화 설정 활성화
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
		// Dell iDRAC REST 로그인 시도하여 세션 토큰 획득
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
		// Supermicro: 내장된 iKVM.jar 실행
		return launchSupermicroIKVM(javaPath, device)
	} else {
		// 기타 벤더 기본 JNLP 주소
		jnlpURL = fmt.Sprintf("%s://%s/viewer.jnlp", proto, device.IpmiIP)
	}

	// javaws.exe 실행
	logger.Infof("[JNLP] javaws 실행 주소: %s", jnlpURL)
	cmd := exec.Command(javaPath, jnlpURL)
	return cmd.Start()
}

// launchSupermicroIKVM Supermicro iKVM.jar 직접 기동
func launchSupermicroIKVM(javawsPath string, device Device) error {
	// javawsPath를 기반으로 java.exe 경로 유추
	javaBin := strings.Replace(javawsPath, "javaws.exe", "java.exe", 1)
	if _, err := os.Stat(javaBin); os.IsNotExist(err) {
		return fmt.Errorf("java.exe를 찾을 수 없습니다. 경로: %s", javaBin)
	}

	// 내장 iKVM.jar 경로 탐색 (Electron 실행 위치 및 소스 위치 기준 하이브리드 검색)
	executablePath, _ := os.Executable()
	execDir := filepath.Dir(executablePath)
	
	jarPaths := []string{
		filepath.Join(execDir, "IPMIVIEW", "2.14.0", "extracted", "D_", "IPMI20", "FILES FOR IPMI VIEW", "iKVM.jar"),
		filepath.Join("C:\\Users\\kuri\\MyProJ\\ipmi-manager\\gui-client", "IPMIVIEW", "2.14.0", "extracted", "D_", "IPMI20", "FILES FOR IPMI VIEW", "iKVM.jar"), // gui-client 하위 경로
		filepath.Join("C:\\Users\\kuri\\MyProJ\\ipmi-manager", "IPMIVIEW", "2.14.0", "extracted", "D_", "IPMI20", "FILES FOR IPMI VIEW", "iKVM.jar"),            // 루트 경로 폴백
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
		"5900",   // KVM Port
		"623",    // VM Port
		"2",      // Encryption Mode (2 = SSL)
		"0",      // Virtual Media Enable Flag (0 = Disable)
	}

	// 포트 분기
	if !device.HTTPS {
		cmdArgs[11] = "1" // Non-SSL Mode Flag
	}
	_ = webPort

	cmd := exec.Command(javaBin, cmdArgs...)
	cmd.Dir = jarDir // Native DLL 로딩을 위해 작업 디렉토리를 JAR 폴더로 설정
	return cmd.Start()
}

// launchWeb 웹 브라우저 기동 (방안 1: Electron 뷰어 기동 / 방안 2: 기본 브라우저 실행 하이브리드)
func launchWeb(device Device) error {
	// 1. 로컬의 Electron 뷰어 실행 파일 탐색
	executablePath, _ := os.Executable()
	execDir := filepath.Dir(executablePath)
	
	electronPaths := []string{
		filepath.Join(execDir, "ipmi-manager.exe"),
		"C:\\Users\\kuri\\MyProJ\\ipmi-manager\\dist\\win-unpacked\\ipmi-manager.exe", // 개발 환경 빌드 경로
	}

	var electronPath string
	for _, p := range electronPaths {
		if _, err := os.Stat(p); err == nil {
			electronPath = p
			break
		}
	}

	if electronPath != "" {
		// 방안 1: 기존 Electron 앱을 기동하여 해당 장비로 자동 로그인 브라우저 화면 오픈
		logger.Infof("[Web] Electron 뷰어 기동: %s (장비 ID: %s)", electronPath, device.ID)
		cmd := exec.Command(electronPath, "--device-id="+device.ID, "--connect-type=web")
		return cmd.Start()
	}

	// 방안 2: Electron 뷰어가 없는 경우 기본 시스템 웹 브라우저로 직접 기동 (폴백)
	proto := "https"
	if !device.HTTPS {
		proto = "http"
	}
	loginUrl := fmt.Sprintf("%s://%s", proto, device.IpmiIP)
	
	// iDRAC 9인 경우 로그인 페이지 지정
	vendor := strings.ToLower(device.Vendor)
	model := strings.ToLower(device.Model)
	if vendor == "dell" && (strings.Contains(model, "r640") || strings.Contains(model, "r740")) {
		loginUrl = fmt.Sprintf("%s://%s/restgui/start.html", proto, device.IpmiIP)
	}

	logger.Infof("[Web] Electron 뷰어를 찾을 수 없어 기본 웹 브라우저로 폴백 오픈: %s", loginUrl)
	cmd := exec.Command("cmd", "/c", "start", "", loginUrl)
	return cmd.Start()
}
