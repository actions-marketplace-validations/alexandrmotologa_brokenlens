# BrokenLens v1.0.0 - High-Throughput Broken Link, Dead Asset & SEO Auditor

We are pleased to introduce **BrokenLens v1.0.0**, an ultra-fast web quality auditor, broken link detector, and SEO mixed-content scanner built in Go with an embedded React 19 interactive report studio.

### Key Capabilities

- **Ultra-High Concurrency Engine**: Asynchronous crawler handling 500+ URLs/second with keep-alive connection pooling and DNS reuse.
- **Deep Asset Extraction**: Validates `<a>`, `<area>`, `<img>`, `srcset`, `<picture>`, `<script>`, `<link rel="stylesheet">`, `<link rel="preload">`, inline CSS `url(...)`, and multimedia tags.
- **HTML Snippet Capture**: Records the exact HTML element for every link and asset, pinpointing broken items directly in your source code.
- **In-Page Fragment Anchor Verification**: Checks that `#section-name` fragments point to real DOM elements (`id` or `<a name>`) on target pages.
- **Technical SEO Auditing**:
  - Validates `<link rel="canonical">` targets for 404s and HTTP-over-HTTPS scheme mismatches.
  - Flags `noindex` and `nofollow` directives preventing search indexing.
  - Audits `<meta property="og:image">` and `twitter:image` preview assets.
  - Detects soft 404s (HTTP 200 with "404 Not Found" title or body) and internal orphan pages.
- **Mixed Content Security Audits**: Flags insecure HTTP assets on HTTPS pages, classifying them as Active (critical) or Passive (warning).
- **Redirect Chain Tracer**: Traces multi-hop redirects (301, 302, 307, 308) and detects circular loops ($A \to B \to A$).
- **CI/CD Quality Gating**:
  - Automated failure with exit code 1 (`--fail-on-broken`).
  - Differential PR gating against baseline reports (`--diff baseline.json` + `--fail-on-new-broken`).
- **Protected Staging Audits**: Supports HTTP Basic Auth (`--auth`), custom headers (`--header`), and session cookies (`--cookie`).
- **Embedded Web Cockpit**: React 19 visual report studio embedded inside the binary via Go `embed.FS` with real-time SSE streaming.
- **Multi-Format Reports**: JSON, CSV, JUnit XML, Markdown, and native `$GITHUB_STEP_SUMMARY` export.

### GitHub Actions Quickstart

```yaml
- name: Run BrokenLens Audit
  id: brokenlens
  uses: alexandrmotologa/brokenlens@v1
  with:
    url: 'https://staging.example.com'
    concurrency: '30'
    max-depth: '5'
    fail-on-broken: 'true'
    check-external: 'true'
    check-seo: 'true'
    output: 'brokenlens-report.json'
    junit: 'brokenlens-junit.xml'
```

### Action Outputs

- `total-scanned`: Total URLs crawled and scanned.
- `total-broken`: Count of broken links and missing assets.
- `health-score`: Overall site health score (0.0 to 100.0).
- `total-redirects`: Number of redirect chains detected.
- `total-mixed-content`: Number of mixed content violations detected.
- `report-path`: Output path of the generated JSON report.
- `junit-path`: Output path of the generated JUnit XML file.
