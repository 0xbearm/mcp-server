package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

const usageText = `Usage:
  mcp-runner [options] <server> [-- <container-arg>...]
  mcp-runner [options] list

Options:
  --config PATH  Configuration file (default: $MCP_RUNNER_CONFIG or the user config directory)
  --dry-run      Print the Docker command instead of running it
`

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		var exitErr interface{ ExitCode() int }
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "mcp-runner: %v\n", err)
		os.Exit(1)
	}
}

func run(arguments []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("mcp-runner", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		fmt.Fprint(stderr, usageText)
	}

	var configOverride string
	var dryRun bool
	flags.StringVar(&configOverride, "config", "", "configuration file")
	flags.BoolVar(&dryRun, "dry-run", false, "print the Docker command")

	if err := flags.Parse(arguments); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	positionals := flags.Args()
	if len(positionals) == 0 {
		flags.Usage()
		return errors.New("a server name is required")
	}

	configPath, err := resolveConfigPath(configOverride)
	if err != nil {
		return err
	}
	config, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	if positionals[0] == "list" {
		if len(positionals) != 1 {
			return errors.New("list does not accept additional arguments")
		}
		names := make([]string, 0, len(config.Servers))
		for name := range config.Servers {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			fmt.Fprintln(stdout, name)
		}
		return nil
	}

	serverName := positionals[0]
	containerArgs := positionals[1:]
	if len(containerArgs) > 0 && containerArgs[0] == "--" {
		containerArgs = containerArgs[1:]
	}

	server, ok := config.Servers[serverName]
	if !ok {
		return fmt.Errorf("server %q is not configured", serverName)
	}

	dockerArgs, err := buildDockerArgs(serverName, server, containerArgs)
	if err != nil {
		return fmt.Errorf("server %q: %w", serverName, err)
	}

	if dryRun {
		fmt.Fprintln(stdout, formatCommand(config.Docker, dockerArgs))
		return nil
	}

	if strings.TrimSpace(config.Docker) == "" {
		return errors.New("Docker executable cannot be empty")
	}
	return executeDocker(config.Docker, dockerArgs)
}
