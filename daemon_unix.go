//go:build !windows

package main

import (
	"errors"
	"os"
	"syscall"
)

func detachedProcAttr() *syscall.SysProcAttr {
	// New session so the daemon survives the terminal closing.
	return &syscall.SysProcAttr{Setsid: true}
}

func processAlive(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}

func terminate(p *os.Process) error {
	return p.Signal(syscall.SIGTERM)
}
