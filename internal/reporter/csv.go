package reporter

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/alexandrmotologa/brokenlens/internal/model"
)

// WriteCSV exports detected audit issues into CSV format.
func WriteCSV(w io.Writer, issues []model.AuditIssue) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{
		"ID",
		"Type",
		"Severity",
		"TargetURL",
		"StatusCode",
		"SourceURL",
		"AssetType",
		"Context",
		"Snippet",
		"IsInternal",
		"IsNew",
		"ErrorMessage",
	}

	if err := writer.Write(header); err != nil {
		return err
	}

	for _, issue := range issues {
		statusStr := ""
		if issue.StatusCode > 0 {
			statusStr = strconv.Itoa(issue.StatusCode)
		}

		isInternalStr := "false"
		if issue.IsInternal {
			isInternalStr = "true"
		}

		isNewStr := "false"
		if issue.IsNew {
			isNewStr = "true"
		}

		row := []string{
			issue.ID,
			string(issue.Type),
			string(issue.Severity),
			issue.TargetURL,
			statusStr,
			issue.SourceURL,
			string(issue.AssetType),
			issue.Context,
			issue.Snippet,
			isInternalStr,
			isNewStr,
			issue.ErrorMessage,
		}

		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// ExportCSV writes detected issues to a file path.
func ExportCSV(filePath string, issues []model.AuditIssue) error {
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer f.Close()
	return WriteCSV(f, issues)
}
