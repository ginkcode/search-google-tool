.PHONY: build up up-http down restart logs status test test-http test-fetch tools help

IMAGE   := searxng-mcp:latest
NETWORK := searxng-network

## Build the MCP server Docker image
build:
	docker compose build mcp

## Start SearXNG + MCP in stdio mode (spawned per session)
up:
	docker compose up -d searxng

## Start SearXNG + MCP in HTTP mode (persistent service on :3000)
up-http:
	docker compose up -d searxng mcp-http

## Stop all services
down:
	docker compose down

## Rebuild image and restart
restart: build
	docker compose up -d --force-recreate mcp-http

## Tail logs (pass s=mcp-http to filter a service)
logs:
	docker compose logs -f $(s)

## Show running services
status:
	docker compose ps

## Smoke-test stdio mode
test:
	@echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"web_search","arguments":{"query":"SearXNG MCP","num_results":3}}}' \
	| docker run --rm -i --network $(NETWORK) \
	  -e SEARXNG_URL=http://searxng:8080 $(IMAGE)

## Smoke-test HTTP mode
test-http:
	curl -s -X POST http://localhost:3333/mcp \
	  -H "Content-Type: application/json" \
	  -H "Accept: application/json, text/event-stream" \
	  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"web_search","arguments":{"query":"SearXNG","num_results":3}}}' \
	  | grep "^data:" | sed 's/^data: //' | jq .

## List available tools and their input schemas
tools:
	curl -s -X POST http://localhost:3333/mcp \
	  -H "Content-Type: application/json" \
	  -H "Accept: application/json, text/event-stream" \
	  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}' \
	  | grep "^data:" | sed 's/^data: //' | jq '.result.tools[] | {name, description, input: .inputSchema.properties}'

## Test fetch_content tool via HTTP mode (pass url=https://... to override)
test-fetch:
	curl -s -X POST http://localhost:3333/mcp \
	  -H "Content-Type: application/json" \
	  -H "Accept: application/json, text/event-stream" \
	  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"fetch_content","arguments":{"url":"$(or $(url),http://ginkcode.com)"}}}' \
	  | grep "^data:" | sed 's/^data: //' | jq .

help:
	@grep -E '^##' Makefile | sed 's/## /  /'
