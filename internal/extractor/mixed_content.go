package extractor

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/brokenlens/internal/model"
)

// CheckMixedContent audits whether a resource loaded on an HTTPS parent page is insecure HTTP.
func CheckMixedContent(parentURL string, assetURL string, assetType model.AssetType) *model.AuditIssue {
	parentLower := strings.ToLower(parentURL)
	assetLower := strings.ToLower(assetURL)

	// Mixed content only occurs when the parent document is served securely via HTTPS
	if !strings.HasPrefix(parentLower, "https://") {
		return nil
	}

	// Insecure resource referenced over plain HTTP
	if !strings.HasPrefix(assetLower, "http://") {
		return nil
	}

	isActive := isActiveMixedContent(assetType)

	var issueType model.IssueType
	var severity model.IssueSeverity
	var description string

	if isActive {
		issueType = model.IssueMixedContentActive
		severity = model.SeverityCritical
		description = fmt.Sprintf("Active mixed content: Insecure %s loaded over HTTP on HTTPS page (blocked by browsers)", assetType)
	} else {
		issueType = model.IssueMixedContentPassive
		severity = model.SeverityWarning
		description = fmt.Sprintf("Passive mixed content: Insecure %s loaded over HTTP on HTTPS page", assetType)
	}

	return &model.AuditIssue{
		ID:           fmt.Sprintf("mixed-%s-%s", parentURL, assetURL),
		Type:         issueType,
		Severity:     severity,
		TargetURL:    assetURL,
		SourceURL:    parentURL,
		ErrorMessage: description,
		AssetType:    assetType,
		Context:      fmt.Sprintf("Parent: %s", parentURL),
	}
}

// isActiveMixedContent determines if an asset belongs to blockable active content.
func isActiveMixedContent(assetType model.AssetType) bool {
	switch assetType {
	case model.AssetScript, model.AssetStylesheet, model.AssetPreload:
		return true
	default:
		return false
	}
}
