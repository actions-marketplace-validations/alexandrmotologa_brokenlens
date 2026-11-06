package reporter

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/alexandrmotologa/brokenlens/internal/model"
)

// WriteMarkdown formats the audit summary into GitHub-flavored Markdown.
func WriteMarkdown(w io.Writer, report *model.AuditReport) error {
	s := report.Summary

	fmt.Fprintf(w, "## BrokenLens Audit Summary: `%s`\n\n", s.TargetURL)
	fmt.Fprintln(w, "| Metric | Value |")
	fmt.Fprintln(w, "| :--- | :--- |")
	fmt.Fprintf(w, "| Target URL | `%s` |\n", s.TargetURL)
	fmt.Fprintf(w, "| Duration | %s (%d ms) |\n", time.Duration(s.DurationMs)*time.Millisecond, s.DurationMs)
	fmt.Fprintf(w, "| Requests / Second | %.1f req/s |\n", s.RequestsPerSecond)
	fmt.Fprintf(w, "| Health Score | %.1f / 100.0 |\n", s.HealthScore)
	fmt.Fprintf(w, "| Total URLs Scanned | %d |\n", s.TotalScanned)
	fmt.Fprintf(w, "| Internal URLs | %d |\n", s.TotalInternal)
	fmt.Fprintf(w, "| External URLs | %d |\n", s.TotalExternal)
	fmt.Fprintf(w, "| Total Broken | %d |\n", s.TotalBroken)
	fmt.Fprintf(w, "| Redirect Chains | %d |\n", s.TotalRedirects)
	fmt.Fprintf(w, "| Mixed Content Warnings | %d |\n", s.TotalMixedContent)
	fmt.Fprintf(w, "| Missing Anchors (#id) | %d |\n", s.TotalMissingAnchors)
	fmt.Fprintln(w)

	if len(report.Issues) > 0 {
		fmt.Fprintf(w, "### Detected Issues (%d total)\n\n", len(report.Issues))
		fmt.Fprintln(w, "| Severity | Type | Target URL | HTTP Status | Source Page | Details |")
		fmt.Fprintln(w, "| :--- | :--- | :--- | :--- | :--- | :--- |")

		limit := 30
		for i, issue := range report.Issues {
			if i >= limit {
				fmt.Fprintf(w, "\n*... and %d more issues omitted from markdown summary.*\n", len(report.Issues)-limit)
				break
			}
			statusStr := fmt.Sprintf("%d", issue.StatusCode)
			if issue.StatusCode == 0 {
				statusStr = "ERR"
			}
			fmt.Fprintf(w, "| **%s** | `%s` | `%s` | %s | `%s` | %s |\n",
				strings.ToUpper(string(issue.Severity)),
				issue.Type,
				issue.TargetURL,
				statusStr,
				issue.SourceURL,
				issue.ErrorMessage,
			)
		}
	} else {
		fmt.Fprintln(w, "### Result")
		fmt.Fprintln(w, "All links, assets, scripts, stylesheets, and anchors passed validation without errors.")
	}

	return nil
}

// ExportMarkdown writes the Markdown report to a file.
func ExportMarkdown(filePath string, report *model.AuditReport) error {
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create Markdown file: %w", err)
	}
	defer f.Close()
	return WriteMarkdown(f, report)
}

// WriteGitHubStepSummary automatically writes to GITHUB_STEP_SUMMARY environment file if present.
func WriteGitHubStepSummary(report *model.AuditReport) error {
	summaryPath := os.Getenv("GITHUB_STEP_SUMMARY")
	if summaryPath == "" {
		return nil
	}

	f, err := os.OpenFile(summaryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	return WriteMarkdown(f, report)
}
