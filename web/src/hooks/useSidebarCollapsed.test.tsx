import { render, screen } from '@testing-library/react';
import { describe, expect, test } from 'vitest';
import { useSidebarCollapsed } from './useSidebarCollapsed';

function Probe() {
  const collapsed = useSidebarCollapsed();
  return <div data-testid="state">{collapsed ? 'collapsed' : 'expanded'}</div>;
}

describe('useSidebarCollapsed', () => {
  test('returns collapsed when width is below 768', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 600 });
    render(<Probe />);
    expect(screen.getByTestId('state')).toHaveTextContent('collapsed');
  });

  test('returns expanded when width is at least 768', () => {
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 1200 });
    render(<Probe />);
    expect(screen.getByTestId('state')).toHaveTextContent('expanded');
  });
});
