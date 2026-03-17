"""Web content fetching with HTML stripping."""

import html
import re
from urllib.parse import urlparse

from curl_cffi import requests as curl_requests
from readability import Document

from .config import FETCH_MAX_CHARS, FLARESOLVERR_URL

_RE_CANONICAL = re.compile(
    r'<link[^>]+rel=["\']canonical["\'][^>]+href=["\']([^"\']+)["\']'
    r'|<link[^>]+href=["\']([^"\']+)["\'][^>]+rel=["\']canonical["\']',
    re.IGNORECASE,
)
_RE_OG_URL = re.compile(
    r'<meta[^>]+property=["\']og:url["\'][^>]+content=["\']([^"\']+)["\']'
    r'|<meta[^>]+content=["\']([^"\']+)["\'][^>]+property=["\']og:url["\']',
    re.IGNORECASE,
)
_RE_JSONLD_URL = re.compile(r'"url"\s*:\s*"(https?://[^"]+)"', re.IGNORECASE)

# Pre-compiled regex patterns for HTML stripping
_RE_SCRIPT = re.compile(r"<script[^>]*>.*?</script>", re.IGNORECASE | re.DOTALL)
_RE_STYLE = re.compile(r"<style[^>]*>.*?</style>", re.IGNORECASE | re.DOTALL)
_RE_TAG = re.compile(r"<[^>]+>")
_RE_SPACES = re.compile(r"\s{2,}")


def _regex_strip(html_text: str) -> str:
    text = _RE_SCRIPT.sub("", html_text)
    text = _RE_STYLE.sub("", text)
    text = _RE_TAG.sub(" ", text)
    text = html.unescape(text)
    text = _RE_SPACES.sub(" ", text)
    return text.strip()


def strip_html(html_text: str) -> str:
    """Extract article text using Readability, fall back to regex stripping."""
    try:
        doc = Document(html_text)
        content = doc.summary(html_partial=True)
        if content:
            text = _regex_strip(content)
            if len(text) > 200:
                return text
    except Exception:
        pass
    return _regex_strip(html_text)


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


def _extract_source_url(html_text: str, original_url: str) -> str | None:
    """Extract canonical/og:url/JSON-LD url if it points to a different domain."""
    original_host = urlparse(original_url).hostname or ""
    for pattern in (_RE_CANONICAL, _RE_OG_URL):
        m = pattern.search(html_text)
        if m:
            candidate = next(g for g in m.groups() if g)
            candidate_host = urlparse(candidate).hostname or ""
            if candidate_host and candidate_host != original_host:
                return candidate
    for m in _RE_JSONLD_URL.finditer(html_text):
        candidate = m.group(1)
        candidate_host = urlparse(candidate).hostname or ""
        if candidate_host and candidate_host != original_host:
            return candidate
    return None


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

    if len(text) < 500 and FLARESOLVERR_URL:
        html_text = _fetch_via_flaresolverr(url)
        text = strip_html(html_text)

        # If still thin, try fetching the source article URL
        if len(text) < 500:
            source_url = _extract_source_url(html_text, url)
            if source_url:
                content = fetch_page_content(source_url)
                return f"{content}\n\nReal source: {source_url}"

    if len(text) > FETCH_MAX_CHARS:
        return f"{text[:FETCH_MAX_CHARS]}\n\n[Truncated — {len(text)} total chars]"

    return text
