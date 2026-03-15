package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	if transport == "http" {
		startHTTP()
	} else {
		startStdio()
	}
}

func startStdio() {
	srv := server.NewStdioServer(createServer())
	log.Printf("SearXNG MCP server running via stdio (SEARXNG_URL=%s)", searxngURL)
	if err := srv.Listen(context.Background(), os.Stdin, os.Stdout); err != nil {
		log.Fatalf("Fatal error: %v", err)
	}
}

func startHTTP() {
	mcpHandler := server.NewStreamableHTTPServer(createServer(),
		server.WithHeartbeatInterval(15*time.Second),
		server.WithStateLess(true),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.Handle("/mcp", mcpHandler)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("SearXNG MCP server running via HTTP on port %d (SEARXNG_URL=%s)", port, searxngURL)
	log.Printf("  MCP endpoint: http://localhost%s/mcp", addr)
	log.Printf("  Health check: http://localhost%s/health", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Fatal error: %v", err)
	}
}
