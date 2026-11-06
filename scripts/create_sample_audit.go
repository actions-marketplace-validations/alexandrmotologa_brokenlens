package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/alexandrmotologa/brokenlens/internal/config"
	"github.com/alexandrmotologa/brokenlens/internal/crawler"
	"github.com/alexandrmotologa/brokenlens/internal/model"
)

func main() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head>
  <title>Acme Cloud Platform — Real-Time Infrastructure</title>
  <link rel="canonical" href="http://acme.local/canonical-insecure" />
  <link rel="stylesheet" href="/assets/app.css" />
  <link rel="preload" href="/fonts/inter.woff2" as="font" />
  <meta name="robots" content="index, follow" />
  <meta property="og:image" content="/images/og-card-missing.jpg" />
</head>
<body>
  <nav class="navbar">
    <a href="/features" class="nav-link">Features</a>
    <a href="/pricing#enterprise-plan" class="nav-link">Pricing Plans</a>
    <a href="/docs#setup-guide" class="nav-link">Documentation</a>
    <a href="/blog" class="nav-link">Blog</a>
    <a href="/status-redirect" class="nav-link">System Status</a>
    <a href="/legacy-api-v1" class="btn-warning">Legacy API (404)</a>
  </nav>

  <main>
    <h1>Next-Gen Cloud Orchestration</h1>
    <p>Deploy containers at microsecond speed with automatic link validation.</p>
    <img src="/images/hero-banner.webp" alt="Cloud Architecture Overview" />
    <img src="/images/dead-partner-logo.png" alt="Missing Partner Logo" />
    <a href="/pricing#non-existent-tier" class="cta-broken">Special Promotion Tier (#missing)</a>
    <a href="/circular-1" class="hidden-link">Redirect Loop Test</a>
  </main>
</body>
</html>`)

		case "/features":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head><title>Features — Acme Cloud</title></head>
<body>
  <h1 id="core-features">Platform Features</h1>
  <a href="/docs#cli-usage">CLI Reference</a>
  <a href="/broken-download.zip">Download Agent SDK (404)</a>
  <img src="/assets/features-diagram.svg" alt="Architecture" />
</body>
</html>`)

		case "/pricing":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head><title>Pricing — Acme Cloud</title></head>
<body>
  <div id="pricing-root">
    <section id="enterprise-plan"><h2>Enterprise Tier</h2></section>
    <section id="starter-plan"><h2>Starter Tier</h2></section>
  </div>
</body>
</html>`)

		case "/docs":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head><title>Documentation — Acme Cloud</title></head>
<body>
  <h1 id="setup-guide">Setup Guide</h1>
  <div id="cli-usage">Usage instructions here...</div>
</body>
</html>`)

		case "/blog":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head><title>Engineering Blog — Acme Cloud</title></head>
<body>
  <h1>Engineering Updates</h1>
  <a href="/blog/post-1">Migrating to Go 1.27</a>
  <a href="/blog/post-soft-404">Deprecated Article</a>
</body>
</html>`)

		case "/blog/post-1":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head><title>Migrating to Go 1.27 — Acme Blog</title></head>
<body>
  <h1>Migrating to Go 1.27</h1>
  <p>Technical details of our recent migration.</p>
</body>
</html>`)

		case "/blog/post-soft-404":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			// Soft 404: Returns 200 OK, but title indicates 404
			fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head><title>Page Not Found - 404 Error</title></head>
<body>
  <h1>This article is no longer available</h1>
  <p>The requested blog post has been removed.</p>
</body>
</html>`)

		case "/assets/app.css":
			w.Header().Set("Content-Type", "text/css")
			fmt.Fprint(w, `body { font-family: sans-serif; background: url('/images/bg-grid.png'); }`)

		case "/assets/features-diagram.svg":
			w.Header().Set("Content-Type", "image/svg+xml")
			fmt.Fprint(w, `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100"><rect width="100" height="100" fill="#3b82f6"/></svg>`)

		case "/images/hero-banner.webp":
			w.Header().Set("Content-Type", "image/webp")
			w.Write([]byte("RIFF....WEBPVP8 "))

		case "/images/bg-grid.png":
			w.Header().Set("Content-Type", "image/png")
			w.Write([]byte("\x89PNG\r\n\x1a\n"))

		case "/fonts/inter.woff2":
			w.Header().Set("Content-Type", "font/woff2")
			w.Write([]byte("wOF2...."))

		// Redirect chain
		case "/status-redirect":
			http.Redirect(w, r, "/status-v2", http.StatusMovedPermanently)
		case "/status-v2":
			http.Redirect(w, r, "/status-live", http.StatusFound)
		case "/status-live":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `<h1>All Systems Operational</h1>`)

		// Circular redirect loop
		case "/circular-1":
			http.Redirect(w, r, "/circular-2", http.StatusMovedPermanently)
		case "/circular-2":
			http.Redirect(w, r, "/circular-1", http.StatusFound)

		// 404 Not Found targets
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	cfg := &config.Config{
		TargetURL:          ts.URL,
		Concurrency:        10,
		MaxDepth:           5,
		Timeout:            5 * time.Second,
		UserAgent:          "BrokenLens/1.0 (+https://github.com/alexandrmotologa/brokenlens)",
		CheckExternal:      false,
		CheckFragments:     true,
		CheckMixedContent:  true,
		CheckSEO:           true,
		FollowRobotsTxt:    false,
		HonorSitemap:       false,
		MaxRedirects:       5,
		FailOnBroken:       false,
		MaxBrokenThreshold: 0,
	}

	engine, err := crawler.NewEngine(cfg, nil)
	if err != nil {
		panic(err)
	}

	report, err := engine.Run(context.Background())
	if err != nil {
		panic(err)
	}

	// Make URLs clean for visual presentation
	report.Summary.TargetURL = "https://acme-cloud.internal"

	// Mark some issues as NEW for differential preview demonstration
	if len(report.Issues) > 2 {
		report.Issues[0].IsNew = true
		report.Issues[1].IsNew = true
	}

	report.Diff = &model.DiffReport{
		BaselineReportPath: "main-branch-report.json",
		NewIssues:          report.Issues[:2],
		ResolvedIssues: []model.AuditIssue{
			{
				ID:           "resolved-1",
				Type:         model.IssueBroken,
				Severity:     model.SeverityError,
				TargetURL:    "https://acme-cloud.internal/docs/old-v0-sdk",
				SourceURL:    "https://acme-cloud.internal/docs",
				StatusCode:   404,
				ErrorMessage: "Not Found",
				AssetType:    model.AssetLink,
			},
		},
		PersistedIssues: report.Issues[2:],
	}

	report.Summary.Diff = &model.DiffSummary{
		BaselineReportPath:   "main-branch-report.json",
		NewIssuesCount:       2,
		ResolvedIssuesCount:  1,
		PersistedIssuesCount: len(report.Issues) - 2,
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("docs/sample_report.json", data, 0644); err != nil {
		panic(err)
	}

	fmt.Printf("✓ Created authentic sample report with %d scanned URLs, %d issues, %d redirects.\n",
		report.Summary.TotalScanned, len(report.Issues), report.Summary.TotalRedirects)
}
