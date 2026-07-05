#!/bin/sh
set -eu

server="${1:-}"
[ -n "$server" ] || {
    echo "Usage: docker run mcp-server <server>" >&2
    exit 2
}
shift

case "$server" in
    python-ssh-mcp)
        cd /mcp/python-ssh-mcp
        echo "Running python-ssh-mcp"
        exec uv run server.py "$@"
        ;;
    postgres-mcp)
        echo "Running postgres-mcp"
        exec uv run postgres-mcp "$@"
        ;;
    ghidra-mcp)
        cd /mcp
        echo "Running ghidra-mcp"
        exec uv run bridge_mcp_ghidra_5.14.2.py "$@"
        ;;
    *)
        echo "Unknown MCP server: $server" >&2
        exit 2
        ;;
esac
