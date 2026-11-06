package crawler

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alexandrmotologa/brokenlens/internal/config"
	"github.com/alexandrmotologa/brokenlens/internal/extractor"
	"github.com/alexandrmotologa/brokenlens/internal/model"
)

// CrawlTask represents a single work unit in the crawl queue.
type CrawlTask struct {
	URL        string
	Depth      int
	IsInternal bool
	AssetType  model.AssetType
	SourceURL  string
	AnchorText string
	Fragment   string
}

// ProgressStats captures live crawl statistics.
type ProgressStats struct {
	TotalScanned   int64
	TotalQueued    int64
	TotalBroken    int64
	ActiveWorkers  int64
	CurrentURL     string
	Duration       time.Duration
	RequestsPerSec float64
}

// ProgressCallback is invoked as crawl progress occurs.
type ProgressCallback func(stats ProgressStats)

// AnchorCheckTask records a fragment to be validated once the target page is parsed.
type AnchorCheckTask struct {
	TargetURL  string
	Fragment   string
	SourceURL  string
	AnchorText string
}

// Engine orchestrates the distributed crawling and asset validation pipeline.
type Engine struct {
	cfg        *config.Config
	client     *HTTPClient
	robots     *RobotsPolicy
	targetHost string

	// State storage
	visitedMu    sync.RWMutex
	visited      map[string]*model.CrawlResult
	inboundMu    sync.RWMutex
	inboundMap   map[string][]model.InboundReference
	anchorMu     sync.Mutex
	anchorChecks []AnchorCheckTask
	issuesMu     sync.Mutex
	issues       []model.AuditIssue

	// Queue and concurrency control
	taskQueue   chan CrawlTask
	pendingWait sync.WaitGroup
	scanned     atomic.Int64
	queued      atomic.Int64
	broken      atomic.Int64
	active      atomic.Int64

	startTime  time.Time
	onProgress ProgressCallback
}

// NewEngine creates a configured crawl engine.
func NewEngine(cfg *config.Config, cb ProgressCallback) (*Engine, error) {
	parsedTarget, err := url.Parse(cfg.TargetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid target url: %w", err)
	}

	client := NewHTTPClient(cfg)

	return &Engine{
		cfg:          cfg,
		client:       client,
		targetHost:   parsedTarget.Host,
		visited:      make(map[string]*model.CrawlResult),
		inboundMap:   make(map[string][]model.InboundReference),
		taskQueue:    make(chan CrawlTask, 50000),
		onProgress:   cb,
		anchorChecks: make([]AnchorCheckTask, 0),
	}, nil
}

// Run executes the audit until completion or context cancellation.
func (e *Engine) Run(ctx context.Context) (*model.AuditReport, error) {
	e.startTime = time.Now()

	// Ingest robots.txt if enabled
	if e.cfg.FollowRobotsTxt {
		e.robots = FetchRobotsTxt(ctx, e.client, e.cfg.TargetURL, e.cfg)
	} else {
		e.robots = &RobotsPolicy{}
	}

	// Ingest sitemaps if enabled
	var sitemapSeeds []string
	if e.cfg.HonorSitemap {
		sitemapSeeds = IngestSitemaps(ctx, e.client, e.cfg.TargetURL, e.robots.Sitemaps)
	}

	// Normalize target root URL
	rootURL, rootFrag, err := NormalizeURL(e.cfg.TargetURL, e.cfg.TargetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize root url: %w", err)
	}

	// Enqueue initial root task
	e.enqueueTask(CrawlTask{
		URL:        rootURL,
		Depth:      0,
		IsInternal: true,
		AssetType:  model.AssetLink,
		Fragment:   rootFrag,
	})

	// Enqueue sitemap seeds as depth 1
	for _, seed := range sitemapSeeds {
		normSeed, frag, err := NormalizeURL(seed, rootURL)
		if err == nil && IsSameDomain(normSeed, e.targetHost) {
			e.enqueueTask(CrawlTask{
				URL:        normSeed,
				Depth:      1,
				IsInternal: true,
				AssetType:  model.AssetLink,
				Fragment:   frag,
			})
		}
	}

	// Launch worker pool
	var workerWg sync.WaitGroup
	for i := 0; i < e.cfg.Concurrency; i++ {
		workerWg.Add(1)
		go func(workerID int) {
			defer workerWg.Done()
			e.workerLoop(ctx, workerID)
		}(i)
	}

	// Wait in background until all pending tasks finish, then close taskQueue
	doneCh := make(chan struct{})
	go func() {
		e.pendingWait.Wait()
		close(e.taskQueue)
		close(doneCh)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-doneCh:
	}

	// Wait for workers to cleanly exit
	workerWg.Wait()

	// Post-crawl: Evaluate all collected anchor fragments
	if e.cfg.CheckFragments {
		e.validateAllAnchors()
	}

	// Assemble final report
	report := e.compileReport()
	return report, nil
}

func (e *Engine) enqueueTask(task CrawlTask) bool {
	// Check URL filters
	if e.cfg.IncludeRegex != nil && !e.cfg.IncludeRegex.MatchString(task.URL) {
		return false
	}
	if e.cfg.ExcludeRegex != nil && e.cfg.ExcludeRegex.MatchString(task.URL) {
		return false
	}

	// Check robots.txt disallow for internal links
	if task.IsInternal && e.robots != nil {
		parsed, err := url.Parse(task.URL)
		if err == nil && !e.robots.IsAllowed(parsed.Path) {
			return false
		}
	}

	// Save anchor check whenever a fragment is present, even if base URL was already queued or visited
	if task.Fragment != "" {
		e.anchorMu.Lock()
		e.anchorChecks = append(e.anchorChecks, AnchorCheckTask{
			TargetURL:  task.URL,
			Fragment:   task.Fragment,
			SourceURL:  task.SourceURL,
			AnchorText: task.AnchorText,
		})
		e.anchorMu.Unlock()
	}

	// Check if already visited
	e.visitedMu.Lock()
	if _, exists := e.visited[task.URL]; exists {
		e.visitedMu.Unlock()
		// Record inbound reference even if already visited
		if task.SourceURL != "" {
			e.inboundMu.Lock()
			e.inboundMap[task.URL] = append(e.inboundMap[task.URL], model.InboundReference{
				SourceURL: task.SourceURL,
				AssetType: task.AssetType,
				Text:      task.AnchorText,
				Fragment:  task.Fragment,
			})
			e.inboundMu.Unlock()
		}
		return false
	}
	// Mark as reserved in visited with nil entry while processing
	e.visited[task.URL] = nil
	e.visitedMu.Unlock()

	// Record initial inbound reference
	if task.SourceURL != "" {
		e.inboundMu.Lock()
		e.inboundMap[task.URL] = append(e.inboundMap[task.URL], model.InboundReference{
			SourceURL: task.SourceURL,
			AssetType: task.AssetType,
			Text:      task.AnchorText,
			Fragment:  task.Fragment,
		})
		e.inboundMu.Unlock()
	}

	e.pendingWait.Add(1)
	e.queued.Add(1)
	e.taskQueue <- task
	return true
}

func (e *Engine) workerLoop(ctx context.Context, id int) {
	for task := range e.taskQueue {
		select {
		case <-ctx.Done():
			e.pendingWait.Done()
			return
		default:
		}

		e.active.Add(1)
		e.processTask(ctx, task)
		e.active.Add(-1)
		e.scanned.Add(1)
		e.pendingWait.Done()

		if e.cfg.RateLimitMs > 0 {
			time.Sleep(time.Duration(e.cfg.RateLimitMs) * time.Millisecond)
		}

		// Emit live progress stats
		if e.onProgress != nil {
			elapsed := time.Since(e.startTime)
			scanned := e.scanned.Load()
			var rps float64
			if elapsed.Seconds() > 0 {
				rps = float64(scanned) / elapsed.Seconds()
			}
			e.onProgress(ProgressStats{
				TotalScanned:   scanned,
				TotalQueued:    e.queued.Load(),
				TotalBroken:    e.broken.Load(),
				ActiveWorkers:  e.active.Load(),
				CurrentURL:     task.URL,
				Duration:       elapsed,
				RequestsPerSec: rps,
			})
		}
	}
}

func (e *Engine) processTask(ctx context.Context, task CrawlTask) {
	// Determine method and whether body should be read
	// For internal HTML pages we read the body to extract links.
	// For external links or media/scripts/styles, we can use HEAD first or read minimal body.
	readBody := task.IsInternal && (task.AssetType == model.AssetLink || task.AssetType == "")
	method := http.MethodGet
	if !task.IsInternal && task.AssetType == model.AssetLink {
		method = http.MethodHead
	}

	resp := e.client.Fetch(ctx, task.URL, method, readBody)

	// Prepare result
	crawlResult := &model.CrawlResult{
		URL:           task.URL,
		RequestedURL:  task.URL,
		FinalURL:      resp.FinalURL,
		StatusCode:    resp.StatusCode,
		StatusText:    resp.StatusText,
		DurationMs:    resp.DurationMs,
		ContentType:   resp.ContentType,
		ContentLength: resp.ContentLength,
		IsInternal:    task.IsInternal,
		AssetType:     task.AssetType,
		RedirectChain: resp.RedirectChain,
		DOMFragments:  make(map[string]bool),
	}

	if resp.Error != nil {
		crawlResult.Error = resp.Error.Error()
	}

	// Detect issues
	e.evaluateIssues(task, resp, crawlResult)

	// If internal HTML page with HTTP 200, extract child assets and links
	if task.IsInternal && resp.StatusCode == http.StatusOK && (IsHTMLContent(resp.ContentType) || len(resp.Body) > 0) {
		ext, err := extractor.ExtractHTMLFromBytes(resp.Body, task.URL)
		if err == nil {
			crawlResult.DOMFragments = ext.DOMFragments
			crawlResult.ExtractedLinks = len(ext.Targets)

			// Enqueue discovered child links and assets
			if e.cfg.MaxDepth == 0 || task.Depth < e.cfg.MaxDepth {
				for _, target := range ext.Targets {
					normURL, frag, err := NormalizeURL(target.RawURL, task.URL)
					if err != nil || normURL == "" {
						continue
					}

					// Audit Mixed Content
					if e.cfg.CheckMixedContent {
						if issue := extractor.CheckMixedContent(task.URL, normURL, target.AssetType); issue != nil {
							e.issuesMu.Lock()
							e.issues = append(e.issues, *issue)
							e.issuesMu.Unlock()
						}
					}

					isChildInternal := IsSameDomain(normURL, e.targetHost)

					// If external and check-external is false, skip
					if !isChildInternal && !e.cfg.CheckExternal {
						continue
					}

					// Max URLs threshold check
					if e.cfg.MaxURLs > 0 && int(e.queued.Load()) >= e.cfg.MaxURLs {
						break
					}

					e.enqueueTask(CrawlTask{
						URL:        normURL,
						Depth:      task.Depth + 1,
						IsInternal: isChildInternal,
						AssetType:  target.AssetType,
						SourceURL:  task.URL,
						AnchorText: target.Text,
						Fragment:   frag,
					})
				}
			}
		}
	}

	// Commit result into visited map
	e.visitedMu.Lock()
	e.visited[task.URL] = crawlResult
	e.visitedMu.Unlock()
}

func (e *Engine) evaluateIssues(task CrawlTask, resp *FetchResponse, result *model.CrawlResult) {
	// Circular Redirect Loop
	if resp.IsRedirectLoop {
		e.broken.Add(1)
		e.issuesMu.Lock()
		e.issues = append(e.issues, model.AuditIssue{
			ID:           fmt.Sprintf("loop-%s", task.URL),
			Type:         model.IssueRedirectLoop,
			Severity:     model.SeverityCritical,
			TargetURL:    task.URL,
			SourceURL:    task.SourceURL,
			StatusCode:   resp.StatusCode,
			ErrorMessage: fmt.Sprintf("Circular redirect loop: %s", resp.Error.Error()),
			AssetType:    task.AssetType,
			Context:      fmt.Sprintf("%d hops before loop", len(resp.RedirectChain)),
		})
		e.issuesMu.Unlock()
		return
	}

	// Too Many Redirects
	if len(resp.RedirectChain) > e.cfg.MaxRedirects {
		e.broken.Add(1)
		e.issuesMu.Lock()
		e.issues = append(e.issues, model.AuditIssue{
			ID:           fmt.Sprintf("toomanyredirects-%s", task.URL),
			Type:         model.IssueExcessiveRedirects,
			Severity:     model.SeverityWarning,
			TargetURL:    task.URL,
			SourceURL:    task.SourceURL,
			StatusCode:   resp.StatusCode,
			ErrorMessage: fmt.Sprintf("Excessive redirects (%d hops exceed limit of %d)", len(resp.RedirectChain), e.cfg.MaxRedirects),
			AssetType:    task.AssetType,
		})
		e.issuesMu.Unlock()
	}

	// Broken HTTP Status or Connection Error
	if resp.StatusCode >= 400 || (resp.StatusCode == 0 && resp.Error != nil) {
		e.broken.Add(1)
		severity := model.SeverityError
		if task.AssetType == model.AssetScript || task.AssetType == model.AssetStylesheet {
			severity = model.SeverityCritical
		}

		errMsg := resp.StatusText
		if resp.Error != nil {
			errMsg = resp.Error.Error()
		}

		context := task.AnchorText
		if context == "" {
			context = string(task.AssetType)
		}

		e.issuesMu.Lock()
		e.issues = append(e.issues, model.AuditIssue{
			ID:           fmt.Sprintf("broken-%s", task.URL),
			Type:         model.IssueBroken,
			Severity:     severity,
			TargetURL:    task.URL,
			SourceURL:    task.SourceURL,
			StatusCode:   resp.StatusCode,
			ErrorMessage: errMsg,
			AssetType:    task.AssetType,
			Context:      context,
		})
		e.issuesMu.Unlock()
	}
}

func (e *Engine) validateAllAnchors() {
	e.anchorMu.Lock()
	defer e.anchorMu.Unlock()

	e.visitedMu.RLock()
	defer e.visitedMu.RUnlock()

	for _, check := range e.anchorChecks {
		res, exists := e.visited[check.TargetURL]
		if !exists || res == nil {
			continue
		}

		if !extractor.ValidateAnchor(check.Fragment, res.DOMFragments) {
			e.issuesMu.Lock()
			e.issues = append(e.issues, extractor.CreateAnchorNotFoundIssue(
				check.TargetURL,
				check.Fragment,
				check.SourceURL,
				check.AnchorText,
			))
			e.issuesMu.Unlock()
		}
	}
}

func (e *Engine) compileReport() *model.AuditReport {
	e.visitedMu.RLock()
	defer e.visitedMu.RUnlock()
	e.inboundMu.RLock()
	defer e.inboundMu.RUnlock()
	e.issuesMu.Lock()
	defer e.issuesMu.Unlock()

	endTime := time.Now()
	duration := endTime.Sub(e.startTime)

	var results []model.CrawlResult
	var redirectChains []model.RedirectChain

	statusDist := make(map[string]int)
	assetDist := make(map[string]int)
	var internalCount, externalCount int
	var missingAnchorCount, mixedContentCount int

	for _, issue := range e.issues {
		switch issue.Type {
		case model.IssueAnchorNotFound:
			missingAnchorCount++
		case model.IssueMixedContentActive, model.IssueMixedContentPassive:
			mixedContentCount++
		}
	}

	for u, res := range e.visited {
		if res == nil {
			continue
		}

		// Attach inbound references
		if inbounds, ok := e.inboundMap[u]; ok {
			res.InboundSources = inbounds
		}

		if res.IsInternal {
			internalCount++
		} else {
			externalCount++
		}

		// Status distribution
		if res.StatusCode > 0 {
			bucket := fmt.Sprintf("%dxx", res.StatusCode/100)
			statusDist[bucket]++
		} else {
			statusDist["error"]++
		}

		// Asset type distribution
		assetDist[string(res.AssetType)]++

		// Redirect chains
		if len(res.RedirectChain) > 0 {
			redirectChains = append(redirectChains, model.RedirectChain{
				InitialURL: res.URL,
				FinalURL:   res.FinalURL,
				Hops:       res.RedirectChain,
			})
		}

		results = append(results, *res)
	}

	totalScanned := len(results)
	totalBroken := int(e.broken.Load())

	// Calculate Health Score (100 - penalties)
	healthScore := 100.0
	if totalScanned > 0 {
		brokenRatio := float64(totalBroken) / float64(totalScanned)
		mixedRatio := float64(mixedContentCount) / float64(totalScanned)
		penalty := (brokenRatio * 70.0) + (mixedRatio * 20.0) + (float64(missingAnchorCount) * 0.5)
		healthScore = 100.0 - penalty
		if healthScore < 0.0 {
			healthScore = 0.0
		}
	}

	var rps float64
	if duration.Seconds() > 0 {
		rps = float64(totalScanned) / duration.Seconds()
	}

	summary := model.CrawlSummary{
		TargetURL:              e.cfg.TargetURL,
		StartTime:              e.startTime,
		EndTime:                endTime,
		DurationMs:             duration.Milliseconds(),
		TotalScanned:           totalScanned,
		TotalUnique:            totalScanned,
		TotalInternal:          internalCount,
		TotalExternal:          externalCount,
		TotalBroken:            totalBroken,
		TotalRedirects:         len(redirectChains),
		TotalMixedContent:      mixedContentCount,
		TotalMissingAnchors:    missingAnchorCount,
		HealthScore:            healthScore,
		StatusCodeDistribution: statusDist,
		AssetTypeDistribution:  assetDist,
		RequestsPerSecond:      rps,
	}

	return &model.AuditReport{
		Summary:        summary,
		Issues:         e.issues,
		RedirectChains: redirectChains,
		Results:        results,
	}
}
