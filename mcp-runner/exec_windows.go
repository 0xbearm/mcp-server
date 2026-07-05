//go:build windows

package main

import (
	"os"
	"os/exec"
)

func executeDocker(executable string, arguments []string) error {
	command := exec.Command(executable, arguments...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}
