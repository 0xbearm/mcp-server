package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var environmentName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func buildDockerArgs(serverName string, server Server, runtimeArgs []string) ([]string, error) {
	args := []string{"run", "--rm", "-i"}

	for _, argument := range server.DockerArgs {
		if argument == "-t" || argument == "--tty" || strings.HasPrefix(argument, "--tty=") {
			return nil, errors.New("TTY allocation is incompatible with MCP stdio")
		}
		if argument == "--" {
			return nil, errors.New("docker_args cannot contain --")
		}
		args = append(args, argument)
	}

	targets := make(map[string]struct{}, len(server.Mounts))
	for index, mount := range server.Mounts {
		mountArgument, err := buildMountArgument(mount)
		if err != nil {
			return nil, fmt.Errorf("mount %d: %w", index+1, err)
		}
		if _, exists := targets[mount.Target]; exists {
			return nil, fmt.Errorf("mount %d: duplicate target %q", index+1, mount.Target)
		}
		targets[mount.Target] = struct{}{}
		args = append(args, "--mount", mountArgument)
	}

	staticEnvironmentNames := make([]string, 0, len(server.Env))
	for name := range server.Env {
		if !environmentName.MatchString(name) {
			return nil, fmt.Errorf("invalid environment variable name %q", name)
		}
		staticEnvironmentNames = append(staticEnvironmentNames, name)
	}
	sort.Strings(staticEnvironmentNames)
	for _, name := range staticEnvironmentNames {
		args = append(args, "--env", name+"="+server.Env[name])
	}

	passedEnvironment := append([]string(nil), server.PassEnv...)
	sort.Strings(passedEnvironment)
	seenEnvironment := make(map[string]struct{}, len(passedEnvironment))
	for _, name := range passedEnvironment {
		if !environmentName.MatchString(name) {
			return nil, fmt.Errorf("invalid pass_env name %q", name)
		}
		if _, exists := server.Env[name]; exists {
			return nil, fmt.Errorf("environment variable %q is defined in both env and pass_env", name)
		}
		if _, exists := seenEnvironment[name]; exists {
			return nil, fmt.Errorf("duplicate pass_env name %q", name)
		}
		if _, exists := os.LookupEnv(name); !exists {
			return nil, fmt.Errorf("pass_env variable %q is not set", name)
		}
		seenEnvironment[name] = struct{}{}
		args = append(args, "--env", name)
	}

	args = append(args, server.Image)
	args = append(args, serverName)
	args = append(args, server.Args...)
	args = append(args, runtimeArgs...)
	return args, nil
}

func buildMountArgument(mount Mount) (string, error) {
	if mount.Source == "" {
		return "", errors.New("source cannot be empty")
	}
	if mount.Target == "" {
		return "", errors.New("target cannot be empty")
	}
	if strings.ContainsAny(mount.Source, ",\n\r") || strings.ContainsAny(mount.Target, ",\n\r") {
		return "", errors.New("source and target cannot contain commas or newlines")
	}

	source, err := expandHome(mount.Source)
	if err != nil {
		return "", err
	}
	source, err = filepath.Abs(source)
	if err != nil {
		return "", fmt.Errorf("resolve source %q: %w", mount.Source, err)
	}
	if _, err := os.Stat(source); err != nil {
		return "", fmt.Errorf("access source %q: %w", source, err)
	}
	if !path.IsAbs(mount.Target) {
		return "", fmt.Errorf("target must be an absolute container path: %q", mount.Target)
	}

	parts := []string{"type=bind", "src=" + source, "dst=" + mount.Target}
	if mount.ReadOnly {
		parts = append(parts, "readonly")
	}
	return strings.Join(parts, ","), nil
}

func formatCommand(executable string, arguments []string) string {
	parts := make([]string, 0, len(arguments)+1)
	parts = append(parts, shellQuote(executable))
	for _, argument := range arguments {
		parts = append(parts, shellQuote(argument))
	}
	return strings.Join(parts, " ")
}

func shellQuote(value string) string {
	if value != "" && strings.IndexFunc(value, func(character rune) bool {
		return !((character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') ||
			strings.ContainsRune("_@%+=:,./-", character))
	}) == -1 {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}
