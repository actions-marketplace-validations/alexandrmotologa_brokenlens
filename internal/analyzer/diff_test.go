package analyzer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/brokenlens/internal/model"
)

func TestCompareWithBaseline(t *testing.T) {
	tempDir := t.TempDir()
	baselinePath := filepath.Join(tempDir, "baseline.json")

	baselineReport := model.AuditReport{
		Issues: []model.AuditIssue{
			{
				ID:        "issue-1",
				Type:      model.IssueBroken,
				TargetURL: "https://example.com/old-broken",
				SourceURL: "https://example.com",
			},
			{
				ID:        "issue-2",
				Type:      model.IssueBroken,
				TargetURL: "https://example.com/persisted-broken",
				SourceURL: "https://example.com",
			},
		},
	}

	baselineData, err := json.Marshal(baselineReport)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if err := os.WriteFile(baselinePath, baselineData, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	currentReport := &model.AuditReport{
		Issues: []model.AuditIssue{
			{
				ID:        "issue-2",
				Type:      model.IssueBroken,
				TargetURL: "https://example.com/persisted-broken",
				SourceURL: "https://example.com",
			},
			{
				ID:        "issue-3",
				Type:      model.IssueBroken,
				TargetURL: "https://example.com/brand-new-broken",
				SourceURL: "https://example.com",
			},
		},
	}

	diff, err := CompareWithBaseline(currentReport, baselinePath)
	if err != nil {
		t.Fatalf("CompareWithBaseline failed: %v", err)
	}

	if len(diff.NewIssues) != 1 || diff.NewIssues[0].TargetURL != "https://example.com/brand-new-broken" {
		t.Errorf("Expected 1 new issue for brand-new-broken, got %d", len(diff.NewIssues))
	}

	if len(diff.ResolvedIssues) != 1 || diff.ResolvedIssues[0].TargetURL != "https://example.com/old-broken" {
		t.Errorf("Expected 1 resolved issue for old-broken, got %d", len(diff.ResolvedIssues))
	}

	if len(diff.PersistedIssues) != 1 || diff.PersistedIssues[0].TargetURL != "https://example.com/persisted-broken" {
		t.Errorf("Expected 1 persisted issue, got %d", len(diff.PersistedIssues))
	}
}
