import React, { useState, useMemo } from 'react';
import { ExternalLink, Search, ChevronLeft, ChevronRight } from 'lucide-react';
import { CrawlResult } from '../types';

interface AllUrlsTableProps {
  results: CrawlResult[];
}

export const AllUrlsTable: React.FC<AllUrlsTableProps> = ({ results }) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const pageSize = 25;

  const filtered = useMemo(() => {
    if (!searchTerm.trim()) return results;
    const term = searchTerm.toLowerCase();
    return results.filter(
      (r) =>
        r.url.toLowerCase().includes(term) ||
        r.statusText.toLowerCase().includes(term) ||
        (r.contentType && r.contentType.toLowerCase().includes(term))
    );
  }, [results, searchTerm]);

  const totalPages = Math.ceil(filtered.length / pageSize) || 1;
  const paged = useMemo(() => {
    const start = (currentPage - 1) * pageSize;
    return filtered.slice(start, start + pageSize);
  }, [filtered, currentPage]);

  return (
    <div className="bg-zinc-900/60 border border-zinc-800/80 rounded-xl overflow-hidden">
      <div className="p-4 border-b border-zinc-800 flex items-center justify-between">
        <div className="relative w-80">
          <Search className="w-4 h-4 text-zinc-500 absolute left-3 top-2.5" />
          <input
            type="text"
            value={searchTerm}
            onChange={(e) => {
              setSearchTerm(e.target.value);
              setCurrentPage(1);
            }}
            placeholder="Search all crawled URLs..."
            className="w-full bg-zinc-950 border border-zinc-800 rounded-lg pl-9 pr-3 py-1.5 text-xs text-zinc-200 placeholder-zinc-500 focus:outline-none focus:border-red-500/50"
          />
        </div>
        <div className="text-xs text-zinc-400">
          Total: <span className="font-semibold text-zinc-200">{results.length}</span> URLs
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-left border-collapse text-xs">
          <thead>
            <tr className="border-b border-zinc-800 bg-zinc-950/40 text-zinc-400 font-medium uppercase tracking-wider text-[10px]">
              <th className="py-2.5 px-4 w-16">Status</th>
              <th className="py-2.5 px-4">URL</th>
              <th className="py-2.5 px-4 w-28">Type</th>
              <th className="py-2.5 px-4 w-24">Scope</th>
              <th className="py-2.5 px-4 w-20">Latency</th>
              <th className="py-2.5 px-4 w-24">Inbound</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-zinc-800/60">
            {paged.length === 0 ? (
              <tr>
                <td colSpan={6} className="py-8 text-center text-zinc-500 text-sm">
                  No URLs match the query.
                </td>
              </tr>
            ) : (
              paged.map((res, i) => (
                <tr key={i} className="hover:bg-zinc-800/30 transition-colors">
                  <td className="py-2.5 px-4 font-mono font-semibold">
                    <span
                      className={
                        res.statusCode >= 200 && res.statusCode < 300
                          ? 'text-emerald-400'
                          : res.statusCode >= 300 && res.statusCode < 400
                          ? 'text-purple-400'
                          : res.statusCode >= 400
                          ? 'text-red-400'
                          : 'text-zinc-500'
                      }
                    >
                      {res.statusCode || 'ERR'}
                    </span>
                  </td>
                  <td className="py-2.5 px-4 font-mono max-w-md truncate" title={res.url}>
                    <a
                      href={res.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-zinc-200 hover:text-red-400 inline-flex items-center gap-1"
                    >
                      <span className="truncate">{res.url}</span>
                      <ExternalLink className="w-3 h-3 opacity-60 flex-shrink-0" />
                    </a>
                  </td>
                  <td className="py-2.5 px-4 text-zinc-400 text-[11px] truncate max-w-xs">
                    {res.contentType ? res.contentType.split(';')[0] : res.assetType}
                  </td>
                  <td className="py-2.5 px-4">
                    <span
                      className={`px-1.5 py-0.5 rounded text-[10px] font-semibold ${
                        res.isInternal
                          ? 'bg-blue-500/10 text-blue-400 border border-blue-500/20'
                          : 'bg-zinc-800 text-zinc-400 border border-zinc-700/60'
                      }`}
                    >
                      {res.isInternal ? 'Internal' : 'External'}
                    </span>
                  </td>
                  <td className="py-2.5 px-4 text-zinc-400 font-mono text-[11px]">{res.durationMs}ms</td>
                  <td className="py-2.5 px-4 text-zinc-400 font-mono text-[11px]">
                    {res.inboundSources ? `${res.inboundSources.length} links` : '—'}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {filtered.length > pageSize && (
        <div className="p-3 border-t border-zinc-800 flex items-center justify-between text-xs text-zinc-400">
          <div>
            Showing {(currentPage - 1) * pageSize + 1} to{' '}
            {Math.min(currentPage * pageSize, filtered.length)} of {filtered.length} URLs
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
