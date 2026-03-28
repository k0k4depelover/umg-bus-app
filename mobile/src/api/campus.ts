import { httpClient } from './http';
import { Campus, LiveLocation } from '../types';

const CAMPUSES_QUERY = `
  query Campuses {
    campuses {
      campusID
      name
      city
      bounds {
        swLat
        swLng
        neLat
        neLng
      }
      activePilot {
        pilotId
        fullName
        isOnline
        lastSeen
      }
    }
  }
`;

const CAMPUS_QUERY = `
  query Campus($id: ID!) {
    campus(id: $id) {
      campusID
      name
      city
      bounds {
        swLat
        swLng
        neLat
        neLng
      }
      activePilot {
        pilotId
        fullName
        isOnline
        lastSeen
      }
    }
  }
`;

const LIVE_LOCATION_QUERY = `
  query LiveLocation($campusId: ID!) {
    liveLocation(CampusId: $campusId) {
      pilot_id
      campus_id
      lat
      lng
      bearing
      speed
      updated_at
    }
  }
`;

export const campusService = {
  async getCampuses(): Promise<Campus[]> {
    const data =
      await httpClient.graphql<{ campuses: Campus[] }>(CAMPUSES_QUERY);
    return data.campuses;
  },

  async getCampus(id: string): Promise<Campus | null> {
    const data = await httpClient.graphql<{ campus: Campus | null }>(
      CAMPUS_QUERY,
      { id },
    );
    return data.campus;
  },

  async getLiveLocation(campusId: string): Promise<LiveLocation | null> {
    const data = await httpClient.graphql<{
      liveLocation: LiveLocation | null;
    }>(LIVE_LOCATION_QUERY, { campusId });
    return data.liveLocation;
  },
};
