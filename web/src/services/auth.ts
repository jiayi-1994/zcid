import { http } from './http';

interface ApiSuccessResponse<T> {
  code: number;
  message: string;
  data: T;
}

export interface LoginResult {
  accessToken: string;
  refreshToken: string;
}

export async function login(username: string, password: string): Promise<LoginResult> {
  const response = await http.post<ApiSuccessResponse<LoginResult>>('/auth/login', {
    username,
    password,
  });

  if (response.data.code !== 0 || !response.data.data?.accessToken || !response.data.data?.refreshToken) {
    throw new Error(response.data.message || '登录失败');
  }

  return response.data.data;
}

export async function logout(refreshToken: string): Promise<void> {
  await http.post<ApiSuccessResponse<{ loggedOut: boolean }>>('/auth/logout', {
    refreshToken,
  });
}
