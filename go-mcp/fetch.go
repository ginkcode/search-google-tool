package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	fhttp "github.com/bogdanfinn/fhttp"
	tlsclient "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

var (
	reScript    = regexp.MustCompile(`(?si)<script.*?</script>`)
	reStyle     = regexp.MustCompile(`(?si)<style.*?</style>`)
	reTag       = regexp.MustCompile(`<[^>]+>`)
	reSpaces    = regexp.MustCompile(`\s{2,}`)
	reCanonical = regexp.MustCompile(`(?i)<link[^>]+rel=["']canonical["'][^>]+href=["']([^"']+)["']|<link[^>]+href=["']([^"']+)["'][^>]+rel=["']canonical["']`)
	reOgURL     = regexp.MustCompile(`(?i)<meta[^>]+property=["']og:url["'][^>]+content=["']([^"']+)["']|<meta[^>]+content=["']([^"']+)["'][^>]+property=["']og:url["']`)

	htmlEntities = strings.NewReplacer(
		"&nbsp;", " ",
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", `"`,
		"&#39;", "'",
	)
)

func stripHTML(html string) string {
	html = reScript.ReplaceAllString(html, "")
	html = reStyle.ReplaceAllString(html, "")
	html = reTag.ReplaceAllString(html, " ")
	html = htmlEntities.Replace(html)
	html = reSpaces.ReplaceAllString(html, " ")
	return strings.TrimSpace(html)
}

func fetchViaFlareSolverr(rawURL string) (string, error) {
	payload, _ := json.Marshal(map[string]any{
		"cmd":        "request.get",
		"url":        rawURL,
		"maxTimeout": 60000,
	})
	resp, err := http.Post(flareSolverrURL+"/v1", "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("FlareSolverr request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Status   string `json:"status"`
		Message  string `json:"message"`
		Solution struct {
			Response string `json:"response"`
		} `json:"solution"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("FlareSolverr response parse failed: %w", err)
	}
	if result.Status != "ok" {
		return "", fmt.Errorf("FlareSolverr error: %s", result.Message)
	}
	return result.Solution.Response, nil
}

func extractSourceURL(htmlText, originalURL string) string {
	originalHost := ""
	if u, err := url.Parse(originalURL); err == nil {
		originalHost = u.Hostname()
	}
	for _, re := range []*regexp.Regexp{reCanonical, reOgURL} {
		m := re.FindStringSubmatch(htmlText)
		if m == nil {
			continue
		}
		candidate := ""
		for _, g := range m[1:] {
			if g != "" {
				candidate = g
				break
			}
		}
		if u, err := url.Parse(candidate); err == nil && u.Hostname() != "" && u.Hostname() != originalHost {
			return candidate
		}
	}
	return ""
}

func fetchPageContent(rawURL string) (string, error) {
	client, err := tlsclient.NewHttpClient(tlsclient.NewNoopLogger(),
		tlsclient.WithClientProfile(profiles.Safari_IOS_17_0),
		tlsclient.WithTimeoutSeconds(10),
	)
	if err != nil {
		return "", err
	}

	req, err := fhttp.NewRequest(fhttp.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "SearXNG-MCP/1.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var htmlText string
	if resp.StatusCode != fhttp.StatusOK {
		if resp.Header.Get("cf-mitigated") != "" && flareSolverrURL != "" {
			htmlText, err = fetchViaFlareSolverr(rawURL)
			if err != nil {
				return "", err
			}
		} else {
			return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}
	} else {
		ct := resp.Header.Get("Content-Type")
		if !strings.Contains(ct, "text/html") && !strings.Contains(ct, "text/plain") {
			return "", fmt.Errorf("unsupported content type: %s", ct)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		htmlText = string(body)
	}

	text := stripHTML(htmlText)
	if utf8.RuneCountInString(text) < 500 && flareSolverrURL != "" {
		htmlText, err = fetchViaFlareSolverr(rawURL)
		if err != nil {
			return "", err
		}
		text = stripHTML(htmlText)

		if utf8.RuneCountInString(text) < 500 {
			if sourceURL := extractSourceURL(htmlText, rawURL); sourceURL != "" {
				return fetchPageContent(sourceURL)
			}
		}
	}
	if utf8.RuneCountInString(text) > fetchMaxChars {
		runes := []rune(text)
		return fmt.Sprintf("%s\n\n[Truncated — %d total chars]", string(runes[:fetchMaxChars]), len(runes)), nil
	}
	return text, nil
}
