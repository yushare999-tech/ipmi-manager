package main

import (
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"
)

var (
	modWtsapi32 = syscall.NewLazyDLL("wtsapi32.dll")
	modKernel32 = syscall.NewLazyDLL("kernel32.dll")
	modAdvapi32 = syscall.NewLazyDLL("advapi32.dll")
	modUserenv  = syscall.NewLazyDLL("userenv.dll")

	procWTSGetActiveConsoleSessionId = modKernel32.NewProc("WTSGetActiveConsoleSessionId")
	procWTSQueryUserToken            = modWtsapi32.NewProc("WTSQueryUserToken")
	procCreateProcessAsUser          = modAdvapi32.NewProc("CreateProcessAsUserW")
	procCreateEnvironmentBlock       = modUserenv.NewProc("CreateEnvironmentBlock")
	procDestroyEnvironmentBlock      = modUserenv.NewProc("DestroyEnvironmentBlock")
	procDuplicateTokenEx             = modAdvapi32.NewProc("DuplicateTokenEx")
	procLookupPrivilegeValue         = modAdvapi32.NewProc("LookupPrivilegeValueW")
	procAdjustTokenPrivileges        = modAdvapi32.NewProc("AdjustTokenPrivileges")
	procOpenProcessToken             = modAdvapi32.NewProc("OpenProcessToken")
	procCreateToolhelp32Snapshot     = modKernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32FirstW              = modKernel32.NewProc("Process32FirstW")
	procProcess32NextW               = modKernel32.NewProc("Process32NextW")
	procOpenProcess                  = modKernel32.NewProc("OpenProcess")
	procGetExitCodeProcess           = modKernel32.NewProc("GetExitCodeProcess")
)

const (
	tokenPrimary             = 1
	securityImpersonation    = 2
	maximumAllowed           = 0x02000000
	tokenAdjustPrivileges    = 0x0020
	tokenQuery               = 0x0008
	tokenDuplicate           = 0x0002
	sePrivilegeEnabled       = 0x00000002
	createNoWindow           = 0x08000000 // 콘솔/CMD 창 완전 숨김 (createNewConsole 대체)
	createUnicodeEnvironment = 0x00000400
	normalPriorityClass      = 0x00000020
	th32csSnapProcess        = 0x00000002
	processQueryInformation  = 0x0400
	processQueryLimited      = 0x1000
	startfUseshowwindow      = 0x00000001 // STARTF_USESHOWWINDOW
	swHideWindow             = 0          // SW_HIDE
)

type luid struct {
	LowPart  uint32
	HighPart int32
}

type luidAndAttributes struct {
	Luid       luid
	Attributes uint32
}

type tokenPrivileges struct {
	PrivilegeCount uint32
	Privileges     [1]luidAndAttributes
}

type startupInfoW struct {
	Cb              uint32
	LpReserved      *uint16
	LpDesktop       *uint16
	LpTitle         *uint16
	DwX             uint32
	DwY             uint32
	DwXSize         uint32
	DwYSize         uint32
	DwXCountChars   uint32
	DwYCountChars   uint32
	DwFillAttribute uint32
	DwFlags         uint32
	WShowWindow     uint16
	CbReserved2     uint16
	LpReserved2     *byte
	HStdInput       syscall.Handle
	HStdOutput      syscall.Handle
	HStdError       syscall.Handle
}

type processInformation struct {
	HProcess    syscall.Handle
	HThread     syscall.Handle
	DwProcessId uint32
	DwThreadId  uint32
}

type processEntry32 struct {
	DwSize              uint32
	CntUsage            uint32
	Th32ProcessID       uint32
	Th32DefaultHeapID   uintptr
	Th32ModuleID        uint32
	CntThreads          uint32
	Th32ParentProcessID uint32
	PcPriClassBase      int32
	DwFlags             uint32
	SzExeFile           [260]uint16
}

// enablePrivilege enables the named privilege on the current process token.
func enablePrivilege(name string) error {
	var token syscall.Token
	curProc, _ := syscall.GetCurrentProcess()
	err := syscall.OpenProcessToken(curProc, tokenAdjustPrivileges|tokenQuery, &token)
	if err != nil {
		return fmt.Errorf("OpenProcessToken: %v", err)
	}
	defer token.Close()

	var privLuid luid
	privNamePtr, _ := syscall.UTF16PtrFromString(name)
	ret, _, err := procLookupPrivilegeValue.Call(0, uintptr(unsafe.Pointer(privNamePtr)), uintptr(unsafe.Pointer(&privLuid)))
	if ret == 0 {
		return fmt.Errorf("LookupPrivilegeValue(%s): %v", name, err)
	}

	tp := tokenPrivileges{
		PrivilegeCount: 1,
		Privileges: [1]luidAndAttributes{{
			Luid:       privLuid,
			Attributes: sePrivilegeEnabled,
		}},
	}
	ret, _, err = procAdjustTokenPrivileges.Call(
		uintptr(token),
		0,
		uintptr(unsafe.Pointer(&tp)),
		0, 0, 0,
	)
	if ret == 0 {
		return fmt.Errorf("AdjustTokenPrivileges(%s): %v", name, err)
	}
	// Check if the privilege was actually assigned
	if err == syscall.Errno(1300) { // ERROR_NOT_ALL_ASSIGNED
		return fmt.Errorf("privilege %s not held by this process token", name)
	}
	return nil
}

// getUserTokenViaExplorer retrieves the user token by finding the explorer.exe
// process running in the active console session and duplicating its token.
// This does NOT require SE_TCB_PRIVILEGE but requires SE_DEBUG_PRIVILEGE.
func getUserTokenViaExplorer(targetSessionID uint32) (syscall.Handle, error) {
	// Enable SE_DEBUG_PRIVILEGE to open system processes
	if err := enablePrivilege("SeDebugPrivilege"); err != nil {
		logger.Warningf("[LaunchAsUser] SeDebugPrivilege 활성화 실패: %v", err)
	}

	// Snapshot all running processes
	snapshot, _, err := procCreateToolhelp32Snapshot.Call(th32csSnapProcess, 0)
	if snapshot == ^uintptr(0) {
		return 0, fmt.Errorf("CreateToolhelp32Snapshot 실패: %v", err)
	}
	defer syscall.CloseHandle(syscall.Handle(snapshot))

	var entry processEntry32
	entry.DwSize = uint32(unsafe.Sizeof(entry))

	ret, _, _ := procProcess32FirstW.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		return 0, fmt.Errorf("Process32First 실패")
	}

	explorerName, _ := syscall.UTF16FromString("explorer.exe")

	for {
		// Check if this process is explorer.exe
		match := true
		for i, c := range explorerName {
			if i >= len(entry.SzExeFile) || entry.SzExeFile[i] != c {
				match = false
				break
			}
		}

		if match && entry.Th32ProcessID != 0 {
			// Open the explorer.exe process
			hProc, _, _ := procOpenProcess.Call(
				processQueryInformation|processQueryLimited,
				0,
				uintptr(entry.Th32ProcessID),
			)
			if hProc != 0 {
				// Get the token of this process
				var procToken syscall.Handle
				ret, _, _ := procOpenProcessToken.Call(hProc, tokenDuplicate|tokenQuery, uintptr(unsafe.Pointer(&procToken)))
				syscall.CloseHandle(syscall.Handle(hProc))
				if ret != 0 {
					// Verify this token belongs to the target session
					// Duplicate to Primary Token
					var dupToken syscall.Handle
					ret2, _, err2 := procDuplicateTokenEx.Call(
						uintptr(procToken),
						maximumAllowed,
						0,
						securityImpersonation,
						tokenPrimary,
						uintptr(unsafe.Pointer(&dupToken)),
					)
					syscall.CloseHandle(procToken)
					if ret2 != 0 {
						logger.Infof("[LaunchAsUser] explorer.exe(PID=%d) 토큰 획득 성공", entry.Th32ProcessID)
						return dupToken, nil
					}
					logger.Warningf("[LaunchAsUser] DuplicateTokenEx 실패 (PID=%d): %v", entry.Th32ProcessID, err2)
				}
			}
		}

		ret, _, _ = procProcess32NextW.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
		if ret == 0 {
			break
		}
	}

	return 0, fmt.Errorf("explorer.exe를 찾을 수 없거나 토큰 획득 실패 (sessionID=%d)", targetSessionID)
}

// launchAsUser attempts to launch a GUI process in the active user session (Session 1)
// even when called from a SYSTEM service (Session 0).
// Falls back to normal exec.Command if token acquisition fails (non-service / no desktop mode).
func launchAsUser(exe string, args []string, workDir string) error {
	// 1. Get the active console session ID
	sessionID, _, _ := procWTSGetActiveConsoleSessionId.Call()
	logger.Infof("[LaunchAsUser] 활성 콘솔 세션 ID: %d", sessionID)

	if sessionID != 0xFFFFFFFF {
		// 2a. Try WTSQueryUserToken first (needs SeTcbPrivilege)
		_ = enablePrivilege("SeTcbPrivilege")
		var userToken syscall.Handle
		ret, _, wtsErr := procWTSQueryUserToken.Call(sessionID, uintptr(unsafe.Pointer(&userToken)))
		if ret != 0 {
			defer syscall.CloseHandle(userToken)
			// Duplicate to Primary
			var dupToken syscall.Handle
			ret2, _, err2 := procDuplicateTokenEx.Call(
				uintptr(userToken),
				maximumAllowed, 0, securityImpersonation, tokenPrimary,
				uintptr(unsafe.Pointer(&dupToken)),
			)
			if ret2 != 0 {
				defer syscall.CloseHandle(dupToken)
				logger.Infof("[LaunchAsUser] WTSQueryUserToken 방식으로 토큰 획득 성공")
				return createProcessWithToken(dupToken, exe, args, workDir)
			}
			_ = err2
		} else {
			logger.Warningf("[LaunchAsUser] WTSQueryUserToken 실패 (%v). Explorer 토큰 방식 시도.", wtsErr)
		}

		// 2b. Fallback: steal token from explorer.exe in the active session
		dupToken, err := getUserTokenViaExplorer(uint32(sessionID))
		if err != nil {
			logger.Warningf("[LaunchAsUser] Explorer 토큰 방식도 실패 (%v). exec.Command 폴백.", err)
		} else {
			defer syscall.CloseHandle(dupToken)
			return createProcessWithToken(dupToken, exe, args, workDir)
		}
	}

	// 3. Last resort: normal exec (works in foreground -run mode)
	logger.Warningf("[LaunchAsUser] 모든 세션 기동 방식 실패. exec.Command 직접 실행.")
	return launchViaExec(exe, args, workDir)
}

// createProcessWithToken launches the process using the given user token.
func createProcessWithToken(token syscall.Handle, exe string, args []string, workDir string) error {
	// Create environment block for this user token
	var envBlock uintptr
	procCreateEnvironmentBlock.Call(uintptr(unsafe.Pointer(&envBlock)), uintptr(token), 0)
	if envBlock != 0 {
		defer procDestroyEnvironmentBlock.Call(envBlock)
	}

	cmdLine, err := syscall.UTF16PtrFromString(buildCmdLine(exe, args))
	if err != nil {
		return fmt.Errorf("UTF16PtrFromString 실패: %v", err)
	}

	desktop, _ := syscall.UTF16PtrFromString("winsta0\\default")
	si := startupInfoW{
		Cb:          uint32(unsafe.Sizeof(startupInfoW{})),
		LpDesktop:   desktop,
		DwFlags:     startfUseshowwindow, // STARTF_USESHOWWINDOW: WShowWindow 필드를 유효하게 만듦
		WShowWindow: swHideWindow,        // SW_HIDE: 프로세스 창을 처음부터 숨김
	}
	var pi processInformation

	var workDirPtr *uint16
	if workDir != "" {
		workDirPtr, _ = syscall.UTF16PtrFromString(workDir)
	}

	flags := uint32(createNoWindow | normalPriorityClass)
	if envBlock != 0 {
		flags |= createUnicodeEnvironment
	}

	ret, _, err := procCreateProcessAsUser.Call(
		uintptr(token),
		0,
		uintptr(unsafe.Pointer(cmdLine)),
		0, 0, 0,
		uintptr(flags),
		envBlock,
		uintptr(unsafe.Pointer(workDirPtr)),
		uintptr(unsafe.Pointer(&si)),
		uintptr(unsafe.Pointer(&pi)),
	)
	if ret == 0 {
		return fmt.Errorf("CreateProcessAsUser 실패: %v", err)
	}

	syscall.CloseHandle(pi.HProcess)
	syscall.CloseHandle(pi.HThread)
	logger.Infof("[LaunchAsUser] ✅ 사용자 세션에서 프로세스 기동 성공: %s", exe)
	return nil
}

// launchViaExec is the fallback for foreground (-run) mode.
func launchViaExec(exe string, args []string, workDir string) error {
	cmd := exec.Command(exe, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}
	// 콘솔/CMD 창 없이 실행 (포그라운드 모드에서도 동일하게 창 숨김)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	logger.Infof("[LaunchViaExec] 직접 실행 (포그라운드 모드): %s %v", exe, args)
	return cmd.Start()
}

// isProcessRunning returns true if the process with the given PID is still alive.
func isProcessRunning(pid uint32) bool {
	h, _, _ := procOpenProcess.Call(processQueryLimited, 0, uintptr(pid))
	if h == 0 {
		return false
	}
	defer syscall.CloseHandle(syscall.Handle(h))
	var exitCode uint32
	ret, _, _ := procGetExitCodeProcess.Call(h, uintptr(unsafe.Pointer(&exitCode)))
	if ret == 0 {
		return false
	}
	return exitCode == 259 // STILL_ACTIVE
}

// launchAsUserWithPID is like launchAsUser but also returns the PID of the spawned process.
func launchAsUserWithPID(exe string, args []string, workDir string) (uint32, error) {
	sessionID, _, _ := procWTSGetActiveConsoleSessionId.Call()

	if sessionID != 0xFFFFFFFF {
		_ = enablePrivilege("SeTcbPrivilege")
		var userToken syscall.Handle
		ret, _, wtsErr := procWTSQueryUserToken.Call(sessionID, uintptr(unsafe.Pointer(&userToken)))
		if ret != 0 {
			defer syscall.CloseHandle(userToken)
			var dupToken syscall.Handle
			ret2, _, _ := procDuplicateTokenEx.Call(
				uintptr(userToken), maximumAllowed, 0, securityImpersonation, tokenPrimary,
				uintptr(unsafe.Pointer(&dupToken)),
			)
			if ret2 != 0 {
				defer syscall.CloseHandle(dupToken)
				return createProcessWithTokenGetPID(dupToken, exe, args, workDir)
			}
		} else {
			logger.Warningf("[LaunchAsUserWithPID] WTSQueryUserToken 실패 (%v). Explorer 토큰 방식 시도.", wtsErr)
		}

		dupToken, err := getUserTokenViaExplorer(uint32(sessionID))
		if err != nil {
			logger.Warningf("[LaunchAsUserWithPID] Explorer 토큰 방식도 실패 (%v).", err)
		} else {
			defer syscall.CloseHandle(dupToken)
			return createProcessWithTokenGetPID(dupToken, exe, args, workDir)
		}
	}
	// Fallback: exec.Command (returns 0 PID as it's not tracked)
	return 0, launchViaExec(exe, args, workDir)
}

// createProcessWithTokenGetPID is like createProcessWithToken but returns the PID.
func createProcessWithTokenGetPID(token syscall.Handle, exe string, args []string, workDir string) (uint32, error) {
	var envBlock uintptr
	procCreateEnvironmentBlock.Call(uintptr(unsafe.Pointer(&envBlock)), uintptr(token), 0)
	if envBlock != 0 {
		defer procDestroyEnvironmentBlock.Call(envBlock)
	}

	cmdLine, err := syscall.UTF16PtrFromString(buildCmdLine(exe, args))
	if err != nil {
		return 0, fmt.Errorf("UTF16PtrFromString 실패: %v", err)
	}

	desktop, _ := syscall.UTF16PtrFromString("winsta0\\default")
	si := startupInfoW{
		Cb:          uint32(unsafe.Sizeof(startupInfoW{})),
		LpDesktop:   desktop,
		DwFlags:     startfUseshowwindow,
		WShowWindow: swHideWindow,
	}
	var pi processInformation

	var workDirPtr *uint16
	if workDir != "" {
		workDirPtr, _ = syscall.UTF16PtrFromString(workDir)
	}

	flags := uint32(createNoWindow | normalPriorityClass)
	if envBlock != 0 {
		flags |= createUnicodeEnvironment
	}

	ret, _, err := procCreateProcessAsUser.Call(
		uintptr(token), 0,
		uintptr(unsafe.Pointer(cmdLine)),
		0, 0, 0,
		uintptr(flags),
		envBlock,
		uintptr(unsafe.Pointer(workDirPtr)),
		uintptr(unsafe.Pointer(&si)),
		uintptr(unsafe.Pointer(&pi)),
	)
	if ret == 0 {
		return 0, fmt.Errorf("CreateProcessAsUser 실패: %v", err)
	}

	pid := pi.DwProcessId
	syscall.CloseHandle(pi.HProcess)
	syscall.CloseHandle(pi.HThread)
	logger.Infof("[LaunchAsUserWithPID] ✅ 프로세스 기동 성공: %s (PID: %d)", exe, pid)
	return pid, nil
}

// buildCmdLine builds a Windows-style quoted command line from exe + args.
func buildCmdLine(exe string, args []string) string {
	cmdLine := fmt.Sprintf(`"%s"`, exe)
	for _, arg := range args {
		if needsQuote(arg) {
			cmdLine += fmt.Sprintf(` "%s"`, arg)
		} else {
			cmdLine += " " + arg
		}
	}
	return cmdLine
}

func needsQuote(s string) bool {
	for _, c := range s {
		if c == ' ' || c == '\t' {
			return true
		}
	}
	return false
}
