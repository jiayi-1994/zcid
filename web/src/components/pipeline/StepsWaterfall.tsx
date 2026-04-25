import type { StepExecution } from '../../services/pipelineRun';
import { StatusBadge } from '../ui/StatusBadge';

function formatDuration(ms: number): string {
  const seconds = Math.max(0, Math.round(ms / 1000));
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ${seconds % 60}s`;
  return `${Math.floor(minutes / 60)}h ${minutes % 60}m`;
}

function durationMs(item: StepExecution, now: number): number {
  if (typeof item.durationMs === 'number') return item.durationMs;
  if (!item.startedAt) return 0;
  const end = item.finishedAt ? new Date(item.finishedAt).getTime() : now;
  return Math.max(0, end - new Date(item.startedAt).getTime());
}

function groupByTaskRun(items: StepExecution[]): Array<[string, StepExecution[]]> {
  const groups = new Map<string, StepExecution[]>();
  [...items]
    .sort((a, b) => a.taskRunName.localeCompare(b.taskRunName) || a.stepIndex - b.stepIndex)
    .forEach((item) => {
      const key = item.taskRunName || 'unknown-task';
      groups.set(key, [...(groups.get(key) ?? []), item]);
    });
  return [...groups.entries()];
}

const STATUS_COLOR: Record<string, string> = {
  succeeded: 'var(--green)',
  failed: 'var(--red)',
  cancelled: 'var(--amber)',
  interrupted: 'var(--amber)',
  running: 'var(--blue-ink)',
  pending: 'var(--z-300)',
  queued: 'var(--z-300)',
};

export function StepsWaterfall({ items, error }: { items: StepExecution[]; error?: string | null }) {
  const now = Date.now();
  if (error) {
    return <div role="alert" style={{ color: 'var(--red-ink)', fontSize: 13 }}>{error}</div>;
  }
  if (items.length === 0) {
    return <div style={{ color: 'var(--z-400)', fontSize: 13 }}>Step timing will appear here once the run starts reporting task status.</div>;
  }

  const durations = items.map((item) => durationMs(item, now));
  const maxDuration = Math.max(1, ...durations);

  return (
    <div>
      <style>{`
        @keyframes zcid-step-pulse { 0%, 100% { opacity: .55; } 50% { opacity: 1; } }
        @media (prefers-reduced-motion: reduce) { .step-waterfall__bar--running { animation: none !important; } }
        @media (max-width: 640px) { .step-waterfall__row { min-width: 620px; } }
      `}</style>
      <div role="list" style={{ display: 'flex', flexDirection: 'column', gap: 16, overflowX: 'auto' }}>
        {groupByTaskRun(items).map(([taskRun, steps]) => (
          <section key={taskRun} aria-label={taskRun}>
            <div className="mono" style={{ fontSize: 11, color: 'var(--z-500)', marginBottom: 8 }}>{taskRun}</div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              {steps.map((step) => {
                const ms = durationMs(step, now);
                const label = `${step.stepName}, ${step.status}, ${formatDuration(ms)}`;
                const pct = Math.max(0, Math.min(100, (ms / maxDuration) * 100));
                const width = `max(4px, ${pct.toFixed(3)}%)`;
                const isRunning = !step.finishedAt && step.status === 'running';
                return (
                  <div
                    key={`${step.taskRunName}-${step.stepName}`}
                    role="listitem"
                    aria-label={label}
                    title={label}
                    className="step-waterfall__row"
                    style={{ display: 'grid', gridTemplateColumns: '160px minmax(220px, 1fr) 70px', gap: 10, alignItems: 'center' }}
                  >
                    <div style={{ minWidth: 0 }}>
                      <div style={{ fontSize: 12, fontWeight: 650, color: 'var(--z-800)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{step.stepName}</div>
                      <StatusBadge status={step.status} />
                    </div>
                    <div style={{ height: 18, background: 'var(--z-100)', borderRadius: 999, overflow: 'hidden' }}>
                      <div
                        className={isRunning ? 'step-waterfall__bar--running' : undefined}
                        data-testid="step-waterfall-bar"
                        data-duration-ms={ms}
                        style={{
                          width,
                          height: '100%',
                          minWidth: 4,
                          borderRadius: 999,
                          background: STATUS_COLOR[step.status] ?? 'var(--z-400)',
                          animation: isRunning ? 'zcid-step-pulse 1.4s ease-in-out infinite' : undefined,
                        }}
                      />
                    </div>
                    <span className="mono" style={{ color: 'var(--z-500)', fontSize: 12 }}>{formatDuration(ms)}</span>
                  </div>
                );
              })}
            </div>
          </section>
        ))}
      </div>
      <div style={{ marginTop: 14, color: 'var(--z-400)', fontSize: 12 }}>Per-step logs are available in the Build Output section below.</div>
    </div>
  );
}