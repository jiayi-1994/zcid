import { render, screen } from '@testing-library/react';
import { describe, expect, test } from 'vitest';
import { PageSkeleton } from './PageSkeleton';

describe('PageSkeleton', () => {
  test('renders skeleton placeholder container', () => {
    render(<PageSkeleton />);
    expect(screen.getByTestId('page-skeleton')).toBeInTheDocument();
  });
});
