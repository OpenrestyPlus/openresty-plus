//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

func configureDaemonProcess(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: windows.DETACHED_PROCESS | windows.CREATE_NEW_PROCESS_GROUP,
	}
}

func showLogsOS(logPath string, follow bool) error {
	windowsPath := strings.ReplaceAll(logPath, "'", "''")
	script := "Get-Content -LiteralPath '" + windowsPath + "' -Tail 100"
	if follow {
		script += " -Wait"
	}
	command := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}

func stopProcess(pid int) error {
	return exec.Command("taskkill.exe", "/PID", fmt.Sprint(pid), "/T", "/F").Run()
}

func processIDExists(pid int) bool {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	_ = windows.CloseHandle(handle)
	return true
}

func processMatchesOS(record pidRecord) bool {
	if !processIDExists(record.PID) {
		return false
	}
	script := fmt.Sprintf("$p=Get-CimInstance Win32_Process -Filter 'ProcessId=%d' -ErrorAction SilentlyContinue; if ($p -and $p.CommandLine -like '*--run-id=%s*') { exit 0 } else { exit 1 }", record.PID, record.Token)
	return exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script).Run() == nil
}
