//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func setDetachedProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

func sendTermSignal(proc *os.Process) error {
	return exec.Command("taskkill", "/T", "/PID", fmt.Sprint(proc.Pid)).Run()
}

func sendKillSignal(proc *os.Process) error {
	return exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprint(proc.Pid)).Run()
}

func isProcessRunning(proc *os.Process) bool {
	const PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
	handle, err := syscall.OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(proc.Pid))
	if err != nil {
		return false
	}
	syscall.CloseHandle(handle)
	return true
}
