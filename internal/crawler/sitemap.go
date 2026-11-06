package crawler

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type urlSet struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Loc string `xml:"loc"`
}

type sitemapIndex struct {
	XMLName  xml.Name `xml:"sitemapindex"`
	Sitemaps []sitemapIndexEntry `xml:"sitemap"`
}

type sitemapIndexEntry struct {
	Loc string `xml:"loc"`
}

// IngestSitemaps fetches and extracts URLs from all specified or discovered sitemaps.
func IngestSitemaps(ctx context.Context, client *HTTPClient, targetOrigin string, additionalSitemaps []string) []string {
	var candidateSitemaps []string

	parsed, err := url.Parse(targetOrigin)
	if err == nil {
		candidateSitemaps = append(candidateSitemaps,
			fmt.Sprintf("%s://%s/sitemap.xml", parsed.Scheme, parsed.Host),
			fmt.Sprintf("%s://%s/sitemap_index.xml", parsed.Scheme, parsed.Host),
		)
	}

	candidateSitemaps = append(candidateSitemaps, additionalSitemaps...)

	seenSitemaps := make(map[string]bool)
	var discoveredURLs []string
	seenURLs := make(map[string]bool)

	for _, smURL := range candidateSitemaps {
		if seenSitemaps[smURL] {
			continue
		}
		seenSitemaps[smURL] = true

		urls := parseSitemapRecursive(ctx, client, smURL, seenSitemaps, 0)
		for _, u := range urls {
			if !seenURLs[u] {
				seenURLs[u] = true
				discoveredURLs = append(discoveredURLs, u)
			}
		}
	}

	return discoveredURLs
}

func parseSitemapRecursive(ctx context.Context, client *HTTPClient, sitemapURL string, seen map[string]bool, depth int) []string {
	if depth > 5 {
		return nil
	}

	resp := client.Fetch(ctx, sitemapURL, http.MethodGet, true)
	if resp.StatusCode != http.StatusOK || len(resp.Body) == 0 {
		return nil
	}

	body := resp.Body

	// Handle Gzip if .gz suffix or gzip magic bytes
	if strings.HasSuffix(strings.ToLower(sitemapURL), ".gz") || (len(body) > 2 && body[0] == 0x1f && body[1] == 0x8b) {
		gzReader, err := gzip.NewReader(bytes.NewReader(body))
		if err == nil {
			decompressed, err := io.ReadAll(gzReader)
			_ = gzReader.Close()
			if err == nil {
				body = decompressed
			}
		}
	}

	var urls []string

	// Check if this is a sitemapindex
	var idx sitemapIndex
	if err := xml.Unmarshal(body, &idx); err == nil && len(idx.Sitemaps) > 0 {
		for _, sm := range idx.Sitemaps {
			subURL := strings.TrimSpace(sm.Loc)
			if subURL != "" && !seen[subURL] {
				seen[subURL] = true
				subURLs := parseSitemapRecursive(ctx, client, subURL, seen, depth+1)
				urls = append(urls, subURLs...)
			}
		}
		return urls
	}

	// Try parsing standard urlset
	var us urlSet
	if err := xml.Unmarshal(body, &us); err == nil {
		for _, u := range us.URLs {
			loc := strings.TrimSpace(u.Loc)
			if loc != "" {
				urls = append(urls, loc)
			}
		}
	}

	return urls
}
