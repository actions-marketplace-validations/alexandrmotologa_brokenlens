package reporter

import (
	"encoding/json"
	"io"
	"os"

	"github.com/alexandrmotologa/brokenlens/internal/model"
)

// WriteJSON formats the audit report as indented JSON and writes to an io.Writer.
func WriteJSON(w io.Writer, report *model.AuditReport) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// ExportJSON writes the report to a target file path.
func ExportJSON(filePath string, report *model.AuditReport) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return WriteJSON(f, report)
}
