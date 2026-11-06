import React from 'react';
import { AlertCircle, CheckCircle2, ShieldAlert, Download, RefreshCw, Globe, Play } from 'lucide-react';
import { AuditReport, LiveProgress } from '../types';

interface HeaderProps {
  report: AuditReport | null;
  progress: LiveProgress | null;
  onScan: (url: string) => void;
  onExportJSON: () => void;
  onExportCSV: () => void;
}

export const Header: React.FC<HeaderProps> = ({
  report,
  progress,
  onScan,
  onExportJSON,
  onExportCSV,
}) => {
  const [targetInput, setTargetInput] = React.useState(report?.summary.targetUrl || '');

  React.useEffect(() => {
    if (report?.summary.targetUrl && !targetInput) {
      setTargetInput(report.summary.targetUrl);
    }
  }, [report?.summary.targetUrl]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (targetInput.trim()) {
      onScan(targetInput.trim());
    }
  };

  const healthScore = report?.summary.healthScore ?? 100;
  const getHealthBadge = () => {
    if (healthScore >= 90) {
      return (
        <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
          <CheckCircle2 className="w-3.5 h-3.5" />
          Health {healthScore.toFixed(1)}%
        </span>
      );
    }
    if (healthScore >= 70) {
      return (
        <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/20">
          <AlertCircle className="w-3.5 h-3.5" />
          Health {healthScore.toFixed(1)}%
        </span>
      );
    }
    return (
      <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-red-500/10 text-red-400 border border-red-500/20">
        <ShieldAlert className="w-3.5 h-3.5" />
        Health {healthScore.toFixed(1)}%
      </span>
    );
  };

  return (
    <header className="border-b border-zinc-800 bg-zinc-900/60 backdrop-blur sticky top-0 z-30">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-3.5 flex flex-col md:flex-row items-center justify-between gap-4">
        {/* Brand */}
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-lg bg-red-500/10 border border-red-500/30 flex items-center justify-center text-red-400 shadow-sm shadow-red-950">
            <Globe className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="font-bold text-base tracking-tight text-zinc-100">brokenlens</span>
              <span className="text-[10px] font-mono uppercase tracking-wider px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-400 border border-zinc-700/60">
                v1.0
              </span>
            </div>
            <p className="text-xs text-zinc-400 hidden sm:block">
              Web Quality, Dead Link & Mixed-Content Auditor
            </p>
          </div>
        </div>

        {/* Scan Input */}
        <form onSubmit={handleSubmit} className="flex-1 max-w-xl w-full flex items-center gap-2">
          <div className="relative flex-1">
            <input
              type="text"
              value={targetInput}
              onChange={(e) => setTargetInput(e.target.value)}
              placeholder="Enter site URL (e.g. https://example.com)..."
              disabled={progress?.isCrawling}
              className="w-full bg-zinc-950 border border-zinc-800 rounded-lg px-3.5 py-1.5 text-sm text-zinc-200 placeholder-zinc-500 focus:outline-none focus:border-red-500/50 focus:ring-1 focus:ring-red-500/50 disabled:opacity-50"
            />
          </div>
          <button
            type="submit"
            disabled={progress?.isCrawling || !targetInput.trim()}
            className="inline-flex items-center gap-1.5 px-4 py-1.5 bg-red-600 hover:bg-red-500 disabled:bg-zinc-800 disabled:text-zinc-500 text-white text-xs font-semibold rounded-lg shadow transition-colors"
          >
            {progress?.isCrawling ? (
              <>
                <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                Auditing...
              </>
            ) : (
              <>
                <Play className="w-3.5 h-3.5 fill-current" />
                Start Audit
              </>
            )}
          </button>
        </form>

        {/* Meta & Exports */}
        <div className="flex items-center gap-2.5">
          {report && getHealthBadge()}
          <div className="flex items-center gap-1 border-l border-zinc-800 pl-2.5">
            <button
              onClick={onExportJSON}
              disabled={!report}
              title="Download JSON Report"
              className="p-1.5 rounded-lg text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 transition-colors disabled:opacity-40"
            >
              <Download className="w-4 h-4" />
            </button>
            <button
              onClick={onExportCSV}
              disabled={!report}
              title="Download CSV of Issues"
              className="text-xs font-medium px-2 py-1 rounded-lg text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 transition-colors disabled:opacity-40"
            >
              CSV
            </button>
          </div>
        </div>
      </div>
    </header>
  );
};
