import { type ReactNode } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { type PermissionKey, useAuthStore } from '../../stores/auth';

interface RequirePermissionProps {
  permission: PermissionKey;
  children: ReactNode;
}

export function RequirePermission({ permission, children }: RequirePermissionProps) {
  const hasPermission = useAuthStore((state) => state.hasPermission(permission));
  const location = useLocation();

  if (!hasPermission) {
    return <Navigate to="/403" replace state={{ from: location.pathname }} />;
  }

  return <>{children}</>;
}
