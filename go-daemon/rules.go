package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Rule KVM 접속 방식 매칭 규칙 정의
type Rule struct {
	ID           string `json:"id"`
	Vendor       string `json:"vendor"`        // 대상 벤더 (예: dell, supermicro, hp, *)
	ModelPattern string `json:"model_pattern"` // 모델명 매칭 키워드 (예: R630, X10, *)
	ConnectType  string `json:"connect_type"`  // 실행할 방식 (ikvm, jnlp, web)
	Priority     int    `json:"priority"`      // 우선순위 (낮을수록 우선 매칭)
	Description  string `json:"description"`
}

// RulesConfig 규칙 파일 전체 구조 (Js-Proxy 설정 연동 포함)
type RulesConfig struct {
	Rules        []Rule `json:"rules"`
	JsProxyURL   string `json:"js_proxy_url"`   // Js-Proxy API 엔드포인트
	JsProxyToken string `json:"js_proxy_token"` // API 인증 토큰
}

var rulesConfigPath string

func init() {
	// AppData 내에 ipmi-manager 폴더 경로 확인
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = "."
	}
	rulesConfigPath = filepath.Join(appData, "ipmi-manager", "rules-config.json")
}

// GetDefaultRules 기본 내장 매칭 규칙 반환
func GetDefaultRules() []Rule {
	return []Rule{
		{
			ID:           "rule_default_supermicro_ikvm",
			Vendor:       "supermicro",
			ModelPattern: "x10",
			ConnectType:  "ikvm",
			Priority:     1,
			Description:  "Supermicro X10 세대 이상 iKVM.jar 직접 구동",
		},
		{
			ID:           "rule_default_supermicro_ikvm_x11",
			Vendor:       "supermicro",
			ModelPattern: "x11",
			ConnectType:  "ikvm",
			Priority:     2,
			Description:  "Supermicro X11 세대 iKVM.jar 직접 구동",
		},
		{
			ID:           "rule_default_dell_idrac8",
			Vendor:       "dell",
			ModelPattern: "r630",
			ConnectType:  "jnlp",
			Priority:     3,
			Description:  "Dell iDRAC 8 장비 JNLP (자바 웹 스타트) 구동",
		},
		{
			ID:           "rule_default_dell_idrac8_r730",
			Vendor:       "dell",
			ModelPattern: "r730",
			ConnectType:  "jnlp",
			Priority:     4,
			Description:  "Dell iDRAC 8 (R730) 장비 JNLP 구동",
		},
		{
			ID:           "rule_default_dell_idrac9",
			Vendor:       "dell",
			ModelPattern: "r640",
			ConnectType:  "web",
			Priority:     5,
			Description:  "Dell iDRAC 9 (R640) HTML5 웹 콘솔 직접 로그인",
		},
		{
			ID:           "rule_default_dell_idrac9_r740",
			Vendor:       "dell",
			ModelPattern: "r740",
			ConnectType:  "web",
			Priority:     6,
			Description:  "Dell iDRAC 9 (R740) HTML5 웹 콘솔 직접 로그인",
		},
		{
			ID:           "rule_default_fallback",
			Vendor:       "*",
			ModelPattern: "*",
			ConnectType:  "web",
			Priority:     99,
			Description:  "매칭되는 규칙이 없을 때 기본 웹 자동 로그인으로 폴백",
		},
	}
}

// LoadRulesConfig 규칙 및 프록시 설정 파일 로드
func LoadRulesConfig() (RulesConfig, error) {
	var config RulesConfig

	// 디렉토리가 없으면 생성
	dir := filepath.Dir(rulesConfigPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		_ = os.MkdirAll(dir, 0755)
	}

	// 파일이 없으면 기본 설정 파일 생성
	if _, err := os.Stat(rulesConfigPath); os.IsNotExist(err) {
		config.Rules = GetDefaultRules()
		config.JsProxyURL = "http://127.0.0.1:3000/api/device" // 기본값 예시
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

	// 우선순위 정렬
	SortRules(config.Rules)

	return config, nil
}

// SaveRulesConfig 규칙 및 프록시 설정 파일 저장
func SaveRulesConfig(config RulesConfig) error {
	// 저장하기 전에 우선순위 정렬
	SortRules(config.Rules)

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

// MatchRoute 장비 스펙(Vendor, Model)을 바탕으로 매칭되는 규칙의 ConnectType 반환
func MatchRoute(rules []Rule, vendor, model string) string {
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

		// 둘 다 매칭되면 해당 룰의 ConnectType 반환
		if vendorMatched && modelMatched {
			return rule.ConnectType
		}
	}

	// 최악의 경우 기본값인 web 반환
	return "web"
}
