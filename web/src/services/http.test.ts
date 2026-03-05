import axios, { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from 'axios';
import { afterEach, beforeEach, describe, expect, test, vi } from 'vitest';
import { http } from './http';
import { useAuthStore } from '../stores/auth';

function getAuthHeader(config: InternalAxiosRequestConfig): string | undefined {
  if (config.headers instanceof AxiosHeaders) {
    const value = config.headers.get('Authorization');
    return typeof value === 'string' ? value : undefined;
  }

  const headers = config.headers as Record<string, unknown> | undefined;
  const value = headers?.Authorization ?? headers?.authorization;
  return typeof value === 'string' ? value : undefined;
}

function createUnauthorizedError(config: InternalAxiosRequestConfig): AxiosError {
  return new AxiosError(
    'Unauthorized',
    'ERR_BAD_REQUEST',
    config,
    {},
    {
      status: 401,
      statusText: 'Unauthorized',
      headers: {},
      config,
      data: {},
    },
  );
}

describe('http 401 refresh interceptor', () => {
  const replaceMock = vi.fn();

  beforeEach(() => {
    useAuthStore.getState().clearSession();
    window.localStorage.clear();
    replaceMock.mockReset();
    vi.stubGlobal('location', {
      ...window.location,
      replace: replaceMock,
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
    useAuthStore.getState().clearSession();
    http.defaults.adapter = undefined;
  });

  test('refresh success updates access token and retries original request', async () => {
    useAuthStore.getState().setSession({
      accessToken: 'old-access-token',
      refreshToken: 'old-refresh-token',
      user: { username: 'alice', role: 'member' },
    });

    vi.spyOn(axios, 'post').mockResolvedValue({
      data: {
        code: 0,
        message: 'ok',
        data: {
          accessToken: 'new-access-token',
        },
      },
    } as never);

    const authHeaders: Array<string | undefined> = [];
    let protectedRequestCount = 0;

    http.defaults.adapter = async (config) => {
      if (config.url === '/protected') {
        protectedRequestCount += 1;
        authHeaders.push(getAuthHeader(config));

        if (protectedRequestCount === 1) {
          throw createUnauthorizedError(config);
        }

        return {
          status: 200,
          statusText: 'OK',
          config,
          headers: {},
          data: { ok: true },
        };
      }

      return {
        status: 200,
        statusText: 'OK',
        config,
        headers: {},
        data: {},
      };
    };

    const response = await http.get('/protected');

    expect(response.status).toBe(200);
    expect(protectedRequestCount).toBe(2);
    expect(authHeaders).toEqual(['Bearer old-access-token', 'Bearer new-access-token']);
    expect(useAuthStore.getState().accessToken).toBe('new-access-token');
    expect(replaceMock).not.toHaveBeenCalled();
  });

  test('refresh failure clears session and redirects to login', async () => {
    useAuthStore.getState().setSession({
      accessToken: 'old-access-token',
      refreshToken: 'old-refresh-token',
      user: { username: 'alice', role: 'member' },
    });

    vi.spyOn(axios, 'post').mockRejectedValue(new Error('refresh failed'));

    http.defaults.adapter = async (config) => {
      if (config.url === '/protected') {
        throw createUnauthorizedError(config);
      }

      return {
        status: 200,
        statusText: 'OK',
        config,
        headers: {},
        data: {},
      };
    };

    await expect(http.get('/protected')).rejects.toBeInstanceOf(Error);

    expect(useAuthStore.getState().accessToken).toBeNull();
    expect(useAuthStore.getState().refreshToken).toBeNull();
    expect(useAuthStore.getState().user).toBeNull();
    expect(replaceMock).toHaveBeenCalledWith('/login');
  });
});
