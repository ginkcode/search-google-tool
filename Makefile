.PHONY: ts-build ts-up ts-up-http ts-restart ts-test ts-test-http ts-test-fetch ts-tools \
        go-build go-up-http go-restart go-test-http go-test-fetch go-tools \
        py-build py-up-http py-restart py-test py-test-http py-test-fetch py-tools \
        down logs status help

NODE_IMAGE  := searxng-mcp:latest
GO_IMAGE    := searxng-go-mcp:latest
PYTHON_IMAGE := searxng-py-mcp:latest
NETWORK     := searxng-network

# ── TypeScript MCP ─────────────────────────────────────────────────────────────

## Build the TypeScript MCP server Docker image
ts-build:
	docker compose build mcp

## Start SearXNG + TypeScript MCP in stdio mode
ts-up:
	docker compose up -d searxng

## Start SearXNG + TypeScript MCP in HTTP mode (persistent on :3333)
ts-up-http:
	docker compose up -d searxng mcp-http

## Rebuild TypeScript image and restart HTTP service
ts-restart: ts-build
	docker compose up -d --force-recreate mcp-http

## Smoke-test TypeScript stdio mode
ts-test:
	@echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"web_search","arguments":{"query":"SearXNG MCP","num_results":3}}}' \
	| docker run --rm -i --network $(NETWORK) \
	  -e SEARXNG_URL=http://searxng:8080 $(NODE_IMAGE)

## Smoke-test TypeScript HTTP mode
ts-test-http:
	curl -s -X POST http://localhost:3333/mcp \
	  -H "Content-Type: application/json" \
	  -H "Accept: application/json, text/event-stream" \
	  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"web_search","arguments":{"query":"SearXNG","num_results":3}}}' \
	  | grep "^data:" | sed 's/^data: //' | jq .

## List TypeScript MCP tools and their input schemas
ts-tools:
	curl -s -X POST http://localhost:3333/mcp \
	  -H "Content-Type: application/json" \
	  -H "Accept: application/json, text/event-stream" \
	  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' \
	  | grep "^data:" | sed 's/^data: //' | jq '.result.tools[] | {name, description, input: .inputSchema.properties}'

## Test TypeScript fetch_content via HTTP (pass url=https://... to override)
ts-test-fetch:
	curl -s -X POST http://localhost:3333/mcp \
	  -H "Content-Type: application/json" \
	  -H "Accept: application/json, text/event-stream" \
	  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"fetch_content","arguments":{"url":"$(or $(url),http://ginkcode.com)"}}}' \
	  | grep "^data:" | sed 's/^data: //' | jq .

# ── Go MCP ─────────────────────────────────────────────────────────────────────

## Build the Go MCP server Docker image
go-build:
	docker compose build go-mcp

## Start SearXNG + Go MCP in HTTP mode (persistent on :3334)
go-up-http:
	docker compose up -d searxng go-mcp-http

## Rebuild Go image and restart HTTP service
go-restart: go-build
	docker compose up -d --force-recreate go-mcp-http

## Smoke-test Go MCP HTTP mode
go-test-http:
	curl -s -X POST http://localhost:3334/mcp \
	  -H "Content-Type: application/json" \
	  -H "Accept: application/json, text/event-stream" \
	  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"web_search","arguments":{"query":"SearXNG","num_results":3}}}' \
	  | jq .

## List Go MCP tools and their input schemas
go-tools:
	curl -s -X POST http://localhost:3334/mcp \
	  -H "Content-Type: application/json" \
	  -H "Accept: application/json, text/event-stream" \
	  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' \
	  | jq '.result.tools[] | {name, description, input: .inputSchema.properties}'

## Test Go fetch_content via HTTP (pass url=https://... to override)
go-test-fetch:
	curl -s -X POST http://localhost:3334/mcp \
	  -H "Content-Type: application/json" \
	  -H "Accept: application/json, text/event-stream" \
	  -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"fetch_content","arguments":{"url":"$(or $(url),http://ginkcode.com)"}}}' \
	  | jq .

# ── Python MCP ─────────────────────────────────────────────────────────────────

## Build the Python MCP server Docker image
py-build:
	docker compose build py-mcp

## Start SearXNG + Python MCP in HTTP mode (persistent on :3335)
py-up-http:
	docker compose up -d searxng py-mcp-http

## Rebuild Python image and restart HTTP service
py-restart: py-build
	docker compose up -d --force-recreate py-mcp-http

## Smoke-test Python stdio mode
py-test:
	@echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"web_search","arguments":{"query":"SearXNG MCP","num_results":3}}}' \
	| docker run --rm -i --network $(NETWORK) \
	  -e SEARXNG_URL=http://searxng:8080 $(PYTHON_IMAGE)

## Smoke-test Python HTTP mode
py-test-http:
	curl -s -X POST http://localhost:3335/mcp \
	  -H "Content-Type: application/json" \
	  -H "Accept: application/json, text/event-stream" \
	  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"web_search","arguments":{"query":"SearXNG","num_results":3}}}' \
	  | grep "^data:" | sed 's/^data: //' | jq .

## List Python MCP tools and their input schemas
py-tools:
	curl -s -X POST http://localhost:3335/mcp \
	  -H "Content-Type: application/json" \
	  -H "Accept: application/json, text/event-stream" \
	  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' \
	  | grep "^data:" | sed 's/^data: //' | jq '.result.tools[] | {name, description, input: .inputSchema.properties}'

## Test Python fetch_content via HTTP (pass url=https://... to override)
py-test-fetch:
	curl -s -X POST http://localhost:3335/mcp \
	  -H "Content-Type: application/json" \
	  -H "Accept: application/json, text/event-stream" \
	  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"fetch_content","arguments":{"url":"$(or $(url),http://ginkcode.com)"}}}' \
	  | grep "^data:" | sed 's/^data: //' | jq .

# ── Shared ─────────────────────────────────────────────────────────────────────

## Stop all services
down:
	docker compose down

## Tail logs (pass s=<service> to filter, e.g. s=go-mcp-http)
logs:
	docker compose logs -f $(s)

## Show running services
status:
	docker compose ps

help:
	@awk '/^##/{desc=substr($$0,4); next} desc && /^[a-zA-Z0-9_-]+:/{printf "  \033[36m%-20s\033[0m %s\n", substr($$1,1,length($$1)-1), desc; desc=""}' Makefile
