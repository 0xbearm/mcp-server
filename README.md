# mcp-server

This repository packages multiple stdio MCP servers into one Docker image.
The image entrypoint selects a server by name, while `mcp-runner` keeps Docker
flags, bind mounts, and environment forwarding out of MCP client
configuration.

## Included servers

- `python-ssh-mcp`: manages SSH sessions and executes commands on configured
  hosts.
- `postgres-mcp`: the upstream PostgreSQL MCP server, launched with
- `ghidra-mcp`: bridges MCP stdio to the GhidraMCP plugin running in a host
  Ghidra instance.

## Requirements

- Docker
- Go 1.23 or newer, when building `mcp-runner`

## Build

Build the container image from the repository root:

```sh
docker build -t mcp-server:latest .
```

Build and test the runner:

```sh
cd mcp-runner
go test ./...
go vet ./...
CGO_ENABLED=0 go build -trimpath -o mcp-runner .
cd ..
```

## Configure

Create the runner configuration:

```sh
mkdir -p ~/.config/mcp-runner
cp mcp-runner/config.example.json ~/.config/mcp-runner/config.json
```

Edit the copied file for your machine. The SSH server also requires
`~/.config/python-ssh-mcp/config.json`; see the
[SSH server documentation](mcp/python-ssh-mcp/README.md).
These are machine-local configuration files and should not be committed.

The example forwards `DATABASE_URI` from the runner's environment. Set it to a
PostgreSQL URI whose hostname is reachable from inside the container:

```sh
export DATABASE_URI='postgresql://username:password@database-host:5432/dbname'
```

Inside a container, `localhost` refers to the container itself. Use a Docker
network hostname, `host.docker.internal` where supported, or another reachable
database hostname when PostgreSQL runs outside this container.

See the [mcp-runner documentation](mcp-runner/README.md) for the complete
configuration schema and environment-handling options.

## Run directly

The argument after the image selects the MCP server:

```sh
docker run --rm -i \
  --mount type=bind,src="$HOME/.config/python-ssh-mcp/config.json",dst=/root/.config/python-ssh-mcp/config.json,readonly \
  --mount type=bind,src="$HOME/.ssh",dst=/root/.ssh,readonly \
  mcp-server:latest python-ssh-mcp
```

```sh
docker run --rm -i \
  --env DATABASE_URI \
  mcp-server:latest postgres-mcp
```

```sh
docker run --rm -i \
  --network=host \
  mcp-server:latest ghidra-mcp
```

Do not add `-t`. A pseudo-terminal can interfere with the MCP stdio protocol.

## Configure Codex

After building `mcp-runner`, use the executable's absolute path:

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

The `command` value must point to the executable, not the `mcp-runner`
directory.
