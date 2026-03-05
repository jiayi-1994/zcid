import { useEffect, useState } from 'react';
import { BREAKPOINTS } from '../constants/breakpoints';

export function useSidebarCollapsed(): boolean {
  const [collapsed, setCollapsed] = useState<boolean>(() => {
    if (typeof window === 'undefined') {
      return false;
    }
    return window.innerWidth < BREAKPOINTS.xl;
  });

  useEffect(() => {
    const onResize = () => setCollapsed(window.innerWidth < BREAKPOINTS.xl);
    onResize();
    window.addEventListener('resize', onResize);
    return () => window.removeEventListener('resize', onResize);
  }, []);

  return collapsed;
}
