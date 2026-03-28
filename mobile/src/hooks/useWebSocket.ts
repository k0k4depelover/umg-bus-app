import { useEffect, useRef, useCallback, useState } from 'react';
import { AppState, AppStateStatus } from 'react-native';
import { API_CONFIG } from '../api/config';
import { httpClient } from '../api/http';
import { LiveLocation, PilotPing } from '../types';

interface UseWebSocketOptions {
  role: 'pilot' | 'student';
  onMessage?: (data: LiveLocation) => void;
  onError?: (error: string) => void;
  onConnected?: () => void;
  onDisconnected?: () => void;
}

export function useWebSocket({
  role,
  onMessage,
  onError,
  onConnected,
  onDisconnected,
}: UseWebSocketOptions) {
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const mountedRef = useRef(true);

  const connect = useCallback(() => {
    const token = httpClient.getAccessToken();
    if (!token) {
      onError?.('No access token');
      return;
    }

    const endpoint =
      role === 'pilot'
        ? API_CONFIG.WS_ENDPOINTS.PILOT
        : API_CONFIG.WS_ENDPOINTS.STUDENT;
    const url = `${API_CONFIG.WS_BASE}${endpoint}?token=${token}`;

    try {
      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        if (!mountedRef.current) return;
        setIsConnected(true);
        onConnected?.();
      };

      ws.onmessage = (event: MessageEvent) => {
        if (!mountedRef.current) return;
        try {
          const data = JSON.parse(
            typeof event.data === 'string' ? event.data : '',
          ) as LiveLocation;
          onMessage?.(data);
        } catch {
          // Ignore malformed messages
        }
      };

      ws.onerror = () => {
        if (!mountedRef.current) return;
        onError?.('WebSocket error');
      };

      ws.onclose = () => {
        if (!mountedRef.current) return;
        setIsConnected(false);
        onDisconnected?.();
        // Auto-reconnect after 3 seconds
        reconnectTimer.current = setTimeout(() => {
          if (mountedRef.current) connect();
        }, 3000);
      };
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : 'WebSocket failed';
      onError?.(message);
    }
  }, [role, onMessage, onError, onConnected, onDisconnected]);

  const disconnect = useCallback(() => {
    if (reconnectTimer.current) {
      clearTimeout(reconnectTimer.current);
      reconnectTimer.current = null;
    }
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    setIsConnected(false);
  }, []);

  const sendPing = useCallback((ping: PilotPing) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(ping));
    }
  }, []);

  // Handle app state changes (background/foreground)
  useEffect(() => {
    const subscription = AppState.addEventListener(
      'change',
      (state: AppStateStatus) => {
        if (state === 'active' && !wsRef.current) {
          connect();
        } else if (state === 'background') {
          // Keep connection alive for pilot (GPS streaming)
          // Disconnect for student to save battery
          if (role === 'student') {
            disconnect();
          }
        }
      },
    );

    return () => subscription.remove();
  }, [connect, disconnect, role]);

  useEffect(() => {
    return () => {
      mountedRef.current = false;
      disconnect();
    };
  }, [disconnect]);

  return { isConnected, connect, disconnect, sendPing };
}
