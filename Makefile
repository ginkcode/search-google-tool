.PHONY: build build-go up up-http up-go-http down restart restart-go logs status \
        test test-http test-fetch tools \
        go-test-http go-test-fetch go-tools \
        help

NODE_IMAGE := searxng-mcp:latest
GO_IMAGE   := searxng-go-mcp:latest
NETWORK    := searxng-network

# ── Node.js MCP ────────────────────────────────────────────────────────────────

## Build the Node.js MCP server Docker image
build:
	docker compose build mcp

## Start SearXNG + Node.js MCP in stdio mode
up:
	docker compose up -d searxng

## Start SearXNG + Node.js MCP in HTTP mode (persistent on :3333)
up-http:
	docker compose up -d searxng mcp-http

## Rebuild Node.js image and restart HTTP service
restart: build
	docker compose up -d --force-recreate mcp-http

## Smoke-test Node.js stdio mode
test:
	@echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"web_search","arguments":{"query":"SearXNG MCP","num_results":3}}}' \
	| docker run --rm -i --network $(NETWORK) \
	  -e SEARXNG_URL=http://searxng:8080 $(NODE_IMAGE)

## Smoke-test Node.js HTTP mode
test-http:
	curl -s -X POST http://localhost:3333/mcp \
	  -H "Content-Type: application/json" \
	  -H "Accept: application/json, text/event-stream" \
	  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"web_search","arguments":{"query":"SearXNG","num_results":3}}}' \
	  | grep "^data:" | sed 's/^data: //' | jq .

## List Node.js MCP tools and their input schemas
tools:
	curl -s -X POST http://localhost:3333/mcp \
	  -H "Content-Type: application/json" \
	  -H "Accept: application/json, text/event-stream" \
	  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' \
	  | grep "^data:" | sed 's/^data: //' | jq '.result.tools[] | {name, description, input: .inputSchema.properties}'

## Test Node.js fetch_content via HTTP (pass url=https://... to override)
test-fetch:
	curl -s -X POST http://localhost:3333/mcp \
	  -H "Content-Type: application/json" \
	  -H "Accept: application/json, text/event-stream" \
	  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"fetch_content","arguments":{"url":"$(or $(url),http://ginkcode.com)"}}}' \
	  | grep "^data:" | sed 's/^data: //' | jq .

# ── Go MCP ─────────────────────────────────────────────────────────────────────

## Build the Go MCP server Docker image
build-go:
	docker compose build go-mcp

## Start SearXNG + Go MCP in HTTP mode (persistent on :3334)
up-go-http:
	docker compose up -d searxng go-mcp-http

## Rebuild Go image and restart HTTP service
restart-go: build-go
	docker compose up -d --force-recreate go-mcp-http

## Smoke-test Go MCP HTTP mode
go-test-http:
	$(eval SESSION := $(shell curl -si -X POST http://localhost:3334/mcp \
	  -H "Content-Type: application/json" \
	  -H "Accept: application/json, text/event-stream" \
	  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"make","version":"1.0"}}}' \
	  | grep -i mcp-session-id | awk '{print $$2}' | tr -d '\r'))
	curl -s -X POST http://localhost:3334/mcp \
	  -H "Content-Type: application/json" \
	  -H "Accept: application/json, text/event-stream" \
	  -H "Mcp-Session-Id: $(SESSION)" \
	  -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"web_search","arguments":{"query":"SearXNG","num_results":3}}}' \
	  | jq .

## List Go MCP tools and their input schemas
go-tools:
	$(eval SESSION := $(shell curl -si -X POST http://localhost:3334/mcp \
	  -H "Content-Type: application/json" \
	  -H "Accept: application/json, text/event-stream" \
	  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"make","version":"1.0"}}}' \
	  | grep -i mcp-session-id | awk '{print $$2}' | tr -d '\r'))
	curl -s -X POST http://localhost:3334/mcp \
	  -H "Content-Type: application/json" \
	  -H "Accept: application/json, text/event-stream" \
	  -H "Mcp-Session-Id: $(SESSION)" \
	  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' \
	  | jq '.result.tools[] | {name, description, input: .inputSchema.properties}'

## Test Go fetch_content via HTTP (pass url=https://... to override)
go-test-fetch:
	$(eval SESSION := $(shell curl -si -X POST http://localhost:3334/mcp \
	  -H "Content-Type: application/json" \
	  -H "Accept: application/json, text/event-stream" \
	  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"make","version":"1.0"}}}' \
	  | grep -i mcp-session-id | awk '{print $$2}' | tr -d '\r'))
	curl -s -X POST http://localhost:3334/mcp \
	  -H "Content-Type: application/json" \
	  -H "Accept: application/json, text/event-stream" \
	  -H "Mcp-Session-Id: $(SESSION)" \
	  -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"fetch_content","arguments":{"url":"$(or $(url),http://ginkcode.com)"}}}' \
	  | jq .

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
	@grep -E '^##' Makefile | sed 's/## /  /'
