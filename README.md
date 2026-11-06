<p align="center">
  <img src="docs/images/logo.png?raw=true" alt="BrokenLens Logo" width="130" style="border-radius: 24px;" />
</p>

<h1 align="center">BrokenLens</h1>

<p align="center">
  <strong>High-Throughput Broken Link, Dead Asset &amp; SEO Mixed-Content Auditor in a Single Binary</strong>
</p>

<p align="center">
  <img src="docs/images/brokenlens_demo.gif?raw=true" alt="BrokenLens Live Audit Demo" width="100%" style="border-radius: 12px; border: 1px solid #27272a;" />
</p>

---

## Overview

BrokenLens audits websites and web applications for broken hyperlinks (HTTP 404), missing images (including `srcset` candidates), broken scripts and stylesheets, missing in-page fragment anchors (`#id`), circular redirect loops, and mixed-content security violations (insecure HTTP resources embedded in HTTPS pages).

It runs as a standalone single-binary command-line tool, executes crawls at high concurrency (configurable from 1 to 250 parallel workers), generates machine-readable reports (JSON, CSV, JUnit XML, Markdown), and provides an embedded visual report studio served directly from memory.

---

## Key Capabilities

- **High-Concurrency Engine**: Asynchronous crawler written in Go with connection pooling, keep-alive transport, and DNS reuse.
- **Deep HTML Asset Extraction**: Inspects `<a>`, `<area>`, `<img>`, `srcset`, `<picture>`, `<script>`, `<link rel="stylesheet">`, `<link rel="preload">`, inline CSS `url(...)`, and `<video>/<audio>` tags.
- **HTML Snippet Capture**: Records the exact HTML element string for every discovered link or asset, pinpointing broken items in source templates.
- **In-Page Fragment Validation**: Verifies that `#section-name` fragments point to actual DOM `id` or `<a name>` elements on the target page.
- **Technical SEO Auditing**:
  - **Canonical Consistency**: Validates `<link rel="canonical">` targets for 404s and HTTP-over-HTTPS scheme mismatches.
  - **Meta Robots & X-Robots-Tag**: Flags `noindex` and `nofollow` directives preventing page discovery.
  - **Social Media Cards**: Audits `<meta property="og:image">` and `twitter:image` preview assets.
  - **Soft 404 Detection**: Flags pages returning HTTP 200 OK with "404 Not Found" page titles or empty error bodies.
  - **Orphan Page Detection**: Identifies internal pages with zero inbound links.
- **Mixed Content Detection**: Flags insecure `http://` assets referenced on `https://` parent pages, categorizing them into Active (critical) and Passive (warning) violations.
- **Redirect Chain Tracer**: Traces multi-hop redirects (301, 302, 307, 308) and detects circular loops ($A \to B \to A$).
- **Differential PR Gating**: Compares current scans against baseline reports (`--diff baseline.json`) with `--fail-on-new-broken` to fail CI only when a pull request introduces new regressions.
- **Protected Staging Audits**: Supports HTTP Basic Auth (`--auth`), custom headers (`--header`), and session cookies (`--cookie`).
- **Resilient Rate Limiting**: Automatically handles HTTP 429 and 503 responses with `Retry-After` backoff.
- **Embedded Web Studio**: React 19 dashboard embedded inside the binary via `embed.FS` with real-time Server-Sent Events (SSE) updates, interactive filtering, and code snippet inspection.

---

## Visual Studio & Responsive Cockpit

BrokenLens provides an embedded dark-mode cockpit built with React 19 and Tailwind CSS, responsive across desktop, tablet, and mobile viewports:

<p align="center">
  <img src="docs/images/brokenlens_dashboard.png?raw=true" alt="BrokenLens Desktop Dashboard" width="100%" style="border-radius: 8px; border: 1px solid #27272a; margin-bottom: 16px;" />
</p>

### Responsive Multi-Device Adaptability

| Desktop Redirect Inspector | Mobile Responsive Cockpit (390px) |
| :---: | :---: |
| <img src="docs/images/brokenlens_redirects.png?raw=true" alt="Redirect Chains View" width="560" style="border-radius: 8px; border: 1px solid #27272a;" /> | <img src="docs/images/brokenlens_mobile.png?raw=true" alt="Mobile View" width="280" style="border-radius: 8px; border: 1px solid #27272a;" /> |

---

## Installation

### From Source

```bash
git clone https://github.com/alexandrmotologa/brokenlens.git
cd brokenlens
make build
```

This compiles the React 19 frontend into `web/dist` and produces the `brokenlens` binary in the current directory.

---

## Quick Start

### Basic Scan

```bash
./brokenlens https://example.com
```

### Staging Scan with Authentication & Custom Headers

```bash
./brokenlens https://staging.example.com \
  --auth "admin:secretPass" \
  --header "X-Staging-Access: allowed" \
  --cookie "session_id=xyz"
```

### CI/CD Differential PR Gating

Compare the current pull request branch against a baseline report from `main`:

```bash
./brokenlens https://pr-123.preview.example.com \
  --diff baseline-main.json \
  --fail-on-new-broken \
  --output report.json
```

The process exits with code 1 only if the pull request introduces new broken items, ignoring pre-existing legacy issues.

### Launch Interactive Report Studio

```bash
./brokenlens https://example.com --serve --port 8080
```

Open `http://127.0.0.1:8080` in your browser to view live crawl progress, search broken assets, inspect redirect chains, view HTML snippets, and export CSV/JSON reports.

### Inspect Saved Report in Web Studio

```bash
./brokenlens serve --report report.json --port 8080
```

---

## CLI Options

| Flag | Shorthand | Default | Description |
| :--- | :--- | :--- | :--- |
| `--concurrency` | `-c` | `25` | Number of concurrent worker routines (1 to 250) |
| `--max-depth` | `-d` | `10` | Maximum link crawl depth (0 = unlimited) |
| `--max-urls` | | `0` | Maximum number of URLs to crawl (0 = unlimited) |
| `--timeout` | | `10` | Per-request timeout in seconds |
| `--auth` | | `""` | HTTP Basic Auth credentials (`user:password`) |
| `--header` | | `""` | Custom HTTP headers (`Key: Value`, multiple separated by `\|\|`) |
| `--cookie` | | `""` | Custom Cookie header string |
| `--diff` | | `""` | Path to baseline report JSON file for differential comparison |
| `--fail-on-broken` | | `false` | Exit with code 1 if broken links or critical issues are found |
| `--fail-on-new-broken` | | `false` | Exit with code 1 only if new broken items appear relative to baseline |
| `--check-seo` | | `true` | Audit canonical tags, meta robots, social images, and soft 404s |
| `--check-external` | | `true` | Verify HTTP status of external links |
| `--check-fragments` | | `true` | Verify in-page `#anchor` existence in target DOM |
| `--check-mixed-content` | | `true` | Detect HTTP assets on HTTPS pages |
| `--robots` | | `true` | Respect robots.txt exclusion rules |
| `--sitemap` | | `true` | Auto-discover and crawl URLs from sitemap.xml |
| `--max-redirects` | | `10` | Maximum redirect hops before reporting failure |
| `--max-broken` | | `0` | Maximum allowed broken items before exiting with failure |
| `--output` | `-o` | `""` | Export report file path (.json, .csv) |
| `--junit` | | `""` | Export JUnit XML report file path |
| `--markdown` | | `""` | Export Markdown step summary file path |
| `--serve` | | `false` | Start embedded React visual report studio |
| `--port` | | `8080` | Port for the report studio server |
| `--host` | | `127.0.0.1` | Host address for the report studio server |
| `--report` | | `""` | Load existing report JSON file for visual server viewer |
| `--delay` | | `0` | Delay in milliseconds between requests per worker |
| `--quiet` | `-q` | `false` | Suppress terminal live progress output |

---

## GitHub Actions (Marketplace)

BrokenLens is available directly as a GitHub Action on GitHub Marketplace:

```yaml
name: Pull Request Link & SEO Audit

on: [pull_request]

jobs:
  audit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Download Baseline Report
        run: curl -s https://example.com/reports/main.json -o baseline.json || echo "{}" > baseline.json

      - name: Run BrokenLens Differential Gate
        id: brokenlens
        uses: alexandrmotologa/brokenlens@v1
        with:
          url: 'https://staging.example.com'
          concurrency: '30'
          max-depth: '5'
          fail-on-broken: 'false'
          fail-on-new-broken: 'true'
          diff: 'baseline.json'
          output: 'pr-report.json'
          junit: 'junit.xml'

      - name: Verify Results
        run: |
          echo "Health Score: ${{ steps.brokenlens.outputs.health-score }}/100"
          echo "Total Scanned: ${{ steps.brokenlens.outputs.total-scanned }}"
          echo "New Broken Items: ${{ steps.brokenlens.outputs.total-broken }}"
```

---

## Documentation

- [Architecture & Internal Design](docs/ARCHITECTURE.md)
- [CI/CD Integration Guide](docs/CI_CD_INTEGRATION.md)
- [Master Plan & Roadmap](MASTER_PLAN.md)

---

## License

MIT License. See [LICENSE](LICENSE) for details.
