package crawler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexandrmotologa/brokenlens/internal/config"
	"github.com/alexandrmotologa/brokenlens/internal/model"
)

func TestEngineIntegration(t *testing.T) {
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<body>
  <div id="home-root">Home</div>
  <a href="/valid-page">Valid Page</a>
  <a href="/valid-page#team-section">Existing Anchor</a>
  <a href="/valid-page#missing-section">Missing Anchor</a>
  <a href="/broken-404">Dead Link</a>
  <img src="/dead-photo.png" alt="Missing Photo" />
  <a href="/loop-1">Redirect Loop</a>
</body>
</html>
`)
		case "/valid-page":
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<body>
  <h2 id="team-section">Our Team</h2>
  <p>Content here</p>
</body>
</html>
`)
		case "/broken-404":
			http.NotFound(w, r)

		case "/dead-photo.png":
			http.NotFound(w, r)

		case "/loop-1":
			http.Redirect(w, r, "/loop-2", http.StatusMovedPermanently)

		case "/loop-2":
			http.Redirect(w, r, "/loop-1", http.StatusFound)

		case "/robots.txt":
			http.NotFound(w, r)

		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	cfg := &config.Config{
		TargetURL:          ts.URL,
		Concurrency:        5,
		MaxDepth:           3,
		Timeout:            5 * time.Second,
		UserAgent:          "BrokenLens-Test/1.0",
		CheckExternal:      false,
		CheckFragments:     true,
		CheckMixedContent:  false,
		FollowRobotsTxt:    false,
		HonorSitemap:       false,
		MaxRedirects:       5,
		FailOnBroken:       true,
		MaxBrokenThreshold: 0,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Config validation failed: %v", err)
	}

	engine, err := NewEngine(cfg, nil)
	if err != nil {
		t.Fatalf("NewEngine failed: %v", err)
	}

	report, err := engine.Run(context.Background())
	if err != nil {
		t.Fatalf("engine.Run failed: %v", err)
	}

	if report == nil {
		t.Fatal("Expected report, got nil")
	}

	// Verify Summary
	if report.Summary.TotalScanned < 4 {
		t.Errorf("Expected at least 4 scanned URLs, got %d", report.Summary.TotalScanned)
	}

	// Verify Broken items count (404 link + 404 image + loop)
	if report.Summary.TotalBroken < 2 {
		t.Errorf("Expected at least 2 broken items, got %d", report.Summary.TotalBroken)
	}

	// Verify Issues
	found404Link := false
	found404Img := false
	foundMissingAnchor := false
	foundLoop := false

	for _, issue := range report.Issues {
		switch issue.Type {
		case model.IssueBroken:
			if issue.AssetType == model.AssetLink && issue.StatusCode == 404 {
				found404Link = true
			}
			if issue.AssetType == model.AssetImage && issue.StatusCode == 404 {
				found404Img = true
			}
		case model.IssueAnchorNotFound:
			if issue.TargetURL == ts.URL+"/valid-page#missing-section" {
				foundMissingAnchor = true
			}
		case model.IssueRedirectLoop:
			foundLoop = true
		}
	}

	if !found404Link {
		t.Error("Failed to detect 404 broken link /broken-404")
	}
	if !found404Img {
		t.Error("Failed to detect 404 broken image /dead-photo.png")
	}
	if !foundMissingAnchor {
		t.Error("Failed to detect missing DOM anchor #missing-section")
	}
	if !foundLoop {
		t.Error("Failed to detect circular redirect loop /loop-1 <-> /loop-2")
	}

	// Verify Health Score was penalized
	if report.Summary.HealthScore >= 100.0 {
		t.Errorf("Health score should have been penalized for broken items, got %f", report.Summary.HealthScore)
	}
}
