import { create } from 'zustand';

const STORAGE_KEY = 'zcid.auth';

export type SystemRole = 'admin' | 'project_admin' | 'member';

export type PermissionKey =
  | 'route:dashboard:view'
  | 'route:admin-users:view'
  | 'route:admin-variables:view'
  | 'route:admin-integrations:view'
  | 'route:admin-audit:view'
  | 'route:admin-settings:view'
  | 'route:projects:view'
  | 'action:user:create';

const ROLE_PERMISSIONS: Record<SystemRole, PermissionKey[]> = {
  admin: ['route:dashboard:view', 'route:admin-users:view', 'route:admin-variables:view', 'route:admin-integrations:view', 'route:admin-audit:view', 'route:admin-settings:view', 'route:projects:view', 'action:user:create'],
  project_admin: ['route:dashboard:view', 'route:admin-audit:view', 'route:projects:view'],
  member: ['route:dashboard:view', 'route:projects:view'],
};

export interface AuthUser {
  username: string;
  role: SystemRole;
}

export interface AuthSession {
  accessToken: string;
  refreshToken: string;
  user?: AuthUser | null;
}

interface StoredAuthSession {
  accessToken: string;
  refreshToken: string;
  user: AuthUser | null;
}

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  user: AuthUser | null;
  permissions: PermissionKey[];
  setSession: (session: AuthSession) => void;
  setAccessToken: (accessToken: string) => void;
  clearSession: () => void;
  isAuthenticated: () => boolean;
  hasPermission: (permission: PermissionKey) => boolean;
}

interface AccessTokenClaims {
  username?: string;
  role?: string;
}

function parseTokenClaims(accessToken: string): AccessTokenClaims | null {
  const parts = accessToken.split('.');
  if (parts.length !== 3) {
    return null;
  }

  try {
    const payload = parts[1];
    const normalized = payload.replace(/-/g, '+').replace(/_/g, '/');
    const padded = normalized.padEnd(normalized.length + ((4 - (normalized.length % 4)) % 4), '=');
    const decoded = atob(padded);
    return JSON.parse(decoded) as AccessTokenClaims;
  } catch {
    return null;
  }
}

function parseRole(rawRole: string | undefined): SystemRole {
  if (rawRole === 'admin' || rawRole === 'project_admin' || rawRole === 'member') {
    return rawRole;
  }
  return 'member';
}

function buildUser(session: AuthSession): AuthUser | null {
  if (session.user) {
    return {
      username: session.user.username,
      role: parseRole(session.user.role),
    };
  }

  const claims = parseTokenClaims(session.accessToken);
  if (!claims?.username) {
    return null;
  }

  return {
    username: claims.username,
    role: parseRole(claims.role),
  };
}

function readStoredSession(): StoredAuthSession | null {
  if (typeof window === 'undefined') {
    return null;
  }

  const raw = window.localStorage.getItem(STORAGE_KEY);
  if (!raw) {
    return null;
  }

  try {
    const parsed = JSON.parse(raw) as StoredAuthSession;
    if (!parsed.accessToken || !parsed.refreshToken) {
      return null;
    }

    const claims = parseTokenClaims(parsed.accessToken);
    const username = parsed.user?.username ?? claims?.username;
    if (!username) {
      return null;
    }

    return {
      accessToken: parsed.accessToken,
      refreshToken: parsed.refreshToken,
      user: {
        username,
        role: parseRole(parsed.user?.role ?? claims?.role),
      },
    };
  } catch {
    return null;
  }
}

function persistSession(session: StoredAuthSession) {
  if (typeof window === 'undefined') {
    return;
  }

  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(session));
}

function clearStoredSession() {
  if (typeof window === 'undefined') {
    return;
  }

  window.localStorage.removeItem(STORAGE_KEY);
}

const initialSession = readStoredSession();

export const useAuthStore = create<AuthState>((set, get) => ({
  accessToken: initialSession?.accessToken ?? null,
  refreshToken: initialSession?.refreshToken ?? null,
  user: initialSession?.user ?? null,
  permissions: initialSession?.user ? ROLE_PERMISSIONS[initialSession.user.role] : [],
  setSession: (session) => {
    const user = buildUser(session);
    const next: StoredAuthSession = {
      accessToken: session.accessToken,
      refreshToken: session.refreshToken,
      user,
    };

    persistSession(next);
    set({
      accessToken: next.accessToken,
      refreshToken: next.refreshToken,
      user: next.user,
      permissions: next.user ? ROLE_PERMISSIONS[next.user.role] : [],
    });
  },
  setAccessToken: (accessToken) => {
    const refreshToken = get().refreshToken;
    if (!refreshToken) {
      return;
    }

    const claims = parseTokenClaims(accessToken);
    const currentUser = get().user;
    const username = currentUser?.username ?? claims?.username;
    const role = parseRole(claims?.role ?? currentUser?.role);
    const user = username ? { username, role } : null;

    const next: StoredAuthSession = {
      accessToken,
      refreshToken,
      user,
    };

    persistSession(next);
    set({
      accessToken,
      user,
      permissions: user ? ROLE_PERMISSIONS[user.role] : [],
    });
  },
  clearSession: () => {
    clearStoredSession();
    set({ accessToken: null, refreshToken: null, user: null, permissions: [] });
  },
  isAuthenticated: () => Boolean(get().accessToken && get().refreshToken),
  hasPermission: (permission) => get().permissions.includes(permission),
}));

