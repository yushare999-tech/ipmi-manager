package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Profile 자바 환경 및 iKVM.jar 관련 버전별 프로필 정의
type Profile struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	JavaPath    string `json:"java_path"`      // javaws.exe 경로
	IkvmJarPath string `json:"ikvm_jar_path"`  // iKVM.jar 경로
	IsDefault   bool   `json:"is_default"`     // 기본 프로필 여부
	Description string `json:"description"`
}

// Rule KVM 접속 방식 매칭 규칙 정의 (프로필 ID 연동)
type Rule struct {
	ID           string `json:"id"`
	Vendor       string `json:"vendor"`        // 대상 벤더 (예: dell, supermicro, hp, *)
	ModelPattern string `json:"model_pattern"` // 모델명 매칭 키워드 (예: R630, X10, *)
	ConnectType  string `json:"connect_type"`  // 실행할 방식 (ikvm, jnlp, WEB)
	ProfileID    string `json:"profile_id"`    // 연동할 프로필 ID
	Priority     int    `json:"priority"`      // 우선순위 (낮을수록 우선 매칭)
	Description  string `json:"description"`
}

// RulesConfig 규칙 파일 전체 구조 (프로필 및 프록시 설정 포함)
type RulesConfig struct {
	Rules        []Rule    `json:"rules"`
	Profiles     []Profile `json:"profiles"`
	JsProxyURL   string    `json:"js_proxy_url"`   // Js-Proxy API 엔드포인트
	JsProxyToken string    `json:"js_proxy_token"` // API 인증 토큰
}

var rulesConfigPath string

func init() {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = "."
	}
	rulesConfigPath = filepath.Join(appData, "ipmi-manager", "rules-config.json")
}

// GetDefaultProfiles 기본 실행 프로필 생성 (Node 버전 기본 탑재 사양 적용)
func GetDefaultProfiles() []Profile {
	return []Profile{
		{
			ID:          "profile_default",
			Name:        "기본 Java JRE 8 환경 (Default)",
			JavaPath:    "C:\\Program Files (x86)\\Java\\jre1.8.0_291\\bin\\javaws.exe",
			IkvmJarPath: "C:\\Users\\kuri\\MyProJ\\ipmi-manager\\IPMIVIEW\\2.14.0\\extracted\\D_\\IPMI20\\FILES FOR IPMI VIEW\\iKVM.jar",
			IsDefault:   true,
			Description: "기본 자바 8 및 내장 Supermicro iKVM.jar(v2.14.0) 구동용 프로필",
		},
	}
}

// GetDefaultRules Node.js 버전 제어 로직 기준 기본 매칭 규칙 구성
func GetDefaultRules() []Rule {
	return []Rule{
		{
			ID:           "rule_supermicro_x9_ikvm",
			Vendor:       "supermicro",
			ModelPattern: "x9",
			ConnectType:  "ikvm",
			ProfileID:    "profile_default",
			Priority:     1,
			Description:  "Supermicro X9 세대 장비 iKVM.jar 직접 구동",
		},
		{
			ID:           "rule_supermicro_x10_ikvm",
			Vendor:       "supermicro",
			ModelPattern: "x10",
			ConnectType:  "ikvm",
			ProfileID:    "profile_default",
			Priority:     2,
			Description:  "Supermicro X10 세대 장비 iKVM.jar 직접 구동 (UAC 우회)",
		},
		{
			ID:           "rule_supermicro_x11_ikvm",
			Vendor:       "supermicro",
			ModelPattern: "x11",
			ConnectType:  "ikvm",
			ProfileID:    "profile_default",
			Priority:     3,
			Description:  "Supermicro X11 세대 장비 iKVM.jar 직접 구동",
		},
		{
			ID:           "rule_dell_idrac8_r630",
			Vendor:       "dell",
			ModelPattern: "r630",
			ConnectType:  "jnlp",
			ProfileID:    "profile_default",
			Priority:     4,
			Description:  "Dell iDRAC 8 (R630) - REST 세션 토큰 연동 JNLP 구동",
		},
		{
			ID:           "rule_dell_idrac8_r730",
			Vendor:       "dell",
			ModelPattern: "r730",
			ConnectType:  "jnlp",
			ProfileID:    "profile_default",
			Priority:     5,
			Description:  "Dell iDRAC 8 (R730) - REST 세션 토큰 연동 JNLP 구동",
		},
		{
			ID:           "rule_dell_idrac9_r640",
			Vendor:       "dell",
			ModelPattern: "r640",
			ConnectType:  "WEB",
			ProfileID:    "profile_default",
			Priority:     6,
			Description:  "Dell iDRAC 9 (R640) - 내장 HTML5 웹 콘솔 직접 연결",
		},
		{
			ID:           "rule_dell_idrac9_r740",
			Vendor:       "dell",
			ModelPattern: "r740",
			ConnectType:  "WEB",
			ProfileID:    "profile_default",
			Priority:     7,
			Description:  "Dell iDRAC 9 (R740) - 내장 HTML5 웹 콘솔 직접 연결",
		},
		{
			ID:           "rule_hp_ilo3_jnlp",
			Vendor:       "hp",
			ModelPattern: "ilo3",
			ConnectType:  "jnlp",
			ProfileID:    "profile_default",
			Priority:     8,
			Description:  "HP iLO 3 장비 JNLP (자바 웹 스타트) 구동",
		},
		{
			ID:           "rule_hp_ilo4_jnlp",
			Vendor:       "hp",
			ModelPattern: "ilo4",
			ConnectType:  "jnlp",
			ProfileID:    "profile_default",
			Priority:     9,
			Description:  "HP iLO 4 장비 JNLP (자바 웹 스타트) 구동",
		},
		{
			ID:           "rule_hp_ilo5_web",
			Vendor:       "hp",
			ModelPattern: "ilo5",
			ConnectType:  "WEB",
			ProfileID:    "profile_default",
			Priority:     10,
			Description:  "HP iLO 5 장비 HTML5 웹 콘솔 직접 연결",
		},
		{
			ID:           "rule_fallback_default",
			Vendor:       "*",
			ModelPattern: "*",
			ConnectType:  "WEB",
			ProfileID:    "profile_default",
			Priority:     99,
			Description:  "매칭 조건이 없거나 예외 발생 시 기본 WEB 방식으로 폴백",
		},
	}
}

// LoadRulesConfig 규칙 및 프로필 설정 파일 로드
func LoadRulesConfig() (RulesConfig, error) {
	var config RulesConfig

	dir := filepath.Dir(rulesConfigPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		_ = os.MkdirAll(dir, 0755)
	}

	if _, err := os.Stat(rulesConfigPath); os.IsNotExist(err) {
		config.Profiles = GetDefaultProfiles()
		config.Rules = GetDefaultRules()
		config.JsProxyURL = "https://js-proxy.jscomz.net/api/devices" // 실제 API 엔드포인트 기본값 지정
		config.JsProxyToken = ""
		err = SaveRulesConfig(config)
		if err != nil {
			return config, err
		}
		return config, nil
	}

	fileBytes, err := ioutil.ReadFile(rulesConfigPath)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(fileBytes, &config)
	if err != nil {
		return config, err
	}

	// 규칙 우선순위 정렬
	SortRules(config.Rules)

	return config, nil
}

// SaveRulesConfig 규칙 및 프로필 설정 파일 저장
func SaveRulesConfig(config RulesConfig) error {
	SortRules(config.Rules)

	if len(config.Profiles) == 0 {
		config.Profiles = GetDefaultProfiles()
	}

	fileBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(rulesConfigPath, fileBytes, 0644)
}

// SortRules 규칙 슬라이스를 우선순위 오름차순으로 정렬
func SortRules(rules []Rule) {
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority < rules[j].Priority
	})
}

// MatchRoute 장비 스펙(Vendor, Model)을 기반으로 매칭되는 규칙의 ConnectType과 ProfileID 반환
func MatchRoute(rules []Rule, vendor, model string) (string, string) {
	vLower := strings.ToLower(strings.TrimSpace(vendor))
	mLower := strings.ToLower(strings.TrimSpace(model))

	for _, rule := range rules {
		rVendor := strings.ToLower(strings.TrimSpace(rule.Vendor))
		rModel := strings.ToLower(strings.TrimSpace(rule.ModelPattern))

		// 1. 벤더 매칭 체크
		vendorMatched := false
		if rVendor == "*" || rVendor == "" {
			vendorMatched = true
		} else if strings.Contains(vLower, rVendor) {
			vendorMatched = true
		}

		// 2. 모델 매칭 체크
		modelMatched := false
		if rModel == "*" || rModel == "" {
			modelMatched = true
		} else if strings.Contains(mLower, rModel) {
			modelMatched = true
		}

		// 둘 다 매칭되면 해당 규칙의 ConnectType 및 ProfileID 반환
		if vendorMatched && modelMatched {
			return rule.ConnectType, rule.ProfileID
		}
	}

	// 기본 폴백 방식 반환
	return "WEB", ""
}
