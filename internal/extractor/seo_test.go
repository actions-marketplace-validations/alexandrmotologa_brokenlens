package extractor

import (
	"testing"

	"github.com/alexandrmotologa/brokenlens/internal/model"
)

func TestAuditCanonical(t *testing.T) {
	// Canonical mismatch: HTTPS parent with HTTP canonical
	mismatch := AuditCanonical("https://example.com/blog", "http://example.com/blog", nil)
	if mismatch == nil || mismatch.Type != model.IssueCanonicalMismatch {
		t.Fatalf("Expected IssueCanonicalMismatch, got %v", mismatch)
	}

	// Broken canonical: target returned 404
	brokenRes := &model.CrawlResult{
		StatusCode: 404,
	}
	broken := AuditCanonical("https://example.com/blog", "https://example.com/canonical-target", brokenRes)
	if broken == nil || broken.Type != model.IssueBrokenCanonical {
		t.Fatalf("Expected IssueBrokenCanonical, got %v", broken)
	}
}

func TestAuditMetaRobots(t *testing.T) {
	issueMeta := AuditMetaRobots("https://example.com/private", "noindex, nofollow", "")
	if issueMeta == nil || issueMeta.Type != model.IssueNoIndexDirective {
		t.Fatalf("Expected IssueNoIndexDirective from meta tag, got %v", issueMeta)
	}

	issueHeader := AuditMetaRobots("https://example.com/private", "", "noindex")
	if issueHeader == nil || issueHeader.Type != model.IssueNoIndexDirective {
		t.Fatalf("Expected IssueNoIndexDirective from X-Robots-Tag, got %v", issueHeader)
	}
}

func TestDetectSoft404(t *testing.T) {
	// Soft 404 via title
	issueTitle := DetectSoft404("https://example.com/product/123", 200, "Product - 404 Not Found", []byte("Some content"))
	if issueTitle == nil || issueTitle.Type != model.IssueSoft404 {
		t.Fatalf("Expected IssueSoft404 via title, got %v", issueTitle)
	}

	// Normal valid page: should NOT trigger soft 404
	normal := DetectSoft404("https://example.com/about", 200, "About Our Company", []byte("Welcome to our website!"))
	if normal != nil {
		t.Fatalf("Did not expect soft 404 for normal page, got %v", normal)
	}
}
