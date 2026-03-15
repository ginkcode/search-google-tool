"""SearXNG API integration."""

from dataclasses import dataclass
from typing import Any
import urllib.request
import urllib.parse
import json

from .config import SEARXNG_URL, DEFAULT_LANGUAGE


@dataclass
class SearxngResult:
    title: str
    url: str
    content: str
    published_date: str | None = None
    engine: str | None = None
    engines: list[str] | None = None


@dataclass
class SearxngResponse:
    query: str
    results: list[SearxngResult]
    answers: list[Any] | None = None
    infoboxes: list[dict[str, str]] | None = None
    suggestions: list[str] | None = None


def search_searxng(
    query: str,
    *,
    categories: str = "general",
    language: str | None = None,
    num_results: int = 10,
    time_range: str | None = None,
) -> SearxngResponse:
    """Search SearXNG and return parsed results."""
    params = {
        "q": query,
        "format": "json",
        "categories": categories,
    }
    if language:
        params["language"] = language
    if time_range:
        params["time_range"] = time_range

    url = f"{SEARXNG_URL}/search?{urllib.parse.urlencode(params)}"
    headers = {
        "Accept": "application/json",
        "Accept-Language": "en-US,en;q=0.9",
        "User-Agent": "Mozilla/5.0 (compatible; SearXNG-MCP/1.0)",
    }

    req = urllib.request.Request(url, headers=headers, method="GET")
    with urllib.request.urlopen(req) as resp:
        if resp.status != 200:
            raise Exception(f"SearXNG returned {resp.status}")
        data = json.loads(resp.read().decode("utf-8"))

    results = [
        SearxngResult(
            title=r.get("title", ""),
            url=r.get("url", ""),
            content=r.get("content", ""),
            published_date=r.get("publishedDate"),
            engine=r.get("engine"),
            engines=r.get("engines"),
        )
        for r in data.get("results", [])[:num_results]
    ]

    return SearxngResponse(
        query=data.get("query", ""),
        results=results,
        answers=data.get("answers"),
        infoboxes=data.get("infoboxes"),
        suggestions=data.get("suggestions"),
    )


def format_results(data: SearxngResponse) -> str:
    """Format SearXNG response as markdown text."""
    lines: list[str] = []

    if data.answers:
        answer = data.answers[0]
        if isinstance(answer, dict):
            answer = answer.get("answer", "")
        lines.append(f"**Direct answer:** {answer}\n")

    if data.infoboxes:
        infobox = data.infoboxes[0]
        lines.append(f"**{infobox.get('infobox', 'Info')}:** {infobox.get('content', '')}\n")

    if not data.results:
        return "No results found."

    for i, r in enumerate(data.results, 1):
        lines.append(f"{i}. **{r.title}**")
        lines.append(f"   URL: {r.url}")
        if r.content:
            lines.append(f"   {r.content}")
        if r.published_date:
            lines.append(f"   Published: {r.published_date}")
        lines.append("")

    if data.suggestions:
        lines.append(f"\n**Related searches:** {', '.join(data.suggestions[:5])}")

    return "\n".join(lines)