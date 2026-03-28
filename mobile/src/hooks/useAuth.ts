import { create } from 'zustand';
import { JWTClaims } from '../types';
import { httpClient } from '../api/http';
import { authService } from '../api/auth';

function decodeJWT(token: string): JWTClaims | null {
  try {
    const payload = token.split('.')[1];
    const decoded = JSON.parse(atob(payload));
    return decoded as JWTClaims;
  } catch {
    return null;
  }
}

interface AuthState {
  isLoading: boolean;
  isAuthenticated: boolean;
  claims: JWTClaims | null;
  error: string | null;

  init: () => Promise<void>;
  login: (
    username: string,
    password: string,
    role: 'pilot' | 'student',
  ) => Promise<void>;
  logout: () => Promise<void>;
  clearError: () => void;
}

export const useAuth = create<AuthState>((set) => ({
  isLoading: true,
  isAuthenticated: false,
  claims: null,
  error: null,

  init: async () => {
    try {
      await httpClient.init();
      const token = httpClient.getAccessToken();
      if (token) {
        const claims = decodeJWT(token);
        if (claims && claims.exp * 1000 > Date.now()) {
          set({ isAuthenticated: true, claims, isLoading: false });
          return;
        }
        // Token expired - try refresh
        const refreshToken = httpClient.getRefreshToken();
        if (refreshToken) {
          try {
            // Attempt a health check to trigger refresh
            await httpClient.fetch('/health-check');
            const newToken = httpClient.getAccessToken();
            if (newToken) {
              const newClaims = decodeJWT(newToken);
              set({
                isAuthenticated: true,
                claims: newClaims,
                isLoading: false,
              });
              return;
            }
          } catch {
            await httpClient.clearTokens();
          }
        }
      }
      set({ isAuthenticated: false, claims: null, isLoading: false });
    } catch {
      set({ isAuthenticated: false, claims: null, isLoading: false });
    }
  },

  login: async (username, password, role) => {
    set({ error: null });
    try {
      const response = await authService.login({ username, password, role });
      const claims = decodeJWT(response.access_token);
      set({ isAuthenticated: true, claims, error: null });
    } catch (e: unknown) {
      const message =
        e instanceof Error ? e.message : 'Error al iniciar sesión';
      set({ error: message });
      throw e;
    }
  },

  logout: async () => {
    await authService.logout();
    set({ isAuthenticated: false, claims: null, error: null });
  },

  clearError: () => set({ error: null }),
}));
