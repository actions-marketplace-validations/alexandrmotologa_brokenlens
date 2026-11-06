import React from 'react';

interface StatusChartProps {
  distribution: Record<string, number>;
  totalScanned: number;
}

export const StatusChart: React.FC<StatusChartProps> = ({ distribution, totalScanned }) => {
  const buckets = [
    { key: '2xx', label: '2xx Success', color: 'bg-emerald-500', textColor: 'text-emerald-400' },
    { key: '3xx', label: '3xx Redirect', color: 'bg-purple-500', textColor: 'text-purple-400' },
    { key: '4xx', label: '4xx Client Error', color: 'bg-red-500', textColor: 'text-red-400' },
    { key: '5xx', label: '5xx Server Error', color: 'bg-orange-500', textColor: 'text-orange-400' },
    { key: 'error', label: 'Network Errors', color: 'bg-zinc-600', textColor: 'text-zinc-400' },
  ];

  const total = totalScanned || Object.values(distribution).reduce((a, b) => a + b, 0) || 1;

  return (
    <div className="bg-zinc-900/60 border border-zinc-800/80 rounded-xl p-5 mb-6">
      <h3 className="text-sm font-semibold text-zinc-200 mb-3">HTTP Status Code Distribution</h3>

      {/* Progress Bar */}
      <div className="h-3 w-full bg-zinc-800 rounded-full overflow-hidden flex mb-4">
        {buckets.map((b) => {
          const count = distribution[b.key] || 0;
          const pct = (count / total) * 100;
          if (pct === 0) return null;
          return (
            <div
              key={b.key}
              style={{ width: `${pct}%` }}
              className={`${b.color} transition-all duration-300`}
              title={`${b.label}: ${count} (${pct.toFixed(1)}%)`}
            />
          );
        })}
      </div>

      {/* Legend & Stats */}
      <div className="grid grid-cols-2 sm:grid-cols-5 gap-3">
        {buckets.map((b) => {
          const count = distribution[b.key] || 0;
          const pct = ((count / total) * 100).toFixed(1);
          return (
            <div key={b.key} className="bg-zinc-950/60 border border-zinc-800/60 rounded-lg p-2.5">
              <div className="flex items-center gap-1.5 mb-1">
                <span className={`w-2 h-2 rounded-full ${b.color}`} />
                <span className="text-[11px] text-zinc-400 font-medium">{b.label}</span>
              </div>
              <div className="flex items-baseline justify-between">
                <span className={`text-base font-bold font-mono ${b.textColor}`}>{count}</span>
                <span className="text-[10px] text-zinc-500">{pct}%</span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
