"""Configuration loaded from environment variables."""

import os

SEARXNG_URL = os.environ.get("SEARXNG_URL", "http://localhost:8080")
DEFAULT_LANGUAGE = os.environ.get("SEARXNG_LANGUAGE", "")
try:
    PORT = int(os.environ.get("PORT", "3000"))
except ValueError:
    PORT = 3000
TRANSPORT = os.environ.get("TRANSPORT", "stdio")
FETCH_MAX_CHARS = 20000
FLARESOLVERR_URL = os.environ.get("FLARESOLVERR_URL", "")
