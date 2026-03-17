# Issues & Solutions

## Fetch tool returning 403 on protected sites (e.g. Forbes)

**Root cause:** Bot-detection systems fingerprint HTTP clients via TLS ClientHello (JA3/JA4 fingerprint), HTTP version, and User-Agent. The default HTTP libraries in all three MCPs were detectable as non-browser clients.

---

### Python MCP

**Library:** `urllib.request` (standard library)

**Problem:** Python's `ssl` module produces a recognizable TLS fingerprint and only speaks HTTP/1.1. Sites like Forbes block it with 403.

**Solution:** Replaced `urllib.request` with `curl_cffi` using `impersonate="safari17_0"`.

- `curl_cffi` wraps libcurl and reproduces a real browser's TLS fingerprint (cipher suites, extension order, ALPN).
- The `safari17_0` profile was chosen after testing — Chrome profiles were blocked but Safari/Firefox profiles returned 200.

**Changes:**
- `py-mcp/pyproject.toml`: added `curl_cffi>=0.7.0`
- `py-mcp/src/fetch.py`: replaced `urllib.request` with `curl_cffi.requests`

---

### Go MCP

**Library:** `net/http` (standard library)

**Problem:** Go's `crypto/tls` produces a distinct TLS fingerprint that bot-detection systems identify and block.

**Solution:** Replaced `net/http` with `bogdanfinn/tls-client` using `profiles.Safari_IOS_17_0`.

- `bogdanfinn/tls-client` wraps `refraction-networking/utls` (pure Go — compatible with `CGO_ENABLED=0`) and reproduces browser TLS fingerprints.
- `fhttp` (bogdanfinn's fork of `net/http`) is required for request construction as it is the type the client expects.

**Changes:**
- `go-mcp/go.mod`: added `github.com/bogdanfinn/tls-client`
- `go-mcp/fetch.go`: replaced `net/http` with `fhttp` + `tlsclient`

---

### Node MCP

**Library:** `fetch` (built-in, via undici)

**Problem:** Node.js's built-in `fetch` (undici) has a recognizable TLS fingerprint. Additionally, a custom `User-Agent: SearXNG-MCP/1.0` header was being sent — Forbes blocks non-standard User-Agents unless the TLS fingerprint matches a known browser.

**Solution:** Replaced `fetch()` with a `curl` subprocess call via `child_process.execFile`.

- System `curl` uses its own TLS stack and User-Agent (`curl/x.x.x`) which Forbes allows.
- The custom `User-Agent` header was removed so curl uses its default.
- HTTP status code is extracted via `--write-out "\n__STATUS__%{http_code}"` appended to stdout.

**Changes:**
- `node-mcp/Dockerfile`: added `apk add --no-cache curl` to the runtime image
- `node-mcp/src/fetch.ts`: replaced `fetch()` with `execFileAsync("curl", [...])`

---

## Fetch tool blocked by Cloudflare Managed Challenge (e.g. en-hrana.org)

**Root cause:** Some sites use Cloudflare's Managed Challenge (`cf-mitigated: challenge`), which requires executing JavaScript in a real browser to obtain a valid session cookie. No plain HTTP client can solve this — not even one with a correct TLS fingerprint.

**Solution:** Added [FlareSolverr](https://github.com/FlareSolverr/FlareSolverr) as a sidecar service. It runs headless Chrome and exposes a REST API to solve Cloudflare challenges. Each MCP detects the `cf-mitigated: challenge` response header and automatically falls back to FlareSolverr.

**Flow:**
1. MCP makes a normal request (curl_cffi / bogdanfinn / curl subprocess)
2. If the response has `cf-mitigated` header AND `FLARESOLVERR_URL` is set → POST to `FLARESOLVERR_URL/v1` with `{"cmd":"request.get","url":"..."}`
3. FlareSolverr solves the JS challenge using Chrome and returns the page HTML
4. MCP strips and returns the content as usual

**Changes:**
- `docker-compose.yml`: added `flaresolverr` service (`ghcr.io/flaresolverr/flaresolverr`) with `shm_size: 128m`; added `FLARESOLVERR_URL=http://flaresolverr:8191` to all six MCP service definitions
- `py-mcp/src/config.py`, `go-mcp/config.go`, `node-mcp/src/config.ts`: added `FLARESOLVERR_URL` env var
- `py-mcp/src/fetch.py`: checks `resp.headers.get("cf-mitigated")` on non-200; calls FlareSolverr via `curl_cffi.requests.post`
- `go-mcp/fetch.go`: checks `resp.Header.Get("cf-mitigated")` on non-200; calls FlareSolverr via standard `net/http.Post`
- `node-mcp/src/fetch.ts`: switched curl from `--write-out` status-only to `--include` (dumps headers inline); parses `cf-mitigated` header from output; calls FlareSolverr via built-in `fetch()`

**Note:** Alpine's curl build (8.17.0 musl) does not support `%{header.X}` write-out variables despite the version being ≥7.84. Using `--include` is the portable alternative.
