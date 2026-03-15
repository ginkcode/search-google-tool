#!/usr/bin/env node
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { StreamableHTTPServerTransport } from "@modelcontextprotocol/sdk/server/streamableHttp.js";
import http from "node:http";
import { PORT, SEARXNG_URL, TRANSPORT } from "./config.js";
import { createServer } from "./server.js";

async function startStdio() {
  const transport = new StdioServerTransport();
  await createServer().connect(transport);
  console.error(`SearXNG MCP server running via stdio (SEARXNG_URL=${SEARXNG_URL})`);
}

async function startHttp() {
  const httpServer = http.createServer(async (req, res) => {
    if (req.method === "GET" && req.url === "/health") {
      res.writeHead(200).end(JSON.stringify({ status: "ok" }));
      return;
    }

    if (req.url !== "/mcp") {
      res.writeHead(404).end();
      return;
    }

    const transport = new StreamableHTTPServerTransport({
      sessionIdGenerator: undefined,
    });
    await createServer().connect(transport);
    await transport.handleRequest(req, res);
  });

  httpServer.listen(PORT, () => {
    console.error(`SearXNG MCP server running via HTTP on port ${PORT} (SEARXNG_URL=${SEARXNG_URL})`);
    console.error(`  MCP endpoint: http://localhost:${PORT}/mcp`);
    console.error(`  Health check: http://localhost:${PORT}/health`);
  });
}

if (TRANSPORT === "http") {
  startHttp().catch((err) => { console.error("Fatal error:", err); process.exit(1); });
} else {
  startStdio().catch((err) => { console.error("Fatal error:", err); process.exit(1); });
}
