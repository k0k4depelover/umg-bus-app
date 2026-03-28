import { httpClient } from './http';
import { API_CONFIG } from './config';
import { LoginRequest, LoginResponse } from '../types';

export const authService = {
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    const res = await fetch(
      `${API_CONFIG.HTTP_BASE}${API_CONFIG.ENDPOINTS.LOGIN}`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(credentials),
      },
    );

    if (!res.ok) {
      const error = await res
        .json()
        .catch(() => ({ error: 'Login failed' }));
      throw new Error(error.error || `HTTP ${res.status}`);
    }

    const data: LoginResponse = await res.json();
    await httpClient.setTokens(data.access_token, data.refresh_token);
    return data;
  },

  async logout(): Promise<void> {
    const refreshToken = httpClient.getRefreshToken();
    if (refreshToken) {
      try {
        await fetch(
          `${API_CONFIG.HTTP_BASE}${API_CONFIG.ENDPOINTS.LOGOUT}`,
          {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ refresh_token: refreshToken }),
          },
        );
      } catch {
        // Ignore logout errors - clear tokens anyway
      }
    }
    await httpClient.clearTokens();
  },
};
