package crawler

import (
	"testing"
)

func TestParseRobotsTxt(t *testing.T) {
	robotsContent := `
User-agent: *
Disallow: /admin/
Disallow: /private/
Allow: /private/public-doc
Sitemap: https://example.com/sitemap.xml
Sitemap: https://example.com/sitemap-news.xml

User-agent: Googlebot
Disallow: /google-skip/
`

	policy := ParseRobotsTxt([]byte(robotsContent), "BrokenLens/1.0")

	if len(policy.Sitemaps) != 2 {
		t.Fatalf("Expected 2 sitemaps, got %d", len(policy.Sitemaps))
	}

	if policy.Sitemaps[0] != "https://example.com/sitemap.xml" {
		t.Errorf("Unexpected sitemap %s", policy.Sitemaps[0])
	}

	if policy.IsAllowed("/admin/users") {
		t.Error("Expected /admin/users to be disallowed")
	}

	if !policy.IsAllowed("/public/page") {
		t.Error("Expected /public/page to be allowed")
	}

	if !policy.IsAllowed("/private/public-doc") {
		t.Error("Expected /private/public-doc to be allowed via explicit Allow")
	}
}
