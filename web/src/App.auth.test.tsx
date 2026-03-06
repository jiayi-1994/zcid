import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, test, vi } from 'vitest';
import App from './App';
import { login, logout } from './services/auth';
import { useAuthStore } from './stores/auth';

vi.mock('./services/auth', () => ({
  login: vi.fn(),
  logout: vi.fn(),
}));

describe('App auth routing', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    cleanup();
    useAuthStore.getState().clearSession();
    window.localStorage.clear();
  });

  test('redirects unauthenticated user to login page', async () => {
    window.history.pushState({}, '', '/dashboard');

    render(<App />);

    expect(await screen.findByText('欢迎回来')).toBeInTheDocument();
  });

  test('shows dashboard for authenticated member and hides restricted entry/action', async () => {
    useAuthStore.getState().setSession({
      accessToken: 'access-token',
      refreshToken: 'refresh-token',
      user: { username: 'alice', role: 'member' },
    });

    window.history.pushState({}, '', '/dashboard');

    render(<App />);

    expect(await screen.findByRole('heading', { name: 'Dashboard' })).toBeInTheDocument();
    expect(screen.getByText('alice')).toBeInTheDocument();
    expect(screen.getByText('普通成员')).toBeInTheDocument();
    expect(screen.queryByText('用户管理')).not.toBeInTheDocument();
  });

  test('direct access to restricted route shows 403 for member', async () => {
    useAuthStore.getState().setSession({
      accessToken: 'access-token',
      refreshToken: 'refresh-token',
      user: { username: 'alice', role: 'member' },
    });

    window.history.pushState({}, '', '/admin/users');

    render(<App />);

    expect(await screen.findByText('403')).toBeInTheDocument();
    expect(screen.getByText('无权限访问')).toBeInTheDocument();
  });

  test('logs in and navigates to dashboard', async () => {
    vi.mocked(login).mockResolvedValue({
      accessToken: 'new-access-token',
      refreshToken: 'new-refresh-token',
    });

    window.history.pushState({}, '', '/login');

    render(<App />);

    fireEvent.change(screen.getByPlaceholderText('请输入用户名'), {
      target: { value: 'alice' },
    });
    fireEvent.change(screen.getByPlaceholderText('请输入密码'), {
      target: { value: 'pass123' },
    });

    fireEvent.click(screen.getByRole('button', { name: '登录' }));

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Dashboard' })).toBeInTheDocument();
    });

    expect(useAuthStore.getState().accessToken).toBe('new-access-token');
    expect(useAuthStore.getState().refreshToken).toBe('new-refresh-token');
  });

  test('logs out and blocks protected routes afterward', async () => {
    vi.mocked(logout).mockResolvedValue();
    useAuthStore.getState().setSession({
      accessToken: 'access-token',
      refreshToken: 'refresh-token',
      user: { username: 'alice', role: 'member' },
    });

    window.history.pushState({}, '', '/dashboard');
    render(<App />);

    const userEntry = await screen.findByText('alice');
    fireEvent.click(userEntry.closest('.user-entry')!);
    fireEvent.click(await screen.findByRole('menuitem', { name: '退出登录' }));

    await waitFor(() => {
      expect(logout).toHaveBeenCalledWith('refresh-token');
      expect(useAuthStore.getState().isAuthenticated()).toBe(false);
      expect(screen.getByText('欢迎回来')).toBeInTheDocument();
    });

    cleanup();
    window.history.pushState({}, '', '/dashboard');
    render(<App />);
    expect(await screen.findByText('欢迎回来')).toBeInTheDocument();
  });

  test('still clears session and redirects when logout api fails', async () => {
    vi.mocked(logout).mockRejectedValue(new Error('network error'));
    useAuthStore.getState().setSession({
      accessToken: 'access-token',
      refreshToken: 'refresh-token',
      user: { username: 'alice', role: 'member' },
    });

    window.history.pushState({}, '', '/dashboard');
    render(<App />);

    const userEntry = await screen.findByText('alice');
    fireEvent.click(userEntry.closest('.user-entry')!);
    fireEvent.click(await screen.findByRole('menuitem', { name: '退出登录' }));

    await waitFor(() => {
      expect(logout).toHaveBeenCalledWith('refresh-token');
      expect(useAuthStore.getState().accessToken).toBeNull();
      expect(useAuthStore.getState().refreshToken).toBeNull();
      expect(useAuthStore.getState().user).toBeNull();
      expect(useAuthStore.getState().permissions).toEqual([]);
      expect(screen.getByText('欢迎回来')).toBeInTheDocument();
    });
  });
});
