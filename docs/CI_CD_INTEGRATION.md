# BrokenLens CI/CD Integration Guide

BrokenLens provides automated quality gating for continuous integration pipelines. It returns exit code `1` when broken assets or critical issues are detected, preventing regressions from deploying to production.

---

## Exit Code Behavior

| Exit Code | Condition |
| :--- | :--- |
| `0` | Audit completed successfully with zero broken items or below threshold. |
| `1` | Broken resources or critical issues detected with `--fail-on-broken`. |
| `2` | Configuration or CLI flag parsing error. |

---

## 1. GitHub Actions

### Option A: Using the Composite Action

```yaml
name: Production Link & Asset Audit

on:
  deployment_status:
  schedule:
    - cron: '0 4 * * 1' # Run every Monday at 4:00 AM UTC
  workflow_dispatch:

jobs:
  audit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Audit Website
        uses: alexandrmotologa/brokenlens@main
        with:
          url: 'https://example.com'
          concurrency: '30'
          max-depth: '5'
          fail-on-broken: 'true'
          output: 'brokenlens-report.json'
          junit: 'brokenlens-junit.xml'

      - name: Publish Test Results
        uses: EnricoMi/publish-unit-test-result-action@v2
        if: always()
        with:
          files: 'brokenlens-junit.xml'
```

### Option B: Direct Go Run or Binary Execution

```yaml
name: Staging Gate

on: [pull_request]

jobs:
  verify-links:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Run BrokenLens Gating
        run: |
          go run ./cmd/brokenlens https://staging.example.com \
            --concurrency 40 \
            --max-depth 4 \
            --check-external \
            --fail-on-broken \
            --junit results.xml
```

---

## 2. GitLab CI

```yaml
stages:
  - post-deploy

broken_link_audit:
  stage: post-deploy
  image: golang:1.24-alpine
  script:
    - go run ./cmd/brokenlens https://example.com --fail-on-broken --junit junit.xml
  artifacts:
    when: always
    reports:
      junit: junit.xml
  rules:
    - if: $CI_COMMIT_BRANCH == "main"
```

---

## 3. Docker Container Execution

You can run BrokenLens headlessly inside Docker without installing Go on the host:

```bash
docker run --rm \
  -v $(pwd)/reports:/reports \
  ghcr.io/alexandrmotologa/brokenlens:latest https://example.com \
    --concurrency 30 \
    --fail-on-broken \
    --output /reports/report.json
```
