package reporter

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/alexandrmotologa/brokenlens/internal/crawler"
	"github.com/alexandrmotologa/brokenlens/internal/model"
)

// TerminalReporter renders live spinner and formatted audit summaries to stdout.
type TerminalReporter struct {
	w       io.Writer
	isQuiet bool
}

// NewTerminalReporter creates a new terminal reporter.
func NewTerminalReporter(w io.Writer, isQuiet bool) *TerminalReporter {
	if w == nil {
		w = os.Stdout
	}
	return &TerminalReporter{w: w, isQuiet: isQuiet}
}

// OnProgress renders a dynamic single-line crawl status update.
func (t *TerminalReporter) OnProgress(stats crawler.ProgressStats) {
	if t.isQuiet {
		return
	}

	displayURL := stats.CurrentURL
	if len(displayURL) > 42 {
		displayURL = displayURL[:39] + "..."
	}

	fmt.Fprintf(t.w, "\r\033[K[brokenlens] %4d scanned | %3d broken | %5.1f req/s | %s",
		stats.TotalScanned, stats.TotalBroken, stats.RequestsPerSec, displayURL)
}

// RenderSummary renders the complete terminal audit report after crawl completion.
func (t *TerminalReporter) RenderSummary(report *model.AuditReport) {
	s := report.Summary

	fmt.Fprintln(t.w, "\r\033[K")
	fmt.Fprintln(t.w, "================================================================================")
	fmt.Fprintf(t.w, "  BROKENLENS AUDIT REPORT: %s\n", s.TargetURL)
	fmt.Fprintln(t.w, "================================================================================")
	fmt.Fprintf(t.w, "  Duration:          %s (%d ms)\n", time.Duration(s.DurationMs)*time.Millisecond, s.DurationMs)
	fmt.Fprintf(t.w, "  Throughput:        %.1f requests/sec\n", s.RequestsPerSecond)
	fmt.Fprintf(t.w, "  Health Score:      %.1f / 100.0\n", s.HealthScore)
	fmt.Fprintln(t.w, "--------------------------------------------------------------------------------")
	fmt.Fprintf(t.w, "  Total URLs Scanned: %-6d  |  Internal URLs:     %-6d\n", s.TotalScanned, s.TotalInternal)
	fmt.Fprintf(t.w, "  Total Broken:       %-6d  |  External URLs:     %-6d\n", s.TotalBroken, s.TotalExternal)
	fmt.Fprintf(t.w, "  Redirect Chains:    %-6d  |  Mixed Content:     %-6d\n", s.TotalRedirects, s.TotalMixedContent)
	fmt.Fprintf(t.w, "  Missing Anchors:    %-6d  |  Orphan Pages:      %-6d\n", s.TotalMissingAnchors, s.TotalOrphanPages)
	fmt.Fprintf(t.w, "  Soft 404 Pages:     %-6d  |  Canonical Issues:  %-6d\n", s.TotalSoft404, s.TotalCanonicalIssues)
	fmt.Fprintln(t.w, "--------------------------------------------------------------------------------")

	// Differential Summary if present
	if s.Diff != nil {
		fmt.Fprintf(t.w, "  DIFFERENTIAL PR AUDIT (vs %s)\n", s.Diff.BaselineReportPath)
		fmt.Fprintf(t.w, "  - NEW Issues:       %-6d (introduced in this run)\n", s.Diff.NewIssuesCount)
		fmt.Fprintf(t.w, "  - RESOLVED Issues:  %-6d (fixed since baseline)\n", s.Diff.ResolvedIssuesCount)
		fmt.Fprintf(t.w, "  - Persisted Issues: %-6d (pre-existing)\n", s.Diff.PersistedIssuesCount)
		fmt.Fprintln(t.w, "--------------------------------------------------------------------------------")
	}

	// Status code breakdown
	fmt.Fprint(t.w, "  Status Breakdown:  ")
	for bucket, count := range s.StatusCodeDistribution {
		fmt.Fprintf(t.w, "[%s: %d] ", bucket, count)
	}
	fmt.Fprintln(t.w)

	// Issues table (if any)
	if len(report.Issues) > 0 {
		fmt.Fprintln(t.w, "================================================================================")
		fmt.Fprintf(t.w, "  DETECTED ISSUES (%d total)\n", len(report.Issues))
		fmt.Fprintln(t.w, "================================================================================")

		limit := 15
		for i, issue := range report.Issues {
			if i >= limit {
				fmt.Fprintf(t.w, "  ... and %d more issues (export to JSON/CSV for full list)\n", len(report.Issues)-limit)
				break
			}

			statusStr := fmt.Sprintf("%d", issue.StatusCode)
			if issue.StatusCode == 0 {
				statusStr = "ERR"
			}

			newTag := ""
			if issue.IsNew {
				newTag = "[NEW] "
			}

			fmt.Fprintf(t.w, "  %s[%s] [%s] %s (HTTP %s)\n",
				newTag,
				strings.ToUpper(string(issue.Severity)),
				issue.Type,
				issue.TargetURL,
				statusStr,
			)
			if issue.SourceURL != "" {
				fmt.Fprintf(t.w, "        Found on: %s\n", issue.SourceURL)
			}
			if issue.Context != "" {
				fmt.Fprintf(t.w, "        Context:  %s\n", issue.Context)
			}
			if issue.Snippet != "" {
				fmt.Fprintf(t.w, "        Snippet:  %s\n", issue.Snippet)
			}
			if issue.ErrorMessage != "" {
				fmt.Fprintf(t.w, "        Reason:   %s\n", issue.ErrorMessage)
			}
			fmt.Fprintln(t.w)
		}
	} else {
		fmt.Fprintln(t.w, "\n  No broken links, dead assets, or mixed-content issues detected.")
	}

	fmt.Fprintln(t.w, "================================================================================")
}
