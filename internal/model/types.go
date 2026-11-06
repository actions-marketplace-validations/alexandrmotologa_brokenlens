package model

import "time"

// AssetType identifies the category of an extracted HTML resource.
type AssetType string

const (
	AssetLink       AssetType = "link"
	AssetImage      AssetType = "image"
	AssetScript     AssetType = "script"
	AssetStylesheet AssetType = "stylesheet"
	AssetPreload    AssetType = "preload"
	AssetMedia      AssetType = "media"
	AssetCSS        AssetType = "css_url"
	AssetCanonical  AssetType = "canonical"
	AssetFavicon    AssetType = "favicon"
)

// IssueType identifies specific failure classifications.
type IssueType string

const (
	IssueBroken             IssueType = "broken_resource"
	IssueTimeout            IssueType = "timeout"
	IssueRedirectLoop       IssueType = "redirect_loop"
	IssueExcessiveRedirects IssueType = "excessive_redirects"
	IssueAnchorNotFound     IssueType = "anchor_not_found"
	IssueMixedContentActive IssueType = "mixed_content_active"
	IssueMixedContentPassive IssueType = "mixed_content_passive"
	IssueInvalidURL         IssueType = "invalid_url"
)

// IssueSeverity indicates how severe an issue is.
type IssueSeverity string

const (
	SeverityCritical IssueSeverity = "critical"
	SeverityError    IssueSeverity = "error"
	SeverityWarning  IssueSeverity = "warning"
	SeverityInfo     IssueSeverity = "info"
)

// InboundReference stores information about where an asset was referenced.
type InboundReference struct {
	SourceURL string    `json:"sourceUrl"`
	AssetType AssetType `json:"assetType"`
	Text      string    `json:"text,omitempty"`
	Fragment  string    `json:"fragment,omitempty"`
}

// RedirectHop represents a single hop in an HTTP redirection chain.
type RedirectHop struct {
	FromURL    string `json:"fromUrl"`
	ToURL      string `json:"toUrl"`
	StatusCode int    `json:"statusCode"`
}

// RedirectChain contains all hops from an initial URL to its final destination.
type RedirectChain struct {
	InitialURL string        `json:"initialUrl"`
	FinalURL   string        `json:"finalUrl"`
	Hops       []RedirectHop `json:"hops"`
	IsLoop     bool          `json:"isLoop"`
}

// CrawlResult holds inspection data for a single visited URL.
type CrawlResult struct {
	URL             string             `json:"url"`
	RequestedURL    string             `json:"requestedUrl"`
	FinalURL        string             `json:"finalUrl"`
	StatusCode      int                `json:"statusCode"`
	StatusText      string             `json:"statusText"`
	DurationMs      int64              `json:"durationMs"`
	ContentType     string             `json:"contentType"`
	ContentLength   int64              `json:"contentLength"`
	IsInternal      bool               `json:"isInternal"`
	AssetType       AssetType          `json:"assetType"`
	RedirectChain   []RedirectHop      `json:"redirectChain,omitempty"`
	Error           string             `json:"error,omitempty"`
	InboundSources  []InboundReference `json:"inboundSources,omitempty"`
	ExtractedLinks  int                `json:"extractedLinks"`
	ExtractedAssets int                `json:"extractedAssets"`
	DOMFragments    map[string]bool    `json:"-"` // Internal set of IDs and name attributes found
}

// AuditIssue represents a detected problem during the crawl.
type AuditIssue struct {
	ID           string        `json:"id"`
	Type         IssueType     `json:"type"`
	Severity     IssueSeverity `json:"severity"`
	TargetURL    string        `json:"targetUrl"`
	SourceURL    string        `json:"sourceUrl"`
	StatusCode   int           `json:"statusCode,omitempty"`
	ErrorMessage string        `json:"errorMessage"`
	AssetType    AssetType     `json:"assetType"`
	Context      string        `json:"context,omitempty"`
}

// CrawlSummary provides aggregate metrics for the entire audit session.
type CrawlSummary struct {
	TargetURL              string         `json:"targetUrl"`
	StartTime              time.Time      `json:"startTime"`
	EndTime                time.Time      `json:"endTime"`
	DurationMs             int64          `json:"durationMs"`
	TotalScanned           int            `json:"totalScanned"`
	TotalUnique            int            `json:"totalUnique"`
	TotalInternal          int            `json:"totalInternal"`
	TotalExternal          int            `json:"totalExternal"`
	TotalBroken            int            `json:"totalBroken"`
	TotalRedirects         int            `json:"totalRedirects"`
	TotalMixedContent      int            `json:"totalMixedContent"`
	TotalMissingAnchors    int            `json:"totalMissingAnchors"`
	HealthScore            float64        `json:"healthScore"`
	StatusCodeDistribution map[string]int `json:"statusCodeDistribution"`
	AssetTypeDistribution  map[string]int `json:"assetTypeDistribution"`
	RequestsPerSecond      float64        `json:"requestsPerSecond"`
}

// AuditReport is the full serialized report data model.
type AuditReport struct {
	Summary        CrawlSummary    `json:"summary"`
	Issues         []AuditIssue    `json:"issues"`
	RedirectChains []RedirectChain `json:"redirectChains"`
	Results        []CrawlResult   `json:"results,omitempty"`
}
