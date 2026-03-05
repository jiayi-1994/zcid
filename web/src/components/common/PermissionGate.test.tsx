import { render, screen } from '@testing-library/react';
import { describe, expect, test } from 'vitest';
import { PermissionGate } from './PermissionGate';

describe('PermissionGate', () => {
  test('renders children when allowed', () => {
    render(
      <PermissionGate allowed>
        <div>visible-content</div>
      </PermissionGate>,
    );

    expect(screen.getByText('visible-content')).toBeInTheDocument();
  });

  test('renders nothing when not allowed', () => {
    render(
      <PermissionGate allowed={false}>
        <div>hidden-content</div>
      </PermissionGate>,
    );

    expect(screen.queryByText('hidden-content')).not.toBeInTheDocument();
  });
});
