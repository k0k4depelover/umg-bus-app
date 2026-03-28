// Auth
export interface LoginRequest {
  username: string;
  password: string;
  role: 'pilot' | 'student';
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
}

export interface RefreshResponse {
  access_token: string;
}

export interface JWTClaims {
  user_id: string;
  campus_id: string;
  role: 'pilot' | 'student';
  exp: number;
  iat: number;
}

// Campus & Map
export interface MapBounds {
  swLat: number;
  swLng: number;
  neLat: number;
  neLng: number;
}

export interface Campus {
  campusID: string;
  name: string;
  city: string;
  bounds: MapBounds;
  activePilot?: PilotStatus | null;
}

export interface PilotStatus {
  pilotId: string;
  fullName: string;
  isOnline: boolean;
  lastSeen: string;
}

// Location (WebSocket)
export interface PilotPing {
  lat: number;
  lng: number;
  bearing: number;
  speed: number;
}

export interface LiveLocation {
  pilot_id: string;
  campus_id: string;
  lat: number;
  lng: number;
  bearing: number;
  speed: number;
  updated_at: string;
}

// GraphQL
export interface GraphQLResponse<T> {
  data?: T;
  errors?: Array<{ message: string; path?: string[] }>;
}

export interface CampusesData {
  campuses: Campus[];
}

export interface CampusData {
  campus: Campus | null;
}

export interface LiveLocationData {
  liveLocation: LiveLocation | null;
}

// Navigation
export type RootStackParamList = {
  Auth: undefined;
  PilotMain: undefined;
  StudentMain: undefined;
};

export type AuthStackParamList = {
  Login: undefined;
  // Register: undefined; // Future
};

export type PilotTabParamList = {
  PilotMap: undefined;
  PilotProfile: undefined;
  PilotSettings: undefined;
};

export type StudentTabParamList = {
  StudentMap: undefined;
  StudentProfile: undefined;
  StudentSettings: undefined;
  StudentCampus: undefined;
};
