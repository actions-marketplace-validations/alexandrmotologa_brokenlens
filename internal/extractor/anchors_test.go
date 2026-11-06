package extractor

import (
	"testing"

	"github.com/alexandrmotologa/brokenlens/internal/model"
)

func TestValidateAnchor(t *testing.T) {
	domFragments := map[string]bool{
		"section-1":   true,
		"pricing-faq": true,
		"HeaderTitle": true,
	}

	if !ValidateAnchor("section-1", domFragments) {
		t.Error("Expected exact match for 'section-1'")
	}

	if !ValidateAnchor("headertitle", domFragments) {
		t.Error("Expected case-insensitive match for 'headertitle'")
	}

	if ValidateAnchor("non-existent-section", domFragments) {
		t.Error("Did not expect match for 'non-existent-section'")
	}

	if !ValidateAnchor("", domFragments) {
		t.Error("Empty fragment should be considered valid")
	}
}

func TestCheckMixedContent(t *testing.T) {
	// Active mixed content: HTTP script on HTTPS page
	issueActive := CheckMixedContent("https://secure.example.com/page", "http://cdn.example.com/lib.js", model.AssetScript)
	if issueActive == nil {
		t.Fatal("Expected active mixed content issue, got nil")
	}
	if issueActive.Severity != model.SeverityCritical {
		t.Errorf("Expected Critical severity for active mixed content, got %s", issueActive.Severity)
	}

	// Passive mixed content: HTTP image on HTTPS page
	issuePassive := CheckMixedContent("https://secure.example.com/page", "http://cdn.example.com/photo.jpg", model.AssetImage)
	if issuePassive == nil {
		t.Fatal("Expected passive mixed content issue, got nil")
	}
	if issuePassive.Severity != model.SeverityWarning {
		t.Errorf("Expected Warning severity for passive mixed content, got %s", issuePassive.Severity)
	}

	// Insecure parent page (plain HTTP): should NOT trigger mixed content
	issueHTTPParent := CheckMixedContent("http://plain.example.com/page", "http://cdn.example.com/photo.jpg", model.AssetImage)
	if issueHTTPParent != nil {
		t.Errorf("Did not expect mixed content alert for plain HTTP parent, got %v", issueHTTPParent)
	}

	// Secure resource on HTTPS: should NOT trigger mixed content
	issueHTTPS := CheckMixedContent("https://secure.example.com/page", "https://cdn.example.com/photo.jpg", model.AssetImage)
	if issueHTTPS != nil {
		t.Errorf("Did not expect mixed content alert for secure resource, got %v", issueHTTPS)
	}
}
