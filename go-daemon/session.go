package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Device 장비 스펙 정의 (ipmi-config.json 및 Js-Proxy 연동 규격)
type Device struct {
	ID           string `json:"js_serial"`    // js_serial -> id
	Name         string `json:"hostname"`     // hostname -> name
	IpmiIP       string `json:"ip"`           // ip -> ipmi_ip
	MAC          string `json:"mac"`          // mac address
	Vendor       string `json:"vendor"`
	Model        string `json:"model"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Note         string `json:"network_memo"` // network_memo -> note
	HTTPS        bool   `json:"https"`
	UsedProtocol string `json:"used_protocol"`
	Type         string `json:"type"`         // 장비 제어 타입 (예: ipmi, ssh 등)
}

// Config 데이터 스펙
type AppConfig struct {
	Devices    []Device `json:"devices"`
	JavaPath   string   `json:"javaPath"`   // javaws.exe 경로
	ApiBaseUrl string   `json:"apiBaseUrl"`
	ApiToken   string   `json:"apiToken"`
}

// getLegacyHTTPClient 구형 iDRAC 6/7/8 SSL/TLS 규격을 지원하는 HTTP 클라이언트 생성
func getLegacyHTTPClient() *http.Client {
	// TLS 1.0 및 취약한 구형 암호화 스위트들 명시적 활성화
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS10,
		CipherSuites: []uint16{
			tls.TLS_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA, // iDRAC 6 대응
			tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,       // iDRAC 6 대응
		},
	}

	transport := &http.Transport{
		TLSClientConfig:   tlsConfig,
		DisableKeepAlives: true,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}
}

// IdracLogin Dell iDRAC REST 로그인 및 ST1/ST2 세션 토큰 문자열 반환
func IdracLogin(device Device) (string, error) {
	client := getLegacyHTTPClient()
	proto := "https"
	if !device.HTTPS {
		proto = "http"
	}

	loginURL := fmt.Sprintf("%s://%s/data/login", proto, device.IpmiIP)

	form := url.Values{}
	form.Add("user", device.Username)
	form.Add("password", device.Password)

	req, err := http.NewRequest("POST", loginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error status: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	body := string(bodyBytes)
	
	// 응답 XML 또는 JSON 분석하여 <forwardUrl> 내의 ST1, ST2 토큰 추출
	// 예: <forwardUrl>index.html?ST1=...,ST2=...</forwardUrl>
	if !strings.Contains(body, "ST1=") {
		return "", fmt.Errorf("login response does not contain ST1 token. (Response: %s)", body)
	}

	// 단순 문자열 파싱으로 ST1, ST2 파라미터 추출
	startIdx := strings.Index(body, "ST1=")
	if startIdx == -1 {
		return "", fmt.Errorf("failed to locate ST1 parameter")
	}

	// 끝점 찾기 (< 또는 " 등 구분자)
	endIdx := len(body)
	for i := startIdx; i < len(body); i++ {
		c := body[i]
		if c == '<' || c == '"' || c == '\'' || c == ' ' || c == '}' || c == ']' {
			endIdx = i
			break
		}
	}

	tokenString := body[startIdx:endIdx]
	return tokenString, nil
}
