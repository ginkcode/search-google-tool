package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func createServer() *server.MCPServer {
	s := server.NewMCPServer("searxng-mcp", "1.0.0",
		server.WithToolCapabilities(true),
	)

	s.AddTool(
		mcp.NewTool("web_search",
			mcp.WithDescription("Search the web using SearXNG, which aggregates results from Google, DuckDuckGo and more. Returns a list of results with titles, URLs, and short snippets. Use fetch_content to read the full text of any result."),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("The search query"),
			),
			mcp.WithNumber("num_results",
				mcp.DefaultNumber(10),
				mcp.Description("Number of results to return (default: 10, max: 20)"),
			),
			mcp.WithString("language",
				mcp.Description(`Language code for search results. Detect from the query language and pass the matching locale code (e.g. "vi-VN" for Vietnamese, "en-US" for English, "fr-FR" for French, "ja-JP" for Japanese). Omit only if the language is ambiguous.`),
			),
			mcp.WithString("time_range",
				mcp.Description("Filter results by time range"),
				mcp.Enum("day", "week", "month", "year"),
			),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			query := req.GetString("query", "")
			numResults := min(req.GetInt("num_results", 10), 20)
			lang := req.GetString("language", defaultLang)
			timeRange := req.GetString("time_range", "")

			data, err := searchSearxng(query, searchOptions{
				categories: "general",
				numResults: numResults,
				language:   lang,
				timeRange:  timeRange,
			})
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(formatResults(data)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("news_search",
			mcp.WithDescription("Search for recent news articles using SearXNG. Returns titles, URLs, and short snippets. Use fetch_content to read the full text of any article."),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("The news search query"),
			),
			mcp.WithNumber("num_results",
				mcp.DefaultNumber(10),
				mcp.Description("Number of results to return (default: 10, max: 20)"),
			),
			mcp.WithString("language",
				mcp.Description(`Language code for results. Detect from the query language (e.g. "vi-VN", "en-US", "fr-FR"). Omit if ambiguous.`),
			),
			mcp.WithString("time_range",
				mcp.DefaultString("week"),
				mcp.Description("Filter by time range (default: week)"),
				mcp.Enum("day", "week", "month", "year"),
			),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			query := req.GetString("query", "")
			numResults := min(req.GetInt("num_results", 10), 20)
			lang := req.GetString("language", defaultLang)
			timeRange := req.GetString("time_range", "week")

			data, err := searchSearxng(query, searchOptions{
				categories: "news",
				numResults: numResults,
				language:   lang,
				timeRange:  timeRange,
			})
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(formatResults(data)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("fetch_content",
			mcp.WithDescription("Fetch and return the full text content of a web page. Use this after web_search or news_search when a result snippet is too short and you need the complete article or page content."),
			mcp.WithString("url",
				mcp.Required(),
				mcp.Description("The URL of the page to fetch"),
			),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			rawURL := req.GetString("url", "")
			text, err := fetchPageContent(rawURL)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(text), nil
		},
	)

	return s
}
