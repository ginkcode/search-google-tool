package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type searxngResult struct {
	Title         string   `json:"title"`
	URL           string   `json:"url"`
	Content       string   `json:"content"`
	PublishedDate string   `json:"publishedDate,omitempty"`
	Engine        string   `json:"engine,omitempty"`
	Engines       []string `json:"engines,omitempty"`
}

type searxngResponse struct {
	Query   string          `json:"query"`
	Results []searxngResult `json:"results"`
	Answers []string        `json:"answers,omitempty"`
	Infoboxes []struct {
		Infobox string `json:"infobox"`
		Content string `json:"content"`
	} `json:"infoboxes,omitempty"`
	Suggestions []string `json:"suggestions,omitempty"`
}

type searchOptions struct {
	categories string
	language   string
	numResults int
	timeRange  string
}

func searchSearxng(query string, opts searchOptions) (*searxngResponse, error) {
	if opts.categories == "" {
		opts.categories = "general"
	}

	params := url.Values{
		"q":          {query},
		"format":     {"json"},
		"categories": {opts.categories},
	}
	if opts.language != "" {
		params.Set("language", opts.language)
	}
	if opts.timeRange != "" {
		params.Set("time_range", opts.timeRange)
	}

	req, err := http.NewRequest(http.MethodGet, searxngURL+"/search?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; SearXNG-MCP/1.0)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SearXNG returned %d: %s", resp.StatusCode, resp.Status)
	}

	var data searxngResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if opts.numResults > 0 && len(data.Results) > opts.numResults {
		data.Results = data.Results[:opts.numResults]
	}

	return &data, nil
}

func formatResults(data *searxngResponse) string {
	var sb strings.Builder

	if len(data.Answers) > 0 {
		fmt.Fprintf(&sb, "**Direct answer:** %s\n\n", data.Answers[0])
	}
	if len(data.Infoboxes) > 0 {
		fmt.Fprintf(&sb, "**%s:** %s\n\n", data.Infoboxes[0].Infobox, data.Infoboxes[0].Content)
	}

	if len(data.Results) == 0 {
		return "No results found."
	}

	for i, r := range data.Results {
		fmt.Fprintf(&sb, "%d. **%s**\n", i+1, r.Title)
		fmt.Fprintf(&sb, "   URL: %s\n", r.URL)
		if r.Content != "" {
			fmt.Fprintf(&sb, "   %s\n", r.Content)
		}
		if r.PublishedDate != "" {
			fmt.Fprintf(&sb, "   Published: %s\n", r.PublishedDate)
		}
		sb.WriteString("\n")
	}

	if len(data.Suggestions) > 0 {
		end := 5
		if len(data.Suggestions) < end {
			end = len(data.Suggestions)
		}
		fmt.Fprintf(&sb, "\n**Related searches:** %s", strings.Join(data.Suggestions[:end], ", "))
	}

	return sb.String()
}
