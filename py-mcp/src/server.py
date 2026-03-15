"""MCP server and tool definitions."""

from mcp.server.fastmcp import FastMCP

from .config import DEFAULT_LANGUAGE, PORT, TRANSPORT
from .fetch import fetch_page_content
from .searxng import search_searxng, format_results

# Configure server based on transport mode
mcp = FastMCP(
    "searxng-mcp",
    host="0.0.0.0",
    port=PORT,
    stateless_http=(TRANSPORT == "http"),
)


@mcp.tool()
def web_search(
    query: str,
    num_results: int = 10,
    language: str | None = None,
    time_range: str | None = None,
) -> str:
    """Search the web using SearXNG, which aggregates results from Google, DuckDuckGo and more. Returns a list of results with titles, URLs, and short snippets. Use fetch_content to read the full text of any result.

    Args:
        query: The search query
        num_results: Number of results to return (default: 10, max: 20)
        language: Language code for search results (e.g. "vi-VN", "en-US", "fr-FR")
        time_range: Filter results by time range ("day", "week", "month", "year")
    """
    num_results = min(num_results, 20)
    lang = language or (DEFAULT_LANGUAGE or None)
    data = search_searxng(
        query,
        categories="general",
        num_results=num_results,
        language=lang,
        time_range=time_range,
    )
    return format_results(data)


@mcp.tool()
def news_search(
    query: str,
    num_results: int = 10,
    language: str | None = None,
    time_range: str = "week",
) -> str:
    """Search for recent news articles using SearXNG. Returns titles, URLs, and short snippets. Use fetch_content to read the full text of any article.

    Args:
        query: The news search query
        num_results: Number of results to return (default: 10, max: 20)
        language: Language code for results (e.g. "vi-VN", "en-US", "fr-FR")
        time_range: Filter by time range (default: "week")
    """
    num_results = min(num_results, 20)
    lang = language or (DEFAULT_LANGUAGE or None)
    data = search_searxng(
        query,
        categories="news",
        num_results=num_results,
        language=lang,
        time_range=time_range,
    )
    return format_results(data)


@mcp.tool()
def fetch_content(url: str) -> str:
    """Fetch and return the full text content of a web page. Use this after web_search or news_search when a result snippet is too short and you need the complete article or page content.

    Args:
        url: The URL of the page to fetch
    """
    return fetch_page_content(url)