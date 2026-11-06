package extractor

import (
	"strings"
	"testing"

	"github.com/alexandrmotologa/brokenlens/internal/model"
)

func TestExtractHTML(t *testing.T) {
	htmlContent := `
<!DOCTYPE html>
<html>
<head>
  <title>Test Page Title</title>
  <link rel="canonical" href="https://example.com/canonical-test" />
  <link rel="stylesheet" href="/assets/style.css" />
  <link rel="preload" href="/fonts/font.woff2" as="font" />
  <link rel="icon" href="/favicon.ico" />
  <style>
    .hero { background-image: url('/images/hero-bg.jpg'); }
  </style>
</head>
<body>
  <div id="main-content">
    <h1 id="welcome-heading">Welcome</h1>
    <a href="/about#team" name="top-anchor">Our Team</a>
    <a href="https://external.com/partner">Partner</a>
    <img src="/images/logo.png" alt="Company Logo" srcset="/images/logo-1x.png 1x, /images/logo-2x.png 2x" />
    <picture>
      <source srcset="/images/hero-large.webp 1200w, /images/hero-small.webp 600w" />
      <img src="/images/hero-fallback.jpg" alt="Hero" />
    </picture>
    <div style="background: url(/images/inline-banner.png)"></div>
    <video src="/videos/intro.mp4" poster="/images/poster.jpg"></video>
    <script src="/js/bundle.js"></script>
    <script>console.log("inline script");</script>
  </div>
</body>
</html>
`

	result, err := ExtractHTML(strings.NewReader(htmlContent), "https://example.com")
	if err != nil {
		t.Fatalf("ExtractHTML failed: %v", err)
	}

	if result.PageTitle != "Test Page Title" {
		t.Errorf("Expected title 'Test Page Title', got %q", result.PageTitle)
	}

	if result.CanonicalURL != "https://example.com/canonical-test" {
		t.Errorf("Expected canonical URL 'https://example.com/canonical-test', got %q", result.CanonicalURL)
	}

	// Verify DOM Fragments
	expectedIDs := []string{"main-content", "welcome-heading", "top-anchor"}
	for _, id := range expectedIDs {
		if !result.DOMFragments[id] {
			t.Errorf("Expected DOM fragment %q was not found in result", id)
		}
	}

	// Verify extracted targets
	hasAsset := func(rawURL string, aType model.AssetType) bool {
		for _, tgt := range result.Targets {
			if tgt.RawURL == rawURL && tgt.AssetType == aType {
				return true
			}
		}
		return false
	}

	if !hasAsset("/assets/style.css", model.AssetStylesheet) {
		t.Error("Missing stylesheet target /assets/style.css")
	}

	if !hasAsset("/js/bundle.js", model.AssetScript) {
		t.Error("Missing script target /js/bundle.js")
	}

	if !hasAsset("/about#team", model.AssetLink) {
		t.Error("Missing link target /about#team")
	}

	if !hasAsset("/images/logo.png", model.AssetImage) {
		t.Error("Missing image target /images/logo.png")
	}

	if !hasAsset("/images/logo-2x.png", model.AssetImage) {
		t.Error("Missing srcset target /images/logo-2x.png")
	}

	if !hasAsset("/images/hero-bg.jpg", model.AssetCSS) {
		t.Error("Missing CSS url target /images/hero-bg.jpg")
	}

	if !hasAsset("/images/inline-banner.png", model.AssetCSS) {
		t.Error("Missing inline style CSS target /images/inline-banner.png")
	}
}
