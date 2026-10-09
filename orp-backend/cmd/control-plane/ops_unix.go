//go:build !windows

package main

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

func configureDaemonProcess(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

func showLogsOS(logPath string, follow bool) error {
	tailArgs := []string{"-n", "100"}
	if follow {
		tailArgs = append(tailArgs, "-f")
	}
	command := exec.Command("tail", append(tailArgs, logPath)...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}

func stopProcess(pid int) error {
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil && err != syscall.ESRCH {
		return err
	}
	return nil
}

func processIDExists(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}

func processMatchesOS(record pidRecord) bool {
	if !processIDExists(record.PID) {
		return false
	}
	output, err := exec.Command("ps", "-p", strconv.Itoa(record.PID), "-o", "command=").Output()
	return err == nil && strings.Contains(string(output), "--run-id="+record.Token)
}
