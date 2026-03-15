import { SEARXNG_URL } from "./config.js";

export interface SearxngResult {
  title: string;
  url: string;
  content: string;
  score?: number;
  category?: string;
  publishedDate?: string;
  engine?: string;
  engines?: string[];
}

export interface SearxngResponse {
  query: string;
  results: SearxngResult[];
  answers?: string[];
  infoboxes?: Array<{ infobox: string; content: string }>;
  suggestions?: string[];
  number_of_results?: number;
}

export async function searchSearxng(
  query: string,
  options: {
    categories?: string;
    language?: string;
    numResults?: number;
    timeRange?: string;
  } = {}
): Promise<SearxngResponse> {
  const params = new URLSearchParams({
    q: query,
    format: "json",
    categories: options.categories ?? "general",
    ...(options.language && { language: options.language }),
    ...(options.timeRange && { time_range: options.timeRange }),
  });

  const response = await fetch(`${SEARXNG_URL}/search?${params}`, {
    headers: {
      "Accept": "application/json",
      "Accept-Language": "en-US,en;q=0.9",
      "User-Agent": "Mozilla/5.0 (compatible; SearXNG-MCP/1.0)",
    },
  });

  if (!response.ok) {
    throw new Error(`SearXNG returned ${response.status}: ${response.statusText}`);
  }

  const data = (await response.json()) as SearxngResponse;

  if (options.numResults) {
    data.results = data.results.slice(0, options.numResults);
  }

  return data;
}

export function formatResults(data: SearxngResponse): string {
  const lines: string[] = [];

  if (data.answers && data.answers.length > 0) {
    lines.push(`**Direct answer:** ${data.answers[0]}\n`);
  }

  if (data.infoboxes && data.infoboxes.length > 0) {
    lines.push(`**${data.infoboxes[0].infobox}:** ${data.infoboxes[0].content}\n`);
  }

  if (data.results.length === 0) {
    return "No results found.";
  }

  data.results.forEach((result, i) => {
    lines.push(`${i + 1}. **${result.title}**`);
    lines.push(`   URL: ${result.url}`);
    if (result.content) {
      lines.push(`   ${result.content}`);
    }
    if (result.publishedDate) {
      lines.push(`   Published: ${result.publishedDate}`);
    }
    lines.push("");
  });

  if (data.suggestions && data.suggestions.length > 0) {
    lines.push(`\n**Related searches:** ${data.suggestions.slice(0, 5).join(", ")}`);
  }

  return lines.join("\n");
}
