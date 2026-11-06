package reporter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/alexandrmotologa/brokenlens/internal/model"
)

func makeTestReport() *model.AuditReport {
	return &model.AuditReport{
		Summary: model.CrawlSummary{
			TargetURL:         "https://example.com",
			StartTime:         time.Now(),
			EndTime:           time.Now().Add(500 * time.Millisecond),
			DurationMs:        500,
			TotalScanned:      10,
			TotalUnique:       10,
			TotalInternal:     8,
			TotalExternal:     2,
			TotalBroken:       1,
			TotalRedirects:    1,
			TotalMixedContent: 1,
			HealthScore:       85.5,
			RequestsPerSecond: 20.0,
			StatusCodeDistribution: map[string]int{
				"2xx": 8,
				"3xx": 1,
				"4xx": 1,
			},
		},
		Issues: []model.AuditIssue{
			{
				ID:           "test-issue-1",
				Type:         model.IssueBroken,
				Severity:     model.SeverityError,
				TargetURL:    "https://example.com/broken",
				SourceURL:    "https://example.com",
				StatusCode:   404,
				ErrorMessage: "Not Found",
				AssetType:    model.AssetLink,
				Context:      "Broken link",
			},
			{
				ID:           "test-issue-2",
				Type:         model.IssueMixedContentActive,
				Severity:     model.SeverityCritical,
				TargetURL:    "http://cdn.example.com/script.js",
				SourceURL:    "https://example.com",
				ErrorMessage: "Insecure script",
				AssetType:    model.AssetScript,
			},
		},
		RedirectChains: []model.RedirectChain{
			{
				InitialURL: "https://example.com/old",
				FinalURL:   "https://example.com/new",
				Hops: []model.RedirectHop{
					{
						FromURL:    "https://example.com/old",
						ToURL:      "https://example.com/new",
						StatusCode: 301,
					},
				},
			},
		},
	}
}

func TestWriteJSON(t *testing.T) {
	rep := makeTestReport()
	var buf bytes.Buffer
	if err := WriteJSON(&buf, rep); err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "https://example.com/broken") {
		t.Errorf("JSON output does not contain expected target URL: %s", out)
	}
}

func TestWriteCSV(t *testing.T) {
	rep := makeTestReport()
	var buf bytes.Buffer
	if err := WriteCSV(&buf, rep.Issues); err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "test-issue-1") || !strings.Contains(out, "404") {
		t.Errorf("CSV output missing issue data: %s", out)
	}
}

func TestWriteJUnit(t *testing.T) {
	rep := makeTestReport()
	var buf bytes.Buffer
	if err := WriteJUnit(&buf, rep); err != nil {
		t.Fatalf("WriteJUnit failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `<testsuites name="brokenlens"`) || !strings.Contains(out, `failures="2"`) {
		t.Errorf("JUnit XML output missing expected structure: %s", out)
	}
}

func TestWriteMarkdown(t *testing.T) {
	rep := makeTestReport()
	var buf bytes.Buffer
	if err := WriteMarkdown(&buf, rep); err != nil {
		t.Fatalf("WriteMarkdown failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "BrokenLens Audit Summary") || !strings.Contains(out, "85.5 / 100.0") {
		t.Errorf("Markdown output missing summary content: %s", out)
	}
}
