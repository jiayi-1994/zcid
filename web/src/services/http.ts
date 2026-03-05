import axios, {
  AxiosError,
  type AxiosInstance,
  type InternalAxiosRequestConfig,
} from 'axios';
import { useAuthStore } from '../stores/auth';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1';

type RetriableRequestConfig = InternalAxiosRequestConfig & {
  _retry?: boolean;
};

interface RefreshResponse {
  code: number;
  message: string;
  data: {
    accessToken: string;
  };
}

const http: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
});

let isRefreshing = false;
let pendingRequests: Array<(token: string | null) => void> = [];

function flushPendingRequests(token: string | null) {
  for (const resolve of pendingRequests) {
    resolve(token);
  }
  pendingRequests = [];
}

async function refreshAccessToken(refreshToken: string): Promise<string> {
  const response = await axios.post<RefreshResponse>(`${API_BASE_URL}/auth/refresh`, {
    refreshToken,
  });

  if (response.data.code !== 0 || !response.data.data?.accessToken) {
    throw new Error(response.data.message || 'refresh failed');
  }

  return response.data.data.accessToken;
}

http.interceptors.request.use((config) => {
  const token = useAuthStore.getState().accessToken;
  if (!token) {
    return config;
  }

  const nextConfig = config;
  nextConfig.headers.set('Authorization', `Bearer ${token}`);
  return nextConfig;
});

http.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as RetriableRequestConfig | undefined;

    if (!originalRequest || error.response?.status !== 401) {
      return Promise.reject(error);
    }

    if (originalRequest.url?.includes('/auth/refresh') || originalRequest._retry) {
      useAuthStore.getState().clearSession();
      window.location.replace('/login');
      return Promise.reject(error);
    }

    const refreshToken = useAuthStore.getState().refreshToken;
    if (!refreshToken) {
      useAuthStore.getState().clearSession();
      window.location.replace('/login');
      return Promise.reject(error);
    }

    if (isRefreshing) {
      return new Promise((resolve, reject) => {
        pendingRequests.push((token) => {
          if (!token) {
            reject(error);
            return;
          }

          originalRequest.headers.set('Authorization', `Bearer ${token}`);
          resolve(http(originalRequest));
        });
      });
    }

    originalRequest._retry = true;
    isRefreshing = true;

    try {
      const newAccessToken = await refreshAccessToken(refreshToken);
      useAuthStore.getState().setAccessToken(newAccessToken);
      flushPendingRequests(newAccessToken);
      originalRequest.headers.set('Authorization', `Bearer ${newAccessToken}`);
      return http(originalRequest);
    } catch (refreshError) {
      flushPendingRequests(null);
      useAuthStore.getState().clearSession();
      window.location.replace('/login');
      return Promise.reject(refreshError);
    } finally {
      isRefreshing = false;
    }
  },
);

export { http };
