package extractor

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/brokenlens/internal/model"
)

// AuditCanonical checks whether canonical URL directives are well-formed and consistent.
func AuditCanonical(pageURL string, canonicalURL string, canonicalResult *model.CrawlResult) *model.AuditIssue {
	if canonicalURL == "" {
		return nil
	}

	pageLower := strings.ToLower(pageURL)
	canonLower := strings.ToLower(canonicalURL)

	// Flag HTTPS page with insecure HTTP canonical
	if strings.HasPrefix(pageLower, "https://") && strings.HasPrefix(canonLower, "http://") {
		return &model.AuditIssue{
			ID:           fmt.Sprintf("canon-mismatch-%s", pageURL),
			Type:         model.IssueCanonicalMismatch,
			Severity:     model.SeverityWarning,
			TargetURL:    canonicalURL,
			SourceURL:    pageURL,
			ErrorMessage: fmt.Sprintf("Insecure HTTP canonical URL specified on HTTPS page: %s", canonicalURL),
			AssetType:    model.AssetCanonical,
			Context:      "Canonical URL mismatch",
		}
	}

	// Flag broken canonical target
	if canonicalResult != nil && (canonicalResult.StatusCode >= 400 || canonicalResult.StatusCode == 0) {
		return &model.AuditIssue{
			ID:           fmt.Sprintf("canon-broken-%s", pageURL),
			Type:         model.IssueBrokenCanonical,
			Severity:     model.SeverityError,
			TargetURL:    canonicalURL,
			SourceURL:    pageURL,
			StatusCode:   canonicalResult.StatusCode,
			ErrorMessage: fmt.Sprintf("Canonical URL target resolves to HTTP %d error: %s", canonicalResult.StatusCode, canonicalURL),
			AssetType:    model.AssetCanonical,
			Context:      "Broken canonical URL",
		}
	}

	return nil
}

// AuditMetaRobots detects noindex or nofollow directives in HTML meta tags or HTTP headers.
func AuditMetaRobots(pageURL string, robotsMeta string, xRobotsTag string) *model.AuditIssue {
	combined := strings.ToLower(robotsMeta + " " + xRobotsTag)

	if strings.Contains(combined, "noindex") {
		directive := robotsMeta
		if xRobotsTag != "" {
			directive += fmt.Sprintf(" (X-Robots-Tag: %s)", xRobotsTag)
		}
		return &model.AuditIssue{
			ID:           fmt.Sprintf("robots-noindex-%s", pageURL),
			Type:         model.IssueNoIndexDirective,
			Severity:     model.SeverityInfo,
			TargetURL:    pageURL,
			SourceURL:    pageURL,
			ErrorMessage: fmt.Sprintf("Page declares noindex directive: %s", strings.TrimSpace(directive)),
			AssetType:    model.AssetLink,
			Context:      "Indexation blocked via robots directive",
		}
	}

	return nil
}

// DetectSoft404 identifies pages that return HTTP 200 OK but display an error or missing-page message.
func DetectSoft404(pageURL string, statusCode int, title string, body []byte) *model.AuditIssue {
	if statusCode != 200 {
		return nil
	}

	titleLower := strings.ToLower(title)

	// Check title clues
	isSoft404 := false
	var reason string

	if strings.Contains(titleLower, "404 not found") ||
		strings.Contains(titleLower, "page not found") ||
		titleLower == "404" ||
		strings.Contains(titleLower, "pagina nu a fost gasita") ||
		strings.Contains(titleLower, "error 404") {
		isSoft404 = true
		reason = fmt.Sprintf("Page title indicates 404 error despite HTTP 200 status: %q", title)
	} else if len(body) > 0 && len(body) < 180 {
		bodyLower := strings.ToLower(string(body))
		if strings.Contains(bodyLower, "not found") || strings.Contains(bodyLower, "404") {
			isSoft404 = true
			reason = "Minimal HTML body (<180 bytes) containing '404' or 'not found' text"
		}
	}

	if isSoft404 {
		return &model.AuditIssue{
			ID:           fmt.Sprintf("soft404-%s", pageURL),
			Type:         model.IssueSoft404,
			Severity:     model.SeverityWarning,
			TargetURL:    pageURL,
			SourceURL:    pageURL,
			StatusCode:   200,
			ErrorMessage: reason,
			AssetType:    model.AssetLink,
			Context:      "Soft 404 Detected",
		}
	}

	return nil
}
