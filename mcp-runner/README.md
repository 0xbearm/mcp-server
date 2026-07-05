# mcp-runner

`mcp-runner` starts containerized stdio MCP servers from definitions in one
user-level JSON file. It uses only the Go standard library.

## Requirements

- Docker available on `PATH`, unless the configuration selects another Docker
  executable
- Go 1.23 or newer to build from source

## Build and test

From this directory:

```sh
go test ./...
go vet ./...
CGO_ENABLED=0 go build -trimpath -o mcp-runner .
```

Verify the binary:

```sh
./mcp-runner --help
```

Alternatively, install it into the Go binary directory:

```sh
go install .
```

When using `go install`, ensure `$(go env GOPATH)/bin` is in `PATH`.

## Configure

By default, the runner reads:

```text
~/.config/mcp-runner/config.json
```

On platforms with a different user configuration directory, it uses the path
reported by Go's `os.UserConfigDir`.

Create the Linux configuration from the included example:

```sh
mkdir -p ~/.config/mcp-runner
cp config.example.json ~/.config/mcp-runner/config.json
```

The example contains:

```json
{
  "servers": {
    "python-ssh-mcp": {
      "mounts": [
        {
          "source": "~/.config/python-ssh-mcp/config.json",
          "target": "/root/.config/python-ssh-mcp/config.json",
          "read_only": true
        },
        {
          "source": "~/.ssh",
          "target": "/root/.ssh",
          "read_only": true
        }
      ]
    },
    "postgres-mcp": {
      "env": {
        "LOG_LEVEL": "info",
        "DATABASE_URI": "postgresql://postgres:password@localhost:5432/somedatabase"
      }
    },
    "ghidra-mcp": {
      "docker_args": [
        "--network=host"
      ]
    }
  }
}
```

Override the configuration path with either:

```sh
MCP_RUNNER_CONFIG=/path/to/config.json mcp-runner list
mcp-runner --config /path/to/config.json list
```

The command-line option takes precedence over `MCP_RUNNER_CONFIG`.

All server definitions use the same image. The default is
`mcp-server:latest`, so the example omits it. To use another tag or registry,
set `image` once at the top level:

```json
{
  "image": "registry.example/mcp-server:v1",
  "servers": {
    "ghidra-mcp": {
      "docker_args": ["--network=host"]
    }
  }
}
```

### Configuration fields

Top-level fields:

- `docker`: optional Docker executable; defaults to `docker`.
- `image`: optional Docker image containing all MCP servers; defaults to
  `mcp-server:latest`.
- `servers`: required object mapping MCP server names to definitions.

Server fields:

- `mounts`: optional bind mounts.
- `pass_env`: optional names of variables forwarded from the runner's
  environment.
- `env`: optional literal environment values.
- `docker_args`: optional Docker arguments inserted before mounts and the
  image. Container names cannot be set here because the runner generates a
  unique name for each launch.
- `args`: optional arguments appended after the MCP server selector.

Mount fields:

- `source`: required host path. `~` and `~/...` are expanded.
- `target`: required absolute container path.
- `read_only`: optional boolean.

The runner validates bind-mount sources before starting Docker and rejects TTY
allocation because a TTY can corrupt MCP stdio traffic.

Each container is named after its selected server with a random hexadecimal
suffix, such as `ghidra-mcp-a1b2c3d4e5f60718`. This keeps container names
recognizable while allowing multiple instances of the same server to run.

The example gives `ghidra-mcp` access to the host network so its connection to
`127.0.0.1:8089` reaches the Ghidra plugin running on the host. Start the
GhidraMCP server in Ghidra before launching the runner:

```sh
./mcp-runner ghidra-mcp
```

### Environment variables

Use `pass_env` for credentials and other sensitive values:

```json
{
  "pass_env": [
    "DATABASE_URI"
  ]
}
```

Set the value in the environment that starts the runner:

```sh
export DATABASE_URI='postgresql://username:password@database-host:5432/dbname'
./mcp-runner postgres-mcp
```

The runner verifies that every `pass_env` variable exists and passes only its
name to Docker:

```sh
docker run --env DATABASE_URI ...
```

This keeps the value out of Docker's command-line arguments and `--dry-run`
output.

Use `env` only for non-secret literal values:

```json
{
  "env": {
    "LOG_LEVEL": "info"
  }
}
```

Literal `env` values appear in the generated Docker arguments and in
`--dry-run` output.

## Run

List configured servers:

```sh
./mcp-runner list
```

Inspect a generated Docker command:

```sh
./mcp-runner --dry-run python-ssh-mcp
```

Print the generated command to stderr and then start the server:

```sh
./mcp-runner --debug python-ssh-mcp
```

Debug output goes to stderr so it does not interfere with the MCP stdio
transport.

Start a server:

```sh
./mcp-runner python-ssh-mcp
./mcp-runner postgres-mcp
```

The selected configuration key is also passed to the image entrypoint. For
example, `mcp-runner python-ssh-mcp` generates the equivalent of:

```sh
docker run --rm -i --name python-ssh-mcp-[random] [configured options] mcp-server:latest python-ssh-mcp
```

Additional arguments follow the selector:

```sh
./mcp-runner python-ssh-mcp -- --verbose
```

## Configure Codex

Use the executable's absolute path:

```toml
[mcp_servers.python-ssh-mcp]
command = "/absolute/path/to/mcp-server/mcp-runner/mcp-runner"
args = ["python-ssh-mcp"]
startup_timeout_sec = 20
enabled = true

[mcp_servers.postgres-mcp]
command = "/absolute/path/to/mcp-server/mcp-runner/mcp-runner"
args = ["postgres-mcp"]
startup_timeout_sec = 20
enabled = true
```

The `command` must name an executable file. Pointing it at the `mcp-runner`
directory produces a permission-denied startup error.

If installed with `go install` and the Go binary directory is in `PATH`, use:

```toml
command = "mcp-runner"
```
