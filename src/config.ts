export const SEARXNG_URL = process.env.SEARXNG_URL ?? "http://localhost:8080";
export const DEFAULT_LANGUAGE = process.env.SEARXNG_LANGUAGE ?? "";
export const PORT = parseInt(process.env.PORT ?? "3000", 10);
export const TRANSPORT = process.env.TRANSPORT ?? "stdio";
export const FETCH_MAX_CHARS = 20000;
