import { Platform } from 'react-native';

// Android emulator uses 10.0.2.2 to reach host machine localhost
const getBaseUrl = () => {
  if (__DEV__) {
    return Platform.select({
      android: 'http://10.0.2.2',
      ios: 'http://localhost',
      default: 'http://localhost',
    });
  }
  // Production URL - replace with actual server
  return 'https://api.umgbus.app';
};

const BASE = getBaseUrl();

export const API_CONFIG = {
  HTTP_BASE: `${BASE}:8084`,
  WS_BASE: `${BASE?.replace('http', 'ws')}:8085`,
  ENDPOINTS: {
    LOGIN: '/auth/login',
    REFRESH: '/auth/refresh',
    LOGOUT: '/auth/logout',
    GRAPHQL: '/graphql',
    HEALTH: '/health-check',
    CAMPUSES: '/campus',
  },
  WS_ENDPOINTS: {
    PILOT: '/ws/pilot',
    STUDENT: '/ws/student',
  },
} as const;
