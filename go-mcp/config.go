package main

import (
	"os"
	"strconv"
)

var (
	searxngURL      = getEnv("SEARXNG_URL", "http://localhost:8080")
	defaultLang     = getEnv("SEARXNG_LANGUAGE", "")
	port            = getEnvInt("PORT", 3000)
	transport       = getEnv("TRANSPORT", "stdio")
	flareSolverrURL = getEnv("FLARESOLVERR_URL", "")
	fetchMaxChars   = 20000
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
