FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    curl \
    ca-certificates \
    git \
    python3 \
    python3-pip \
    pipx \
    nodejs \
    npm \
    && rm -rf /var/lib/apt/lists/*

ADD https://astral.sh/uv/install.sh /uv-installer.sh
RUN sh /uv-installer.sh && rm /uv-installer.sh
ENV PATH="/root/.local/bin/:$PATH"

WORKDIR /mcp
COPY docker-entrypoint.sh .
RUN chmod +x /mcp/docker-entrypoint.sh

COPY mcp .

WORKDIR /mcp/python-ssh-mcp
RUN uv sync --frozen --no-dev

# Install from pip or other places
RUN uv pip install postgres-mcp

ENTRYPOINT [ "/mcp/docker-entrypoint.sh" ]
