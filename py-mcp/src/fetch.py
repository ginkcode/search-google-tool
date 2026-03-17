"""Web content fetching with HTML stripping."""

import json
import re
import html

from curl_cffi import requests as curl_requests

from .config import FETCH_MAX_CHARS, FLARESOLVERR_URL

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


def _fetch_via_flaresolverr(url: str) -> str:
    resp = curl_requests.post(
        f"{FLARESOLVERR_URL}/v1",
        json={"cmd": "request.get", "url": url, "maxTimeout": 60000},
        timeout=70,
    )
    data = resp.json()
    if data.get("status") != "ok":
        raise Exception(f"FlareSolverr error: {data.get('message', 'unknown')}")
    return data["solution"]["response"]


def fetch_page_content(url: str) -> str:
    """Fetch and extract text content from a web page."""
    resp = curl_requests.get(url, impersonate="safari17_0", timeout=10)

    if resp.status_code != 200:
        if resp.headers.get("cf-mitigated") and FLARESOLVERR_URL:
            html_text = _fetch_via_flaresolverr(url)
        else:
            raise Exception(f"HTTP Error {resp.status_code}: {resp.reason}")
    else:
        content_type = resp.headers.get("content-type", "")
        if "text/html" not in content_type and "text/plain" not in content_type:
            raise Exception(f"Unsupported content type: {content_type}")
        html_text = resp.text

    text = strip_html(html_text)

    if len(text) > FETCH_MAX_CHARS:
        return f"{text[:FETCH_MAX_CHARS]}\n\n[Truncated — {len(text)} total chars]"

    return text
