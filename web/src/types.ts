export type AssetType =
  | 'link'
  | 'image'
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
  | 'invalid_url';

export type IssueSeverity = 'critical' | 'error' | 'warning' | 'info';

export interface InboundReference {
  sourceUrl: string;
  assetType: AssetType;
  text?: string;
  fragment?: string;
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
  healthScore: number;
  statusCodeDistribution: Record<string, number>;
  assetTypeDistribution: Record<string, number>;
  requestsPerSecond: number;
}

export interface AuditReport {
  summary: CrawlSummary;
  issues: AuditIssue[];
  redirectChains: RedirectChain[];
  results?: CrawlResult[];
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
