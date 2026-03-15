"""Entry point for the SearXNG MCP server."""

import sys

from .config import PORT, SEARXNG_URL, TRANSPORT
from .server import mcp


def main():
    """Main entry point."""
    print(f"SearXNG MCP server running via {TRANSPORT} (SEARXNG_URL={SEARXNG_URL})", file=sys.stderr)
    if TRANSPORT == "http":
        print(f"  MCP endpoint: http://localhost:{PORT}/mcp", file=sys.stderr)
        mcp.run(transport="streamable-http")
    else:
        mcp.run()


if __name__ == "__main__":
    main()