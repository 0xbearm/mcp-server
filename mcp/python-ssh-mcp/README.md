# python-ssh-mcp

`python-ssh-mcp` is a stdio MCP server for opening SSH sessions and running
commands on preconfigured hosts.

## Configure

The server reads:

```text
~/.config/python-ssh-mcp/config.json
```

Create the directory and configuration file before starting the server:

```sh
mkdir -p ~/.config/python-ssh-mcp
```

Example:

```json
{
  "connections": [
    {
      "name": "key-example",
      "host": "ssh.example.com",
      "user": "operator",
      "port": 22,
      "auth": {
        "type": "key",
        "path": "~/.ssh/id_ed25519"
      }
    },
    {
      "name": "password-example",
      "host": "ssh.example.com",
      "user": "operator",
      "port": 22,
      "auth": {
        "type": "password",
        "password": "$SSH_PASSWORD"
      }
    }
  ]
}
```

A password beginning with `$` names an environment variable. In the example,
the server reads the value from `SSH_PASSWORD`. Forward it through
`mcp-runner` without storing the value:

```json
{
  "pass_env": [
    "SSH_PASSWORD"
  ]
}
```

For key authentication, the runner configuration must mount the referenced key
or its containing SSH directory into the container. The repository's runner
example mounts `~/.ssh` read-only at `/root/.ssh`.

The current SSH implementation disables known-host verification. Use it only
with hosts and networks you trust.

## Run in the container

Build the image from the repository root, then select this server:

```sh
docker build -t mcp-server:latest .
docker run --rm -i \
  --mount type=bind,src="$HOME/.config/python-ssh-mcp/config.json",dst=/root/.config/python-ssh-mcp/config.json,readonly \
  --mount type=bind,src="$HOME/.ssh",dst=/root/.ssh,readonly \
  mcp-server:latest python-ssh-mcp
```

For normal MCP client configuration, use
[`mcp-runner`](../../mcp-runner/README.md) instead of repeating these Docker
arguments.

## Run locally for development

From this directory:

```sh
uv sync --frozen
uv run server.py
```

Test with MCP Inspector:

```sh
npx @modelcontextprotocol/inspector uv run server.py
```

## Add a connection

`config.py` can append a connection to an existing configuration file:

```sh
uv run config.py \
  --name example \
  --host ssh.example.com \
  --user operator \
  --port 22 \
  --connection-type key
```
