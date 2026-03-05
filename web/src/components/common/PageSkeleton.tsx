import { Skeleton } from '@arco-design/web-react';

export function PageSkeleton() {
  return (
    <div data-testid="page-skeleton">
      <Skeleton text={{ rows: 6 }} animation />
    </div>
  );
}
