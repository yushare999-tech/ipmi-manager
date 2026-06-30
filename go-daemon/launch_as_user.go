package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var (
	modWtsapi32               = syscall.NewLazyDLL("wtsapi32.dll")
	modKernel32               = syscall.NewLazyDLL("kernel32.dll")
	modAdvapi32               = syscall.NewLazyDLL("advapi32.dll")
	modUserenv                = syscall.NewLazyDLL("userenv.dll")

	procWTSGetActiveConsoleSessionId = modKernel32.NewProc("WTSGetActiveConsoleSessionId")
	procWTSQueryUserToken            = modWtsapi32.NewProc("WTSQueryUserToken")
	procCreateProcessAsUser          = modAdvapi32.NewProc("CreateProcessAsUserW")
	procCreateEnvironmentBlock       = modUserenv.NewProc("CreateEnvironmentBlock")
	procDestroyEnvironmentBlock      = modUserenv.NewProc("DestroyEnvironmentBlock")
	procDuplicateTokenEx             = modAdvapi32.NewProc("DuplicateTokenEx")
	procCloseHandle                  = modKernel32.NewProc("CloseHandle")
)

const (
	TokenPrimary       = 1
	SecurityImpersonation = 2
	MAXIMUM_ALLOWED    = 0x02000000
	CREATE_NEW_CONSOLE = 0x00000010
	CREATE_UNICODE_ENVIRONMENT = 0x00000400
	NORMAL_PRIORITY_CLASS = 0x00000020
)

type STARTUPINFOW struct {
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

type PROCESS_INFORMATION struct {
	HProcess    syscall.Handle
	HThread     syscall.Handle
	DwProcessId uint32
	DwThreadId  uint32
}

// launchAsUser attempts to launch a process in the active user session (Session 1)
// even when called from a SYSTEM service (Session 0).
// Falls back to exec.Command if the WTS token acquisition fails (e.g., non-service mode).
func launchAsUser(exe string, args []string, workDir string) error {
	// 1. Get the active console session ID (the user's desktop session)
	sessionID, _, _ := procWTSGetActiveConsoleSessionId.Call()
	if sessionID == 0xFFFFFFFF {
		// No active console session: fall back to normal exec
		logger.Warningf("[LaunchAsUser] No active console session found. Falling back to exec.Command for: %s", exe)
		return launchFallback(exe, args, workDir)
	}

	// 2. Get the user token for that session
	var userToken syscall.Handle
	ret, _, err := procWTSQueryUserToken.Call(sessionID, uintptr(unsafe.Pointer(&userToken)))
	if ret == 0 {
		logger.Warningf("[LaunchAsUser] WTSQueryUserToken failed (err=%v). Falling back to exec.Command.", err)
		return launchFallback(exe, args, workDir)
	}
	defer procCloseHandle.Call(uintptr(userToken))

	// 3. Duplicate the token so we can use CreateProcessAsUser
	var dupToken syscall.Handle
	ret, _, err = procDuplicateTokenEx.Call(
		uintptr(userToken),
		MAXIMUM_ALLOWED,
		0,
		SecurityImpersonation,
		TokenPrimary,
		uintptr(unsafe.Pointer(&dupToken)),
	)
	if ret == 0 {
		logger.Warningf("[LaunchAsUser] DuplicateTokenEx failed (err=%v). Falling back to exec.Command.", err)
		return launchFallback(exe, args, workDir)
	}
	defer procCloseHandle.Call(uintptr(dupToken))

	// 4. Create environment block for the user
	var envBlock uintptr
	procCreateEnvironmentBlock.Call(uintptr(unsafe.Pointer(&envBlock)), uintptr(dupToken), 0)
	if envBlock != 0 {
		defer procDestroyEnvironmentBlock.Call(envBlock)
	}

	// 5. Build command line string
	cmdLine := syscall.StringToUTF16Ptr(buildCmdLine(exe, args))

	// 6. Set startup info to use the interactive desktop (WinSta0\Default)
	desktop, _ := syscall.UTF16PtrFromString("winsta0\\default")
	si := STARTUPINFOW{
		Cb:        uint32(unsafe.Sizeof(STARTUPINFOW{})),
		LpDesktop: desktop,
	}
	var pi PROCESS_INFORMATION

	// 7. Build working directory
	var workDirPtr *uint16
	if workDir != "" {
		if _, statErr := os.Stat(workDir); statErr == nil {
			workDirPtr, _ = syscall.UTF16PtrFromString(workDir)
		}
	}

	// 8. Launch the process in the user session
	flags := uint32(CREATE_NEW_CONSOLE | NORMAL_PRIORITY_CLASS)
	if envBlock != 0 {
		flags |= CREATE_UNICODE_ENVIRONMENT
	}

	ret, _, err = procCreateProcessAsUser.Call(
		uintptr(dupToken),
		0, // lpApplicationName (null, using cmdLine)
		uintptr(unsafe.Pointer(cmdLine)),
		0, // lpProcessAttributes
		0, // lpThreadAttributes
		0, // bInheritHandles (false)
		uintptr(flags),
		envBlock,
		uintptr(unsafe.Pointer(workDirPtr)),
		uintptr(unsafe.Pointer(&si)),
		uintptr(unsafe.Pointer(&pi)),
	)
	if ret == 0 {
		logger.Warningf("[LaunchAsUser] CreateProcessAsUser failed (err=%v). Falling back to exec.Command.", err)
		return launchFallback(exe, args, workDir)
	}

	// Close process/thread handles (we don't wait for the process)
	procCloseHandle.Call(uintptr(pi.HProcess))
	procCloseHandle.Call(uintptr(pi.HThread))

	logger.Infof("[LaunchAsUser] Process launched successfully in user session %d: %s", sessionID, exe)
	return nil
}

// buildCmdLine builds a Windows-style quoted command line from exe + args.
func buildCmdLine(exe string, args []string) string {
	cmdLine := fmt.Sprintf(`"%s"`, exe)
	for _, arg := range args {
		// Quote args that contain spaces
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

// launchFallback is used when WTS token acquisition fails (non-service mode).
func launchFallback(exe string, args []string, workDir string) error {
	cmd := make([]string, 0, len(args))
	cmd = append(cmd, args...)
	c := syscallExecCmd(exe, cmd, workDir)
	return c.Start()
}

// syscallExecCmd builds an os/exec.Cmd for the fallback path.
func syscallExecCmd(exe string, args []string, workDir string) interface{ Start() error } {
	import_exec_cmd := struct {
		path    string
		args    []string
		workDir string
	}{exe, args, workDir}
	_ = import_exec_cmd
	// Use a proper exec.Cmd via adapter
	return &execAdapter{exe: exe, args: args, workDir: workDir}
}

type execAdapter struct {
	exe     string
	args    []string
	workDir string
}

func (e *execAdapter) Start() error {
	import_os_exec_package_cmd_struct := struct {
		Path string
		Args []string
		Dir  string
	}{e.exe, e.args, e.workDir}
	_ = import_os_exec_package_cmd_struct

	// Direct syscall-level process start (no import cycle)
	argv, err := syscall.UTF16PtrFromString(buildCmdLine(e.exe, e.args))
	if err != nil {
		return err
	}
	var si syscall.StartupInfo
	si.Cb = uint32(unsafe.Sizeof(si))
	var pi syscall.ProcessInformation
	var workDirPtr *uint16
	if e.workDir != "" {
		workDirPtr, _ = syscall.UTF16PtrFromString(e.workDir)
	}
	err = syscall.CreateProcess(
		nil,
		argv,
		nil,
		nil,
		false,
		CREATE_NEW_CONSOLE,
		nil,
		workDirPtr,
		&si,
		&pi,
	)
	if err != nil {
		return err
	}
	syscall.CloseHandle(pi.Process)
	syscall.CloseHandle(pi.Thread)
	return nil
}
