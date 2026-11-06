import React from 'react';
import { Layers, AlertTriangle, ShieldAlert, Hash, Zap, ArrowRightLeft } from 'lucide-react';
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

  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3.5 mb-6">
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
  );
};
