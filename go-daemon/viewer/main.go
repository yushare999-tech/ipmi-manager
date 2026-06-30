package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/jchv/go-webview2"
)

// Version Web Viewer Version (Auto incremented by build script)
const Version = "1.5.13"

func main() {
	// Bypass SSL/TLS security warnings, self-signed certificate errors and allow legacy TLS 1.0/1.1
	os.Setenv("WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS", "--ignore-certificate-errors --ignore-ssl-errors --ssl-version-min=tls1 --tls-version-min=tls1")

	// Parse CLI arguments
	targetURL := flag.String("url", "", "IPMI Web Console URL")
	username := flag.String("user", "", "IPMI Username")
	password := flag.String("pass", "", "IPMI Password")
	vendor := flag.String("vendor", "", "Hardware Vendor (dell, supermicro, hp, etc.)")
	debugMode := flag.Bool("debug", false, "Enable Edge DevTools (Inspect Element)")
	flag.Parse()

	if *targetURL == "" {
		log.Fatal("Error: --url parameter is required")
	}

	// Validate and sanitize URL
	parsedURL, err := url.Parse(*targetURL)
	if err != nil {
		log.Fatalf("Error parsing URL: %v", err)
	}
	if parsedURL.Scheme == "" {
		*targetURL = "https://" + *targetURL
	}

	// Create WebView2 instance (1280x800, debug flag controls DevTools availability)
	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:     *debugMode,
		AutoFocus: true,
		WindowOptions: webview2.WindowOptions{
			Title: "IPMI Web Viewer v" + Version,
		},
	})
	if w == nil {
		log.Fatal("Failed to create WebView2 instance. Make sure WebView2 Runtime is installed.")
	}
	defer w.Destroy()

	w.SetSize(1280, 800, webview2.HintNone)

	// Build JavaScript for auto-filling credentials and clicking login with console tracking
	jsTemplate := `
	(function() {
		var user = '%s';
		var pass = '%s';
		var vendor = '%s';
		console.log("[IPMI-Viewer] Auto-login script started. Vendor: " + vendor + ", User: " + user);

		function injectCredentials() {
			var filled = false;

			// 1. Dell iDRAC 9 (HTML5)
			var drac9User = document.getElementById('user');
			var drac9Pass = document.getElementById('password');
			var drac9Btn = document.getElementById('submit-button');
			if (drac9User && drac9Pass) {
				console.log("[IPMI-Viewer] Dell iDRAC 9 input elements detected.");
				drac9User.value = user;
				drac9Pass.value = pass;
				drac9User.dispatchEvent(new Event('input', { bubbles: true }));
				drac9Pass.dispatchEvent(new Event('input', { bubbles: true }));
				filled = true;
				if (drac9Btn) {
					console.log("[IPMI-Viewer] Triggering Dell iDRAC 9 login click.");
					setTimeout(function() { drac9Btn.click(); }, 500);
					return true;
				}
			}

			// 2. Dell iDRAC 7 / 8
			var drac8User = document.getElementById('username');
			var drac8Pass = document.getElementById('password');
			var drac8Btn = document.getElementById('loginBtn');
			if (drac8User && drac8Pass) {
				console.log("[IPMI-Viewer] Dell iDRAC 7/8 input elements detected.");
				drac8User.value = user;
				drac8Pass.value = pass;
				drac8User.dispatchEvent(new Event('input', { bubbles: true }));
				drac8Pass.dispatchEvent(new Event('input', { bubbles: true }));
				filled = true;
				if (drac8Btn) {
					console.log("[IPMI-Viewer] Triggering Dell iDRAC 7/8 login click.");
					setTimeout(function() { drac8Btn.click(); }, 500);
					return true;
				}
			}

			// 3. Supermicro IPMI
			var smUser = document.querySelector('input[name="username"]');
			var smPass = document.querySelector('input[name="password"]');
			var smBtn = document.querySelector('input[type="submit"]') || document.getElementById('login_btn') || document.querySelector('.login-btn');
			if (smUser && smPass) {
				console.log("[IPMI-Viewer] Supermicro IPMI input elements detected.");
				smUser.value = user;
				smPass.value = pass;
				filled = true;
				if (smBtn) {
					console.log("[IPMI-Viewer] Triggering Supermicro IPMI login click.");
					setTimeout(function() { smBtn.click(); }, 500);
					return true;
				}
			}

			// 4. HP iLO 3 / 4 / 5
			var iloUser = document.getElementById('ilo_user') || document.getElementById('username') || document.querySelector('input[name="username"]');
			var iloPass = document.getElementById('ilo_password') || document.getElementById('password') || document.querySelector('input[name="password"]');
			var iloBtn = document.getElementById('login_button') || document.getElementById('loginBtn') || document.querySelector('button[type="submit"]');
			if (iloUser && iloPass) {
				console.log("[IPMI-Viewer] HP iLO input elements detected.");
				iloUser.value = user;
				iloPass.value = pass;
				iloUser.dispatchEvent(new Event('input', { bubbles: true }));
				iloPass.dispatchEvent(new Event('input', { bubbles: true }));
				filled = true;
				if (iloBtn) {
					console.log("[IPMI-Viewer] Triggering HP iLO login click.");
					setTimeout(function() { iloBtn.click(); }, 500);
					return true;
				}
			}

			// 5. Generic Fallback
			var genUser = document.querySelector('input[type="text"]') || document.querySelector('input[name*="user" i]') || document.querySelector('input[id*="user" i]');
			var genPass = document.querySelector('input[type="password"]');
			var genBtn = document.querySelector('input[type="submit"]') || document.querySelector('button[type="submit"]');
			if (genUser && genPass && !filled) {
				console.log("[IPMI-Viewer] Generic login input elements detected.");
				genUser.value = user;
				genPass.value = pass;
				genUser.dispatchEvent(new Event('input', { bubbles: true }));
				genPass.dispatchEvent(new Event('input', { bubbles: true }));
				if (genBtn) {
					console.log("[IPMI-Viewer] Triggering Generic login click.");
					setTimeout(function() { genBtn.click(); }, 500);
					return true;
				}
			}
			return false;
		}

		// Detect elements periodically (max 8 seconds)
		var attempts = 0;
		var intervalId = setInterval(function() {
			var success = injectCredentials();
			attempts++;
			if (success || attempts > 16) {
				console.log("[IPMI-Viewer] Autologin attempt stopped. Success: " + success + ", Attempts: " + attempts);
				clearInterval(intervalId);
			}
		}, 500);
	})();
	`

	// Escape special characters and bind data
	escapedUser := strings.ReplaceAll(*username, "'", "\\'")
	escapedPass := strings.ReplaceAll(*password, "'", "\\'")
	escapedVendor := strings.ReplaceAll(*vendor, "'", "\\'")
	jsCode := fmt.Sprintf(jsTemplate, escapedUser, escapedPass, escapedVendor)

	// Inject script to run on page load
	w.Init(jsCode)

	// Navigate to target and run main loop
	w.Navigate(*targetURL)
	w.Run()
}
