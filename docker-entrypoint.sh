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
        exec uv run server.py "$@"
        ;;
    postgres-mcp)
        exec uv run postgres-mcp --access-mode=unrestricted "$@"
        ;;
    *)
        echo "Unknown MCP server: $server" >&2
        exit 2
        ;;
esac
