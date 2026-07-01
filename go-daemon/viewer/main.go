package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/jchv/go-webview2"
)

var (
	modUser32                  = syscall.NewLazyDLL("user32.dll")
	modKernel32win             = syscall.NewLazyDLL("kernel32.dll")
	procShowWindow             = modUser32.NewProc("ShowWindow")
	procFindWindowW            = modUser32.NewProc("FindWindowW")
	procSetForegroundWin       = modUser32.NewProc("SetForegroundWindow")
	procBringWindowToTop       = modUser32.NewProc("BringWindowToTop")
	procSetWindowPos           = modUser32.NewProc("SetWindowPos")
	procGetCursorPos           = modUser32.NewProc("GetCursorPos")
	procMonitorFromPoint       = modUser32.NewProc("MonitorFromPoint")
	procGetMonitorInfoW        = modUser32.NewProc("GetMonitorInfoW")
	procGetSystemMetrics       = modUser32.NewProc("GetSystemMetrics")
	procCreateMutexW           = modKernel32win.NewProc("CreateMutexW")
	procEnumWindows            = modUser32.NewProc("EnumWindows")
	procGetWindowThreadProcID  = modUser32.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible        = modUser32.NewProc("IsWindowVisible")
	procGetWindowTextLenW      = modUser32.NewProc("GetWindowTextLengthW")
)

const (
	swHide                  = 0
	swShowMinimized         = 2
	swShowNormal            = 1
	swpNoSize               = 0x0001
	swpNoMove               = 0x0002
	swpShowWindow           = 0x0040
	hwndTopmost             = ^uintptr(0) // HWND_TOPMOST  = -1
	hwndNoTopmost           = ^uintptr(1) // HWND_NOTOPMOST = -2
	monitorDefaultToNearest = 0x00000002
)

type winPOINT struct {
	X, Y int32
}

type winRECT struct {
	Left, Top, Right, Bottom int32
}

type winMONITORINFO struct {
	CbSize    uint32
	RcMonitor winRECT
	RcWork    winRECT // work area excludes taskbar
	DwFlags   uint32
}

// Version Web Viewer Version (Auto incremented by build script)
const Version = "1.5.31"

// landingPageHTML is the dark-themed loading page shown before navigating to the IPMI URL.
const landingPageHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>IPMI Manager - Connecting</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{
  background:linear-gradient(135deg,#060d1f 0%,#0f1a2e 50%,#060d1f 100%);
  font-family:'Segoe UI',system-ui,sans-serif;
  display:flex;flex-direction:column;align-items:center;justify-content:center;
  height:100vh;overflow:hidden;color:#e2e8f0;
  user-select:none;
}
.logo-row{display:flex;align-items:center;gap:10px;margin-bottom:8px}
.logo-icon{
  width:36px;height:36px;
  background:linear-gradient(135deg,#38bdf8,#818cf8);
  border-radius:10px;display:flex;align-items:center;justify-content:center;
  font-size:20px;box-shadow:0 0 18px rgba(56,189,248,0.35);
}
.logo-label{font-size:13px;font-weight:600;letter-spacing:3px;color:#64748b;text-transform:uppercase}
h1{
  font-size:22px;font-weight:700;margin:12px 0 4px;
  background:linear-gradient(120deg,#38bdf8 30%,#818cf8 100%);
  -webkit-background-clip:text;-webkit-text-fill-color:transparent;background-clip:text;
}
.sub{font-size:11px;color:#334155;letter-spacing:1px;margin-bottom:30px;text-transform:uppercase}
.ip-chip{
  display:inline-flex;align-items:center;gap:7px;
  background:#0f1d35;border:1px solid #1e3a5f;
  border-radius:99px;padding:5px 16px;
  font-size:12px;color:#7dd3fc;letter-spacing:0.5px;
  margin-bottom:30px;font-family:monospace;
}
.dot{width:7px;height:7px;background:#22d3ee;border-radius:50%;animation:pulse 1.2s ease-in-out infinite}
@keyframes pulse{0%,100%{opacity:1;box-shadow:0 0 0 0 rgba(34,211,238,.4)}50%{opacity:.5;box-shadow:0 0 0 5px rgba(34,211,238,0)}}
.prog-wrap{width:260px}
.prog-track{
  height:4px;background:#0f1d35;border-radius:99px;overflow:hidden;
  border:1px solid #1e3a5f;
}
.prog-bar{
  height:100%;width:0%;
  background:linear-gradient(90deg,#38bdf8,#818cf8);
  border-radius:99px;
  transition:width 0.05s linear;
  box-shadow:0 0 10px rgba(56,189,248,.6);
}
.prog-labels{display:flex;justify-content:space-between;margin-top:10px}
.status-txt{font-size:10px;color:#334155;letter-spacing:.5px}
.pct-txt{font-size:10px;color:#1e3a5f;font-family:monospace}
.grid{
  position:fixed;inset:0;
  background-image:linear-gradient(rgba(56,189,248,.03) 1px,transparent 1px),
    linear-gradient(90deg,rgba(56,189,248,.03) 1px,transparent 1px);
  background-size:32px 32px;
  pointer-events:none;
}
</style>
</head>
<body>
<div class="grid"></div>
<div class="logo-row">
  <div class="logo-icon">?裕?/div>
  <div class="logo-label">IPMI Manager</div>
</div>
<h1>?癒?봄??뽯선 ?꾩꼷???怨뚭퍙</h1>
<div class="sub">Secure Management Console</div>
<div class="ip-chip">
  <div class="dot"></div>
  <span id="ip-text">Resolving host...</span>
</div>
<div class="prog-wrap">
  <div class="prog-track"><div class="prog-bar" id="bar"></div></div>
  <div class="prog-labels">
    <span class="status-txt" id="status">Initializing connection</span>
    <span class="pct-txt" id="pct">0%</span>
  </div>
</div>
<script>
(function(){
  // Extract target URL from global variable
  var target = window.__IPMI_TARGET__ || '';
  var ip = target.replace(/https?:\/\//,'').split('/')[0].split('?')[0];
  document.getElementById('ip-text').textContent = ip || 'Unknown';
  
  var bar = document.getElementById('bar');
  var statusEl = document.getElementById('status');
  var pctEl = document.getElementById('pct');
  var msgs = ['Initializing connection','Authenticating session','Loading console','Ready'];
  var p = 0;
  var iv = setInterval(function(){
    p += 2;
    bar.style.width = p + '%';
    pctEl.textContent = p + '%';
    if(p===25) statusEl.textContent = msgs[1];
    if(p===60) statusEl.textContent = msgs[2];
    if(p===90) statusEl.textContent = msgs[3];
    if(p>=100){ clearInterval(iv); if(target) window.location.href = target; }
  }, 25);
})();
</script>
</body>
</html>`

func main() {
	// Bypass SSL/TLS security warnings, self-signed certificate errors and allow legacy TLS 1.0/1.1
	os.Setenv("WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS", "--ignore-certificate-errors --ignore-ssl-errors --ssl-version-min=tls1 --tls-version-min=tls1")

	// ???? --activate-pid 筌뤴뫀諭? Java ?袁⑥쨮?紐꾨뮞 筌≪럩??筌≪럩??筌≪럩????욧탢??뤿연 筌ㅼ뮇?억쭖?곸몵嚥?揶쎛?紐꾩긾 ????????????????????????????????????????????????
	// ??筌뤴뫀諭??WebView2 ?λ뜃由????곸뵠 筌앸맩????쎈뻬??랁??ル굝利??몃빍??
	for _, arg := range os.Args[1:] {
		if len(arg) > 14 && arg[:14] == "--activate-pid" {
			var targetPID uint32
			fmt.Sscanf(arg[15:], "%d", &targetPID)
			if targetPID > 0 {
				activateWindowByPID(targetPID)
			}
			return
		}
	}
	// ??????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????

	// Parse CLI arguments
	targetURL := flag.String("url", "", "IPMI Web Console URL")
	username := flag.String("user", "", "IPMI Username")
	password := flag.String("pass", "", "IPMI Password")
	vendor := flag.String("vendor", "", "Hardware Vendor (dell, supermicro, hp, etc.)")
	ipFlag := flag.String("ip", "", "Device IP for deduplication")
	debugMode := flag.Bool("debug", false, "Enable Edge DevTools (Inspect Element)")
	minimized := flag.Bool("minimized", false, "Start window minimized to taskbar")
	hidden := flag.Bool("hidden", false, "Start window completely hidden")
	flag.Parse()

	// Determine the IP key: prefer --ip flag, fallback to extracting from URL
	deviceIP := *ipFlag
	if deviceIP == "" && *targetURL != "" {
		p, _ := url.Parse(*targetURL)
		deviceIP = p.Hostname()
	}

	// ???? Duplicate Prevention ??????????????????????????????????????????????????????????????????????????????????????????????????
	// Create a named Global mutex keyed by device IP.
	// If the mutex already exists, an ipmi-viewer for this IP is already running:
	// bring that window to the front and exit immediately.
	winTitle := "IPMI Viewer - " + deviceIP
	safeName := strings.ReplaceAll(deviceIP, ".", "-")
	safeName = strings.ReplaceAll(safeName, ":", "-")
	mutexName, _ := syscall.UTF16PtrFromString("Global\\IPMI-Viewer-" + safeName)
	hMutex, _, mutexErr := procCreateMutexW.Call(0, 1, uintptr(unsafe.Pointer(mutexName)))
	if mutexErr == syscall.ERROR_ALREADY_EXISTS {
		// An existing viewer for this IP is open ??activate it and quit
		existingTitle, _ := syscall.UTF16PtrFromString(winTitle)
		hwndExisting, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(existingTitle)))
		if hwndExisting != 0 {
			procShowWindow.Call(hwndExisting, swShowNormal)
			procSetWindowPos.Call(hwndExisting, hwndTopmost, 0, 0, 0, 0, swpNoSize|swpNoMove)
			procSetForegroundWin.Call(hwndExisting)
			procBringWindowToTop.Call(hwndExisting)
			time.Sleep(100 * time.Millisecond)
			procSetWindowPos.Call(hwndExisting, hwndNoTopmost, 0, 0, 0, 0, swpNoSize|swpNoMove)
		}
		// Close the duplicate handle and exit ??the original mutex holder stays open
		if hMutex != 0 {
			syscall.CloseHandle(syscall.Handle(hMutex))
		}
		return
	}
	// Hold the mutex for the lifetime of this viewer process
	if hMutex != 0 {
		defer syscall.CloseHandle(syscall.Handle(hMutex))
	}
	// ??????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????

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

	// Create WebView2 instance ??640??80 default size
	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:     *debugMode,
		AutoFocus: true,
		WindowOptions: webview2.WindowOptions{
			Title: winTitle, // unique per device IP
		},
	})
	if w == nil {
		log.Fatal("Failed to create WebView2 instance. Make sure WebView2 Runtime is installed.")
	}
	defer w.Destroy()

	w.SetSize(640, 480, webview2.HintNone)

	// Window state logic: asynchronously wait for valid HWND, center on active monitor with fallback, and bring to front
	go func() {
		const winW, winH = 800, 600
		var hwnd uintptr

		// Poll for native window creation (max 2 seconds)
		for i := 0; i < 40; i++ {
			hwnd = uintptr(w.Window())
			if hwnd != 0 {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}

		if hwnd == 0 {
			log.Println("[IPMI-Viewer] [Warning] Failed to retrieve native HWND. Skip window positioning.")
			return
		}

		if *hidden {
			procShowWindow.Call(hwnd, swHide)
			return
		} else if *minimized {
			procShowWindow.Call(hwnd, swShowMinimized)
			return
		}

		// Calculate center coordinates
		var posX, posY int32
		positioned := false

		// 1. Try to find active monitor work area (excluding taskbar)
		var pt winPOINT
		if ret, _, _ := procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt))); ret != 0 {
			hMon, _, _ := procMonitorFromPoint.Call(uintptr(pt.X), uintptr(pt.Y), monitorDefaultToNearest)
			if hMon != 0 {
				var mi winMONITORINFO
				mi.CbSize = uint32(unsafe.Sizeof(mi))
				if retInfo, _, _ := procGetMonitorInfoW.Call(hMon, uintptr(unsafe.Pointer(&mi))); retInfo != 0 && mi.RcWork.Right > mi.RcWork.Left {
					workW := mi.RcWork.Right - mi.RcWork.Left
					workH := mi.RcWork.Bottom - mi.RcWork.Top
					posX = mi.RcWork.Left + (workW-winW)/2
					posY = mi.RcWork.Top + (workH-winH)/2
					positioned = true
				}
			}
		}

		// 2. Fail-safe Fallback: Use system screen metrics if monitor API fails
		if !positioned {
			scrW, _, _ := procGetSystemMetrics.Call(0) // SM_CXSCREEN
			scrH, _, _ := procGetSystemMetrics.Call(1) // SM_CYSCREEN
			if scrW > 0 && scrH > 0 {
				posX = (int32(scrW) - winW) / 2
				posY = (int32(scrH) - winH) / 2
			} else {
				// Absolute default fallback coordinates
				posX = 300
				posY = 200
			}
		}

		// Apply window sizing & position
		procShowWindow.Call(hwnd, swShowNormal)
		procSetWindowPos.Call(
			hwnd,
			hwndTopmost,
			uintptr(posX), uintptr(posY),
			winW, winH,
			swpShowWindow,
		)

		// Set focus & foreground
		procSetForegroundWin.Call(hwnd)
		procBringWindowToTop.Call(hwnd)

		// Restore normal Z-order in background to prevent locking user out
		time.Sleep(200 * time.Millisecond)
		procSetWindowPos.Call(hwnd, hwndNoTopmost, uintptr(posX), uintptr(posY), winW, winH, swpShowWindow)
	}()

	// Escape credential values for safe JS embedding
	escapedURL := strings.ReplaceAll(*targetURL, "'", "\\'")
	escapedUser := strings.ReplaceAll(*username, "'", "\\'")
	escapedPass := strings.ReplaceAll(*password, "'", "\\'")
	escapedVendor := strings.ReplaceAll(*vendor, "'", "\\'")

	// Inject global vars (used by landing page redirect + autologin on the IPMI page)
	w.Init(fmt.Sprintf(`window.__IPMI_TARGET__='%s';`, escapedURL))

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
	jsCode := fmt.Sprintf(jsTemplate, escapedUser, escapedPass, escapedVendor)

	// Inject autologin script (runs on every page load including the IPMI target page)
	w.Init(jsCode)

	// Navigate to the landing page first (base64 encoded data URI ??no length limit issues)
	landingB64 := base64.StdEncoding.EncodeToString([]byte(landingPageHTML))
	w.Navigate("data:text/html;base64," + landingB64)

	// Run main loop (landing page JS will redirect to targetURL after animation)
	w.Run()
}

// activateWindowByPID finds the first visible top-level window owned by the given PID
// and brings it to the foreground, centered on the active monitor.
// Called via: ipmi-viewer.exe --activate-pid=<PID>
func activateWindowByPID(pid uint32) {
	// Retry for up to 10 seconds waiting for the Java window to appear
	var foundHWND uintptr
	for attempt := 0; attempt < 20 && foundHWND == 0; attempt++ {
		if attempt > 0 {
			time.Sleep(500 * time.Millisecond)
		}

		// EnumWindows callback: find a visible, titled window belonging to this PID
		cb := syscall.NewCallback(func(hwnd, _ uintptr) uintptr {
			var winPID uint32
			procGetWindowThreadProcID.Call(hwnd, uintptr(unsafe.Pointer(&winPID)))
			if winPID != pid {
				return 1 // continue
			}
			vis, _, _ := procIsWindowVisible.Call(hwnd)
			if vis == 0 {
				return 1 // continue
			}
			// prefer windows that have a title
			titleLen, _, _ := procGetWindowTextLenW.Call(hwnd)
			if titleLen == 0 && foundHWND != 0 {
				return 1 // skip untitled if we already have a candidate
			}
			foundHWND = hwnd
			return 0 // stop enumeration
		})
		procEnumWindows.Call(cb, 0)
	}

	if foundHWND == 0 {
		log.Printf("[Activate] PID %d ?????媛?쒖쟻 李쎌쓣 李얠? 紐삵뻽?듬땲??", pid)
		return
	}

	// 1. Restore if minimized
	procShowWindow.Call(foundHWND, uintptr(swShowNormal))
	time.Sleep(80 * time.Millisecond)

	// 2. Get active monitor info for centering
	var cursorPt struct{ X, Y int32 }
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&cursorPt)))
	hMonitor, _, _ := procMonitorFromPoint.Call(
		uintptr(cursorPt.X), uintptr(cursorPt.Y),
		2, // MONITOR_DEFAULTTONEAREST
	)

	type monitorInfo struct {
		CbSize    uint32
		RcMonitor struct{ Left, Top, Right, Bottom int32 }
		RcWork    struct{ Left, Top, Right, Bottom int32 }
		DwFlags   uint32
	}
	var mi monitorInfo
	mi.CbSize = uint32(unsafe.Sizeof(mi))
	procGetMonitorInfoW.Call(hMonitor, uintptr(unsafe.Pointer(&mi)))

	mW := mi.RcWork.Right - mi.RcWork.Left
	mH := mi.RcWork.Bottom - mi.RcWork.Top
	mX := mi.RcWork.Left
	mY := mi.RcWork.Top

	// Fallback to GetSystemMetrics if monitor info unavailable
	if mW == 0 || mH == 0 {
		mW32, _, _ := procGetSystemMetrics.Call(0)
		mH32, _, _ := procGetSystemMetrics.Call(1)
		mW = int32(mW32)
		mH = int32(mH32)
	}

	// Center with reasonable Java KVM window size
	winW := int32(800)
	winH := int32(600)
	x := mX + (mW-winW)/2
	y := mY + (mH-winH)/2

	// 3. Move to center of active monitor
	const swpNoSize = 0x0001
	const swpShowWindow = 0x0040
	procSetWindowPos.Call(foundHWND, uintptr(hwndTopmost), uintptr(x), uintptr(y), uintptr(winW), uintptr(winH), swpShowWindow)
	time.Sleep(50 * time.Millisecond)

	// 4. Bring to front
	procSetForegroundWin.Call(foundHWND)
	procBringWindowToTop.Call(foundHWND)
	time.Sleep(100 * time.Millisecond)

	// 5. Remove always-on-top
	procSetWindowPos.Call(foundHWND, uintptr(hwndNoTopmost), 0, 0, 0, 0, uintptr(swpNoSize|swpNoMove))

	log.Printf("[Activate] PID %d 李??쒖꽦???꾨즺.", pid)
}
