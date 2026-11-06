# `brokenlens` - Architecture & Engineering Master Plan

> **High-Throughput Broken Link, Dead Asset & SEO Mixed-Content Auditor in a Single Binary**  
> *Screaming Frog and Ahrefs Site Audit Alternative with CI/CD Gating (Go + React 19)*

---

## 1. Executive Summary & Problem Space

### The Problem
Broken links (HTTP 404), missing images, broken script bundles, infinite redirect loops, and mixed-content warnings (insecure HTTP assets loaded on HTTPS pages) hurt user experience, degrade search engine rankings, and break client-side applications.

Current tools have clear constraints:
1. **Screaming Frog SEO Spider**: Desktop Java application requiring graphical environments. The free tier caps crawling at 500 URLs, and paid licenses cost 209 GBP per year. It cannot be easily integrated headlessly into modern GitHub Actions or GitLab CI pipelines.
2. **Cloud SEO Scanners (Ahrefs, Semrush)**: Paid subscription platforms (99 to 499+ USD per month) that cannot crawl local development environments, staging subdomains, or private internal networks behind firewalls or VPNs.

### The Solution: `brokenlens`
`brokenlens` is a single-binary, high-concurrency web quality auditor written in Go with an embedded React 19 interactive report studio. It crawls websites at speeds exceeding 500 URLs per second, validating internal and external hyperlinks, image assets (including `srcset`), scripts, stylesheets, inline CSS `url()` assets, missing in-page fragment anchors (`#id`), redirect chains, and mixed-content violations.

With native CI/CD gating (`--fail-on-broken`), `brokenlens` terminates with an exit code of `1` when critical broken assets are detected, preventing broken pages from reaching production.

---

## 2. System Architecture & Crawl Engine

```
[ Target Website / Staging URL ]
               │
               │  HTTP / HTTPS Parallel GET & HEAD Requests
               ▼
┌────────────────────────────────────────────────────────────────────────┐
│                        brokenlens Audit Engine                         │
│                                                                        │
│  ┌───────────────────────┐         ┌────────────────────────────────┐  │
│  │ Robots.txt & Sitemap  │ ──────► │ URL Normalizer & Scope Filter  │  │
│  │ Ingestion Engine      │         │ (Internal vs External Domain)  │  │
│  └───────────────────────┘         └────────────────┬───────────────┘  │
│                                                     │                  │
│                                                     ▼                  │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ Concurrency Worker Pool (Configurable 10 to 250 Workers)         │  │
│  └──────────────────────────────────┬───────────────────────────────┘  │
│                                     │                                  │
│                                     ▼                  │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ Streaming HTML Tokenizer & Asset Extractor                       │  │
│  │ - Hyperlinks: <a href>, <area href>, canonical links             │  │
│  │ - Visual Assets: <img src>, <img srcset>, <picture>, <video>     │  │
│  │ - Scripts & Styles: <script src>, <link rel="stylesheet">        │  │
│  │ - In-Page Anchors: Checks #target exists in DOM id attributes    │  │
│  │ - Mixed Content Detector: HTTP assets on HTTPS pages             │  │
│  │ - CSS URL Parser: url(...) in style tags and attributes          │  │
│  └──────────────────────────────────┬───────────────────────────────┘  │
│                                     │                                  │
│                                     ▼                                  │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ HTTP Status Code Evaluator & Redirect Loop Detector              │  │
│  │ - Tracks 301/302/307/308 redirect hops (max redirect limit)      │  │
│  │ - Flags 400 Bad Request, 404 Not Found, 410 Gone, 500 Server Err │  │
│  │ - External Link Optimizer: HEAD request with GET fallback        │  │
│  └──────────────────────────────────┬───────────────────────────────┘  │
│                                     │                                  │
│         ┌───────────────────────────┴──────────────────────────┐       │
│         ▼                                                      ▼       │
│  ┌───────────────────────────────┐           ┌──────────────────────┐  │
│  │ CI/CD Gating & Exit Codes     │           │ Embedded React 19 UI │  │
│  │ (--fail-on-broken -> exit 1)  │           │ (Interactive Report) │  │
│  │ JSON / CSV / JUnit / Markdown │           │ Real-time SSE Stream │  │
│  └───────────────────────────────┘           └──────────────────────┘  │
└────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Technology Stack

| Layer | Component | Specification |
| :--- | :--- | :--- |
| **Language & Runtime** | Go 1.23+ (Go 1.27) | Lock-free worker channels, ultra-low GC overhead, fast parallel network I/O |
| **HTTP Client** | Custom `net/http` Transport | Connection pooling, TLS keep-alives, custom User-Agent, configurable timeout |
| **HTML Parsing Engine** | `golang.org/x/net/html` | High-speed streaming tokenizer parsing DOM nodes without full tree overhead |
| **State & Deduplication** | Thread-Safe Set & Maps | Fast concurrent visited set with URL canonicalization |
| **Export Formats** | Multi-Format Reporters | JSON, CSV, JUnit XML, Markdown step summary, and ANSI terminal table |
| **Embedded Web UI** | React 19 + TypeScript + Vite | Built as a single bundle and served via Go `embed.FS` |
| **UI Styling** | Tailwind CSS + Lucide Icons | Clean modern dark-mode developer cockpit |

---

## 4. Repository & Directory Structure

```
B:\workgit\brokenlens\
├── cmd\
│   └── brokenlens\
│       └── main.go                 # CLI entrypoint, flag parsing, command routing
├── internal\
│   ├── config\
│   │   └── config.go               # Flags, defaults, validation
│   ├── model\
│   │   └── types.go                # Core data models, crawl results, issues
│   ├── crawler\
│   │   ├── engine.go               # Master crawl coordinator and worker dispatcher
│   │   ├── worker.go               # Network fetcher, rate limiting, and response handling
│   │   ├── client.go               # Custom HTTP client, TLS config, and redirect handler
│   │   ├── robots.go               # Robots.txt adherence and parser
│   │   └── sitemap.go              # XML Sitemap and Sitemap Index parser
│   ├── extractor\
│   │   ├── html.go                 # Streaming HTML tokenizer extracting all asset tags
│   │   ├── anchors.go              # In-page #hash anchor tag verification
│   │   ├── css.go                  # CSS url(...) asset extractor
│   │   └── mixed_content.go        # HTTP-over-HTTPS asset violation detection
│   ├── analyzer\
│   │   ├── redirects.go            # Redirect chain tracer and circular loop detector
│   │   ├── status.go               # HTTP status classifier (2xx, 3xx, 4xx, 5xx)
│   │   └── summary.go              # Aggregate statistics and health score calculator
│   ├── reporter\
│   │   ├── terminal.go             # ANSI terminal spinner and tabular summary
│   │   ├── json.go                 # Machine-readable JSON output
│   │   ├── csv.go                  # Spreadsheet CSV exporter
│   │   ├── junit.go                # CI/CD test runner JUnit XML output
│   │   └── markdown.go             # GitHub step summary generator
│   └── server\
│       ├── server.go               # Embedded HTTP server serving React 19 report UI
│       └── api.go                  # REST and SSE API endpoints (/api/scan, /api/report, /api/events)
├── web\
│   ├── src\
│   │   ├── components\
│   │   │   ├── Header.tsx          # Top navigation bar and controls
│   │   │   ├── SummaryCards.tsx    # High-level metrics: Scanned, broken, mixed content, health score
│   │   │   ├── IssuesTable.tsx     # Filterable and searchable table of broken links and assets
│   │   │   ├── StatusChart.tsx     # Status code distribution breakdown
│   │   │   ├── RedirectTree.tsx    # Multi-hop visual redirect chain explorer
│   │   │   ├── AnchorList.tsx      # Missing in-page DOM target inspector
│   │   │   └── MixedContentList.tsx# Active and passive mixed-content alerts
│   │   ├── types.ts                # TypeScript interface definitions mirroring Go models
│   │   ├── App.tsx                 # Root audit report studio
│   │   ├── main.tsx                # React 19 entrypoint
│   │   └── index.css               # Tailwind CSS declarations
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
├── docs\                           # Architecture guides and CLI manuals
├── action.yml                      # GitHub Actions composite action definition
├── Makefile                        # Multi-platform build, test, and release targets
├── Dockerfile                      # Scratch container image (< 25MB)
├── README.md                       # Documentation, usage examples, CI/CD recipes
└── MASTER_PLAN.md                  # This master specification
```

---

## 5. Inspection Matrix & CLI Specification

### Asset Inspection Capabilities

| Asset Category | Target HTML Elements & Attributes | Verification Criteria |
| :--- | :--- | :--- |
| **Hyperlinks** | `<a href="...">`, `<area href="...">` | Resolves to HTTP 200/300; flags 404, 500, network timeouts |
| **In-Page Fragments** | `<a href="/page#section-id">` | Fetches target page and verifies `<* id="section-id">` or `<a name="...">` exists in DOM |
| **Images & Graphics** | `<img src>`, `<img srcset>`, `<picture>`, `<source srcset>` | Image URL returns 200 OK and valid image MIME type |
| **Scripts & Stylesheets** | `<script src="...">`, `<link rel="stylesheet">` | Core rendering assets resolve without 404 or MIME mismatch |
| **Preload & Icons** | `<link rel="preload">`, `<link rel="icon">` | Linked resource resolves to 200 OK |
| **Embedded Media** | `<video src>`, `<audio src>`, `poster="..."` | Media files resolve to 200 OK |
| **CSS Backgrounds** | `url(...)` in style tags and attributes | Stylesheet assets resolve to 200 OK |
| **Security / Mixed Content**| All tags referenced with `http://` on `https://` | Flags insecure passive (images/audio) and active (scripts/styles) content |
| **Redirect Chains** | HTTP 301, 302, 307, 308 | Detects chains greater than 3 hops and circular redirect loops (A to B to A) |

### CLI Command Options

```bash
# Basic scan
brokenlens https://example.com

# Deep concurrent scan with CI/CD failure gating
brokenlens https://example.com \
  --concurrency 50 \
  --max-depth 5 \
  --check-external \
  --fail-on-broken \
  --output report.json

# Launch embedded visual report server
brokenlens https://example.com --serve --port 8080

# Standalone report viewer from previously generated JSON
brokenlens serve --report report.json --port 8080
```

---

## 6. Implementation Roadmap

### Phase 1: Core Scaffolding, Models & Go HTTP Client
- [x] Create public repository on GitHub (`alexandrmotologa/brokenlens`).
- [x] Couple repository metadata: Description and 13 targeted topics across 4 dimensions.
- [x] Register project in MTLG Portfolio ecosystem (`projects.json` in `mtlg-site` and `mtlglabs-space`).
- [x] Initialize Go module `github.com/alexandrmotologa/brokenlens`.
- [x] Implement data models (`internal/model/types.go`): URL targets, crawl issues, redirect hops, summary statistics.
- [x] Implement configuration module (`internal/config/config.go`) with flag parsing and validation.
- [x] Build custom HTTP transport client with keep-alive pooling, TLS configuration, and redirect tracking.

### Phase 2: Async Crawler Engine & Extractors
- [x] Implement URL normalizer, deduplication set, and domain scope filter (internal vs external).
- [x] Implement streaming HTML tokenizer (`internal/extractor/html.go`) for links, images, scripts, stylesheets, and media.
- [x] Implement in-page fragment `#hash` target validator (`internal/extractor/anchors.go`).
- [x] Implement active vs passive mixed content detector (`internal/extractor/mixed_content.go`).
- [x] Implement CSS `url(...)` asset parser (`internal/extractor/css.go`).
- [x] Implement robots.txt parser and XML Sitemap ingestion engine (`internal/crawler/robots.go`, `sitemap.go`).
- [x] Implement concurrent worker pool and coordinator (`internal/crawler/engine.go`, `worker.go`).

### Phase 3: Analyzers, Reporters & CI/CD Gating
- [x] Implement redirect chain analyzer and circular loop detector.
- [x] Implement status classifier and link health score calculator.
- [x] Implement live CLI terminal progress output and summary table (`internal/reporter/terminal.go`).
- [x] Implement JSON, CSV, JUnit XML, and Markdown step summary exporters (`internal/reporter/`).
- [x] Implement `--fail-on-broken` exit-code gating logic for CI/CD pipelines.
- [x] Create GitHub Actions composite action (`action.yml`) and sample workflow.

### Phase 4: Embedded React 19 Visual Studio
- [x] Scaffold Vite + React 19 + Tailwind CSS frontend in `web/`.
- [x] Build high-level metrics cards (Scanned, Broken, Dead Assets, Mixed Content, Health Score).
- [x] Build filterable and searchable issues table with pagination and sorting.
- [x] Build status code distribution charts and redirect chain visualizer.
- [x] Build in-page anchor diagnostics and mixed content security alerts.
- [x] Implement HTTP server in Go (`internal/server/`) serving the embedded React application and JSON/SSE APIs.
- [x] Enable single-binary packaging with Go `embed.FS`.

### Phase 5: Verification, Benchmarks & Documentation
- [x] Write unit tests and integration tests for HTML extraction, redirect loop detection, and anchor matching.
- [x] Run live end-to-end audit against realistic web test fixtures and local HTTP test servers.
- [x] Write factual, humanized documentation in `README.md` and `docs/`.
- [x] Add Makefile and Dockerfile.
- [x] Push clean codebase to GitHub.
