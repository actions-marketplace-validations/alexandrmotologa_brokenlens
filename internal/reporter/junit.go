package reporter

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"

	"github.com/alexandrmotologa/brokenlens/internal/model"
)

type jUnitProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type jUnitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

type jTestCase struct {
	Classname string        `xml:"classname,attr"`
	Name      string        `xml:"name,attr"`
	Time      string        `xml:"time,attr"`
	Failure   *jUnitFailure `xml:"failure,omitempty"`
}

type jTestSuite struct {
	Name       string          `xml:"name,attr"`
	Tests      int             `xml:"tests,attr"`
	Failures   int             `xml:"failures,attr"`
	Errors     int             `xml:"errors,attr"`
	Time       string          `xml:"time,attr"`
	Properties []jUnitProperty `xml:"properties>property,omitempty"`
	TestCases  []jTestCase     `xml:"testcase"`
}

type jTestSuites struct {
	XMLName  xml.Name     `xml:"testsuites"`
	Name     string       `xml:"name,attr"`
	Tests    int          `xml:"tests,attr"`
	Failures int          `xml:"failures,attr"`
	Errors   int          `xml:"errors,attr"`
	Time     string       `xml:"time,attr"`
	Suites   []jTestSuite `xml:"testsuite"`
}

// WriteJUnit generates a JUnit XML representation of the audit report.
func WriteJUnit(w io.Writer, report *model.AuditReport) error {
	s := report.Summary
	totalTests := s.TotalScanned
	if totalTests == 0 {
		totalTests = 1
	}

	failuresCount := len(report.Issues)
	timeSec := fmt.Sprintf("%.3f", float64(s.DurationMs)/1000.0)

	suite := jTestSuite{
		Name:     "Web Quality & Broken Link Audit",
		Tests:    totalTests,
		Failures: failuresCount,
		Errors:   0,
		Time:     timeSec,
		Properties: []jUnitProperty{
			{Name: "target_url", Value: s.TargetURL},
			{Name: "health_score", Value: fmt.Sprintf("%.1f", s.HealthScore)},
			{Name: "total_scanned", Value: fmt.Sprintf("%d", s.TotalScanned)},
		},
	}

	for _, issue := range report.Issues {
		tc := jTestCase{
			Classname: fmt.Sprintf("brokenlens.%s", issue.Type),
			Name:      fmt.Sprintf("%s (%s)", issue.TargetURL, issue.AssetType),
			Time:      "0.001",
			Failure: &jUnitFailure{
				Message: issue.ErrorMessage,
				Type:    string(issue.Severity),
				Content: fmt.Sprintf("Source: %s\nTarget: %s\nStatus: %d\nContext: %s",
					issue.SourceURL, issue.TargetURL, issue.StatusCode, issue.Context),
			},
		}
		suite.TestCases = append(suite.TestCases, tc)
	}

	// If no failures, add a passing dummy testcase
	if len(report.Issues) == 0 {
		suite.TestCases = append(suite.TestCases, jTestCase{
			Classname: "brokenlens.audit",
			Name:      "all_resources_valid",
			Time:      timeSec,
		})
	}

	suites := jTestSuites{
		Name:     "brokenlens",
		Tests:    suite.Tests,
		Failures: suite.Failures,
		Errors:   suite.Errors,
		Time:     timeSec,
		Suites:   []jTestSuite{suite},
	}

	w.Write([]byte(xml.Header))
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	return encoder.Encode(suites)
}

// ExportJUnit writes the JUnit XML report to a file.
func ExportJUnit(filePath string, report *model.AuditReport) error {
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create JUnit XML file: %w", err)
	}
	defer f.Close()
	return WriteJUnit(f, report)
}
