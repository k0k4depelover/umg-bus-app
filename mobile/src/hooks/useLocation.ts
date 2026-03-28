import { useState, useEffect, useRef, useCallback } from 'react';
import { Platform, PermissionsAndroid, Alert } from 'react-native';
import Geolocation from 'react-native-geolocation-service';
import { PilotPing } from '../types';

interface LocationState {
  latitude: number;
  longitude: number;
  bearing: number;
  speed: number;
}

export function useLocation() {
  const [location, setLocation] = useState<LocationState | null>(null);
  const [isTracking, setIsTracking] = useState(false);
  const [hasPermission, setHasPermission] = useState(false);
  const watchIdRef = useRef<number | null>(null);

  const requestPermission = useCallback(async (): Promise<boolean> => {
    if (Platform.OS === 'ios') {
      const status = await Geolocation.requestAuthorization('whenInUse');
      const granted = status === 'granted';
      setHasPermission(granted);
      return granted;
    }

    if (Platform.OS === 'android') {
      const granted = await PermissionsAndroid.request(
        PermissionsAndroid.PERMISSIONS.ACCESS_FINE_LOCATION,
        {
          title: 'Permiso de ubicación',
          message:
            'UMG Bus necesita acceder a tu ubicación para transmitir la posición del bus.',
          buttonPositive: 'Permitir',
          buttonNegative: 'Cancelar',
        },
      );
      const isGranted = granted === PermissionsAndroid.RESULTS.GRANTED;
      setHasPermission(isGranted);
      return isGranted;
    }

    return false;
  }, []);

  const startTracking = useCallback(
    (onLocationUpdate: (ping: PilotPing) => void) => {
      if (!hasPermission) {
        Alert.alert(
          'Permiso requerido',
          'Necesitas otorgar permiso de ubicación.',
        );
        return;
      }

      watchIdRef.current = Geolocation.watchPosition(
        (position) => {
          const { latitude, longitude, heading, speed } = position.coords;
          const loc: LocationState = {
            latitude,
            longitude,
            bearing: heading || 0,
            speed: speed ? speed * 3.6 : 0, // m/s to km/h
          };
          setLocation(loc);

          const ping: PilotPing = {
            lat: latitude,
            lng: longitude,
            bearing: heading || 0,
            speed: speed ? speed * 3.6 : 0,
          };
          onLocationUpdate(ping);
        },
        (error) => {
          console.warn('Location error:', error.message);
        },
        {
          enableHighAccuracy: true,
          distanceFilter: 5, // meters
          interval: 2000, // Android: ms between updates
          fastestInterval: 1000,
          showsBackgroundLocationIndicator: true,
        },
      );

      setIsTracking(true);
    },
    [hasPermission],
  );

  const stopTracking = useCallback(() => {
    if (watchIdRef.current !== null) {
      Geolocation.clearWatch(watchIdRef.current);
      watchIdRef.current = null;
    }
    setIsTracking(false);
  }, []);

  // Request permission on mount
  useEffect(() => {
    requestPermission();
  }, [requestPermission]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (watchIdRef.current !== null) {
        Geolocation.clearWatch(watchIdRef.current);
      }
    };
  }, []);

  return {
    location,
    isTracking,
    hasPermission,
    startTracking,
    stopTracking,
    requestPermission,
  };
}
