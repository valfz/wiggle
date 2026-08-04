package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func pidFilePath() string { return filepath.Join(os.TempDir(), "wiggle.pid") }
func logFilePath() string { return filepath.Join(os.TempDir(), "wiggle.log") }

func startDaemon() error {
	if pid, ok := runningPID(); ok {
		return fmt.Errorf("already running (pid %d)", pid)
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	logFile, err := os.OpenFile(logFilePath(), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer logFile.Close()

	cmd := exec.Command(exe, "run")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = detachedProcAttr()
	if err := cmd.Start(); err != nil {
		return err
	}

	pid := cmd.Process.Pid
	if err := os.WriteFile(pidFilePath(), []byte(strconv.Itoa(pid)), 0o600); err != nil {
		_ = cmd.Process.Kill()
		return err
	}
	_ = cmd.Process.Release()

	fmt.Printf("wiggle started (pid %d), log: %s\n", pid, logFilePath())
	return nil
}

func stopDaemon() error {
	pid, ok := runningPID()
	if !ok {
		_ = os.Remove(pidFilePath())
		return errors.New("not running")
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		_ = os.Remove(pidFilePath())
		return errors.New("not running")
	}
	if err := terminate(proc); err != nil {
		return err
	}
	_ = os.Remove(pidFilePath())

	fmt.Printf("wiggle stopped (pid %d)\n", pid)
	return nil
}

func runningPID() (int, bool) {
	data, err := os.ReadFile(pidFilePath())
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 || !processAlive(pid) {
		return 0, false
	}
	return pid, true
}
