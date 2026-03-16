import { execFile } from "node:child_process";
import { promisify } from "node:util";
import { FETCH_MAX_CHARS } from "./config.js";

const execFileAsync = promisify(execFile);

function stripHtml(html: string): string {
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

export async function fetchPageContent(url: string): Promise<string> {
  let stdout: string;
  try {
    ({ stdout } = await execFileAsync(
      "curl",
      [
        "--silent",
        "--location",
        "--max-time", "10",
        "--write-out", "\n__STATUS__%{http_code}",
        "--header", "Accept: text/html,application/xhtml+xml",
        "--header", "Accept-Language: en-US,en;q=0.9",
        url,
      ],
      { maxBuffer: 10 * 1024 * 1024 }
    ));
  } catch (e: any) {
    throw new Error(`curl failed: ${e.stderr || e.message}`);
  }

  const statusMarker = "\n__STATUS__";
  const markerIdx = stdout.lastIndexOf(statusMarker);
  const statusCode = parseInt(stdout.slice(markerIdx + statusMarker.length), 10);
  const body = stdout.slice(0, markerIdx);

  if (statusCode !== 200) {
    throw new Error(`HTTP Error ${statusCode}`);
  }

  const text = stripHtml(body);

  return text.length > FETCH_MAX_CHARS
    ? text.slice(0, FETCH_MAX_CHARS) + `\n\n[Truncated — ${text.length} total chars]`
    : text;
}
