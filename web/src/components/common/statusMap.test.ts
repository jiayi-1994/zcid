import { describe, expect, test } from 'vitest';
import { STATUS_MAP } from '../../constants/statusMap';

describe('STATUS_MAP', () => {
  test('contains seven required statuses', () => {
    expect(Object.keys(STATUS_MAP)).toEqual([
      'success',
      'running',
      'failed',
      'warning',
      'pending',
      'cancelled',
      'timeout',
    ]);
  });

  test('defines color/bg/icon/label for each status', () => {
    for (const value of Object.values(STATUS_MAP)) {
      expect(typeof value.color).toBe('string');
      expect(typeof value.bg).toBe('string');
      expect(typeof value.icon).toBe('string');
      expect(typeof value.label).toBe('string');
    }
  });
});
