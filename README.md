# SearXNG MCP Server

Free web search for agentic apps — wraps [SearXNG](https://searxng.github.io/searxng/) to aggregate results from Google, Bing, DuckDuckGo, and more.

Works with any MCP client in any programming language.

---

## Quick Start

```bash
make build      # build the Docker image
make up-http    # start SearXNG + MCP HTTP server
```

MCP is now available at `http://localhost:3333/mcp`.

---

## Two transport modes

| Mode | Command | When to use |
|------|---------|-------------|
| **HTTP** (recommended) | `make up-http` | Persistent service, shared by all apps |
| **stdio** | `make up` | Spawned per session by the MCP client |

### HTTP mode — one running service, all apps connect to it

```
App A ──┐
App B ──┼──► http://localhost:3333/mcp ──► SearXNG
App C ──┘
```

Start once, use from anywhere — no Docker knowledge required in the client app.

### stdio mode — container spawned per session

```
App session starts ──► docker run searxng-mcp ──► SearXNG
App session ends   ──► container exits
```

Useful when you can't expose a port or want full isolation per session.

---

## Connecting from your app

### HTTP mode (url-based — simplest)

No Docker needed on the client side. Just point at the running service.

**Claude Desktop / Claude Code** (`~/.claude/claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "web-search": {
      "url": "http://localhost:3333/mcp"
    }
  }
}
```

**Python — `claude-agent-sdk`**:
```python
from claude_agent_sdk import query, ClaudeAgentOptions

async for message in query(
    prompt="...",
    options=ClaudeAgentOptions(
        mcp_servers={
            "web-search": { "url": "http://localhost:3333/mcp" }
        }
    )
):
    ...
```

**TypeScript — `claude-agent-sdk`**:
```typescript
import { query } from "@anthropic-ai/claude-agent-sdk";

for await (const message of query({
  prompt: "...",
  options: {
    mcpServers: {
      "web-search": { url: "http://localhost:3333/mcp" }
    }
  }
})) { ... }
```

**Python — raw `mcp` client**:
```python
from mcp.client.streamable_http import streamablehttp_client
from mcp import ClientSession

async with streamablehttp_client("http://localhost:3333/mcp") as (read, write, _):
    async with ClientSession(read, write) as session:
        await session.initialize()
        result = await session.call_tool("web_search", {"query": "AI news"})
```

**TypeScript — raw `@modelcontextprotocol/sdk` client**:
```typescript
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StreamableHTTPClientTransport } from "@modelcontextprotocol/sdk/client/streamableHttp.js";

const client = new Client({ name: "my-app", version: "1.0.0" }, {});
await client.connect(new StreamableHTTPClientTransport(new URL("http://localhost:3333/mcp")));

const result = await client.callTool({ name: "web_search", arguments: { query: "AI news" } });
```

**LangChain (Python)**:
```python
from langchain_mcp_adapters.client import MultiServerMCPClient

async with MultiServerMCPClient({
    "web-search": { "url": "http://localhost:3333/mcp", "transport": "streamable_http" }
}) as client:
    tools = client.get_tools()
```

---

### stdio mode (docker-based)

Requires Docker on the machine running the client.

**Claude Desktop / Claude Code**:
```json
{
  "mcpServers": {
    "web-search": {
      "command": "docker",
      "args": ["run", "--rm", "-i",
        "--network", "searxng-network",
        "-e", "SEARXNG_URL=http://searxng:8080",
        "searxng-mcp:latest"
      ]
    }
  }
}
```

**Python / TypeScript**: same pattern — pass `command: "docker"` with the same args to `mcp_servers`.

---

## Tools

| Tool | Description |
|------|-------------|
| `web_search` | General web search (Google, Bing, DDG aggregated) |
| `news_search` | Search recent news articles |

### `web_search` parameters
| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `query` | yes | — | Search query |
| `num_results` | no | 10 | Number of results (max 20) |
| `language` | no | auto-detected | Locale code e.g. `vi-VN`, `en-US` |
| `time_range` | no | — | `day` \| `week` \| `month` \| `year` |

### `news_search` parameters
| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `query` | yes | — | Search query |
| `num_results` | no | 10 | Number of results (max 20) |
| `language` | no | auto-detected | Locale code e.g. `vi-VN`, `en-US` |
| `time_range` | no | `week` | `day` \| `week` \| `month` \| `year` |

The `language` parameter is auto-detected from the query — searching in Vietnamese automatically returns Vietnamese results.

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SEARXNG_URL` | `http://localhost:8080` | SearXNG instance URL |
| `SEARXNG_LANGUAGE` | _(none)_ | Default language fallback (e.g. `vi-VN`) |
| `TRANSPORT` | `stdio` | `stdio` or `http` |
| `PORT` | `3000` | HTTP port (only used when `TRANSPORT=http`) |

---

## Makefile reference

| Command | Description |
|---------|-------------|
| `make build` | Build the Docker image |
| `make up` | Start SearXNG (stdio mode) |
| `make up-http` | Start SearXNG + MCP HTTP service on `:3333` |
| `make down` | Stop all services |
| `make restart` | Rebuild and restart |
| `make test` | Smoke-test stdio mode |
| `make test-http` | Smoke-test HTTP mode |
| `make logs` | Tail logs |
| `make status` | Show running containers |

---

## Development

```bash
npm install
npm run dev      # run stdio mode without building
npm run build    # compile TypeScript to dist/
```
