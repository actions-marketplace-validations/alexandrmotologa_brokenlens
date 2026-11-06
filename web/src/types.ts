export type AssetType =
  | 'link'
  | 'image'
  | 'social_image'
  | 'script'
  | 'stylesheet'
  | 'preload'
  | 'media'
  | 'css_url'
  | 'canonical'
  | 'favicon';

export type IssueType =
  | 'broken_resource'
  | 'timeout'
  | 'redirect_loop'
  | 'excessive_redirects'
  | 'anchor_not_found'
  | 'mixed_content_active'
  | 'mixed_content_passive'
  | 'invalid_url'
  | 'broken_canonical'
  | 'canonical_mismatch'
  | 'noindex_directive'
  | 'soft_404'
  | 'orphan_page';

export type IssueSeverity = 'critical' | 'error' | 'warning' | 'info';

export interface InboundReference {
  sourceUrl: string;
  assetType: AssetType;
  text?: string;
  fragment?: string;
  snippet?: string;
}

export interface RedirectHop {
  fromUrl: string;
  toUrl: string;
  statusCode: number;
}

export interface RedirectChain {
  initialUrl: string;
  finalUrl: string;
  hops: RedirectHop[];
  isLoop?: boolean;
}

export interface CrawlResult {
  url: string;
  requestedUrl: string;
  finalUrl: string;
  statusCode: number;
  statusText: string;
  durationMs: number;
  contentType: string;
  contentLength: number;
  isInternal: boolean;
  assetType: AssetType;
  redirectChain?: RedirectHop[];
  error?: string;
  inboundSources?: InboundReference[];
  extractedLinks: number;
  extractedAssets: number;
  robotsMeta?: string;
  canonicalUrl?: string;
  pageTitle?: string;
}

export interface AuditIssue {
  id: string;
  type: IssueType;
  severity: IssueSeverity;
  targetUrl: string;
  sourceUrl: string;
  statusCode?: number;
  errorMessage: string;
  assetType: AssetType;
  context?: string;
  snippet?: string;
  isInternal?: boolean;
  isNew?: boolean;
}

export interface DiffSummary {
  baselineReportPath: string;
  newIssuesCount: number;
  resolvedIssuesCount: number;
  persistedIssuesCount: number;
}

export interface DiffReport {
  baselineReportPath: string;
  newIssues: AuditIssue[];
  resolvedIssues: AuditIssue[];
  persistedIssues: AuditIssue[];
}

export interface CrawlSummary {
  targetUrl: string;
  startTime: string;
  endTime: string;
  durationMs: number;
  totalScanned: number;
  totalUnique: number;
  totalInternal: number;
  totalExternal: number;
  totalBroken: number;
  totalRedirects: number;
  totalMixedContent: number;
  totalMissingAnchors: number;
  totalOrphanPages: number;
  totalSoft404: number;
  totalCanonicalIssues: number;
  healthScore: number;
  statusCodeDistribution: Record<string, number>;
  assetTypeDistribution: Record<string, number>;
  requestsPerSecond: number;
  diff?: DiffSummary;
}

export interface AuditReport {
  summary: CrawlSummary;
  issues: AuditIssue[];
  redirectChains: RedirectChain[];
  results?: CrawlResult[];
  diff?: DiffReport;
}

export interface LiveProgress {
  totalScanned: number;
  totalQueued: number;
  totalBroken: number;
  activeWorkers: number;
  currentUrl: string;
  durationMs: number;
  requestsPerSec: number;
  isCrawling: boolean;
}
