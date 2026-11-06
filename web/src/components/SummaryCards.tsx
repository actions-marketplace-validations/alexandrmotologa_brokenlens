import React from 'react';
import { Layers, AlertTriangle, ShieldAlert, Hash, Zap, ArrowRightLeft, FileQuestion, SearchX, Sparkles } from 'lucide-react';
import { CrawlSummary, LiveProgress } from '../types';

interface SummaryCardsProps {
  summary: CrawlSummary | null;
  progress: LiveProgress | null;
}

export const SummaryCards: React.FC<SummaryCardsProps> = ({ summary, progress }) => {
  const isCrawling = progress?.isCrawling;
  const totalScanned = isCrawling ? progress.totalScanned : summary?.totalScanned ?? 0;
  const totalBroken = isCrawling ? progress.totalBroken : summary?.totalBroken ?? 0;
  const mixedContent = summary?.totalMixedContent ?? 0;
  const missingAnchors = summary?.totalMissingAnchors ?? 0;
  const redirects = summary?.totalRedirects ?? 0;
  const rps = isCrawling ? progress.requestsPerSec : summary?.requestsPerSecond ?? 0;
  const orphans = summary?.totalOrphanPages ?? 0;
  const soft404 = summary?.totalSoft404 ?? 0;
  const diff = summary?.diff;

  return (
    <div className="space-y-3.5 mb-6">
      {/* Differential PR Banner if diff is active */}
      {diff && (
        <div className="bg-zinc-900/80 border border-zinc-800 rounded-xl p-4 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 shadow-sm">
          <div className="flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-lg bg-red-500/10 border border-red-500/20 flex items-center justify-center text-red-400">
              <Sparkles className="w-4 h-4" />
            </div>
            <div>
              <div className="text-xs font-semibold text-zinc-200">
                Differential PR Audit vs <code className="text-zinc-400 font-mono text-[11px]">{diff.baselineReportPath}</code>
              </div>
              <div className="text-[11px] text-zinc-500">
                Comparing current pull request against baseline scan
              </div>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <span className="px-2.5 py-1 rounded-full text-xs font-semibold bg-red-500/20 text-red-400 border border-red-500/30">
              +{diff.newIssuesCount} NEW
            </span>
            <span className="px-2.5 py-1 rounded-full text-xs font-semibold bg-emerald-500/20 text-emerald-400 border border-emerald-500/30">
              -{diff.resolvedIssuesCount} FIXED
            </span>
            <span className="px-2.5 py-1 rounded-full text-xs font-semibold bg-zinc-800 text-zinc-400 border border-zinc-700/60">
              {diff.persistedIssuesCount} PERSISTED
            </span>
          </div>
        </div>
      )}

      {/* Main KPI Grid */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3.5">
        {/* Total Scanned */}
        <div className="bg-zinc-900/60 border border-zinc-800/80 rounded-xl p-4 flex flex-col justify-between">
          <div className="flex items-center justify-between text-zinc-400 mb-2">
            <span className="text-xs font-medium uppercase tracking-wider">Scanned URLs</span>
            <Layers className="w-4 h-4 text-zinc-500" />
          </div>
          <div>
            <div className="text-2xl font-bold tracking-tight text-zinc-100">{totalScanned}</div>
            <div className="text-[11px] text-zinc-500 mt-0.5">
              {summary ? `${summary.totalInternal} int / ${summary.totalExternal} ext` : 'Queued: ' + (progress?.totalQueued ?? 0)}
            </div>
          </div>
        </div>

        {/* Broken Resources */}
        <div className="bg-zinc-900/60 border border-zinc-800/80 rounded-xl p-4 flex flex-col justify-between">
          <div className="flex items-center justify-between text-zinc-400 mb-2">
            <span className="text-xs font-medium uppercase tracking-wider">Broken Items</span>
            <AlertTriangle className="w-4 h-4 text-red-400" />
          </div>
          <div>
            <div className={`text-2xl font-bold tracking-tight ${totalBroken > 0 ? 'text-red-400' : 'text-zinc-100'}`}>
              {totalBroken}
            </div>
            <div className="text-[11px] text-zinc-500 mt-0.5">4xx, 5xx, or timeouts</div>
          </div>
        </div>

        {/* Mixed Content */}
        <div className="bg-zinc-900/60 border border-zinc-800/80 rounded-xl p-4 flex flex-col justify-between">
          <div className="flex items-center justify-between text-zinc-400 mb-2">
            <span className="text-xs font-medium uppercase tracking-wider">Mixed Content</span>
            <ShieldAlert className="w-4 h-4 text-amber-400" />
          </div>
          <div>
            <div className={`text-2xl font-bold tracking-tight ${mixedContent > 0 ? 'text-amber-400' : 'text-zinc-100'}`}>
              {mixedContent}
            </div>
            <div className="text-[11px] text-zinc-500 mt-0.5">HTTP on HTTPS pages</div>
          </div>
        </div>

        {/* Missing In-Page Anchors */}
        <div className="bg-zinc-900/60 border border-zinc-800/80 rounded-xl p-4 flex flex-col justify-between">
          <div className="flex items-center justify-between text-zinc-400 mb-2">
            <span className="text-xs font-medium uppercase tracking-wider">Dead Anchors</span>
            <Hash className="w-4 h-4 text-sky-400" />
          </div>
          <div>
            <div className={`text-2xl font-bold tracking-tight ${missingAnchors > 0 ? 'text-sky-400' : 'text-zinc-100'}`}>
              {missingAnchors}
            </div>
            <div className="text-[11px] text-zinc-500 mt-0.5">Missing #id targets</div>
          </div>
        </div>

        {/* Redirect Chains */}
        <div className="bg-zinc-900/60 border border-zinc-800/80 rounded-xl p-4 flex flex-col justify-between">
          <div className="flex items-center justify-between text-zinc-400 mb-2">
            <span className="text-xs font-medium uppercase tracking-wider">Redirects</span>
            <ArrowRightLeft className="w-4 h-4 text-purple-400" />
          </div>
          <div>
            <div className="text-2xl font-bold tracking-tight text-zinc-100">{redirects}</div>
            <div className="text-[11px] text-zinc-500 mt-0.5">301 / 302 hops</div>
          </div>
        </div>

        {/* Throughput */}
        <div className="bg-zinc-900/60 border border-zinc-800/80 rounded-xl p-4 flex flex-col justify-between">
          <div className="flex items-center justify-between text-zinc-400 mb-2">
            <span className="text-xs font-medium uppercase tracking-wider">Throughput</span>
            <Zap className="w-4 h-4 text-emerald-400" />
          </div>
          <div>
            <div className="text-2xl font-bold tracking-tight text-emerald-400">
              {rps.toFixed(1)} <span className="text-xs font-normal text-zinc-400">req/s</span>
            </div>
            <div className="text-[11px] text-zinc-500 mt-0.5">
              {summary ? `${(summary.durationMs / 1000).toFixed(1)}s elapsed` : 'Active workers: ' + (progress?.activeWorkers ?? 0)}
            </div>
          </div>
        </div>
      </div>

      {/* SEO Sub-chips (Orphans, Soft 404s, Canonicals) */}
      {summary && (orphans > 0 || soft404 > 0 || summary.totalCanonicalIssues > 0) && (
        <div className="flex flex-wrap items-center gap-2 pt-1">
          {orphans > 0 && (
            <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-lg text-xs bg-zinc-900/90 border border-zinc-800 text-zinc-300">
              <FileQuestion className="w-3.5 h-3.5 text-amber-400" />
              <span>
                Orphan Pages: <strong className="text-amber-400 font-mono">{orphans}</strong>
              </span>
            </div>
          )}
          {soft404 > 0 && (
            <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-lg text-xs bg-zinc-900/90 border border-zinc-800 text-zinc-300">
              <SearchX className="w-3.5 h-3.5 text-orange-400" />
              <span>
                Soft 404s: <strong className="text-orange-400 font-mono">{soft404}</strong>
              </span>
            </div>
          )}
          {summary.totalCanonicalIssues > 0 && (
            <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-lg text-xs bg-zinc-900/90 border border-zinc-800 text-zinc-300">
              <span>
                Canonical Issues: <strong className="text-sky-400 font-mono">{summary.totalCanonicalIssues}</strong>
              </span>
            </div>
          )}
        </div>
      )}
    </div>
  );
};
