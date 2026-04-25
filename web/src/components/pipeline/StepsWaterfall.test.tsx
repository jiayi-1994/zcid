import { render, screen } from '@testing-library/react';
import { describe, expect, test, vi } from 'vitest';
import { StepsWaterfall } from './StepsWaterfall';
import type { StepExecution } from '../../services/pipelineRun';

function row(overrides: Partial<StepExecution>): StepExecution {
  return {
    id: `${overrides.taskRunName}-${overrides.stepName}`,
    pipelineRunId: 'run-1',
    taskRunName: 'task-a',
    stepName: 'step',
    stepIndex: 0,
    status: 'succeeded',
    createdAt: '2026-04-24T00:00:00Z',
    updatedAt: '2026-04-24T00:00:00Z',
    ...overrides,
  };
}

describe('StepsWaterfall', () => {
  test('renders empty state', () => {
    render(<StepsWaterfall items={[]} />);
    expect(screen.getByText(/Step timing will appear here/)).toBeInTheDocument();
  });

  test('renders ordered accessible bars with duration ratios', () => {
    render(<StepsWaterfall items={[
      row({ taskRunName: 'task-b', stepName: 'push', stepIndex: 1, durationMs: 30_000 }),
      row({ taskRunName: 'task-a', stepName: 'build', stepIndex: 0, durationMs: 180_000 }),
      row({ taskRunName: 'task-a', stepName: 'checkout', stepIndex: 0, durationMs: 10_000 }),
      row({ taskRunName: 'task-b', stepName: 'scan', stepIndex: 0, durationMs: 5_000 }),
    ]} />);

    const items = screen.getAllByRole('listitem');
    expect(items).toHaveLength(4);
    expect(items[0]).toHaveAccessibleName(/build, succeeded, 3m 0s/);
    const bars = screen.getAllByTestId('step-waterfall-bar');
    expect(bars.map((bar) => bar.getAttribute('data-duration-ms'))).toEqual(['180000', '10000', '5000', '30000']);
  });

  test('renders in-flight running bar using current time and pulse class', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-04-24T00:00:42Z'));
    render(<StepsWaterfall items={[row({ status: 'running', startedAt: '2026-04-24T00:00:00Z', finishedAt: null })]} />);

    expect(screen.getByRole('listitem')).toHaveAccessibleName(/step, running, 42s/);
    expect(screen.getByTestId('step-waterfall-bar')).toHaveClass('step-waterfall__bar--running');
    vi.useRealTimers();
  });

  test('renders error row without throwing', () => {
    render(<StepsWaterfall items={[]} error="unavailable" />);
    expect(screen.getByRole('alert')).toHaveTextContent('unavailable');
  });

  test('handles null timing fields without throwing', () => {
    render(<StepsWaterfall items={[row({ startedAt: null, finishedAt: null, durationMs: null })]} />);
    expect(screen.getByRole('listitem')).toHaveAccessibleName(/step, succeeded, 0s/);
  });
});
