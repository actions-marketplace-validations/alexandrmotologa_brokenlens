package extractor

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/brokenlens/internal/model"
)

// ValidateAnchor checks whether a fragment identifier exists within the page's DOM fragments.
func ValidateAnchor(targetFragment string, domFragments map[string]bool) bool {
	if targetFragment == "" {
		return true
	}

	// Exact match
	if domFragments[targetFragment] {
		return true
	}

	// Case-insensitive fallback check
	lowerFragment := strings.ToLower(targetFragment)
	for id := range domFragments {
		if strings.ToLower(id) == lowerFragment {
			return true
		}
	}

	return false
}

// CreateAnchorNotFoundIssue constructs an AuditIssue when a fragment cannot be located.
func CreateAnchorNotFoundIssue(targetURL string, fragment string, sourceURL string, anchorText string) model.AuditIssue {
	context := fmt.Sprintf("Missing target fragment #%s", fragment)
	if anchorText != "" {
		context += fmt.Sprintf(" (anchor text: %q)", anchorText)
	}

	return model.AuditIssue{
		ID:           fmt.Sprintf("anchor-%s-#%s", targetURL, fragment),
		Type:         model.IssueAnchorNotFound,
		Severity:     model.SeverityWarning,
		TargetURL:    fmt.Sprintf("%s#%s", targetURL, fragment),
		SourceURL:    sourceURL,
		ErrorMessage: fmt.Sprintf("Target fragment #%s does not exist on page %s", fragment, targetURL),
		AssetType:    model.AssetLink,
		Context:      context,
	}
}
