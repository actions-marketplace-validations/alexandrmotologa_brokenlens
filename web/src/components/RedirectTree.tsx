import React from 'react';
import { ArrowRight, AlertCircle, RefreshCw } from 'lucide-react';
import { RedirectChain } from '../types';

interface RedirectTreeProps {
  chains: RedirectChain[];
}

export const RedirectTree: React.FC<RedirectTreeProps> = ({ chains }) => {
  if (chains.length === 0) {
    return (
      <div className="bg-zinc-900/60 border border-zinc-800/80 rounded-xl p-8 text-center text-zinc-500 text-sm">
        No HTTP redirect chains detected in this audit.
      </div>
    );
  }

  return (
    <div className="bg-zinc-900/60 border border-zinc-800/80 rounded-xl p-4">
      <h3 className="text-sm font-semibold text-zinc-200 mb-3">
        Detected Redirect Chains ({chains.length})
      </h3>

      <div className="space-y-3">
        {chains.map((chain, idx) => {
          const isLoop = chain.isLoop;
          const isExcessive = chain.hops.length > 3;

          return (
            <div
              key={idx}
              className={`p-3.5 rounded-lg border text-xs ${
                isLoop
                  ? 'bg-red-500/10 border-red-500/30'
                  : isExcessive
                  ? 'bg-amber-500/10 border-amber-500/30'
                  : 'bg-zinc-950/60 border-zinc-800/80'
              }`}
            >
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <span className="font-mono text-zinc-400">Chain #{idx + 1}</span>
                  <span className="px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-300 text-[10px] font-mono">
                    {chain.hops.length} {chain.hops.length === 1 ? 'hop' : 'hops'}
                  </span>
                  {isLoop && (
                    <span className="inline-flex items-center gap-1 text-[10px] font-bold uppercase text-red-400 bg-red-500/20 px-2 py-0.5 rounded border border-red-500/30">
                      <RefreshCw className="w-3 h-3 animate-spin" />
                      Circular Loop
                    </span>
                  )}
                  {isExcessive && !isLoop && (
                    <span className="inline-flex items-center gap-1 text-[10px] font-bold uppercase text-amber-400 bg-amber-500/20 px-2 py-0.5 rounded border border-amber-500/30">
                      <AlertCircle className="w-3 h-3" />
                      Excessive Redirects
                    </span>
                  )}
                </div>
              </div>

              {/* Hop list */}
              <div className="space-y-1.5 font-mono">
                {chain.hops.map((hop, hIdx) => (
                  <div key={hIdx} className="flex items-center gap-2 text-zinc-300">
                    <span className="text-purple-400 font-bold w-8">{hop.statusCode}</span>
                    <span className="truncate max-w-md">{hop.fromUrl}</span>
                    <ArrowRight className="w-3.5 h-3.5 text-zinc-600 flex-shrink-0" />
                    <span className="truncate max-w-md text-zinc-400">{hop.toUrl}</span>
                  </div>
                ))}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
