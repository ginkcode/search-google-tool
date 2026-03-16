package main

import (
	"fmt"
	"regexp"
	"strings"

	fhttp "github.com/bogdanfinn/fhttp"
	tlsclient "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

var (
	reScript = regexp.MustCompile(`(?si)<script.*?</script>`)
	reStyle  = regexp.MustCompile(`(?si)<style.*?</style>`)
	reTag    = regexp.MustCompile(`<[^>]+>`)
	reSpaces = regexp.MustCompile(`\s{2,}`)

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

	if resp.StatusCode != fhttp.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") && !strings.Contains(ct, "text/plain") {
		return "", fmt.Errorf("unsupported content type: %s", ct)
	}

	var sb strings.Builder
	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			sb.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}

	text := stripHTML(sb.String())
	if len(text) > fetchMaxChars {
		return fmt.Sprintf("%s\n\n[Truncated — %d total chars]", text[:fetchMaxChars], len(text)), nil
	}
	return text, nil
}
