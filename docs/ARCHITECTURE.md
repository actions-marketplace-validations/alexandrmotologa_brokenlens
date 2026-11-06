# BrokenLens Architecture & Internal Design

BrokenLens is a high-concurrency web quality auditor packaged as a self-contained Go executable with an embedded React 19 visual report studio. It evaluates hyperlinks, visual assets, stylesheet references, scripts, in-page fragment targets, and security mixed-content violations.

---

## High-Throughput Pipeline

BrokenLens processes web targets through five concurrent stages:

```
Seed URL -> Robots.txt & Sitemap -> URL Normalization -> Worker Pool -> Analysis & Export
                                                               │
                                       ┌───────────────────────┴───────────────────────┐
                                       ▼                                               ▼
                              Streaming HTML Tokenizer                        HTTP Status & Chain
                              - Hyperlinks (<a href>)                         - 2xx, 3xx, 4xx, 5xx
                              - Images & srcset (<img src>)                   - Circular loop detector
                              - Scripts (<script src>)                        - Excessive hops (>3)
                              - Stylesheets (<link rel="stylesheet">)         - HEAD-first external check
                              - In-page anchors (#target in DOM)
                              - Insecure HTTP on HTTPS (Mixed Content)
```

---

## 1. Concurrency Model & Queue Coordination

The crawler engine relies on Go channels and an atomic work-stealing pool:
- **Worker Routines**: A configurable pool of 1 to 250 goroutines consumes tasks from a buffered channel.
- **Deduplication Set**: A concurrent visited map tracks processed URLs. If a URL was already visited, new inbound references are recorded without scheduling redundant network requests.
- **Graceful Termination**: An atomic pending counter tracks enqueued tasks against completed tasks. When the pending count reaches zero, the task queue closes and workers exit cleanly.

---

## 2. Streaming HTML Extraction

Rather than building an in-memory Document Object Model (DOM) tree for entire pages, BrokenLens uses a streaming tokenizer (`golang.org/x/net/html`):
- **Memory Footprint**: Pages are parsed token by token in a single pass.
- **Asset Coverage**:
  - Hyperlinks: `<a>`, `<area>`
  - Visuals: `<img>`, `srcset`, `<picture>`, `<source>`
  - Assets: `<script src>`, `<link rel="stylesheet">`, `<link rel="preload">`, `<link rel="icon">`
  - Inline Styles: CSS `url(...)` declarations in style tags and `style="..."` attributes
  - Embedded Media: `<video>`, `<audio>`, `<track>`, `poster="..."`
- **DOM Fragment Mapping**: Every element `id` and `<a name>` attribute is indexed in a lookup table for fragment target verification.

---

## 3. Deep In-Page Anchor Verification (#id)

When a page links to a section (for example, `/docs#quickstart`):
1. The base URL (`/docs`) is fetched and audited.
2. The page's extracted DOM IDs and anchor names are recorded.
3. Post-crawl, BrokenLens verifies whether `quickstart` exists on the target page. If missing, an `anchor_not_found` warning issue is emitted with the referencing page and anchor text context.

---

## 4. Security & Mixed-Content Auditing

Browsers restrict insecure HTTP resources when embedded in HTTPS pages. BrokenLens classifies mixed content into two categories:
- **Active Mixed Content (Critical)**: Scripts, stylesheets, preloaded assets, and iframes loaded over `http://`. Browsers block these outright, breaking application logic.
- **Passive Mixed Content (Warning)**: Insecure images, videos, and audio. Browsers may load these with security warnings or upgrade them.

---

## 5. Redirect Chain Tracer & Circular Loop Detection

Instead of delegating redirect handling to standard HTTP libraries that discard intermediate hops:
- BrokenLens inspects every `Location` header manually.
- It records each hop: source URL, target URL, and HTTP status (301, 302, 307, 308).
- It detects circular loops ($A \to B \to A$) and reports them as critical issues.
- It flags excessive redirect chains exceeding the configured threshold.

---

## 6. Single Binary Distribution with Embedded Web UI

The React 19 interactive report studio is compiled with Vite and Tailwind CSS. The output directory is compiled into the Go binary using Go standard library `embed.FS`. A single executable contains:
- The command-line auditor with exit-code gating.
- Multi-format exporters (JSON, CSV, JUnit XML, Markdown).
- An embedded HTTP server providing live Server-Sent Events (SSE) and interactive report dashboards.
