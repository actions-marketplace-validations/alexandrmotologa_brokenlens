package extractor

import (
	"regexp"
	"strings"
)

var cssURLRegex = regexp.MustCompile(`(?i)url\(\s*['"]?([^'")]+)['"]?\s*\)`)

// ExtractCSSURLs extracts referenced asset URLs from CSS stylesheet text or inline styles.
func ExtractCSSURLs(cssContent string) []string {
	matches := cssURLRegex.FindAllStringSubmatch(cssContent, -1)
	var urls []string
	seen := make(map[string]bool)

	for _, m := range matches {
		if len(m) > 1 {
			u := strings.TrimSpace(m[1])
			// Ignore data: and empty URIs
			if u == "" || strings.HasPrefix(strings.ToLower(u), "data:") || strings.HasPrefix(strings.ToLower(u), "javascript:") {
				continue
			}
			if !seen[u] {
				seen[u] = true
				urls = append(urls, u)
			}
		}
	}

	return urls
}
