# BrokenLens

High-throughput broken link, dead asset, and SEO mixed-content auditor in a single Go binary with an embedded React 19 interactive report studio and CI/CD failure gating.

---

## Overview

BrokenLens audits websites and web applications for broken hyperlinks (HTTP 404), missing images (including `srcset` candidates), broken scripts and stylesheets, missing in-page fragment anchors (`#id`), circular redirect loops, and mixed-content security violations (insecure HTTP resources embedded in HTTPS pages).

It runs as a standalone single-binary command-line tool, executes crawls at high concurrency (configurable from 1 to 250 parallel workers), generates machine-readable reports (JSON, CSV, JUnit XML, Markdown), and provides an embedded visual report studio served directly from memory.

---

## Key Capabilities

- **High-Concurrency Engine**: Asynchronous crawler written in Go with connection pooling, keep-alive transport, and DNS reuse.
- **Deep HTML Asset Extraction**: Inspects `<a>`, `<area>`, `<img>`, `srcset`, `<picture>`, `<script>`, `<link rel="stylesheet">`, `<link rel="preload">`, inline CSS `url(...)`, and `<video>/<audio>` tags.
- **In-Page Fragment Validation**: Verifies that `#section-name` fragments point to actual DOM `id` or `<a name>` elements on the target page.
- **Mixed Content Detection**: Flags insecure `http://` assets referenced on `https://` parent pages, categorizing them into Active (critical) and Passive (warning) violations.
- **Redirect Chain Tracer**: Traces multi-hop redirects (301, 302, 307, 308) and detects circular loops ($A \to B \to A$).
- **External Link Optimization**: Uses HEAD-first requests with automatic GET fallbacks to check external links without unnecessary bandwidth consumption.
- **CI/CD Quality Gating**: Returns exit code 1 with `--fail-on-broken` when broken items or critical issues are found, integrating directly with GitHub Actions, GitLab CI, and Jenkins.
- **Embedded Web Studio**: React 19 dashboard embedded inside the binary via `embed.FS` with real-time Server-Sent Events (SSE) updates and interactive search and filtering.

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

### CI/CD Pipeline Scan with Gating

```bash
./brokenlens https://example.com \
  --concurrency 50 \
  --max-depth 5 \
  --check-external \
  --fail-on-broken \
  --output report.json \
  --junit junit.xml
```

### Launch Interactive Report Studio

```bash
./brokenlens https://example.com --serve --port 8080
```

Open `http://127.0.0.1:8080` in your browser to view live crawl progress, search broken assets, inspect redirect chains, and export CSV/JSON reports.

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
| `--check-external` | | `true` | Verify HTTP status of external links |
| `--check-fragments` | | `true` | Verify in-page `#anchor` existence in target DOM |
| `--check-mixed-content` | | `true` | Detect HTTP assets on HTTPS pages |
| `--robots` | | `true` | Respect robots.txt exclusion rules |
| `--sitemap` | | `true` | Auto-discover and crawl URLs from sitemap.xml |
| `--max-redirects` | | `10` | Maximum redirect hops before reporting failure |
| `--fail-on-broken` | | `false` | Exit with code 1 if broken links or critical issues are found |
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

## GitHub Actions Example

```yaml
name: Broken Link Check

on:
  push:
    branches: [ main ]
  schedule:
    - cron: '0 3 * * *'

jobs:
  audit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run BrokenLens
        uses: alexandrmotologa/brokenlens@main
        with:
          url: 'https://example.com'
          concurrency: '30'
          max-depth: '5'
          fail-on-broken: 'true'
          output: 'report.json'
          junit: 'junit.xml'
```

---

## Documentation

- [Architecture & Internal Design](docs/ARCHITECTURE.md)
- [CI/CD Integration Guide](docs/CI_CD_INTEGRATION.md)
- [Master Plan & Roadmap](MASTER_PLAN.md)

---

## License

MIT License. See [LICENSE](LICENSE) for details.
