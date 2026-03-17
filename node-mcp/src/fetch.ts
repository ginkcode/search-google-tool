import { execFile } from "node:child_process";
import { promisify } from "node:util";
import { JSDOM } from "jsdom";
import { Readability } from "@mozilla/readability";
import { FETCH_MAX_CHARS, FLARESOLVERR_URL } from "./config.js";

const RE_CANONICAL = /<link[^>]+rel=["']canonical["'][^>]+href=["']([^"']+)["']|<link[^>]+href=["']([^"']+)["'][^>]+rel=["']canonical["']/i;
const RE_OG_URL = /<meta[^>]+property=["']og:url["'][^>]+content=["']([^"']+)["']|<meta[^>]+content=["']([^"']+)["'][^>]+property=["']og:url["']/i;
const RE_JSONLD_URL = /"url"\s*:\s*"(https?:\/\/[^"]+)"/gi;

function extractSourceUrl(html: string, originalUrl: string): string | null {
  const originalHost = new URL(originalUrl).hostname;
  for (const re of [RE_CANONICAL, RE_OG_URL]) {
    const m = html.match(re);
    if (!m) continue;
    const candidate = m.slice(1).find(Boolean);
    if (!candidate) continue;
    try {
      const host = new URL(candidate).hostname;
      if (host && host !== originalHost) return candidate;
    } catch {}
  }
  RE_JSONLD_URL.lastIndex = 0;
  let m: RegExpExecArray | null;
  while ((m = RE_JSONLD_URL.exec(html)) !== null) {
    try {
      const host = new URL(m[1]).hostname;
      if (host && host !== originalHost) return m[1];
    } catch {}
  }
  return null;
}

const execFileAsync = promisify(execFile);

function regexStrip(html: string): string {
  return html
    .replace(/<script[\s\S]*?<\/script>/gi, "")
    .replace(/<style[\s\S]*?<\/style>/gi, "")
    .replace(/<[^>]+>/g, " ")
    .replace(/&nbsp;/g, " ")
    .replace(/&amp;/g, "&")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&quot;/g, '"')
    .replace(/&#39;/g, "'")
    .replace(/\s{2,}/g, " ")
    .trim();
}

function stripHtml(html: string, url?: string): string {
  try {
    const dom = new JSDOM(html, { url: url ?? "http://localhost" });
    const article = new Readability(dom.window.document).parse();
    if (article?.textContent) {
      const text = article.textContent.replace(/\s{2,}/g, " ").trim();
      if (text.length > 200) return text;
    }
  } catch {}
  return regexStrip(html);
}

async function fetchViaFlareSolverr(url: string): Promise<string> {
  const resp = await fetch(`${FLARESOLVERR_URL}/v1`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ cmd: "request.get", url, maxTimeout: 60000 }),
    signal: AbortSignal.timeout(70000),
  });
  const data = await resp.json() as any;
  if (data.status !== "ok") {
    throw new Error(`FlareSolverr error: ${data.message ?? "unknown"}`);
  }
  return data.solution.response as string;
}

function parseResponse(raw: string): { statusCode: number; cfMitigated: string; body: string } {
  // --include prepends "HTTP/x.x STATUS\r\nHeader: value\r\n...\r\n\r\n" before the body.
  // After redirects there may be multiple header blocks; we want the last one.
  const headerBodySep = "\r\n\r\n";
  let lastSepIdx = raw.lastIndexOf(headerBodySep);
  if (lastSepIdx === -1) {
    return { statusCode: 0, cfMitigated: "", body: raw };
  }
  const headers = raw.slice(0, lastSepIdx);
  const body = raw.slice(lastSepIdx + headerBodySep.length);

  // Find the last status line in the header block
  const statusMatch = headers.match(/HTTP\/[\d.]+ (\d+)/g);
  const statusCode = statusMatch
    ? parseInt(statusMatch[statusMatch.length - 1].split(" ")[1], 10)
    : 0;

  const cfLine = headers.split("\r\n").find(l => l.toLowerCase().startsWith("cf-mitigated:"));
  const cfMitigated = cfLine ? cfLine.split(":")[1].trim() : "";

  return { statusCode, cfMitigated, body };
}

export async function fetchPageContent(url: string): Promise<string> {
  let stdout: string;
  try {
    ({ stdout } = await execFileAsync(
      "curl",
      [
        "--silent",
        "--include",
        "--location",
        "--max-time", "10",
        "--header", "Accept: text/html,application/xhtml+xml",
        "--header", "Accept-Language: en-US,en;q=0.9",
        url,
      ],
      { maxBuffer: 10 * 1024 * 1024 }
    ));
  } catch (e: any) {
    throw new Error(`curl failed: ${e.stderr || e.message}`);
  }

  const { statusCode, cfMitigated, body } = parseResponse(stdout);

  let htmlText: string;
  if (statusCode !== 200) {
    if (cfMitigated && FLARESOLVERR_URL) {
      htmlText = await fetchViaFlareSolverr(url);
    } else {
      throw new Error(`HTTP Error ${statusCode}`);
    }
  } else {
    htmlText = body;
  }

  let text = stripHtml(htmlText, url);

  if (text.length < 500 && FLARESOLVERR_URL) {
    htmlText = await fetchViaFlareSolverr(url);
    text = stripHtml(htmlText, url);

    if (text.length < 500) {
      const sourceUrl = extractSourceUrl(htmlText, url);
      if (sourceUrl) {
        const content = await fetchPageContent(sourceUrl);
        return `${content}\n\nReal source: ${sourceUrl}`;
      }
    }
  }

  return text.length > FETCH_MAX_CHARS
    ? text.slice(0, FETCH_MAX_CHARS) + `\n\n[Truncated — ${text.length} total chars]`
    : text;
}
