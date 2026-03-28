import AsyncStorage from '@react-native-async-storage/async-storage';
import { API_CONFIG } from './config';
import { RefreshResponse } from '../types';

const STORAGE_KEYS = {
  ACCESS_TOKEN: '@umgbus:access_token',
  REFRESH_TOKEN: '@umgbus:refresh_token',
};

class HttpClient {
  private accessToken: string | null = null;
  private refreshToken: string | null = null;
  private refreshPromise: Promise<string> | null = null;

  async init() {
    this.accessToken = await AsyncStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN);
    this.refreshToken = await AsyncStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN);
  }

  async setTokens(access: string, refresh: string) {
    this.accessToken = access;
    this.refreshToken = refresh;
    await AsyncStorage.multiSet([
      [STORAGE_KEYS.ACCESS_TOKEN, access],
      [STORAGE_KEYS.REFRESH_TOKEN, refresh],
    ]);
  }

  async clearTokens() {
    this.accessToken = null;
    this.refreshToken = null;
    await AsyncStorage.multiRemove([
      STORAGE_KEYS.ACCESS_TOKEN,
      STORAGE_KEYS.REFRESH_TOKEN,
    ]);
  }

  getAccessToken() {
    return this.accessToken;
  }

  getRefreshToken() {
    return this.refreshToken;
  }

  private async refreshAccessToken(): Promise<string> {
    if (this.refreshPromise) return this.refreshPromise;

    this.refreshPromise = (async () => {
      try {
        const res = await fetch(
          `${API_CONFIG.HTTP_BASE}${API_CONFIG.ENDPOINTS.REFRESH}`,
          {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ refresh_token: this.refreshToken }),
          },
        );
        if (!res.ok) throw new Error('Refresh failed');
        const data: RefreshResponse = await res.json();
        this.accessToken = data.access_token;
        await AsyncStorage.setItem(
          STORAGE_KEYS.ACCESS_TOKEN,
          data.access_token,
        );
        return data.access_token;
      } finally {
        this.refreshPromise = null;
      }
    })();

    return this.refreshPromise;
  }

  async fetch<T>(path: string, options: RequestInit = {}): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    };

    if (this.accessToken) {
      headers['Authorization'] = `Bearer ${this.accessToken}`;
    }

    let res = await fetch(`${API_CONFIG.HTTP_BASE}${path}`, {
      ...options,
      headers,
    });

    // Try refresh on 401
    if (res.status === 401 && this.refreshToken) {
      try {
        await this.refreshAccessToken();
        headers['Authorization'] = `Bearer ${this.accessToken}`;
        res = await fetch(`${API_CONFIG.HTTP_BASE}${path}`, {
          ...options,
          headers,
        });
      } catch {
        await this.clearTokens();
        throw new Error('SESSION_EXPIRED');
      }
    }

    if (!res.ok) {
      const error = await res.json().catch(() => ({ error: res.statusText }));
      throw new Error(error.error || `HTTP ${res.status}`);
    }

    return res.json();
  }

  async graphql<T>(
    query: string,
    variables?: Record<string, unknown>,
  ): Promise<T> {
    const result = await this.fetch<{
      data?: T;
      errors?: Array<{ message: string }>;
    }>(API_CONFIG.ENDPOINTS.GRAPHQL, {
      method: 'POST',
      body: JSON.stringify({ query, variables }),
    });
    if (result.errors?.length) {
      throw new Error(result.errors[0].message);
    }
    return result.data as T;
  }
}

export const httpClient = new HttpClient();
export { STORAGE_KEYS };
