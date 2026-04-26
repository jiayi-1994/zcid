import { render, screen } from '@testing-library/react';
import { describe, expect, test, vi } from 'vitest';
import { StepTimeline } from './StepTimeline';
import type { StepExecution } from '../../services/pipelineRun';

function row(overrides: Partial<StepExecution>): StepExecution {
  return {
    id: `${overrides.taskRunName}-${overrides.stepName}`,
    pipelineRunId: 'run-1',
    taskRunName: 'stage-a',
    stepName: 'step',
    stepIndex: 0,
    status: 'succeeded',
    createdAt: '2026-04-24T00:00:00Z',
    updatedAt: '2026-04-24T00:00:00Z',
    ...overrides,
  };
}

describe('StepTimeline', () => {
  test('renders empty state', () => {
    render(<StepTimeline steps={[]} />);
    expect(screen.getByText(/Step timing will appear here/)).toBeInTheDocument();
  });

  test('groups by task run and positions bars by start and finish time', () => {
    render(<StepTimeline steps={[
      row({ taskRunName: 'stage-b', stepName: 'push', stepIndex: 0, startedAt: '2026-04-24T00:00:20Z', finishedAt: '2026-04-24T00:00:40Z', durationMs: 20_000 }),
      row({ taskRunName: 'stage-a', stepName: 'checkout', stepIndex: 0, startedAt: '2026-04-24T00:00:00Z', finishedAt: '2026-04-24T00:00:10Z', durationMs: 10_000 }),
      row({ taskRunName: 'stage-a', stepName: 'build', stepIndex: 1, startedAt: '2026-04-24T00:00:10Z', finishedAt: '2026-04-24T00:00:30Z', durationMs: 20_000 }),
    ]} />);

    expect(screen.getByText('stage-a')).toBeInTheDocument();
    expect(screen.getByText('stage-b')).toBeInTheDocument();

    const bars = screen.getAllByTestId('step-timeline-bar');
    expect(bars).toHaveLength(3);
    expect(bars.map((bar) => bar.getAttribute('data-left-pct'))).toEqual(['0.000', '25.000', '50.000']);
    expect(bars.map((bar) => bar.getAttribute('data-width-pct'))).toEqual(['25.000', '50.000', '50.000']);
    expect(screen.getByRole('listitem', { name: /build, succeeded, 20s/ })).toBeInTheDocument();
  });

  test('renders in-flight running bars using current time', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-04-24T00:01:00Z'));
    render(<StepTimeline steps={[row({ status: 'running', startedAt: '2026-04-24T00:00:00Z', finishedAt: null, durationMs: null })]} />);

    expect(screen.getByRole('listitem')).toHaveAccessibleName(/step, running, 1m 0s/);
    expect(screen.getByTestId('step-timeline-bar')).toHaveClass('step-timeline__bar--running');
    vi.useRealTimers();
  });

  test('renders error state', () => {
    render(<StepTimeline steps={[]} error="unavailable" />);
    expect(screen.getByRole('alert')).toHaveTextContent('unavailable');
  });
});
