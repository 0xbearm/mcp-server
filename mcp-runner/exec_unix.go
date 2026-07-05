//go:build !windows

package main

import (
	"os"
	"os/exec"
	"syscall"
)

func executeDocker(executable string, arguments []string) error {
	resolved, err := exec.LookPath(executable)
	if err != nil {
		return err
	}
	argv := append([]string{executable}, arguments...)
	return syscall.Exec(resolved, argv, os.Environ())
}
