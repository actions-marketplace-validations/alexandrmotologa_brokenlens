package analyzer

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alexandrmotologa/brokenlens/internal/model"
)

// CompareWithBaseline compares the current audit issues against a previous baseline report.
func CompareWithBaseline(currentReport *model.AuditReport, baselinePath string) (*model.DiffReport, error) {
	data, err := os.ReadFile(baselinePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read baseline report file: %w", err)
	}

	var baseline model.AuditReport
	if err := json.Unmarshal(data, &baseline); err != nil {
		return nil, fmt.Errorf("failed to parse baseline report JSON: %w", err)
	}

	baselineMap := make(map[string]model.AuditIssue)
	for _, issue := range baseline.Issues {
		key := issueKey(issue)
		baselineMap[key] = issue
	}

	currentMap := make(map[string]model.AuditIssue)
	var newIssues []model.AuditIssue
	var persistedIssues []model.AuditIssue

	for i := range currentReport.Issues {
		issue := currentReport.Issues[i]
		key := issueKey(issue)
		currentMap[key] = issue

		if _, exists := baselineMap[key]; exists {
			persistedIssues = append(persistedIssues, issue)
		} else {
			issue.IsNew = true
			currentReport.Issues[i].IsNew = true
			newIssues = append(newIssues, issue)
		}
	}

	var resolvedIssues []model.AuditIssue
	for key, issue := range baselineMap {
		if _, exists := currentMap[key]; !exists {
			resolvedIssues = append(resolvedIssues, issue)
		}
	}

	diffReport := &model.DiffReport{
		BaselineReportPath: baselinePath,
		NewIssues:          newIssues,
		ResolvedIssues:     resolvedIssues,
		PersistedIssues:    persistedIssues,
	}

	currentReport.Diff = diffReport
	currentReport.Summary.Diff = &model.DiffSummary{
		BaselineReportPath:   baselinePath,
		NewIssuesCount:       len(newIssues),
		ResolvedIssuesCount:  len(resolvedIssues),
		PersistedIssuesCount: len(persistedIssues),
	}

	return diffReport, nil
}

func issueKey(issue model.AuditIssue) string {
	return fmt.Sprintf("%s::%s::%s", issue.Type, issue.TargetURL, issue.SourceURL)
}
