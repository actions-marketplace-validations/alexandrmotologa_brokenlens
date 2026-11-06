import React, { useState, useEffect } from 'react';
import { Header } from './components/Header';
import { SummaryCards } from './components/SummaryCards';
import { IssuesTable } from './components/IssuesTable';
import { StatusChart } from './components/StatusChart';
import { RedirectTree } from './components/RedirectTree';
import { AllUrlsTable } from './components/AllUrlsTable';
import { AuditReport, LiveProgress } from './types';
import { AlertTriangle, BarChart3, ArrowRightLeft, ListFilter, Activity } from 'lucide-react';

export const App: React.FC = () => {
  const [report, setReport] = useState<AuditReport | null>(null);
  const [progress, setProgress] = useState<LiveProgress | null>(null);
  const [activeTab, setActiveTab] = useState<'issues' | 'status' | 'redirects' | 'all'>('issues');
  const [loading, setLoading] = useState(true);

  // Fetch report on mount
  useEffect(() => {
    fetchReport();
    setupSSE();
  }, []);

  const fetchReport = async () => {
    try {
      const res = await fetch('/api/report');
      if (res.ok) {
        const data = await res.json();
        setReport(data);
      }
    } catch (e) {
      console.log('No existing report loaded or server offline', e);
    } finally {
      setLoading(false);
    }
  };

  const setupSSE = () => {
    const eventSource = new EventSource('/api/events');

    eventSource.onmessage = (event) => {
      try {
        const payload = JSON.parse(event.data);
        if (payload.type === 'progress') {
          setProgress({
            totalScanned: payload.totalScanned,
            totalQueued: payload.totalQueued,
            totalBroken: payload.totalBroken,
            activeWorkers: payload.activeWorkers,
            currentUrl: payload.currentUrl,
            durationMs: payload.durationMs,
            requestsPerSec: payload.requestsPerSec,
            isCrawling: true,
          });
        } else if (payload.type === 'done') {
          setProgress(null);
          setReport(payload.report);
        }
      } catch (err) {
        console.error('Failed to parse SSE event:', err);
      }
    };

    eventSource.onerror = () => {
      // EventSource reconnects automatically
    };

    return () => {
      eventSource.close();
    };
  };

  const handleStartScan = async (url: string) => {
    try {
      setProgress({
        totalScanned: 0,
        totalQueued: 1,
        totalBroken: 0,
        activeWorkers: 1,
        currentUrl: url,
        durationMs: 0,
        requestsPerSec: 0,
        isCrawling: true,
      });

      const res = await fetch('/api/scan', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url }),
      });

      if (!res.ok) {
        const errData = await res.json();
        alert(`Failed to start audit: ${errData.error || res.statusText}`);
        setProgress(null);
      }
    } catch (e) {
      alert(`Network error starting audit: ${e}`);
      setProgress(null);
    }
  };

  const handleExportJSON = () => {
    if (!report) return;
    const blob = new Blob([JSON.stringify(report, null, 2)], { type: 'application/json' });
    const u = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = u;
    a.download = `brokenlens-${new Date().toISOString().slice(0, 10)}.json`;
    a.click();
    URL.revokeObjectURL(u);
  };

  const handleExportCSV = () => {
    if (!report) return;
    const headers = ['ID', 'Severity', 'Type', 'TargetURL', 'StatusCode', 'SourceURL', 'AssetType', 'Context', 'ErrorMessage'];
    const rows = report.issues.map((i) => [
      i.id,
      i.severity,
      i.type,
      `"${i.targetUrl.replace(/"/g, '""')}"`,
      i.statusCode || '',
      `"${i.sourceUrl.replace(/"/g, '""')}"`,
      i.assetType,
      `"${(i.context || '').replace(/"/g, '""')}"`,
      `"${i.errorMessage.replace(/"/g, '""')}"`,
    ]);

    const csvContent = [headers.join(','), ...rows.map((r) => r.join(','))].join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv' });
    const u = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = u;
    a.download = `brokenlens-issues-${new Date().toISOString().slice(0, 10)}.csv`;
    a.click();
    URL.revokeObjectURL(u);
  };

  return (
    <div className="min-h-screen bg-zinc-950 text-zinc-100 flex flex-col font-sans">
      <Header
        report={report}
        progress={progress}
        onScan={handleStartScan}
        onExportJSON={handleExportJSON}
        onExportCSV={handleExportCSV}
      />

      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6">
        {/* Live Crawl Progress Banner */}
        {progress?.isCrawling && (
          <div className="bg-red-500/10 border border-red-500/30 rounded-xl p-4 mb-6 animate-pulse">
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2 text-red-400 font-semibold text-sm">
                <Activity className="w-4 h-4 animate-spin" />
                Live Audit in Progress
              </div>
              <div className="text-xs font-mono text-zinc-400">
                {progress.totalScanned} scanned • {progress.totalBroken} broken • {progress.requestsPerSec.toFixed(1)} req/s
              </div>
            </div>
            <div className="text-xs font-mono text-zinc-300 truncate">
              Inspecting: <span className="text-zinc-100">{progress.currentUrl}</span>
            </div>
          </div>
        )}

        {/* High-Level KPI Metric Cards */}
        <SummaryCards summary={report?.summary ?? null} progress={progress} />

        {/* Status Chart */}
        {report && (
          <StatusChart
            distribution={report.summary.statusCodeDistribution}
            totalScanned={report.summary.totalScanned}
          />
        )}

        {/* Tab Navigation */}
        <div className="flex items-center gap-1 border-b border-zinc-800 mb-5 text-xs font-medium">
          <button
            onClick={() => setActiveTab('issues')}
            className={`flex items-center gap-2 px-4 py-2.5 border-b-2 transition-colors ${
              activeTab === 'issues'
                ? 'border-red-500 text-red-400 font-semibold'
                : 'border-transparent text-zinc-400 hover:text-zinc-200'
            }`}
          >
            <AlertTriangle className="w-4 h-4" />
            Detected Issues
            {report && report.issues.length > 0 && (
              <span className="px-1.5 py-0.2 rounded-full text-[10px] bg-red-500/20 text-red-400 font-mono">
                {report.issues.length}
              </span>
            )}
          </button>

          <button
            onClick={() => setActiveTab('redirects')}
            className={`flex items-center gap-2 px-4 py-2.5 border-b-2 transition-colors ${
              activeTab === 'redirects'
                ? 'border-red-500 text-red-400 font-semibold'
                : 'border-transparent text-zinc-400 hover:text-zinc-200'
            }`}
          >
            <ArrowRightLeft className="w-4 h-4" />
            Redirect Chains
            {report && report.redirectChains.length > 0 && (
              <span className="px-1.5 py-0.2 rounded-full text-[10px] bg-purple-500/20 text-purple-400 font-mono">
                {report.redirectChains.length}
              </span>
            )}
          </button>

          <button
            onClick={() => setActiveTab('all')}
            className={`flex items-center gap-2 px-4 py-2.5 border-b-2 transition-colors ${
              activeTab === 'all'
                ? 'border-red-500 text-red-400 font-semibold'
                : 'border-transparent text-zinc-400 hover:text-zinc-200'
            }`}
          >
            <ListFilter className="w-4 h-4" />
            All Crawled URLs
            {report && report.results && (
              <span className="px-1.5 py-0.2 rounded-full text-[10px] bg-zinc-800 text-zinc-400 font-mono">
                {report.results.length}
              </span>
            )}
          </button>
        </div>

        {/* Tab Contents */}
        {activeTab === 'issues' && <IssuesTable issues={report?.issues ?? []} />}
        {activeTab === 'redirects' && <RedirectTree chains={report?.redirectChains ?? []} />}
        {activeTab === 'all' && <AllUrlsTable results={report?.results ?? []} />}
      </main>
    </div>
  );
};
