package crawler

import (
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	base := "https://example.com/blog/article-1"

	tests := []struct {
		name         string
		ref          string
		expectedURL  string
		expectedFrag string
		expectErr    bool
	}{
		{
			name:         "Relative path",
			ref:          "../about",
			expectedURL:  "https://example.com/about",
			expectedFrag: "",
		},
		{
			name:         "Relative root path",
			ref:          "/pricing#tiers",
			expectedURL:  "https://example.com/pricing",
			expectedFrag: "tiers",
		},
		{
			name:         "Full URL with fragment and default port",
			ref:          "HTTPS://EXAMPLE.COM:443/docs#setup",
			expectedURL:  "https://example.com/docs",
			expectedFrag: "setup",
		},
		{
			name:        "Ignored scheme mailto",
			ref:         "mailto:team@example.com",
			expectedURL: "",
		},
		{
			name:        "Ignored scheme javascript",
			ref:         "javascript:void(0)",
			expectedURL: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			norm, frag, err := NormalizeURL(tt.ref, base)
			if (err != nil) != tt.expectErr {
				t.Fatalf("NormalizeURL error = %v, expectErr %v", err, tt.expectErr)
			}
			if norm != tt.expectedURL {
				t.Errorf("NormalizeURL = %q, expected %q", norm, tt.expectedURL)
			}
			if frag != tt.expectedFrag {
				t.Errorf("Fragment = %q, expected %q", frag, tt.expectedFrag)
			}
		})
	}
}

func TestIsSameDomain(t *testing.T) {
	if !IsSameDomain("https://example.com/sub/page", "example.com") {
		t.Error("Expected example.com to match example.com")
	}

	if !IsSameDomain("https://example.com:8080/page", "example.com:9000") {
		t.Error("Expected hostnames to match ignoring ports")
	}

	if IsSameDomain("https://otherdomain.com/page", "example.com") {
		t.Error("Did not expect otherdomain.com to match example.com")
	}
}
