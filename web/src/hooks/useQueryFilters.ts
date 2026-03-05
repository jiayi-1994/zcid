import { useMemo, useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';

/**
 * Syncs filter state with URL query params.
 * @param defaults - default values for each filter key
 * @returns [currentFilters, updateFilters]
 */
export function useQueryFilters<T extends Record<string, string>>(
  defaults: T
): [T, (updates: Partial<T>) => void] {
  const [searchParams, setSearchParams] = useSearchParams();

  const filters = useMemo(() => {
    const result = { ...defaults } as Record<string, string>;
    for (const key of Object.keys(defaults) as (keyof T)[]) {
      const v = searchParams.get(String(key));
      if (v !== null && v !== '') {
        result[String(key)] = v;
      }
    }
    return result as T;
  }, [searchParams, defaults]);

  const updateFilters = useCallback(
    (updates: Partial<T>) => {
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev);
          for (const [key, value] of Object.entries(updates)) {
            if (value === '' || value === undefined || value === null) {
              next.delete(key);
            } else {
              next.set(key, String(value));
            }
          }
          return next;
        },
        { replace: true }
      );
    },
    [setSearchParams]
  );

  return [filters, updateFilters];
}
