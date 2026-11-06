package crawler

import (
	"net/url"
	"strings"
)

// NormalizeURL resolves a raw reference against a base URL and cleans it.
// Returns normalized URL (without fragment), the extracted fragment, and an error if unparseable or ignorable.
func NormalizeURL(rawRef string, baseStr string) (normalized string, fragment string, err error) {
	ref := strings.TrimSpace(rawRef)
	if ref == "" {
		return "", "", url.InvalidHostError("empty url")
	}

	lowerRef := strings.ToLower(ref)
	if strings.HasPrefix(lowerRef, "javascript:") ||
		strings.HasPrefix(lowerRef, "mailto:") ||
		strings.HasPrefix(lowerRef, "tel:") ||
		strings.HasPrefix(lowerRef, "data:") ||
		strings.HasPrefix(lowerRef, "sms:") ||
		strings.HasPrefix(lowerRef, "blob:") {
		return "", "", nil // Skipped schema
	}

	baseURL, err := url.Parse(baseStr)
	if err != nil {
		return "", "", err
	}

	parsedRef, err := url.Parse(ref)
	if err != nil {
		return "", "", err
	}

	resolved := baseURL.ResolveReference(parsedRef)

	// Clean fragment and save it
	fragment = resolved.Fragment
	resolved.Fragment = ""

	// Lowercase scheme and host
	resolved.Scheme = strings.ToLower(resolved.Scheme)
	resolved.Host = strings.ToLower(resolved.Host)

	// Clean standard default ports
	if resolved.Scheme == "http" && strings.HasSuffix(resolved.Host, ":80") {
		resolved.Host = strings.TrimSuffix(resolved.Host, ":80")
	} else if resolved.Scheme == "https" && strings.HasSuffix(resolved.Host, ":443") {
		resolved.Host = strings.TrimSuffix(resolved.Host, ":443")
	}

	// Normalize empty path to "/"
	if resolved.Path == "" {
		resolved.Path = "/"
	}

	return resolved.String(), fragment, nil
}

// IsSameDomain checks whether candidateURL has the exact same hostname as targetOrigin.
func IsSameDomain(candidateURL string, targetHost string) bool {
	parsed, err := url.Parse(candidateURL)
	if err != nil {
		return false
	}
	candHost := strings.ToLower(parsed.Host)
	targetLower := strings.ToLower(targetHost)

	// Strip port numbers for domain comparison
	if idx := strings.IndexByte(candHost, ':'); idx != -1 {
		candHost = candHost[:idx]
	}
	if idx := strings.IndexByte(targetLower, ':'); idx != -1 {
		targetLower = targetLower[:idx]
	}

	return candHost == targetLower
}
