package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// getJavaDeploymentDir Java의 Deployment 폴더 경로 획득
func getJavaDeploymentDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "AppData", "LocalLow", "Sun", "Java", "Deployment")
}

// AddJavaExceptionSite Java 예외 사이트 목록에 URL 등록
func AddJavaExceptionSite(site string) error {
	deployDir := getJavaDeploymentDir()
	if deployDir == "" {
		return fmt.Errorf("user home directory not found")
	}

	securityDir := filepath.Join(deployDir, "security")
	if err := os.MkdirAll(securityDir, 0755); err != nil {
		return err
	}

	exceptionFile := filepath.Join(securityDir, "exception.sites")
	
	// 프로토콜 규격 매핑 (http:// 또는 https:// 가 없으면 둘 다 추가)
	var targets []string
	if !strings.HasPrefix(site, "http://") && !strings.HasPrefix(site, "https://") {
		targets = append(targets, "http://"+site, "https://"+site)
	} else {
		targets = append(targets, site)
	}

	// 기존 파일 읽기
	existingSites := make(map[string]bool)
	if _, err := os.Stat(exceptionFile); err == nil {
		file, err := os.Open(exceptionFile)
		if err == nil {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" {
					existingSites[line] = true
				}
			}
		}
	}

	// 누락된 사이트 추가
	modified := false
	for _, t := range targets {
		if !existingSites[t] {
			existingSites[t] = true
			modified = true
		}
	}

	if !modified {
		return nil // 이미 모든 사이트가 등록됨
	}

	// 파일 다시 쓰기
	file, err := os.OpenFile(exceptionFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for s := range existingSites {
		if _, err := writer.WriteString(s + "\n"); err != nil {
			return err
		}
	}
	return writer.Flush()
}

// ApplyLegacyJavaConfig 자바 보안 정책 완화 (TLS 1.0, 1.1 활성화)
func ApplyLegacyJavaConfig() error {
	deployDir := getJavaDeploymentDir()
	if deployDir == "" {
		return fmt.Errorf("user home directory not found")
	}

	propFile := filepath.Join(deployDir, "deployment.properties")
	
	// 기본 설정 주입 대상 항목들
	targetConfigs := map[string]string{
		"deployment.security.TLSv1.1": "true",
		"deployment.security.TLSv1":   "true",
		"deployment.security.SSLv3":   "true",
		"deployment.security.SSLv2Hello": "true",
	}

	existingProps := make(map[string]string)

	// 기존 파일 로드
	if _, err := os.Stat(propFile); err == nil {
		file, err := os.Open(propFile)
		if err == nil {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					existingProps[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}
		}
	}

	// 설정 병합 및 변경점 확인
	modified := false
	for k, v := range targetConfigs {
		if existingProps[k] != v {
			existingProps[k] = v
			modified = true
		}
	}

	if !modified {
		return nil
	}

	// 덮어쓰기 저장
	file, err := os.OpenFile(propFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for k, v := range existingProps {
		if _, err := writer.WriteString(fmt.Sprintf("%s=%s\n", k, v)); err != nil {
			return err
		}
	}
	return writer.Flush()
}

// PatchJavaSecurity java.security 파일 강제 패치 (구형 암호화 알고리즘 허용)
func PatchJavaSecurity(javawsPath string) (map[string]string, error) {
	if javawsPath == "" {
		return nil, fmt.Errorf("javaws path is empty")
	}

	// javawsPath가 bin/javaws.exe 일 경우, java.security의 위치는 ../lib/security/java.security 임
	binDir := filepath.Dir(javawsPath)
	jreDir := filepath.Dir(binDir)
	securityPath := filepath.Join(jreDir, "lib", "security", "java.security")

	if _, err := os.Stat(securityPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("java.security file not found at: %s", securityPath)
	}

	// 백업 파일 생성
	backupPath := securityPath + ".bak"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		if err := copyFile(securityPath, backupPath); err != nil {
			return nil, fmt.Errorf("failed to create backup: %s", err.Error())
		}
	}

	// 파일 읽기 및 수정
	file, err := os.Open(securityPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	disabledAlgorithmsModified := false
	disabledAlgorithmsKey := "jdk.tls.disabledAlgorithms="

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		
		// 제한 리스트 주석 처리 또는 축소
		if strings.HasPrefix(trimmed, disabledAlgorithmsKey) && !strings.Contains(trimmed, "#") {
			// 주석 처리하여 완전 규제 해제
			line = "#" + line
			disabledAlgorithmsModified = true
		}
		lines = append(lines, line)
	}

	if !disabledAlgorithmsModified {
		return map[string]string{"status": "already_patched", "path": securityPath}, nil
	}

	// 쓰기
	outFile, err := os.OpenFile(securityPath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		// 권한 에러 가능성 있음
		return nil, fmt.Errorf("permission denied to write java.security. Please run as administrator. (%s)", err.Error())
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return nil, err
		}
	}
	if err := writer.Flush(); err != nil {
		return nil, err
	}

	return map[string]string{"status": "patched", "path": securityPath}, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
