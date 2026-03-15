"""Web content fetching with HTML stripping."""

import re
import urllib.request
import html

from .config import FETCH_MAX_CHARS

# Pre-compiled regex patterns for HTML stripping
_RE_SCRIPT = re.compile(r"<script[^>]*>.*?</script>", re.IGNORECASE | re.DOTALL)
_RE_STYLE = re.compile(r"<style[^>]*>.*?</style>", re.IGNORECASE | re.DOTALL)
_RE_TAG = re.compile(r"<[^>]+>")
_RE_SPACES = re.compile(r"\s{2,}")


def strip_html(html_text: str) -> str:
    """Strip HTML tags and decode entities."""
    text = _RE_SCRIPT.sub("", html_text)
    text = _RE_STYLE.sub("", text)
    text = _RE_TAG.sub(" ", text)
    text = html.unescape(text)
    text = _RE_SPACES.sub(" ", text)
    return text.strip()


def fetch_page_content(url: str) -> str:
    """Fetch and extract text content from a web page."""
    headers = {
        "User-Agent": "Mozilla/5.0 (compatible; SearXNG-MCP/1.0)",
        "Accept": "text/html,application/xhtml+xml",
        "Accept-Language": "en-US,en;q=0.9",
    }

    req = urllib.request.Request(url, headers=headers, method="GET")
    with urllib.request.urlopen(req, timeout=10) as resp:
        if resp.status != 200:
            raise Exception(f"HTTP {resp.status}")

        content_type = resp.headers.get("Content-Type", "")
        if "text/html" not in content_type and "text/plain" not in content_type:
            raise Exception(f"Unsupported content type: {content_type}")

        html_bytes = resp.read()
        html_text = html_bytes.decode("utf-8", errors="replace")

    text = strip_html(html_text)

    if len(text) > FETCH_MAX_CHARS:
        return f"{text[:FETCH_MAX_CHARS]}\n\n[Truncated — {len(text)} total chars]"

    return text