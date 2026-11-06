import React, { useState, useMemo } from 'react';
import { ExternalLink, Search, ChevronLeft, ChevronRight, Filter, Code2, Sparkles } from 'lucide-react';
import { AuditIssue, IssueSeverity } from '../types';

interface IssuesTableProps {
  issues: AuditIssue[];
}

export const IssuesTable: React.FC<IssuesTableProps> = ({ issues }) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedSeverity, setSelectedSeverity] = useState<string>('all');
  const [selectedType, setSelectedType] = useState<string>('all');
  const [selectedScope, setSelectedScope] = useState<string>('all');
  const [onlyNew, setOnlyNew] = useState<boolean>(false);
  const [currentPage, setCurrentPage] = useState(1);
  const pageSize = 20;

  const hasNewIssues = useMemo(() => issues.some((i) => i.isNew), [issues]);

  const filteredIssues = useMemo(() => {
    return issues.filter((issue) => {
      if (onlyNew && !issue.isNew) {
        return false;
      }
      if (selectedSeverity !== 'all' && issue.severity !== selectedSeverity) {
        return false;
      }
      if (selectedScope === 'internal' && !issue.isInternal) {
        return false;
      }
      if (selectedScope === 'external' && issue.isInternal) {
        return false;
      }
      if (selectedType !== 'all') {
        if (selectedType === 'broken' && issue.type !== 'broken_resource' && issue.type !== 'timeout') {
          return false;
        }
        if (selectedType === 'mixed' && !issue.type.startsWith('mixed_content')) {
          return false;
        }
        if (selectedType === 'anchor' && issue.type !== 'anchor_not_found') {
          return false;
        }
        if (selectedType === 'redirect' && issue.type !== 'redirect_loop' && issue.type !== 'excessive_redirects') {
          return false;
        }
        if (selectedType === 'seo' && issue.type !== 'broken_canonical' && issue.type !== 'canonical_mismatch' && issue.type !== 'soft_404' && issue.type !== 'noindex_directive') {
          return false;
        }
        if (selectedType === 'orphan' && issue.type !== 'orphan_page') {
          return false;
        }
      }
      if (searchTerm.trim()) {
        const term = searchTerm.toLowerCase();
        return (
          issue.targetUrl.toLowerCase().includes(term) ||
          issue.sourceUrl.toLowerCase().includes(term) ||
          issue.errorMessage.toLowerCase().includes(term) ||
          (issue.snippet && issue.snippet.toLowerCase().includes(term)) ||
          (issue.context && issue.context.toLowerCase().includes(term))
        );
      }
      return true;
    });
  }, [issues, searchTerm, selectedSeverity, selectedType, selectedScope, onlyNew]);

  const totalPages = Math.ceil(filteredIssues.length / pageSize) || 1;
  const pagedIssues = useMemo(() => {
    const start = (currentPage - 1) * pageSize;
    return filteredIssues.slice(start, start + pageSize);
  }, [filteredIssues, currentPage]);

  const getSeverityBadge = (severity: IssueSeverity, isNew?: boolean) => {
    return (
      <div className="flex items-center gap-1.5">
        {isNew && (
          <span className="px-1.5 py-0.5 rounded text-[9px] font-bold uppercase bg-red-600 text-white shadow-sm flex items-center gap-0.5">
            <Sparkles className="w-2.5 h-2.5" />
            NEW
          </span>
        )}
        {severity === 'critical' && (
          <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase bg-red-500/20 text-red-400 border border-red-500/30">
            Critical
          </span>
        )}
        {severity === 'error' && (
          <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase bg-orange-500/20 text-orange-400 border border-orange-500/30">
            Error
          </span>
        )}
        {severity === 'warning' && (
          <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase bg-amber-500/20 text-amber-400 border border-amber-500/30">
            Warning
          </span>
        )}
        {severity === 'info' && (
          <span className="px-2 py-0.5 rounded text-[10px] font-bold uppercase bg-zinc-700 text-zinc-300">
            Info
          </span>
        )}
      </div>
    );
  };

  const getAssetBadge = (type: string) => {
    return (
      <span className="px-1.5 py-0.5 rounded text-[10px] font-mono bg-zinc-800 text-zinc-300 border border-zinc-700/60">
        {type}
      </span>
    );
  };

  return (
    <div className="bg-zinc-900/60 border border-zinc-800/80 rounded-xl overflow-hidden">
      {/* Controls Bar */}
      <div className="p-4 border-b border-zinc-800 flex flex-col md:flex-row gap-3 items-center justify-between">
        <div className="relative w-full md:w-80">
          <Search className="w-4 h-4 text-zinc-500 absolute left-3 top-2.5" />
          <input
            type="text"
            value={searchTerm}
            onChange={(e) => {
              setSearchTerm(e.target.value);
              setCurrentPage(1);
            }}
            placeholder="Search URL, error, context or snippet..."
            className="w-full bg-zinc-950 border border-zinc-800 rounded-lg pl-9 pr-3 py-1.5 text-xs text-zinc-200 placeholder-zinc-500 focus:outline-none focus:border-red-500/50"
          />
        </div>

        <div className="flex flex-wrap items-center gap-2 w-full md:w-auto">
          {hasNewIssues && (
            <button
              onClick={() => {
                setOnlyNew(!onlyNew);
                setCurrentPage(1);
              }}
              className={`px-2.5 py-1.5 rounded-lg text-xs font-semibold flex items-center gap-1 transition-colors ${
                onlyNew
                  ? 'bg-red-600 text-white'
                  : 'bg-zinc-800 text-zinc-300 hover:bg-zinc-700 border border-zinc-700/60'
              }`}
            >
              <Sparkles className="w-3.5 h-3.5" />
              Only New Issues
            </button>
          )}

          <div className="flex items-center gap-1 text-xs text-zinc-400 mr-1">
            <Filter className="w-3.5 h-3.5" />
            <span>Filter:</span>
          </div>

          {/* Scope selector */}
          <select
            value={selectedScope}
            onChange={(e) => {
              setSelectedScope(e.target.value);
              setCurrentPage(1);
            }}
            className="bg-zinc-950 border border-zinc-800 text-zinc-300 text-xs rounded-lg px-2.5 py-1.5 focus:outline-none focus:border-red-500/50"
          >
            <option value="all">All Scopes</option>
            <option value="internal">Internal Only</option>
            <option value="external">External Only</option>
          </select>

          {/* Severity selector */}
          <select
            value={selectedSeverity}
            onChange={(e) => {
              setSelectedSeverity(e.target.value);
              setCurrentPage(1);
            }}
            className="bg-zinc-950 border border-zinc-800 text-zinc-300 text-xs rounded-lg px-2.5 py-1.5 focus:outline-none focus:border-red-500/50"
          >
            <option value="all">All Severities</option>
            <option value="critical">Critical</option>
            <option value="error">Error</option>
            <option value="warning">Warning</option>
            <option value="info">Info</option>
          </select>

          {/* Type selector */}
          <select
            value={selectedType}
            onChange={(e) => {
              setSelectedType(e.target.value);
              setCurrentPage(1);
            }}
            className="bg-zinc-950 border border-zinc-800 text-zinc-300 text-xs rounded-lg px-2.5 py-1.5 focus:outline-none focus:border-red-500/50"
          >
            <option value="all">All Types</option>
            <option value="broken">Broken Links & Assets</option>
            <option value="mixed">Mixed Content</option>
            <option value="anchor">Missing #Anchors</option>
            <option value="seo">SEO (Canonical, Soft 404, Robots)</option>
            <option value="orphan">Orphan Pages</option>
            <option value="redirect">Redirect Issues</option>
          </select>
        </div>
      </div>

      {/* Table */}
      <div className="overflow-x-auto">
        <table className="w-full text-left border-collapse text-xs">
          <thead>
            <tr className="border-b border-zinc-800 bg-zinc-950/40 text-zinc-400 font-medium uppercase tracking-wider text-[10px]">
              <th className="py-2.5 px-4 w-28">Severity</th>
              <th className="py-2.5 px-4 w-16">Status</th>
              <th className="py-2.5 px-4 w-24">Type</th>
              <th className="py-2.5 px-4">Target URL & Snippet</th>
              <th className="py-2.5 px-4">Source Page</th>
              <th className="py-2.5 px-4">Issue Details</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-zinc-800/60">
            {pagedIssues.length === 0 ? (
              <tr>
                <td colSpan={6} className="py-8 text-center text-zinc-500 text-sm">
                  No issues found matching your filters.
                </td>
              </tr>
            ) : (
              pagedIssues.map((issue) => (
                <tr key={issue.id} className="hover:bg-zinc-800/30 transition-colors">
                  <td className="py-3 px-4">{getSeverityBadge(issue.severity, issue.isNew)}</td>
                  <td className="py-3 px-4 font-mono font-semibold">
                    {issue.statusCode ? (
                      <span className={issue.statusCode >= 400 ? 'text-red-400' : 'text-zinc-300'}>
                        {issue.statusCode}
                      </span>
                    ) : (
                      <span className="text-zinc-500">ERR</span>
                    )}
                  </td>
                  <td className="py-3 px-4">{getAssetBadge(issue.assetType)}</td>
                  <td className="py-3 px-4 font-mono max-w-sm truncate" title={issue.targetUrl}>
                    <a
                      href={issue.targetUrl}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-zinc-200 hover:text-red-400 inline-flex items-center gap-1 font-semibold"
                    >
                      <span className="truncate">{issue.targetUrl}</span>
                      <ExternalLink className="w-3 h-3 opacity-60 flex-shrink-0" />
                    </a>

                    {/* HTML Snippet preview */}
                    {issue.snippet && (
                      <div className="mt-1 flex items-center gap-1 text-[11px] text-zinc-400">
                        <Code2 className="w-3 h-3 text-zinc-500 flex-shrink-0" />
                        <code className="bg-zinc-950 px-1.5 py-0.5 rounded border border-zinc-800 text-zinc-300 text-[10px] truncate max-w-xs font-mono">
                          {issue.snippet}
                        </code>
                      </div>
                    )}
                  </td>
                  <td className="py-3 px-4 font-mono max-w-xs truncate text-zinc-400" title={issue.sourceUrl}>
                    {issue.sourceUrl ? (
                      <a
                        href={issue.sourceUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="hover:text-zinc-200 inline-flex items-center gap-1"
                      >
                        <span className="truncate">{issue.sourceUrl}</span>
                        <ExternalLink className="w-3 h-3 opacity-60 flex-shrink-0" />
                      </a>
                    ) : (
                      '—'
                    )}
                  </td>
                  <td className="py-3 px-4">
                    <div className="text-zinc-300 font-medium">{issue.errorMessage}</div>
                    {issue.context && (
                      <div className="text-[11px] text-zinc-500 mt-0.5 truncate max-w-sm" title={issue.context}>
                        {issue.context}
                      </div>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination Footer */}
      {filteredIssues.length > pageSize && (
        <div className="p-3 border-t border-zinc-800 flex items-center justify-between text-xs text-zinc-400">
          <div>
            Showing {(currentPage - 1) * pageSize + 1} to{' '}
            {Math.min(currentPage * pageSize, filteredIssues.length)} of {filteredIssues.length} issues
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
              disabled={currentPage === 1}
              className="p-1 rounded bg-zinc-800 hover:bg-zinc-700 disabled:opacity-40"
            >
              <ChevronLeft className="w-4 h-4" />
            </button>
            <span>
              Page {currentPage} of {totalPages}
            </span>
            <button
              onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
              disabled={currentPage === totalPages}
              className="p-1 rounded bg-zinc-800 hover:bg-zinc-700 disabled:opacity-40"
            >
              <ChevronRight className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}
    </div>
  );
};
