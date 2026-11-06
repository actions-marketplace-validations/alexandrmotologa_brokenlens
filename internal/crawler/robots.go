package crawler

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/alexandrmotologa/brokenlens/internal/config"
)

// RobotsPolicy stores parsed robots.txt rules and declared sitemaps.
type RobotsPolicy struct {
	DisallowedPaths []string
	AllowedPaths    []string
	Sitemaps        []string
}

// FetchRobotsTxt fetches and parses the robots.txt file for a target origin.
func FetchRobotsTxt(ctx context.Context, client *HTTPClient, targetURL string, cfg *config.Config) *RobotsPolicy {
	policy := &RobotsPolicy{}

	parsed, err := url.Parse(targetURL)
	if err != nil {
		return policy
	}

	robotsURL := fmt.Sprintf("%s://%s/robots.txt", parsed.Scheme, parsed.Host)
	resp := client.Fetch(ctx, robotsURL, http.MethodGet, true)

	if resp.StatusCode != http.StatusOK || len(resp.Body) == 0 {
		return policy
	}

	return ParseRobotsTxt(resp.Body, cfg.UserAgent)
}

// ParseRobotsTxt parses raw robots.txt bytes according to the Robots Exclusion Standard.
func ParseRobotsTxt(data []byte, userAgent string) *RobotsPolicy {
	policy := &RobotsPolicy{}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	uaMatches := false
	uaActive := false

	uaClean := strings.ToLower(userAgent)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Strip comments
		if idx := strings.IndexByte(line, '#'); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		directive := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch directive {
		case "sitemap":
			if value != "" {
				policy.Sitemaps = append(policy.Sitemaps, value)
			}

		case "user-agent":
			agent := strings.ToLower(value)
			if agent == "*" {
				uaMatches = true
				uaActive = true
			} else if strings.Contains(uaClean, agent) || agent == "brokenlens" {
				uaMatches = true
				uaActive = true
			} else {
				uaActive = false
			}

		case "disallow":
			if uaActive && value != "" {
				policy.DisallowedPaths = append(policy.DisallowedPaths, value)
			}

		case "allow":
			if uaActive && value != "" {
				policy.AllowedPaths = append(policy.AllowedPaths, value)
			}
		}
	}

	_ = uaMatches
	return policy
}

// IsAllowed evaluates whether a URL path is permissible to crawl under this policy.
func (r *RobotsPolicy) IsAllowed(rawPath string) bool {
	if len(r.DisallowedPaths) == 0 {
		return true
	}

	// Check explicit allows first
	for _, allowed := range r.AllowedPaths {
		if strings.HasPrefix(rawPath, allowed) {
			return true
		}
	}

	// Check disallows
	for _, disallowed := range r.DisallowedPaths {
		if disallowed == "/" {
			return false
		}
		if strings.HasPrefix(rawPath, disallowed) {
			return false
		}
	}

	return true
}
