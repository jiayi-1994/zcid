import { Tooltip } from '@arco-design/web-react';
import type { StepExecution } from '../../services/pipelineRun';
import { StatusBadge } from '../ui/StatusBadge';

const STATUS_COLOR: Record<string, string> = {
  succeeded: 'var(--green)',
  failed: 'var(--red)',
  running: 'var(--blue)',
  pending: 'var(--z-300)',
  queued: 'var(--z-300)',
  cancelled: 'var(--z-400)',
  interrupted: 'var(--z-400)',
};

interface StepTimelineProps {
  steps: StepExecution[];
  error?: string | null;
  runStartedAt?: string | null;
}

interface TimelineStep {
  step: StepExecution;
  startMs: number;
  endMs: number;
  durationMs: number;
  leftPct: number;
  widthPct: number;
}

function formatDuration(ms: number): string {
  const seconds = Math.max(0, Math.round(ms / 1000));
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ${seconds % 60}s`;
  return `${Math.floor(minutes / 60)}h ${minutes % 60}m`;
}

function parseTime(value?: string | null): number | null {
  if (!value) return null;
  const parsed = new Date(value).getTime();
  return Number.isFinite(parsed) ? parsed : null;
}

function taskRunLabel(value: string): string {
  if (!value) return 'unknown-stage';
  const parts = value.split('-');
  return parts.length > 3 ? parts.slice(-2).join('-') : value;
}

function buildTimeline(steps: StepExecution[], runStartedAt?: string | null, now = Date.now()): TimelineStep[] {
  const fallbackRunStart = parseTime(runStartedAt);
  const timings = steps.map((step) => {
    const started = parseTime(step.startedAt) ?? fallbackRunStart ?? parseTime(step.createdAt) ?? now;
    const finished = parseTime(step.finishedAt) ?? (step.status === 'running' ? now : started + Math.max(0, step.durationMs ?? 0));
    const end = Math.max(started, finished);
    return { step, startMs: started, endMs: end, durationMs: Math.max(0, end - started) };
  });

  if (timings.length === 0) return [];
  const minStart = Math.min(...timings.map((item) => item.startMs));
  const maxEnd = Math.max(...timings.map((item) => item.endMs));
  const total = Math.max(1, maxEnd - minStart);

  return timings.map((item) => ({
    ...item,
    leftPct: ((item.startMs - minStart) / total) * 100,
    widthPct: Math.max(0, (item.durationMs / total) * 100),
  }));
}

function groupByTaskRun(items: TimelineStep[]): Array<[string, TimelineStep[]]> {
  const groups = new Map<string, TimelineStep[]>();
  [...items]
    .sort((a, b) => a.step.taskRunName.localeCompare(b.step.taskRunName) || a.step.stepIndex - b.step.stepIndex || a.startMs - b.startMs)
    .forEach((item) => {
      const key = item.step.taskRunName || 'unknown-stage';
      groups.set(key, [...(groups.get(key) ?? []), item]);
    });
  return [...groups.entries()];
}

export function StepTimeline({ steps, error, runStartedAt }: StepTimelineProps) {
  if (error) {
    return <div role="alert" style={{ color: 'var(--red-ink)', fontSize: 13 }}>{error}</div>;
  }
  if (steps.length === 0) {
    return <div style={{ color: 'var(--z-400)', fontSize: 13 }}>Step timing will appear here once the run starts reporting task status.</div>;
  }

  const timeline = buildTimeline(steps, runStartedAt);
  const grouped = groupByTaskRun(timeline);

  return (
    <div>
      <style>{`
        @keyframes zcid-step-timeline-pulse { 0%, 100% { opacity: .56; } 50% { opacity: 1; } }
        @media (prefers-reduced-motion: reduce) { .step-timeline__bar--running { animation: none !important; } }
        @media (max-width: 720px) { .step-timeline__row { min-width: 720px; } }
      `}</style>
      <div style={{ display: 'grid', gridTemplateColumns: '160px 1fr 76px', gap: 10, marginBottom: 10, color: 'var(--z-400)', fontSize: 11, textTransform: 'uppercase', letterSpacing: '0.04em' }}>
        <span>Stage / Step</span>
        <span>Relative execution window</span>
        <span>Duration</span>
      </div>
      <div role="list" style={{ display: 'flex', flexDirection: 'column', gap: 14, overflowX: 'auto' }}>
        {grouped.map(([taskRun, items]) => (
          <section key={taskRun} aria-label={taskRun}>
            <div className="mono" style={{ fontSize: 11, color: 'var(--z-500)', marginBottom: 8 }}>{taskRunLabel(taskRun)}</div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              {items.map(({ step, durationMs, leftPct, widthPct }) => {
                const label = `${step.stepName}, ${step.status}, ${formatDuration(durationMs)}`;
                const isRunning = step.status === 'running' && !step.finishedAt;
                const tooltip = (
                  <div style={{ fontSize: 12 }}>
                    <div style={{ fontWeight: 600, marginBottom: 4 }}>{step.stepName}</div>
                    <div>Status: {step.status}</div>
                    <div>Duration: {formatDuration(durationMs)}</div>
                  </div>
                );
                return (
                  <div
                    key={`${step.taskRunName}-${step.stepName}`}
                    role="listitem"
                    aria-label={label}
                    className="step-timeline__row"
                    style={{ display: 'grid', gridTemplateColumns: '160px minmax(320px, 1fr) 76px', gap: 10, alignItems: 'center' }}
                  >
                    <div style={{ minWidth: 0, display: 'flex', flexDirection: 'column', gap: 4 }}>
                      <div style={{ fontSize: 12, fontWeight: 650, color: 'var(--z-800)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{step.stepName}</div>
                      <StatusBadge status={step.status} />
                    </div>
                    <div style={{ position: 'relative', height: 22, borderRadius: 999, background: 'linear-gradient(90deg, var(--z-100), var(--z-50))', overflow: 'hidden', border: '1px solid var(--z-150)' }}>
                      <Tooltip content={tooltip} mini>
                        <div
                          className={isRunning ? 'step-timeline__bar--running' : undefined}
                          data-testid="step-timeline-bar"
                          data-left-pct={leftPct.toFixed(3)}
                          data-width-pct={widthPct.toFixed(3)}
                          style={{
                            position: 'absolute',
                            left: `${Math.max(0, Math.min(100, leftPct)).toFixed(3)}%`,
                            width: `max(4px, ${Math.max(0, Math.min(100, widthPct)).toFixed(3)}%)`,
                            height: '100%',
                            minWidth: 4,
                            borderRadius: 999,
                            background: STATUS_COLOR[step.status] ?? 'var(--z-400)',
                            boxShadow: '0 1px 0 rgba(255,255,255,.35) inset',
                            animation: isRunning ? 'zcid-step-timeline-pulse 1.4s ease-in-out infinite' : undefined,
                          }}
                        />
                      </Tooltip>
                    </div>
                    <span className="mono" style={{ color: 'var(--z-500)', fontSize: 12 }}>{formatDuration(durationMs)}</span>
                  </div>
                );
              })}
            </div>
          </section>
        ))}
      </div>
      <div style={{ marginTop: 14, color: 'var(--z-400)', fontSize: 12 }}>Bars are positioned by observed step start and finish times.</div>
    </div>
  );
}

export default StepTimeline;
