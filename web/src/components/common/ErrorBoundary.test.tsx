import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, test, vi } from 'vitest';
import { ErrorBoundary } from './ErrorBoundary';

let shouldThrow = true;

function FlakyChild() {
  if (shouldThrow) {
    throw new Error('boom');
  }

  return <div>render-success</div>;
}

describe('ErrorBoundary', () => {
  test('shows fallback UI and retries successfully', () => {
    shouldThrow = true;
    const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    render(
      <ErrorBoundary>
        <FlakyChild />
      </ErrorBoundary>,
    );

    expect(screen.getByRole('alert')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '重试' })).toBeInTheDocument();

    shouldThrow = false;
    fireEvent.click(screen.getByRole('button', { name: '重试' }));

    expect(screen.getByText('render-success')).toBeInTheDocument();
    errorSpy.mockRestore();
  });
});
