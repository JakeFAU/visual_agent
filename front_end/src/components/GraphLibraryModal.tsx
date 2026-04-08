import React from 'react';
import { BookOpenText, FolderOpen, Loader2, RefreshCw, Sparkles, X } from 'lucide-react';
import { ExampleGraph } from '../examples/library';

interface GraphLibraryModalProps {
  isOpen: boolean;
  savedGraphs: string[];
  examples: ExampleGraph[];
  isLoadingSavedGraphs: boolean;
  error: string | null;
  activeItemKey: string | null;
  onClose: () => void;
  onRefreshSavedGraphs: () => void;
  onLoadSavedGraph: (name: string) => void;
  onLoadExampleGraph: (example: ExampleGraph) => void;
}

export const GraphLibraryModal: React.FC<GraphLibraryModalProps> = ({
  isOpen,
  savedGraphs,
  examples,
  isLoadingSavedGraphs,
  error,
  activeItemKey,
  onClose,
  onRefreshSavedGraphs,
  onLoadSavedGraph,
  onLoadExampleGraph,
}) => {
  if (!isOpen) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-[70] flex items-center justify-center bg-black/70 p-4 backdrop-blur-sm">
      <div className="w-full max-w-5xl overflow-hidden rounded-2xl border border-gray-800 bg-gray-950 shadow-2xl">
        <div className="flex items-start justify-between gap-4 border-b border-gray-800 px-6 py-5">
          <div>
            <div className="text-[11px] font-bold uppercase tracking-[0.22em] text-blue-400">Graph Library</div>
            <h2 className="mt-2 text-xl font-semibold text-white">Open a saved workflow or start from an example</h2>
            <p className="mt-1 max-w-2xl text-sm text-gray-400">
              Examples are versioned with the repo so screenshots, docs, and the actual product stay aligned.
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="rounded-full border border-gray-800 bg-gray-900 p-2 text-gray-500 transition-colors hover:border-gray-700 hover:text-gray-200"
            aria-label="Close graph library"
          >
            <X size={16} />
          </button>
        </div>

        <div className="grid gap-0 lg:grid-cols-[1.05fr,1.2fr]">
          <section className="border-b border-gray-800 px-6 py-5 lg:border-b-0 lg:border-r">
            <div className="flex items-center justify-between gap-3">
              <div className="flex items-center gap-2">
                <FolderOpen size={16} className="text-cyan-400" />
                <h3 className="text-sm font-semibold text-white">Saved Graphs</h3>
              </div>
              <button
                type="button"
                onClick={onRefreshSavedGraphs}
                className="inline-flex items-center gap-2 rounded-md border border-gray-800 bg-gray-900 px-3 py-2 text-[11px] font-bold uppercase tracking-widest text-gray-300 transition-colors hover:border-gray-700 hover:text-white"
              >
                <RefreshCw size={12} className={isLoadingSavedGraphs ? 'animate-spin' : ''} />
                Refresh
              </button>
            </div>

            <div className="mt-4 space-y-3">
              {error && (
                <div className="rounded-xl border border-red-500/20 bg-red-950/20 p-3 text-sm text-red-100/90">
                  {error}
                </div>
              )}

              {isLoadingSavedGraphs ? (
                <div className="flex items-center gap-3 rounded-xl border border-gray-800 bg-gray-900/70 p-4 text-sm text-gray-400">
                  <Loader2 size={16} className="animate-spin text-blue-400" />
                  Loading saved workflows...
                </div>
              ) : savedGraphs.length > 0 ? (
                savedGraphs.map((name) => {
                  const itemKey = `saved:${name}`;
                  const isActive = activeItemKey === itemKey;

                  return (
                    <button
                      key={name}
                      type="button"
                      onClick={() => onLoadSavedGraph(name)}
                      disabled={activeItemKey !== null}
                      className="flex w-full items-center justify-between gap-3 rounded-xl border border-gray-800 bg-gray-900/70 p-4 text-left transition-colors hover:border-gray-700 hover:bg-gray-900 disabled:cursor-wait disabled:opacity-70"
                    >
                      <div>
                        <div className="text-sm font-semibold text-white">{name}</div>
                        <div className="mt-1 text-xs text-gray-500">Stored on the local backend and ready to reopen.</div>
                      </div>
                      {isActive ? <Loader2 size={14} className="animate-spin text-blue-400" /> : null}
                    </button>
                  );
                })
              ) : (
                <div className="rounded-xl border border-dashed border-gray-800 bg-gray-900/40 p-4 text-sm text-gray-500">
                  No saved workflows yet. Load an example below, then save your own variant.
                </div>
              )}
            </div>
          </section>

          <section className="px-6 py-5">
            <div className="flex items-center gap-2">
              <BookOpenText size={16} className="text-emerald-400" />
              <h3 className="text-sm font-semibold text-white">Examples</h3>
            </div>

            <div className="mt-4 grid gap-3">
              {examples.map((example) => {
                const itemKey = `example:${example.id}`;
                const isActive = activeItemKey === itemKey;

                return (
                  <button
                    key={example.id}
                    type="button"
                    onClick={() => onLoadExampleGraph(example)}
                    disabled={activeItemKey !== null}
                    className="rounded-xl border border-gray-800 bg-gradient-to-br from-gray-900 via-gray-900 to-gray-950 p-4 text-left transition-colors hover:border-gray-700 disabled:cursor-wait disabled:opacity-70"
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <div className="text-sm font-semibold text-white">{example.title}</div>
                        <div className="mt-1 text-sm leading-6 text-gray-400">{example.description}</div>
                      </div>
                      {isActive ? <Loader2 size={14} className="mt-1 animate-spin text-blue-400" /> : <Sparkles size={14} className="mt-1 text-emerald-400" />}
                    </div>
                    <div className="mt-3 flex flex-wrap gap-2">
                      {example.badges.map((badge) => (
                        <span
                          key={badge}
                          className="rounded-full border border-gray-800 bg-gray-950 px-2.5 py-1 text-[10px] font-bold uppercase tracking-[0.18em] text-gray-400"
                        >
                          {badge}
                        </span>
                      ))}
                    </div>
                  </button>
                );
              })}
            </div>
          </section>
        </div>
      </div>
    </div>
  );
};
