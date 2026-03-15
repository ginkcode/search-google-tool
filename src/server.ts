import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { z } from "zod";
import { DEFAULT_LANGUAGE } from "./config.js";
import { fetchPageContent } from "./fetch.js";
import { searchSearxng, formatResults } from "./searxng.js";

const TIME_RANGE = z.enum(["day", "week", "month", "year"]).optional();
const LANGUAGE = z.string().optional();

export function createServer(): McpServer {
  const server = new McpServer({ name: "searxng-mcp", version: "1.0.0" });

  server.registerTool(
    "web_search",
    {
      description:
        "Search the web using SearXNG, which aggregates results from Google, DuckDuckGo and more. Returns a list of results with titles, URLs, and short snippets. Use fetch_content to read the full text of any result.",
      inputSchema: {
        query: z.string().describe("The search query"),
        num_results: z
          .number()
          .default(10)
          .describe("Number of results to return (default: 10, max: 20)"),
        language: LANGUAGE.describe(
          'Language code for search results. Detect from the query language and pass the matching locale code (e.g. "vi-VN" for Vietnamese, "en-US" for English, "fr-FR" for French, "ja-JP" for Japanese). Omit only if the language is ambiguous.',
        ),
        time_range: TIME_RANGE.describe("Filter results by time range"),
      },
    },
    async ({ query, num_results, language, time_range }) => {
      const numResults = Math.min(num_results ?? 10, 20);
      const lang = language ?? (DEFAULT_LANGUAGE || undefined);
      const data = await searchSearxng(query, {
        categories: "general",
        numResults,
        language: lang,
        timeRange: time_range,
      });
      return { content: [{ type: "text", text: formatResults(data) }] };
    },
  );

  server.registerTool(
    "news_search",
    {
      description:
        "Search for recent news articles using SearXNG. Returns titles, URLs, and short snippets. Use fetch_content to read the full text of any article.",
      inputSchema: {
        query: z.string().describe("The news search query"),
        num_results: z
          .number()
          .default(10)
          .describe("Number of results to return (default: 10, max: 20)"),
        language: LANGUAGE.describe(
          'Language code for results. Detect from the query language (e.g. "vi-VN", "en-US", "fr-FR"). Omit if ambiguous.',
        ),
        time_range: TIME_RANGE.default("week").describe(
          "Filter by time range (default: week)",
        ),
      },
    },
    async ({ query, num_results, language, time_range }) => {
      const numResults = Math.min(num_results ?? 10, 20);
      const lang = language ?? (DEFAULT_LANGUAGE || undefined);
      const data = await searchSearxng(query, {
        categories: "news",
        numResults,
        timeRange: time_range,
        language: lang,
      });
      return { content: [{ type: "text", text: formatResults(data) }] };
    },
  );

  server.registerTool(
    "fetch_content",
    {
      description:
        "Fetch and return the full text content of a web page. Use this after web_search or news_search when a result snippet is too short and you need the complete article or page content.",
      inputSchema: {
        url: z.string().describe("The URL of the page to fetch"),
      },
    },
    async ({ url }) => {
      const text = await fetchPageContent(url);
      return { content: [{ type: "text", text }] };
    },
  );

  return server;
}
